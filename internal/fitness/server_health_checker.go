package fitness

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ServerHealthChecker performs health checks on game server
type ServerHealthChecker struct {
	serverAddr string
	timeout    time.Duration
}

// NewServerHealthChecker creates a new health checker
func NewServerHealthChecker(serverAddr string, timeout time.Duration) *ServerHealthChecker {
	return &ServerHealthChecker{
		serverAddr: serverAddr,
		timeout:    timeout,
	}
}

// CheckServerHealth performs a basic health check on the game server
func (shc *ServerHealthChecker) CheckServerHealth() error {
	zap.L().Info("Checking game server health", zap.String("server", shc.serverAddr))

	// Create a temporary connection for health check
	client := NewTCPClient(shc.serverAddr)

	// Try to connect with timeout
	if err := client.Connect(); err != nil {
		return fmt.Errorf("server health check failed: unable to connect to %s: %w", shc.serverAddr, err)
	}

	// Connection successful, verify server is responsive
	// Try to send a connect message to ensure server is properly responding
	connectReq := ConnectRequest{
		Type:         string(Connect),
		AgentType:    "health_check",
		OpponentType: "none",
	}

	if err := client.SendMessage(connectReq); err != nil {
		if err := client.Disconnect(); err != nil {
			zap.L().Warn("Failed to disconnect client", zap.Error(err))
		}
		return fmt.Errorf("server health check failed: unable to send health check message: %w", err)
	}

	// Clean up connection
	if err := client.Disconnect(); err != nil {
		zap.L().Warn("Failed to disconnect health check connection", zap.Error(err))
	}

	zap.L().Info("Server health check passed", zap.String("server", shc.serverAddr))
	return nil
}

// CheckServerHealthWithRetry performs health check with a single retry
func (shc *ServerHealthChecker) CheckServerHealthWithRetry() error {
	err := shc.CheckServerHealth()
	if err != nil {
		zap.L().Warn("Server health check failed, retrying once", zap.Error(err))
		time.Sleep(1 * time.Second) // Brief delay before retry
		return shc.CheckServerHealth()
	}
	return nil
}
