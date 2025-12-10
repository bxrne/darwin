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
// It only verifies TCP connectivity without sending a Connect message (which starts a game)
func (shc *ServerHealthChecker) CheckServerHealth() error {
	zap.L().Info("Checking game server health", zap.String("server", shc.serverAddr))

	// Create a temporary connection for health check
	client := NewTCPClient(shc.serverAddr)

	// Try to connect with timeout - this verifies the server is listening
	// We do NOT send a Connect message as that would start a game
	if err := client.Connect(); err != nil {
		return fmt.Errorf("server health check failed: unable to connect to %s: %w", shc.serverAddr, err)
	}

	// Connection successful - server is listening and accepting connections
	// Just disconnect without sending any game-starting messages
	if err := client.Disconnect(); err != nil {
		zap.L().Warn("Failed to disconnect health check connection", zap.Error(err))
		// Don't fail the health check if disconnect fails - connection was successful
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
