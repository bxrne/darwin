package fitness

import (
	"fmt"
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
func NewActionTreeFitnessCalculator(serverAddr string, opponentType string, actions []individual.ActionTuple, maxSteps int, populations []*[]individual.Evolvable, selectionPercentage float64, poolSize int, timeout time.Duration) *ActionTreeFitnessCalculator {
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

	wi, wiok := evolvable.(*individual.WeightsIndividual)
	at, atok := evolvable.(*individual.ActionTreeIndividual)
	if !wiok && !atok {
		logmgr.Error("Expected ActionTreeIndividual or weightsIndividual", logmgr.Field("type", fmt.Sprintf("%T", evolvable)))
		evolvable.SetFitness(0.0)
		return
	}

	if wiok {
		weightsFitnesses := make([]float64, 0, len(*atfc.actionTreePopulation))
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
				logmgr.Info("Game fitness calculated",
					logmgr.Field("fitness", fitness),
					logmgr.Field("weights_id", fmt.Sprintf("%p", wi)),
					logmgr.Field("tree_id", fmt.Sprintf("%p", tree)),
					logmgr.Field("game_index", len(weightsFitnesses)))
				weightsFitnesses = append(weightsFitnesses, fitness)
			}
		}
		if len(weightsFitnesses) == 0 {
			wi.SetFitness(0.0)
			return
		}
		weightsCount := int(float64(len(weightsFitnesses)) * atfc.selectionPercentage)
		if weightsCount == 0 {
			weightsCount = 1
		}
		sort.Float64s(weightsFitnesses)
		// Take the best (largest) fitnesses - for negative fitnesses, largest = best
		total_fitness := 0.0
		startIdx := len(weightsFitnesses) - weightsCount
		for i := startIdx; i < len(weightsFitnesses); i++ {
			total_fitness = total_fitness + weightsFitnesses[i]
		}
		finalFitness := total_fitness / float64(weightsCount)
		logmgr.Info("Weights fitness aggregation",
			logmgr.Field("total_fitness", total_fitness),
			logmgr.Field("weights_count", weightsCount),
			logmgr.Field("final_fitness", finalFitness),
			logmgr.Field("fitnesses", weightsFitnesses))
		wi.SetFitness(finalFitness)
		return

	}

	if atok {
		actionTreeFitnesses := make([]float64, 0, len(*atfc.weightsPopulation))
		desc := at.Describe()
		descLen := len(desc)
		if descLen > 100 {
			descLen = 100
		}
		logmgr.Info("Starting ActionTree fitness calculation",
			logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
			logmgr.Field("weights_population_size", len(*atfc.weightsPopulation)),
			logmgr.Field("action_tree_description", desc[:descLen])) // Log first 100 chars of description
		for idx, evolvable := range *atfc.weightsPopulation {
			weights, ok := (evolvable).(*individual.WeightsIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}
			fitness, err := atfc.SetupGameAndRun(weights, at)
			if err != nil {
				logmgr.Error("Failed to setup game and run", logmgr.Field("error", err.Error()))
				actionTreeFitnesses = append(actionTreeFitnesses, 0.0)
				logmgr.Info("Added 0.0 fitness due to error",
					logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
					logmgr.Field("weight_index", idx),
					logmgr.Field("error", err.Error()))
			} else {
				logmgr.Info("Game fitness calculated for ActionTree",
					logmgr.Field("fitness", fitness),
					logmgr.Field("weights_id", fmt.Sprintf("%p", weights)),
					logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
					logmgr.Field("weight_index", idx),
					logmgr.Field("total_games_so_far", len(actionTreeFitnesses)),
					logmgr.Field("raw_fitness_value", fitness))
				actionTreeFitnesses = append(actionTreeFitnesses, fitness)
				logmgr.Info("Appended fitness to slice",
					logmgr.Field("fitness", fitness),
					logmgr.Field("slice_length", len(actionTreeFitnesses)),
					logmgr.Field("all_fitnesses_so_far", fmt.Sprintf("%v", actionTreeFitnesses)))
			}
		}
		if len(actionTreeFitnesses) == 0 {
			at.SetFitness(0.0)
			return
		}
		actionTreeCount := int(float64(len(actionTreeFitnesses)) * atfc.selectionPercentage)
		if actionTreeCount == 0 {
			actionTreeCount = 1
		}
		// Ensure we don't select more than available
		if actionTreeCount > len(actionTreeFitnesses) {
			actionTreeCount = len(actionTreeFitnesses)
		}
		sort.Float64s(actionTreeFitnesses)
		// Take the best (largest) fitnesses - for negative fitnesses, largest = best
		total_fitness := 0.0
		startIdx := len(actionTreeFitnesses) - actionTreeCount
		if startIdx < 0 {
			startIdx = 0
		}
		// CRITICAL: Only sum the selected fitnesses, not all of them!
		selectedFitnesses := actionTreeFitnesses[startIdx:]
		for i := 0; i < len(selectedFitnesses); i++ {
			total_fitness = total_fitness + selectedFitnesses[i]
		}
		if actionTreeCount == 0 {
			logmgr.Error("actionTreeCount is 0, this should not happen!")
			at.SetFitness(0.0)
			return
		}
		// Double-check: total_fitness should be sum of only actionTreeCount values
		if len(selectedFitnesses) != actionTreeCount {
			logmgr.Error("SELECTED FITNESSES COUNT MISMATCH!",
				logmgr.Field("expected_count", actionTreeCount),
				logmgr.Field("actual_count", len(selectedFitnesses)),
				logmgr.Field("start_idx", startIdx),
				logmgr.Field("total_fitnesses", len(actionTreeFitnesses)))
		}
		finalFitness := total_fitness / float64(actionTreeCount)
		logmgr.Info("ActionTree fitness aggregation",
			logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
			logmgr.Field("total_fitnesses", len(actionTreeFitnesses)),
			logmgr.Field("selection_percentage", atfc.selectionPercentage),
			logmgr.Field("action_tree_count", actionTreeCount),
			logmgr.Field("start_idx", startIdx),
			logmgr.Field("total_fitness", total_fitness),
			logmgr.Field("final_fitness", finalFitness),
			logmgr.Field("all_fitnesses", fmt.Sprintf("%v", actionTreeFitnesses)),
			logmgr.Field("selected_fitnesses", fmt.Sprintf("%v", actionTreeFitnesses[startIdx:])))
		at.SetFitness(finalFitness)
		verifiedFitness := at.GetFitness()
		if verifiedFitness != finalFitness {
			logmgr.Error("FITNESS MISMATCH!",
				logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
				logmgr.Field("calculated_fitness", finalFitness),
				logmgr.Field("verified_fitness", verifiedFitness))
		}
		logmgr.Info("ActionTree fitness SET",
			logmgr.Field("tree_id", fmt.Sprintf("%p", at)),
			logmgr.Field("fitness", finalFitness),
			logmgr.Field("verified_fitness", verifiedFitness),
			logmgr.Field("num_games", len(actionTreeFitnesses)),
			logmgr.Field("selected_count", actionTreeCount))
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

	// Connect to game
	connectedResp, err := client.ConnectToGame(atfc.opponentType)
	if err != nil {
		logmgr.Error("Failed to connect to game", logmgr.Field("error", err))
		return 0.0, fmt.Errorf("game connection error: %w", err)
	}

	logmgr.Info("Connected to game",
		logmgr.Field("agent_id", connectedResp.AgentID),
		logmgr.Field("opponent_id", connectedResp.OpponentID))

	// Play game
	fitness := atfc.playGame(client, weightsInd, actionTreeInd)

	logmgr.Info("Fitness calculated",
		logmgr.Field("fitness", fitness),
		logmgr.Field("agent_id", connectedResp.AgentID))
	logmgr.Info("RETURNING FITNESS",
		logmgr.Field("fitness", fitness),
		logmgr.Field("agent_id", connectedResp.AgentID))
	return fitness, nil
}

