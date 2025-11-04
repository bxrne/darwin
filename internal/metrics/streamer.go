package metrics

import (
	"context"
	"sync"
	"sync/atomic"
)

// MetricsStreamer handles async streaming of generation metrics
type MetricsStreamer struct {
	metricsChan <-chan GenerationMetrics
	subscribers []chan GenerationMetrics
	done        chan struct{}
	running     chan struct{} // signals when Start goroutine is done
	mu          sync.Mutex
	stopped     int32 // atomic flag
}

// NewMetricsStreamer creates a new metrics streamer
func NewMetricsStreamer(metricsChan <-chan GenerationMetrics) *MetricsStreamer {
	return &MetricsStreamer{
		metricsChan: metricsChan,
		subscribers: make([]chan GenerationMetrics, 0),
		done:        make(chan struct{}),
		running:     make(chan struct{}),
		stopped:     0,
	}
}

// Subscribe returns a channel that will receive metrics
func (ms *MetricsStreamer) Subscribe() <-chan GenerationMetrics {
	subscriber := make(chan GenerationMetrics, 10) // buffered channel
	ms.mu.Lock()
	ms.subscribers = append(ms.subscribers, subscriber)
	ms.mu.Unlock()
	return subscriber
}

// Start begins streaming metrics to all subscribers
func (ms *MetricsStreamer) Start(ctx context.Context) {
	go func() {
		defer close(ms.running)
		defer ms.closeSubscribers()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ms.done:
				return
			case metrics, ok := <-ms.metricsChan:
				if !ok {
					return
				}
				ms.broadcast(metrics)
			}
		}
	}()
}

// Stop stops the streamer and closes all subscriber channels
func (ms *MetricsStreamer) Stop() {
	ms.mu.Lock()
	if atomic.CompareAndSwapInt32(&ms.stopped, 0, 1) {
		close(ms.done)
	}
	ms.mu.Unlock()
	<-ms.running // wait for the Start goroutine to finish
}

// broadcast sends metrics to all subscribers
func (ms *MetricsStreamer) broadcast(metrics GenerationMetrics) {
	ms.mu.Lock()
	subscribers := make([]chan GenerationMetrics, len(ms.subscribers))
	copy(subscribers, ms.subscribers)
	ms.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- metrics:
		default:
			// Skip if subscriber is not ready (non-blocking)
		}
	}
}

// closeSubscribers closes all subscriber channels
func (ms *MetricsStreamer) closeSubscribers() {
	ms.mu.Lock()
	subscribers := ms.subscribers
	ms.subscribers = nil
	ms.mu.Unlock()

	for _, subscriber := range subscribers {
		close(subscriber)
	}
}
