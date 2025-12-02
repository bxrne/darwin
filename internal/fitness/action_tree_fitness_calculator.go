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
}

// NewActionTreeFitnessCalculator creates a new action tree fitness calculator
func NewActionTreeFitnessCalculator(serverAddr string, opponentType string, actions []string, numInputs int, maxSteps int, populations []*[]individual.Evolvable, selectionPercentage float64) *ActionTreeFitnessCalculator {
	return &ActionTreeFitnessCalculator{
		serverAddr:           serverAddr,
		opponentType:         opponentType,
		maxSteps:             maxSteps,
		actionExecutor:       NewActionExecutor(actions, numInputs),
		weightsPopulation:    populations[0],
		actionTreePopulation: populations[1],
		selectionPercentage:  selectionPercentage,
	}
}

// CalculateFitness evaluates the fitness of an ActionTree individual
func (atfc *ActionTreeFitnessCalculator) CalculateFitness(evolvable individual.Evolvable) {
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
			weightsFitnesses = append(weightsFitnesses, atfc.SetupGameAndRun(wi, tree))
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
			actionTreeFitnesses = append(actionTreeFitnesses, atfc.SetupGameAndRun(weights, at))
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

func (atfc *ActionTreeFitnessCalculator) SetupGameAndRun(weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	// Create TCP client for this evaluation
	client := NewTCPClient(atfc.serverAddr)
	defer func() {
		if err := client.Disconnect(); err != nil {
			logmgr.Error("Failed to disconnect client", logmgr.Field("error", err))
		}
	}()

	// Connect to server
	err := client.Connect()
	if err != nil {
		logmgr.Error("Failed to connect to game server", logmgr.Field("error", err))
		panic("FAILED")
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			logmgr.Error("Failed to disconnect client", logmgr.Field("error", err))
		}
	}()

	// Small delay to ensure connection is stable
	time.Sleep(100 * time.Millisecond)

	// Connect to game
	connectedResp, err := client.ConnectToGame(atfc.opponentType)
	if err != nil {
		logmgr.Error("Failed to connect to game", logmgr.Field("error", err))
		panic("FAILED")
	}

	logmgr.Debug("Connected to game",
		logmgr.Field("agent_id", connectedResp.AgentID),
		logmgr.Field("opponent_id", connectedResp.OpponentID))

	// Play the game
	fitness := atfc.playGame(client, weightsInd, actionTreeInd)

	logmgr.Debug("Fitness calculated",
		logmgr.Field("fitness", fitness),
		logmgr.Field("agent_id", connectedResp.AgentID))
	return fitness
}

// playGame plays a single game and returns the fitness score
func (atfc *ActionTreeFitnessCalculator) playGame(client *TCPClient, weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	var totalReward float64
	var survivalTime int
	var finalScore float64
	var won bool

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
			finalScore = ExtractGameScore(obs.Observation)
			totalReward = obs.Reward

			// Try to get winner info
			if winner, exists := obs.Info["winner"]; exists {
				if winnerStr, ok := winner.(string); ok {
					won = (winnerStr == "client_1") // Assuming we're client_1
				}
			}
			break
		}

		// Extract inputs from observation
		inputs := atfc.extractInputs(obs.Observation)

		// Execute action trees to get action
		action, err := atfc.actionExecutor.ExecuteActionTrees(actionTreeInd, inputs, weightsInd)
		if err != nil {
			logmgr.Error("Failed to execute action trees", logmgr.Field("error", err))
			// Send a default action
			action = "move" // Default action
		}

		// Send action to server
		err = client.SendAction(action)
		if err != nil {
			logmgr.Error("Failed to send action", logmgr.Field("error", err))
			break
		}

		// Small delay to prevent overwhelming the server
		time.Sleep(10 * time.Millisecond)

		survivalTime = step + 1
		totalReward += obs.Reward
	}

	// Calculate fitness based on multiple factors
	fitness := atfc.calculateFitnessScore(totalReward, survivalTime, finalScore, won)

	return fitness
}

// extractInputs extracts numeric inputs from observation data
func (atfc *ActionTreeFitnessCalculator) extractInputs(observation map[string]any) []float64 {
	inputs := make([]float64, 4) // x, y, z, w (max 4 inputs)

	// Try to extract common input fields
	inputFields := []string{"x", "y", "z", "w", "score", "reward", "time", "step"}

	for i, field := range inputFields {
		if i >= len(inputs) {
			break
		}

		if value, exists := observation[field]; exists {
			switch v := value.(type) {
			case float64:
				inputs[i] = v
			case int:
				inputs[i] = float64(v)
			case string:
				if val, err := parseFloat(v); err == nil {
					inputs[i] = val
				}
			}
		}
	}

	// If no specific inputs found, use generic values
	if inputs[0] == 0 && inputs[1] == 0 {
		// Use some observation data as inputs
		if score, ok := observation["score"].(float64); ok {
			inputs[0] = score
		}
		if reward, ok := observation["reward"].(float64); ok {
			inputs[1] = reward
		}
		inputs[2] = 1.0 // Constant
		inputs[3] = 0.5 // Another constant
	}

	return inputs
}

// calculateFitnessScore computes the final fitness based on game performance
func (atfc *ActionTreeFitnessCalculator) calculateFitnessScore(totalReward float64, survivalTime int, finalScore float64, won bool) float64 {
	// Base fitness from total reward
	fitness := totalReward

	// Bonus for survival time
	survivalBonus := float64(survivalTime) * 0.1
	fitness += survivalBonus

	// Bonus for final score
	scoreBonus := finalScore * 0.5
	fitness += scoreBonus

	// Large bonus for winning
	if won {
		fitness += 100.0
	}

	// Penalty for very short games (quick death)
	if survivalTime < 10 {
		fitness *= 0.1
	}

	// Ensure fitness is non-negative
	if fitness < 0 {
		fitness = 0
	}

	// Apply logarithmic scaling to prevent extremely high values
	fitness = math.Log(fitness+1) * 10

	return fitness
}
