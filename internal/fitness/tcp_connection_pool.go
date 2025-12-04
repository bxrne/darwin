package fitness

import (
	"fmt"
	"sync"
	"time"

	"github.com/bxrne/logmgr"
)

// TCPConnectionPool manages a pool of TCP connections to the game server
type TCPConnectionPool struct {
	serverAddr     string
	maxConnections int
	timeout        time.Duration
	mu             sync.RWMutex
	connections    chan *TCPClient
	activeCount    int
	totalCreated   int
	closed         bool
}

// NewTCPConnectionPool creates a new connection pool
func NewTCPConnectionPool(serverAddr string, maxConnections int, timeout time.Duration) *TCPConnectionPool {
	pool := &TCPConnectionPool{
		serverAddr:     serverAddr,
		maxConnections: maxConnections,
		timeout:        timeout,
		connections:    make(chan *TCPClient, maxConnections),
		closed:         false,
	}

	logmgr.Info("Created TCP connection pool",
		logmgr.Field("server", serverAddr),
		logmgr.Field("max_connections", maxConnections),
		logmgr.Field("timeout", timeout))

	return pool
}

// GetConnection retrieves a connection from the pool or creates a new one
func (p *TCPConnectionPool) GetConnection() (*TCPClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	select {
	case client := <-p.connections:
		// Got existing connection, verify it's healthy
		if err := p.healthCheck(client); err != nil {
			client.Disconnect()
			return p.createNewConnection()
		}
		return client, nil

	default:
		// No available connections, create new one if under limit
		if p.activeCount >= p.maxConnections {
			logmgr.Error("Connection pool exhausted",
				logmgr.Field("active", p.activeCount),
				logmgr.Field("max", p.maxConnections),
				logmgr.Field("server", p.serverAddr))
			return nil, fmt.Errorf("connection pool exhausted: %d active connections (max: %d)", p.activeCount, p.maxConnections)
		}
		return p.createNewConnection()
	}
}

// ReturnConnection returns a connection to the pool
func (p *TCPConnectionPool) ReturnConnection(client *TCPClient) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		client.Disconnect()
		return fmt.Errorf("connection pool is closed")
	}

	// Verify connection is still healthy before returning to pool
	if err := p.healthCheck(client); err != nil {
		client.Disconnect()
		p.activeCount--
		return nil
	}

	select {
	case p.connections <- client:
		return nil
	default:
		// Pool full, discard connection
		client.Disconnect()
		p.activeCount--
		return nil
	}
}

// Close closes all connections and shuts down the pool
func (p *TCPConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	// Close all connections in the pool
	for client := range p.connections {
		client.Disconnect()
		p.activeCount--
	}

	logmgr.Info("TCP connection pool closed",
		logmgr.Field("total_created", p.totalCreated),
		logmgr.Field("final_active", p.activeCount))

	return nil
}

// HealthCheck performs a basic health check on a connection
func (p *TCPConnectionPool) HealthCheck() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("connection pool is closed")
	}

	// Try to get a connection and check it
	client, err := p.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection for health check: %w", err)
	}
	defer p.ReturnConnection(client)

	return p.healthCheck(client)
}

// GetStats returns pool statistics
func (p *TCPConnectionPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"server_addr":     p.serverAddr,
		"max_connections": p.maxConnections,
		"active_count":    p.activeCount,
		"available":       len(p.connections),
		"total_created":   p.totalCreated,
		"closed":          p.closed,
	}
}

// createNewConnection creates a new TCP client and connects it
func (p *TCPConnectionPool) createNewConnection() (*TCPClient, error) {
	client := NewTCPClient(p.serverAddr)

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}

	p.activeCount++
	p.totalCreated++

	return client, nil
}

// healthCheck verifies a connection is still valid
func (p *TCPConnectionPool) healthCheck(client *TCPClient) error {
	if client == nil || client.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Simple check: verify the connection is still active
	// We can do a more sophisticated check if needed
	conn := client.conn
	if conn == nil {
		return fmt.Errorf("underlying connection is nil")
	}

	// Check if connection is closed by checking remote addr
	if conn.RemoteAddr() == nil {
		return fmt.Errorf("connection is closed")
	}

	return nil
}
