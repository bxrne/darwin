package metrics

import (
	"context"
	"sync"
)

// MetricsStreamer handles async streaming of generation metrics
type MetricsStreamer struct {
	metricsChan <-chan GenerationMetrics
	subscribers []chan GenerationMetrics
	done        chan struct{}
	wg          sync.WaitGroup
}

// NewMetricsStreamer creates a new metrics streamer
func NewMetricsStreamer(metricsChan <-chan GenerationMetrics) *MetricsStreamer {
	return &MetricsStreamer{
		metricsChan: metricsChan,
		subscribers: make([]chan GenerationMetrics, 0),
		done:        make(chan struct{}),
	}
}

// Subscribe returns a channel that will receive metrics
func (ms *MetricsStreamer) Subscribe() <-chan GenerationMetrics {
	subscriber := make(chan GenerationMetrics, 10) // buffered channel
	ms.subscribers = append(ms.subscribers, subscriber)
	return subscriber
}

// Start begins streaming metrics to all subscribers
func (ms *MetricsStreamer) Start(ctx context.Context) {
	ms.wg.Go(func() {
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
	})
}

// Stop stops the streamer and closes all subscriber channels
func (ms *MetricsStreamer) Stop() {
	select {
	case <-ms.done:
		// already closed
	default:
		close(ms.done)
	}
	ms.wg.Wait()
}

// broadcast sends metrics to all subscribers
func (ms *MetricsStreamer) broadcast(metrics GenerationMetrics) {
	for _, subscriber := range ms.subscribers {
		select {
		case subscriber <- metrics:
		default:
			// Skip if subscriber is not ready (non-blocking)
		}
	}
}

// closeSubscribers closes all subscriber channels
func (ms *MetricsStreamer) closeSubscribers() {
	for _, subscriber := range ms.subscribers {
		close(subscriber)
	}
	ms.subscribers = nil
}
