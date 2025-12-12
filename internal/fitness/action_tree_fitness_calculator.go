package fitness

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"go.uber.org/zap"
)

// ActionTreeFitnessCalculator implements fitness calculation for ActionTree individuals
type ActionTreeFitnessCalculator struct {
	serverAddr           string
	opponentType         string
	maxSteps             int
	actions              []individual.ActionTuple
	weightsPopulation    *[]individual.Evolvable
	actionTreePopulation *[]individual.Evolvable
	selectionPercentage  float64
	testCaseCount        int
	connectionPool       *TCPConnectionPool
	clientId             uint64
}

// NewActionTreeFitnessCalculator creates a new action tree fitness calculator
func NewActionTreeFitnessCalculator(serverAddr string, opponentType string, actions []individual.ActionTuple, maxSteps int, populations []*[]individual.Evolvable, selectionPercentage float64, poolSize int, testCaseCount int, timeout time.Duration) *ActionTreeFitnessCalculator {
	pool := NewTCPConnectionPool(serverAddr, poolSize, timeout)

	return &ActionTreeFitnessCalculator{
		serverAddr:           serverAddr,
		opponentType:         opponentType,
		maxSteps:             maxSteps,
		actions:              actions,
		weightsPopulation:    populations[0],
		actionTreePopulation: populations[1],
		testCaseCount:        testCaseCount,
		selectionPercentage:  selectionPercentage,
		connectionPool:       pool,
		clientId:             0,
	}
}

func (atfc *ActionTreeFitnessCalculator) getClientId() string {
	id := atomic.AddUint64(&atfc.clientId, 1)
	return fmt.Sprintf("client_%d", id)
}

type Client struct {
	ID      string
	Fitness float64
}

// Score computes a fitness-like value from a slice of numbers.
// Positives get diminishing returns (alpha < 1), negatives subtract directly.
func Score(value float64, alpha float64) float64 {
	if value > 0 {
		return math.Pow(value, alpha) // diminishing return for positives
	} else {
		return -1 * math.Pow(math.Abs(value), alpha) // negatives penalize directly
	}
}

// handleTestCase scenario more cleanly and share code
func (atfc *ActionTreeFitnessCalculator) handleTestCases(wi *individual.WeightsIndividual, tree *individual.ActionTreeIndividual, index int, fitnesses []Client) {
	sum := 0.0
	clientId := ""
	successCount := 0
	for range atfc.testCaseCount {
		fitness, currentClientId, err := atfc.SetupGameAndRun(wi, tree)
		if err != nil {
			zap.L().Error("Failed to setup game and run", zap.Error(err))
		} else {
			sum += Score(fitness, 0.5)
			clientId += currentClientId + " " + strconv.FormatFloat(fitness, 'f', -1, 64) + " : "
			successCount++
		}
	}
	// Avoid division by zero - if no successful test cases, use 0.0 fitness
	if successCount == 0 || atfc.testCaseCount == 0 {
		fitnesses[index] = Client{
			ID:      clientId,
			Fitness: 0.0,
		}
	} else {
		fitnesses[index] = Client{
			ID:      clientId,
			Fitness: sum / float64(successCount), // Average over successful test cases
		}
	}
}

// CalculateFitness evaluates the fitness of an ActionTree individual
func (atfc *ActionTreeFitnessCalculator) CalculateFitness(evolvable individual.Evolvable) {

	wi, wiok := evolvable.(*individual.WeightsIndividual)
	at, atok := evolvable.(*individual.ActionTreeIndividual)
	if !wiok && !atok {
		zap.L().Error("Expected ActionTreeIndividual or weightsIndividual", zap.String("type", fmt.Sprintf("%T", evolvable)))
		evolvable.SetFitness(0.0)
		return
	}

	fitnesses := make([]Client, max(len(*atfc.weightsPopulation), len(*atfc.actionTreePopulation)))
	if wiok {
		for index, evolvable := range *atfc.actionTreePopulation {
			tree, ok := (evolvable).(*individual.ActionTreeIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}
			atfc.handleTestCases(wi, tree, index, fitnesses)
		}
		return
	}

	if atok {
		actionTreeFitnesses := make([]Client, len(*atfc.actionTreePopulation))
		for index, evolvable := range *atfc.weightsPopulation {
			weights, ok := (evolvable).(*individual.WeightsIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}

			atfc.handleTestCases(weights, at, index, fitnesses)
		}
		at.SetFitness(actionTreeFitnesses[0].Fitness)
		at.SetClient(actionTreeFitnesses[0].ID)
	}

	sort.Slice(fitnesses, func(i, j int) bool {
		return (fitnesses[i].Fitness > fitnesses[j].Fitness)
	})
	if wiok {
		wi.SetFitness(fitnesses[0].Fitness)
		wi.SetClient(fitnesses[0].ID)
		return
	}

	at.SetFitness(fitnesses[0].Fitness)
	at.SetClient(fitnesses[0].ID)
}