// playGame plays a single game and returns the fitness score
func (atfc *ActionTreeFitnessCalculator) playGame(client *TCPClient, weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	totalReward := 0.0
	finalStep := 0

	logmgr.Debug("Starting game evaluation",
		logmgr.Field("max_steps", atfc.maxSteps),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)),
		logmgr.Field("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	for step := 0; step < atfc.maxSteps; step++ {
		finalStep = step // Track the step we're on
		// Get current observation (first one after connection)
		obs, err := client.ReceiveObservation()
		if err != nil {
			logmgr.Error("Failed to receive observation", logmgr.Field("error", err.Error()))
			logmgr.Error("Game failed due to observation error",
				logmgr.Field("step", step),
				logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)),
				logmgr.Field("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))
			break
		}

		// Log first observation to debug
		if step == 0 {
			keys := make([]string, 0, len(obs.Observation))
			for k := range obs.Observation {
				keys = append(keys, k)
			}
			logmgr.Info("First observation received",
				logmgr.Field("reward", obs.Reward),
				logmgr.Field("terminated", obs.Terminated),
				logmgr.Field("truncated", obs.Truncated),
				logmgr.Field("observation_keys", fmt.Sprintf("%v", keys)))
			logmgr.Info("FULL OBSERVATION STRUCTURE",
				logmgr.Field("observation", fmt.Sprintf("%+v", obs.Observation)))
		}

		// Log game progress every 10 steps
		if step%10 == 0 {
			logmgr.Debug("Game progress",
				logmgr.Field("step", step),
				logmgr.Field("reward", obs.Reward),
				logmgr.Field("terminated", obs.Terminated),
				logmgr.Field("truncated", obs.Truncated),
				logmgr.Field("total_reward", totalReward))
		}

		// Check if game is over
		if obs.Terminated || obs.Truncated {
			logmgr.Info("Game ended early",
				logmgr.Field("step", step),
				logmgr.Field("final_reward", obs.Reward),
				logmgr.Field("total_reward_before_final", totalReward),
				logmgr.Field("total_reward_after_final", totalReward+obs.Reward),
				logmgr.Field("terminated", obs.Terminated),
				logmgr.Field("truncated", obs.Truncated))
			// Extract final game state
			totalReward += obs.Reward
			finalStep = step
			break
		}

		// Execute action trees to get action
		action, err := atfc.actionExecutor.ExecuteActionTreesWithSoftmax(actionTreeInd, weightsInd, obs.Observation)

		if err != nil {
			logmgr.Error("Failed to execute action trees", logmgr.Field("error", err))
			// Send a default action (pass) instead of panicking
			action = []int{0, 0, 0, 0, 0}
		}

		logmgr.Info("Sending action",
			logmgr.Field("step", step),
			logmgr.Field("action", action),
			logmgr.Field("tree_id", fmt.Sprintf("%p", actionTreeInd)),
			logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)))

		// Send action to server
		err = client.SendAction(action)
		if err != nil {
			logmgr.Error("Failed to send action", logmgr.Field("error", err))
			break
		}

		logmgr.Info("Step reward received",
			logmgr.Field("step", step),
			logmgr.Field("obs_reward", obs.Reward),
			logmgr.Field("current_total_reward", totalReward))

		totalReward += obs.Reward
		logmgr.Info("After adding reward",
			logmgr.Field("step", step),
			logmgr.Field("obs_reward", obs.Reward),
			logmgr.Field("new_total_reward", totalReward))
	}

	logmgr.Info("About to return from playGame",
		logmgr.Field("total_reward", totalReward),
		logmgr.Field("steps_completed", finalStep),
		logmgr.Field("tree_id", fmt.Sprintf("%p", actionTreeInd)),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)))

	fitness := totalReward

	logmgr.Info("Game completed - RETURNING FITNESS",
		logmgr.Field("fitness", fitness),
		logmgr.Field("total_reward", totalReward),
		logmgr.Field("steps", finalStep),
		logmgr.Field("max_steps", atfc.maxSteps),
		logmgr.Field("tree_id", fmt.Sprintf("%p", actionTreeInd)),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)))

	return fitness
}

// Close closes the connection pool and cleans up resources
func (atfc *ActionTreeFitnessCalculator) Close() error {
	if atfc.connectionPool != nil {
		return atfc.connectionPool.Close()
	}
	return nil
}
