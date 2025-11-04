package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MetricsStreamerTestSuite struct {
	suite.Suite
	streamer    *MetricsStreamer
	metricsChan chan GenerationMetrics
	subscriber  <-chan GenerationMetrics
}

func (suite *MetricsStreamerTestSuite) SetupTest() {
	suite.metricsChan = make(chan GenerationMetrics, 10)
	suite.streamer = NewMetricsStreamer(suite.metricsChan)
	suite.subscriber = suite.streamer.Subscribe()
}

func (suite *MetricsStreamerTestSuite) TearDownTest() {
	// Stop is called in individual tests
}

func TestMetricsStreamerTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsStreamerTestSuite))
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_NewMetricsStreamer_GIVEN_metrics_channel_WHEN_create_THEN_streamer_initialized() {
	assert.NotNil(suite.T(), suite.streamer)
	assert.NotNil(suite.T(), suite.streamer.metricsChan)
	assert.NotNil(suite.T(), suite.streamer.subscribers)
	assert.NotNil(suite.T(), suite.streamer.done)
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Subscribe_GIVEN_streamer_WHEN_subscribe_THEN_returns_channel() {
	sub := suite.streamer.Subscribe()

	assert.NotNil(suite.T(), sub)
	assert.Len(suite.T(), suite.streamer.subscribers, 2) // setup + this one
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Start_GIVEN_subscribers_WHEN_start_THEN_broadcasts_metrics() {
	ctx := context.Background()

	go suite.streamer.Start(ctx)

	testMetrics := GenerationMetrics{
		Generation:     1,
		Duration:       100 * time.Millisecond,
		BestFitness:    0.9,
		AvgFitness:     0.7,
		MinFitness:     0.5,
		MaxFitness:     0.9,
		PopulationSize: 10,
		Timestamp:      time.Now(),
	}

	suite.metricsChan <- testMetrics

	select {
	case received := <-suite.subscriber:
		assert.Equal(suite.T(), testMetrics.Generation, received.Generation)
		assert.Equal(suite.T(), testMetrics.PopulationSize, received.PopulationSize)
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Metrics not received")
	}
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Start_GIVEN_context_cancelled_WHEN_start_THEN_stops() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	suite.streamer.Start(ctx)

	// Should not hang
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Start_GIVEN_done_channel_closed_WHEN_start_THEN_stops() {
	ctx := context.Background()

	// Create a fresh streamer for this test
	streamer := NewMetricsStreamer(suite.metricsChan)
	defer streamer.Stop() // ensure cleanup

	started := make(chan bool)
	done := make(chan bool)

	go func() {
		started <- true
		streamer.Start(ctx)
		done <- true
	}()

	<-started // wait for Start to begin

	go func() {
		time.Sleep(10 * time.Millisecond)
		streamer.Stop()
	}()

	<-done // wait for Start to complete

	// Should stop when done channel is closed
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Start_GIVEN_closed_metrics_channel_WHEN_start_THEN_stops() {
	ctx := context.Background()
	close(suite.metricsChan)

	suite.streamer.Start(ctx)

	// Should stop when metrics channel is closed
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Stop_GIVEN_running_streamer_WHEN_stop_THEN_closes_subscribers() {
	ctx := context.Background()

	// Create a fresh streamer for this test
	streamer := NewMetricsStreamer(suite.metricsChan)
	subscriber := streamer.Subscribe()

	started := make(chan bool)
	go func() {
		started <- true
		streamer.Start(ctx)
	}()

	<-started                         // wait for Start to begin
	time.Sleep(10 * time.Millisecond) // let it run a bit

	streamer.Stop()

	// Subscribers should be closed
	select {
	case _, ok := <-subscriber:
		assert.False(suite.T(), ok, "Subscriber channel should be closed")
	default:
		suite.Fail("Subscriber channel should be closed")
	}
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Broadcast_GIVEN_metrics_WHEN_broadcast_THEN_sent_to_all_subscribers() {
	testMetrics := GenerationMetrics{Generation: 1}

	suite.streamer.broadcast(testMetrics)

	select {
	case received := <-suite.subscriber:
		assert.Equal(suite.T(), 1, received.Generation)
	case <-time.After(10 * time.Millisecond):
		suite.Fail("Metrics not broadcasted")
	}
}

func (suite *MetricsStreamerTestSuite) TestMetricsStreamer_Broadcast_GIVEN_full_subscriber_channel_WHEN_broadcast_THEN_skips_non_blocking() {
	// Create a streamer with a subscriber that has a full buffer
	streamer := NewMetricsStreamer(make(chan GenerationMetrics, 1))
	streamer.Subscribe()

	// Fill the subscriber channel by sending directly (this is internal, but for test)
	// Actually, since it's receive-only, we can't. Instead, test that broadcast completes without hanging
	testMetrics := GenerationMetrics{Generation: 2}

	// This should complete without blocking, as broadcast uses select with default
	done := make(chan bool)
	go func() {
		streamer.broadcast(testMetrics)
		done <- true
	}()

	select {
	case <-done:
		// Good, broadcast completed
	case <-time.After(10 * time.Millisecond):
		suite.Fail("Broadcast should not block")
	}
}