func (atfc *ActionTreeFitnessCalculator) SetupGameAndRun(weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) (float64, string, error) {
	// Get connection from pool
	client, err := atfc.connectionPool.GetConnection()
	if err != nil {
		zap.L().Error("Failed to get connection from pool", zap.Error(err))
		return 0.0, "", fmt.Errorf("connection pool error: %w", err)
	}
	clientId := atfc.getClientId()
	zap.L().Debug("Got connection for game evaluation",
		zap.String("client_id", clientId),
		zap.String("weights_id", fmt.Sprintf("%p", weightsInd)),
		zap.String("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	// Ensure connection is returned to pool
	defer func() {
		if returnErr := atfc.connectionPool.ReturnConnection(client); returnErr != nil {
			zap.L().Error("Failed to return connection to pool", zap.Error(returnErr))
		} else {
			zap.L().Debug("Connection returned to pool successfully",
				zap.String("client_id", clientId))
		}
	}()

	// Small delay to ensure connection is stable
	time.Sleep(100 * time.Millisecond)

	// Connect to game
	connectedResp, err := client.ConnectToGame(clientId, atfc.opponentType)
	if err != nil {
		zap.L().Error("Failed to connect to game", zap.Error(err))
		return 0.0, "", fmt.Errorf("game connection error: %w", err)
	}

	zap.L().Debug("Connected to game",
		zap.String("agent_id", connectedResp.AgentID),
		zap.String("opponent_id", connectedResp.OpponentID))

	// Play game
	fitness := atfc.playGame(client, weightsInd, actionTreeInd)

	zap.L().Debug("Fitness calculated",
		zap.Float64("fitness", fitness),
		zap.String("agent_id", clientId))
	return fitness, clientId, nil
}

// playGame plays a single game and returns the fitness score
func (atfc *ActionTreeFitnessCalculator) playGame(client *TCPClient, weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	totalReward := 0.0
	zap.L().Debug("Starting game evaluation",
		zap.Int("max_steps", atfc.maxSteps),
		zap.String("weights_id", fmt.Sprintf("%p", weightsInd)),
		zap.String("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	// Get observation from conncetion (Reset)
	obs, err := client.ReceiveObservation()
	if err != nil {
		zap.L().Error("Failed to receive observation", zap.Error(err))
	}
	actionExecutor := NewActionExecutor(atfc.actions)
	actionExecutor.validator.SetMountains(obs.Info)
	constantActionSelectionTracker := make([]bool, 3)

	// Send action to server
	err = client.SendAction([]int{1, 0, 0, 0, 0})
	if err != nil {
		zap.L().Error("Failed to send action", zap.Error(err))
	}
	for step := range atfc.maxSteps {
		totalReward += obs.Reward
		obs, err = client.ReceiveObservation()
		if err != nil {
			zap.L().Error("Failed to receive observation", zap.Error(err))
			break
		}
		// Log game progress every 10 steps
		if step%10 == 0 {
			zap.L().Debug("Game progress",
				zap.Int("step", step),
				zap.Float64("reward", obs.Reward),
				zap.Bool("terminated", obs.Terminated),
				zap.Bool("truncated", obs.Truncated),
				zap.Float64("total_reward", totalReward))
		}

		// Check if game is over
		if obs.Terminated || obs.Truncated {
			zap.L().Debug("Game ended",
				zap.Int("step", step),
				zap.Float64("final_reward", obs.Reward),
				zap.Float64("total_reward", totalReward+obs.Reward),
				zap.Bool("terminated", obs.Terminated),
				zap.Bool("truncated", obs.Truncated))
			// Extract final game state
			totalReward += obs.Reward
			break
		}

		// Execute action trees to get action
		action, err := actionExecutor.ExecuteActionTreesWithSoftmax(actionTreeInd, weightsInd, obs.Observation, obs.Info, &constantActionSelectionTracker)
		if err != nil {
			zap.L().Error("Failed to execute action trees", zap.Error(err))
			// Send a default action (pass) instead of panicking
		}

		// Send action to server
		err = client.SendAction(action)
		if err != nil {
			zap.L().Error("Failed to send action", zap.Error(err))
			break
		}

		totalReward += obs.Reward
	}
	for _, actionIsntConstant := range constantActionSelectionTracker {
		if !actionIsntConstant {
			totalReward -= 10 // Try to reduce them but not totally kill them as genome parts could still be good if one tree is bad
		}
	}
	if totalReward > 30.0 {
		err = client.RequestReplay()
		if err != nil {
			zap.L().Error("Failed to getReplay", zap.Error(err))
		}
	}

	zap.L().Debug("Final fitness calculation",
		zap.Float64("total_reward", totalReward))

	return totalReward
}

// Close closes the connection pool and cleans up resources
func (atfc *ActionTreeFitnessCalculator) Close() error {
	if atfc.connectionPool != nil {
		return atfc.connectionPool.Close()
	}
	return nil
}
