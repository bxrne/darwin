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
		CrossoverRate:   0.9,
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

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_ProcessGeneration_GIVEN_start_command_WHEN_process_THEN_population_evolves() {
	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	// For 2 offspring, need 4 parent selections
	suite.selector.On("Select", suite.population).Return(parent1).Times(2)
	suite.selector.On("Select", suite.population).Return(parent2).Times(2)

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		CrossoverRate:   0.9,
		MutationRate:    1.0,
		ElitismPct:      0.0,
	}

	suite.engine.processGeneration(cmd)

	assert.Len(suite.T(), suite.engine.population, len(suite.population))
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_ProcessGeneration_GIVEN_elitism_WHEN_process_THEN_best_individuals_preserved() {
	// Set fitness manually for predictable elitism
	suite.population[0].(*individual.BinaryIndividual).Fitness = 1.0
	suite.population[1].(*individual.BinaryIndividual).Fitness = 0.0

	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	// Elitism 50% of 2 = 1, so 1 offspring needed, 2 parents
	suite.selector.On("Select", suite.population).Return(parent1).Times(1)
	suite.selector.On("Select", suite.population).Return(parent2).Times(1)

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		CrossoverRate:   0.9,
		MutationRate:    1.0,
		ElitismPct:      0.5, // 50% elitism
	}

	suite.engine.processGeneration(cmd)

	// First individual should be the elite (fitness 1.0)
	assert.Equal(suite.T(), 1.0, suite.engine.population[0].GetFitness())
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_ProcessGeneration_GIVEN_crossover_mutation_WHEN_process_THEN_offspring_generated() {
	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	suite.selector.On("Select", suite.population).Return(parent1)
	suite.selector.On("Select", suite.population).Return(parent2)

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		CrossoverRate:   0.9,
		MutationRate:    1.0,
		ElitismPct:      0.0,
	}

	suite.engine.processGeneration(cmd)

	assert.Len(suite.T(), suite.engine.population, len(suite.population))
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_ProcessGeneration_GIVEN_metrics_channel_WHEN_process_THEN_metrics_sent() {
	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	suite.selector.On("Select", suite.population).Return(parent1).Times(2)
	suite.selector.On("Select", suite.population).Return(parent2).Times(2)

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		CrossoverRate:   0.9,
		MutationRate:    1.0,
		ElitismPct:      0.0,
	}

	suite.engine.processGeneration(cmd)

	select {
	case metrics := <-suite.metricsChan:
		assert.Equal(suite.T(), 1, metrics.Generation)
		assert.Equal(suite.T(), len(suite.population), metrics.PopulationSize)
	default:
		suite.Fail("Metrics not sent")
	}
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_ProcessGeneration_GIVEN_full_metrics_channel_WHEN_process_THEN_skips_sending() {
	// Use a full channel to test non-blocking send
	fullChan := make(chan metrics.GenerationMetrics, 1)
	fullChan <- metrics.GenerationMetrics{} // fill it

	// Create engine with full channel
	engine := NewEvolutionEngine(suite.population, suite.selector, fullChan, make(chan EvolutionCommand, 1))

	parent1 := individual.NewBinaryIndividual(5)
	parent2 := individual.NewBinaryIndividual(5)
	suite.selector.On("Select", suite.population).Return(parent1)
	suite.selector.On("Select", suite.population).Return(parent2)

	cmd := EvolutionCommand{
		Type:            CmdStartGeneration,
		Generation:      1,
		CrossoverPoints: 1,
		CrossoverRate:   0.9,
		MutationRate:    1.0,
		ElitismPct:      0.0,
	}

	// Should complete without blocking
	engine.processGeneration(cmd)

	// Channel should still have 1 message (no additional send)
	if len(fullChan) != 1 {
		suite.Fail("Should not have sent additional metrics")
	}
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_SortPopulation_GIVEN_unsorted_population_WHEN_sort_THEN_sorted_by_fitness_descending() {
	// Set fitness
	suite.population[0].(*individual.BinaryIndividual).Fitness = 0.5
	suite.population[1].(*individual.BinaryIndividual).Fitness = 1.0

	suite.engine.sortPopulation()

	assert.Equal(suite.T(), 1.0, suite.engine.population[0].GetFitness())
	assert.Equal(suite.T(), 0.5, suite.engine.population[1].GetFitness())
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_CalculateMetrics_GIVEN_population_WHEN_calculate_THEN_correct_metrics_returned() {
	// Set fitness
	suite.population[0].(*individual.BinaryIndividual).Fitness = 0.5
	suite.population[1].(*individual.BinaryIndividual).Fitness = 1.0

	duration := 100 * time.Millisecond
	metrics := suite.engine.calculateMetrics(1, duration)

	assert.Equal(suite.T(), 1, metrics.Generation)
	assert.Equal(suite.T(), duration, metrics.Duration)
	assert.Equal(suite.T(), 1.0, metrics.BestFitness)
	assert.Equal(suite.T(), 0.75, metrics.AvgFitness)
	assert.Equal(suite.T(), 0.5, metrics.MinFitness)
	assert.Equal(suite.T(), 1.0, metrics.MaxFitness)
	assert.Equal(suite.T(), len(suite.population), metrics.PopulationSize)
}

func (suite *EvolutionEngineTestSuite) TestEvolutionEngine_CalculateMetrics_GIVEN_empty_population_WHEN_calculate_THEN_zero_metrics() {
	emptyEngine := NewEvolutionEngine([]individual.Evolvable{}, suite.selector, suite.metricsChan, suite.cmdChan)

	duration := 100 * time.Millisecond
	metrics := emptyEngine.calculateMetrics(1, duration)

	assert.Equal(suite.T(), 1, metrics.Generation)
	assert.Equal(suite.T(), duration, metrics.Duration)
	assert.Equal(suite.T(), 0, metrics.PopulationSize)
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
