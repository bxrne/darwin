package fitness

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/logmgr"
)

// ActionTreeFitnessCalculator implements fitness calculation for ActionTree individuals
type ActionTreeFitnessCalculator struct {
	serverAddr           string
	opponentType         string
	maxSteps             int
	actionExecutor       *ActionExecutor
	weightsPopulation    *[]individual.Evolvable
	actionTreePopulation *[]individual.Evolvable
	selectionPercentage  float64
	connectionPool       *TCPConnectionPool
}

// NewActionTreeFitnessCalculator creates a new action tree fitness calculator
func NewActionTreeFitnessCalculator(serverAddr string, opponentType string, actions []string, maxSteps int, populations []*[]individual.Evolvable, selectionPercentage float64, poolSize int, timeout time.Duration) *ActionTreeFitnessCalculator {
	pool := NewTCPConnectionPool(serverAddr, poolSize, timeout)

	return &ActionTreeFitnessCalculator{
		serverAddr:           serverAddr,
		opponentType:         opponentType,
		maxSteps:             maxSteps,
		actionExecutor:       NewActionExecutor(actions),
		weightsPopulation:    populations[0],
		actionTreePopulation: populations[1],
		selectionPercentage:  selectionPercentage,
		connectionPool:       pool,
	}
}

// CalculateFitness evaluates the fitness of an ActionTree individual
func (atfc *ActionTreeFitnessCalculator) CalculateFitness(evolvable individual.Evolvable) {
	// Log connection pool stats periodically
	if atfc.connectionPool != nil {
		stats := atfc.connectionPool.GetStats()
		logmgr.Debug("Connection pool stats",
			logmgr.Field("active", stats["active_count"]),
			logmgr.Field("available", stats["available"]),
			logmgr.Field("total_created", stats["total_created"]))
	}

	wi, wiok := evolvable.(*individual.WeightsIndividual)
	at, atok := evolvable.(*individual.ActionTreeIndividual)
	if !wiok && !atok {
		logmgr.Error("Expected ActionTreeIndividual or weightsIndividual", logmgr.Field("type", fmt.Sprintf("%T", evolvable)))
		evolvable.SetFitness(0.0)
		return
	}

	if wiok {
		weightsFitnesses := make([]float64, len(*atfc.weightsPopulation))
		for _, evolvable := range *atfc.actionTreePopulation {
			tree, ok := (evolvable).(*individual.ActionTreeIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}
			fitness, err := atfc.SetupGameAndRun(wi, tree)
			if err != nil {
				logmgr.Error("Failed to setup game and run", logmgr.Field("error", err.Error()))
				weightsFitnesses = append(weightsFitnesses, 0.0)
			} else {
				weightsFitnesses = append(weightsFitnesses, fitness)
			}
		}
		weightsCount := int(float64(len(weightsFitnesses)) * atfc.selectionPercentage)
		sort.Float64s(weightsFitnesses)
		total_fitness := 0.0
		for i := range weightsCount {
			total_fitness = total_fitness + weightsFitnesses[i]
		}
		wi.SetFitness(total_fitness / float64(weightsCount))
		return

	}

	if atok {
		actionTreeFitnesses := make([]float64, len(*atfc.actionTreePopulation))
		for _, evolvable := range *atfc.weightsPopulation {
			weights, ok := (evolvable).(*individual.WeightsIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}
			fitness, err := atfc.SetupGameAndRun(weights, at)
			if err != nil {
				logmgr.Error("Failed to setup game and run", logmgr.Field("error", err.Error()))
				actionTreeFitnesses = append(actionTreeFitnesses, 0.0)
			} else {
				actionTreeFitnesses = append(actionTreeFitnesses, fitness)
			}
		}
		actionTreeCount := int(float64(len(actionTreeFitnesses)) * atfc.selectionPercentage)
		sort.Float64s(actionTreeFitnesses)
		total_fitness := 0.0
		for i := range actionTreeCount {
			total_fitness = total_fitness + actionTreeFitnesses[i]
		}
		at.SetFitness(total_fitness / float64(actionTreeCount))
		return

	}

}

func (atfc *ActionTreeFitnessCalculator) SetupGameAndRun(weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) (float64, error) {
	// Get connection from pool
	client, err := atfc.connectionPool.GetConnection()
	if err != nil {
		logmgr.Error("Failed to get connection from pool", logmgr.Field("error", err.Error()))
		return 0.0, fmt.Errorf("connection pool error: %w", err)
	}

	logmgr.Debug("Got connection for game evaluation",
		logmgr.Field("client_id", fmt.Sprintf("%p", client)),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)),
		logmgr.Field("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	// Ensure connection is returned to pool
	defer func() {
		if returnErr := atfc.connectionPool.ReturnConnection(client); returnErr != nil {
			logmgr.Error("Failed to return connection to pool", logmgr.Field("error", returnErr.Error()))
		} else {
			logmgr.Debug("Connection returned to pool successfully",
				logmgr.Field("client_id", fmt.Sprintf("%p", client)))
		}
	}()

	// Small delay to ensure connection is stable
	time.Sleep(100 * time.Millisecond)

	// Connect to game
	connectedResp, err := client.ConnectToGame(atfc.opponentType)
	if err != nil {
		logmgr.Error("Failed to connect to game", logmgr.Field("error", err))
		return 0.0, fmt.Errorf("game connection error: %w", err)
	}

	logmgr.Debug("Connected to game",
		logmgr.Field("agent_id", connectedResp.AgentID),
		logmgr.Field("opponent_id", connectedResp.OpponentID))

	// Play game
	fitness := atfc.playGame(client, weightsInd, actionTreeInd)

	logmgr.Debug("Fitness calculated",
		logmgr.Field("fitness", fitness),
		logmgr.Field("agent_id", connectedResp.AgentID))
	return fitness, nil
}

// playGame plays a single game and returns the fitness score
func (atfc *ActionTreeFitnessCalculator) playGame(client *TCPClient, weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	totalReward := 0.0

	for step := 0; step < atfc.maxSteps; step++ {
		// Get current observation (first one after connection)
		obs, err := client.ReceiveObservation()
		if err != nil {
			logmgr.Error("Failed to receive observation", logmgr.Field("error", err))
			break
		}

		// Check if game is over
		if obs.Terminated || obs.Truncated {
			// Extract final game state
			totalReward += obs.Reward
			break
		}

		// Execute action trees to get action
		action, err := atfc.actionExecutor.ExecuteActionTreesWithSoftmax(actionTreeInd, weightsInd, obs.Observation)

		if err != nil {
			logmgr.Error("Failed to execute action trees", logmgr.Field("error", err))
			// Send a default action (pass) instead of panicking
			action = []int{0, 0, 0, 0, 0}
		}

		// Send action to server
		err = client.SendAction(action)
		if err != nil {
			logmgr.Error("Failed to send action", logmgr.Field("error", err))
			break
		}

		// Small delay to prevent overwhelming the server
		time.Sleep(10 * time.Millisecond)

		totalReward += obs.Reward

	}

	fitness := math.Log(totalReward+1) * 10

	return fitness
}

// Close closes the connection pool and cleans up resources
func (atfc *ActionTreeFitnessCalculator) Close() error {
	if atfc.connectionPool != nil {
		return atfc.connectionPool.Close()
	}
	return nil
}
