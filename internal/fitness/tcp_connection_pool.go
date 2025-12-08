package fitness

import (
	"fmt"
	"sync"
	"time"

	"github.com/bxrne/logmgr"
)

type TCPConnectionPool struct {
	serverAddr     string
	maxConnections int
	timeout        time.Duration

	mu          sync.Mutex
	cond        *sync.Cond
	connections chan *TCPClient

	activeCount  int
	totalCreated int
	closed       bool
}

func NewTCPConnectionPool(serverAddr string, maxConnections int, timeout time.Duration) *TCPConnectionPool {
	p := &TCPConnectionPool{
		serverAddr:     serverAddr,
		maxConnections: maxConnections,
		timeout:        timeout,
		connections:    make(chan *TCPClient, maxConnections),
	}
	p.cond = sync.NewCond(&p.mu)

	logmgr.Info("Created TCP connection pool",
		logmgr.Field("server", serverAddr),
		logmgr.Field("max_connections", maxConnections),
		logmgr.Field("timeout", timeout))

	return p
}

// GetPoolStats returns current pool statistics
func (p *TCPConnectionPool) GetPoolStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"active_connections":    p.activeCount,
		"available_connections": len(p.connections),
		"max_connections":       p.maxConnections,
		"total_created":         p.totalCreated,
		"closed":                p.closed,
	}
}

// ─────────────────────────────
//
//	GetConnection() with timeout
//
// ─────────────────────────────
func (p *TCPConnectionPool) GetConnection() (*TCPClient, error) {
	return p.GetConnectionWithTimeout(p.timeout)
}

// GetConnectionWithTimeout gets a connection with specified timeout
func (p *TCPConnectionPool) GetConnectionWithTimeout(timeout time.Duration) (*TCPClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	timeoutChan := time.After(timeout)

	for {
		if p.closed {
			return nil, fmt.Errorf("connection pool is closed")
		}

		select {
		case client := <-p.connections:
			if err := p.healthCheckUnlocked(client); err != nil {
				if err := client.Disconnect(); err != nil {
					logmgr.Warn("Failed to disconnect client", logmgr.Field("err", err))
				}
				p.activeCount--
				continue
			}
			return client, nil

		case <-timeoutChan:
			return nil, fmt.Errorf("timeout getting connection from pool after %v", timeout)

		default:
		}

		if p.activeCount < p.maxConnections {
			return p.createNewConnectionUnlocked()
		}

		// Use cond.WaitWithTimeout to avoid blocking indefinitely
		waitChan := make(chan struct{})
		go func() {
			p.cond.Wait()
			close(waitChan)
		}()

		select {
		case <-waitChan:
			// Wait completed, continue loop
		case <-timeoutChan:
			return nil, fmt.Errorf("timeout waiting for connection from pool after %v", timeout)
		}
	}
}

// ─────────────────────────────
//
//	ReturnConnection
//
// ─────────────────────────────
func (p *TCPConnectionPool) ReturnConnection(client *TCPClient) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		if err := client.Disconnect(); err != nil {
			logmgr.Warn("Failed to disconnect client", logmgr.Field("err", err))
		}
		return fmt.Errorf("connection pool is closed")
	}

	if err := p.healthCheckUnlocked(client); err != nil {
		logmgr.Debug("Connection health check failed, disconnecting",
			logmgr.Field("error", err.Error()),
			logmgr.Field("client_id", fmt.Sprintf("%p", client)))
		if err := client.Disconnect(); err != nil {
			logmgr.Warn("Failed to disconnect client", logmgr.Field("err", err))
		}
		p.activeCount--
		return nil
	}

	select {
	case p.connections <- client:
		logmgr.Debug("Connection returned to pool",
			logmgr.Field("client_id", fmt.Sprintf("%p", client)),
			logmgr.Field("available_connections", len(p.connections)))
		p.cond.Signal()
	default:
		logmgr.Debug("Connection pool full, disconnecting client",
			logmgr.Field("client_id", fmt.Sprintf("%p", client)),
			logmgr.Field("active_count", p.activeCount),
			logmgr.Field("max_connections", p.maxConnections))
		if err := client.Disconnect(); err != nil {
			logmgr.Warn("Failed to disconnect client", logmgr.Field("err", err))
		}
		p.activeCount--
	}

	return nil
}

// ─────────────────────────────
//
//	Close()
//
// ─────────────────────────────
func (p *TCPConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	for client := range p.connections {
		if err := client.Disconnect(); err != nil {
			logmgr.Warn("Failed to disconnect client during close", logmgr.Field("err", err))
		}
		p.activeCount--
	}

	p.cond.Broadcast()

	logmgr.Info("TCP connection pool closed",
		logmgr.Field("total_created", p.totalCreated),
		logmgr.Field("final_active", p.activeCount))

	return nil
}

// ─────────────────────────────
//
//	HealthCheck() external API
//
// ─────────────────────────────
func (p *TCPConnectionPool) HealthCheck() error {
	client, err := p.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		if err := p.ReturnConnection(client); err != nil {
			logmgr.Warn("Failed to return client to pool", logmgr.Field("err", err))
		}
	}()

	return p.healthCheckUnlocked(client)
}

// ─────────────────────────────
//
//	Stats
//
// ─────────────────────────────
func (p *TCPConnectionPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"server_addr":     p.serverAddr,
		"max_connections": p.maxConnections,
		"active_count":    p.activeCount,
		"available":       len(p.connections),
		"total_created":   p.totalCreated,
		"closed":          p.closed,
	}
}

// ─────────────────────────────
//
//	Internal helpers
//
// ─────────────────────────────
func (p *TCPConnectionPool) createNewConnectionUnlocked() (*TCPClient, error) {
	client := NewTCPClient(p.serverAddr)
	logmgr.Debug("Creating new connection to server",
		logmgr.Field("server", p.serverAddr),
		logmgr.Field("client_id", fmt.Sprintf("%p", client)))
	if err := client.Connect(); err != nil {
		logmgr.Error("Failed to create new connection",
			logmgr.Field("server", p.serverAddr),
			logmgr.Field("error", err))
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}
	p.activeCount++
	p.totalCreated++
	logmgr.Info("Created new connection",
		logmgr.Field("client_id", fmt.Sprintf("%p", client)),
		logmgr.Field("active_count", p.activeCount),
		logmgr.Field("total_created", p.totalCreated))
	return client, nil
}

func (p *TCPConnectionPool) healthCheckUnlocked(client *TCPClient) error {
	if client == nil || client.conn == nil {
		return fmt.Errorf("nil connection")
	}
	if client.conn.RemoteAddr() == nil {
		return fmt.Errorf("connection closed")
	}
	return nil
}
