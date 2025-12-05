package fitness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bxrne/logmgr"
)

// MessageType constants matching the Python server
type MessageType string

const (
	Connect     MessageType = "connect"
	Action      MessageType = "action"
	Reset       MessageType = "reset"
	Connected   MessageType = "connected"
	Observation MessageType = "observation"
	Error       MessageType = "error"
	GameOver    MessageType = "game_over"
)

// Message structures matching Python payloads
type ConnectRequest struct {
	Type         string `json:"type"`
	AgentType    string `json:"agent_type"`
	OpponentType string `json:"opponent_type"`
}

type ActionRequest struct {
	Type   string      `json:"type"`
	Action interface{} `json:"action"`
}

type ConnectedResponse struct {
	Type       string `json:"type"`
	AgentID    string `json:"agent_id"`
	OpponentID string `json:"opponent_id"`
	Message    string `json:"message"`
}

type ObservationResponse struct {
	Type        string             `json:"type"`
	Observation map[string]float64 `json:"observation"`
	Reward      float64            `json:"reward"`
	Terminated  bool               `json:"terminated"`
	Truncated   bool               `json:"truncated"`
	Info        [][]bool           `json:"info"`
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type GameOverResponse struct {
	Type         string             `json:"type"`
	Winner       *string            `json:"winner"`
	FinalRewards map[string]float64 `json:"final_rewards"`
	Reason       string             `json:"reason"`
}

// TCPClient handles communication with the game server
type TCPClient struct {
	conn       net.Conn
	reader     *bufio.Reader
	serverAddr string
}

// NewTCPClient creates a new TCP client
func NewTCPClient(serverAddr string) *TCPClient {
	return &TCPClient{
		serverAddr: serverAddr,
	}
}

// Connect establishes connection to the game server
func (tc *TCPClient) Connect() error {
	var err error
	tc.conn, err = net.DialTimeout("tcp", tc.serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", tc.serverAddr, err)
	}

	tc.reader = bufio.NewReader(tc.conn)
	logmgr.Info("Connected to game server", logmgr.Field("address", tc.serverAddr))
	return nil
}

// Disconnect closes the connection
func (tc *TCPClient) Disconnect() error {
	if tc.conn != nil {
		err := tc.conn.Close()
		tc.conn = nil
		tc.reader = nil
		return err
	}
	return nil
}

// SendMessage sends a JSON message to the server
func (tc *TCPClient) SendMessage(message interface{}) error {
	if tc.conn == nil {
		return fmt.Errorf("not connected to server")
	}
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Append newline as required by server protocol
	data = append(data, '\n')

	_, err = tc.conn.Write(data)
	if err != nil {
		// Check for broken pipe specifically
		if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "EPIPE") {
			logmgr.Debug("Broken pipe detected - server likely closed connection")
			return fmt.Errorf("broken pipe: server closed connection")
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// ReceiveMessage receives a JSON message from the server
func (tc *TCPClient) ReceiveMessage() (map[string]interface{}, error) {
	if tc.reader == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	line, err := tc.reader.ReadString('\n')
	if err != nil {
		// Check for connection closed errors
		if strings.Contains(err.Error(), "use of closed network connection") ||
			strings.Contains(err.Error(), "connection reset by peer") {
			logmgr.Debug("Connection closed by peer during read")
		}
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var message map[string]interface{}
	err = json.Unmarshal([]byte(line), &message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return message, nil
}

// ConnectToGame sends connect request and waits for connected response
func (tc *TCPClient) ConnectToGame(opponentType string) (*ConnectedResponse, error) {
	connectReq := ConnectRequest{
		Type:         string(Connect),
		AgentType:    "human",
		OpponentType: opponentType,
	}
	err := tc.SendMessage(connectReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send connect request: %w", err)
	}

	// Wait for connected response
	for {
		msg, err := tc.ReceiveMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to receive connected response: %w", err)
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch MessageType(msgType) {
		case Connected:
			var resp ConnectedResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &resp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse connected response: %w", err)
			}
			return &resp, nil

		case Error:
			var errResp ErrorResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &errResp)
			if err == nil {
				return nil, fmt.Errorf("server error: %s - %s", errResp.Message, errResp.Details)
			}
			return nil, fmt.Errorf("server error: %v", msg)

		default:
			// Ignore other message types for now
			continue
		}
	}
}

// SendAction sends an action to the server
func (tc *TCPClient) SendAction(action interface{}) error {
	actionReq := ActionRequest{
		Type:   string(Action),
		Action: action,
	}

	return tc.SendMessage(actionReq)
}

// ReceiveObservation waits for an observation response
func (tc *TCPClient) ReceiveObservation() (*ObservationResponse, error) {
	for {
		msg, err := tc.ReceiveMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to receive observation: %w", err)
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch MessageType(msgType) {
		case Observation:
			// Debug log raw JSON message
			logmgr.Debug("Raw observation message",
				logmgr.Field("message", fmt.Sprintf("%+v", msg)))
			var resp ObservationResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &resp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse observation response: %w", err)
			}
			// Debug log received observation
			logmgr.Debug("Received observation",
				logmgr.Field("reward", resp.Reward),
				logmgr.Field("terminated", resp.Terminated))
			return &resp, nil

		case GameOver:
			// Game over is also a valid observation response
			var resp ObservationResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &resp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse game over response: %w", err)
			}
			resp.Terminated = true
			return &resp, nil

		case Error:
			var errResp ErrorResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &errResp)
			if err == nil {
				return nil, fmt.Errorf("server error: %s - %s", errResp.Message, errResp.Details)
			}
			return nil, fmt.Errorf("server error: %v", msg)

		default:
			// Ignore other message types
			continue
		}
	}
}

// WaitForGameOver waits specifically for a game over message
func (tc *TCPClient) WaitForGameOver() (*GameOverResponse, error) {
	for {
		msg, err := tc.ReceiveMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to receive game over: %w", err)
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		if MessageType(msgType) == GameOver {
			var resp GameOverResponse
			data, _ := json.Marshal(msg)
			err = json.Unmarshal(data, &resp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse game over response: %w", err)
			}
			return &resp, nil
		}
	}
}

// ExtractGameScore extracts a numeric score from observation data
func ExtractGameScore(observation map[string]interface{}) float64 {
	// Try different common score fields
	if score, ok := observation["score"].(float64); ok {
		return score
	}

	if reward, ok := observation["reward"].(float64); ok {
		return reward
	}

	// Try to extract from nested structures
	if action, ok := observation["action"].(map[string]interface{}); ok {
		if score, ok := action["score"].(float64); ok {
			return score
		}
	}

	// Try to parse from string fields
	if scoreStr, ok := observation["score"].(string); ok {
		if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			return score
		}
	}

	return 0.0
}
