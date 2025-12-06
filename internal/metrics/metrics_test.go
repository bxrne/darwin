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
		Generation:      1,
		Duration:        100 * time.Millisecond,
		BestFitness:     0.9,
		BestDescription: "best_individual",
		MinDepth:        2,
		MaxDepth:        5,
		AvgDepth:        3.5,
		AvgFitness:      0.7,
		MinFitness:      0.5,
		MaxFitness:      0.9,
		PopulationSize:  10,
		Timestamp:       time.Now(),
	}

	suite.metricsChan <- testMetrics

	select {
	case received := <-suite.subscriber:
		assert.Equal(suite.T(), testMetrics.Generation, received.Generation)
		assert.Equal(suite.T(), testMetrics.PopulationSize, received.PopulationSize)
		assert.Equal(suite.T(), testMetrics.BestDescription, received.BestDescription)
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

	<-started // wait for Start to begin

	streamer.Stop()

	// Subscribers should be closed
	select {
	case _, ok := <-subscriber:
		assert.False(suite.T(), ok, "Subscriber channel should be closed")
	default:
		suite.Fail("Subscriber channel should be closed")
	}
}
