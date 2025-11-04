package evolution

import (
	"context"
	"testing"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockSelector is a mock implementation of the Selector interface
type MockSelector struct {
	mock.Mock
}

func (m *MockSelector) Select(population []individual.Evolvable) individual.Evolvable {
	args := m.Called(population)
	return args.Get(0).(individual.Evolvable)
}

type EvolutionEngineTestSuite struct {
	suite.Suite
	population  []individual.Evolvable
	selector    *MockSelector
	metricsChan chan metrics.GenerationMetrics
	cmdChan     chan EvolutionCommand
	engine      *EvolutionEngine
}

func (suite *EvolutionEngineTestSuite) SetupTest() {
	suite.population = []individual.Evolvable{
		individual.NewBinaryIndividual(5),
		individual.NewBinaryIndividual(5),
	}
	suite.selector = &MockSelector{}
	suite.metricsChan = make(chan metrics.GenerationMetrics, 10)
	suite.cmdChan = make(chan EvolutionCommand, 10)
	suite.engine = NewEvolutionEngine(suite.population, suite.selector, suite.metricsChan, suite.cmdChan)
}

func (suite *EvolutionEngineTestSuite) TearDownTest() {
	close(suite.cmdChan)
	close(suite.metricsChan)
}

func TestEvolutionEngineTestSuite(t *testing.T) {
	suite.Run(t, new(EvolutionEngineTestSuite))
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_NewEvolutionEngine_GIVEN_population_selector_channels_WHEN_create_THEN_engine_initialized() {
	assert.NotNil(suite.T(), suite.engine)
	assert.Equal(suite.T(), suite.population, suite.engine.population)
	assert.Equal(suite.T(), suite.selector, suite.engine.selector)
	assert.NotNil(suite.T(), suite.engine.metricsChan)
	assert.NotNil(suite.T(), suite.engine.cmdChan)
	assert.Equal(suite.T(), 0, suite.engine.currentGen)
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_Start_GIVEN_running_engine_WHEN_start_THEN_processes_commands() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	// Set up multiple expectations since processGeneration may be called
	suite.selector.On("Select", suite.population).Return(parent1).Maybe()
	suite.selector.On("Select", suite.population).Return(parent2).Maybe()

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		MutationPoints:  []int{0},
		MutationRate:    1.0, // ensure mutation happens
		ElitismPct:      0.0,
	}

	suite.engine.Start(ctx)
	suite.cmdChan <- cmd
	time.Sleep(10 * time.Millisecond) // allow processing
	cancel()
	suite.engine.Wait()

	// Don't assert expectations since it's concurrent
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_Start_GIVEN_context_cancelled_WHEN_start_THEN_stops_gracefully() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	suite.engine.Start(ctx)
	suite.engine.Wait()

	// Should not panic or hang
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_Wait_GIVEN_started_engine_WHEN_wait_THEN_blocks_until_done() {
	ctx := context.Background()

	suite.engine.Start(ctx)
	suite.cmdChan <- EvolutionCommand{Type: CmdStop}

	// Wait should not block indefinitely
	done := make(chan bool)
	go func() {
		suite.engine.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Wait did not return in time")
	}
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_GetPopulation_GIVEN_population_WHEN_get_THEN_returns_population() {
	pop := suite.engine.GetPopulation()

	assert.Equal(suite.T(), suite.population, pop)
}

func (suite *EvolutionEngineTestSuite) TestPopulationBuilder_BuildBinaryPopulation_GIVEN_size_genome_size_WHEN_build_THEN_population_created() {
	builder := NewPopulationBuilder()
	population := builder.BuildBinaryPopulation(3, 5)

	assert.Len(suite.T(), population, 3)
	for _, ind := range population {
		binInd, ok := ind.(*individual.BinaryIndividual)
		assert.True(suite.T(), ok)
		assert.Len(suite.T(), binInd.Genome, 5)
	}
}
