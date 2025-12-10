package fitness

import (
	"fmt"
	"net"
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

// CheckServerHealth performs a health check on the game server
// It sends a HEALTH message (which does not start a game) and waits for a response
func (shc *ServerHealthChecker) CheckServerHealth() error {
	zap.L().Info("Checking game server health", zap.String("server", shc.serverAddr))

	// Create a temporary connection for health check
	client := NewTCPClient(shc.serverAddr)

	// Try to connect with timeout - this verifies the server is listening
	if err := client.Connect(); err != nil {
		return fmt.Errorf("server health check failed: unable to connect to %s: %w", shc.serverAddr, err)
	}

	// Send health check message (does not start a game)
	if err := client.SendHealthCheck(); err != nil {
		_ = client.Disconnect()
		return fmt.Errorf("server health check failed: failed to send health message: %w", err)
	}

	// Wait for health response with timeout
	// Set a read timeout on the connection
	if tcpConn, ok := client.conn.(*net.TCPConn); ok {
		err_inner := tcpConn.SetReadDeadline(time.Now().Add(shc.timeout))
		if err_inner != nil {
			zap.L().Warn("Failed to set read deadline", zap.Error(err_inner))
		}
	}

	healthResp, err := client.ReceiveHealthResponse()
	if err != nil {
		err_inner := client.Disconnect()
		if err_inner != nil {
			zap.L().Warn("Failed to disconnect", zap.Error(err_inner))
		}
		return fmt.Errorf("server health check failed: failed to receive health response: %w", err)
	}

	// Clear the read deadline
	if tcpConn, ok := client.conn.(*net.TCPConn); ok {
		err_inner := tcpConn.SetReadDeadline(time.Time{})
		if err_inner != nil {
			zap.L().Warn("Failed to set read deadline", zap.Error(err_inner))
		}
	}

	// Verify the response indicates healthy status
	if healthResp.Status != "ok" {
		err_inner := client.Disconnect()
		if err_inner != nil {
			zap.L().Warn("Failed to disconnect", zap.Error(err_inner))
		}
		return fmt.Errorf("server health check failed: server returned non-ok status: %s", healthResp.Status)
	}

	// Disconnect cleanly
	if err := client.Disconnect(); err != nil {
		zap.L().Warn("Failed to disconnect health check connection", zap.Error(err))
		// Don't fail the health check if disconnect fails - health check was successful
	}

	zap.L().Info("Server health check passed", zap.String("server", shc.serverAddr), zap.String("status", healthResp.Status))
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
