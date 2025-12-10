package fitness

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
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

	zap.L().Info("Created TCP connection pool",
		zap.String("server", serverAddr),
		zap.Int("max_connections", maxConnections),
		zap.Duration("timeout", timeout))

	return p
}

// ─────────────────────────────
//
//	BLOCKING GetConnection()
//
// ─────────────────────────────
func (p *TCPConnectionPool) GetConnection() (*TCPClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if p.closed {
			return nil, fmt.Errorf("connection pool is closed")
		}

		select {
		case client := <-p.connections:
			if err := p.healthCheckUnlocked(client); err != nil {
				if err := client.Disconnect(); err != nil {
					zap.L().Warn("Failed to disconnect client", zap.Error(err))
				}
				p.activeCount--
				continue
			}
			return client, nil

		default:
		}

		if p.activeCount < p.maxConnections {
			return p.createNewConnectionUnlocked()
		}

		p.cond.Wait()
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
			zap.L().Warn("Failed to disconnect client", zap.Error(err))
		}
		return fmt.Errorf("connection pool is closed")
	}

	if err := p.healthCheckUnlocked(client); err != nil {
		if err := client.Disconnect(); err != nil {
			zap.L().Warn("Failed to disconnect client", zap.Error(err))
		}
		p.activeCount--
		return nil
	}

	select {
	case p.connections <- client:
		p.cond.Signal()
	default:
		if err := client.Disconnect(); err != nil {
			zap.L().Warn("Failed to disconnect client", zap.Error(err))
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
			zap.L().Warn("Failed to disconnect client during close", zap.Error(err))
		}
		p.activeCount--
	}

	p.cond.Broadcast()

	zap.L().Info("TCP connection pool closed",
		zap.Int("total_created", p.totalCreated),
		zap.Int("final_active", p.activeCount))

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
			zap.L().Warn("Failed to return client to pool", zap.Error(err))
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

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}

	p.activeCount++
	p.totalCreated++
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
