package fitness

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/logmgr"
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
		weightsFitnesses := make([]Client, len(*atfc.weightsPopulation))
		for _, evolvable := range *atfc.actionTreePopulation {
			tree, ok := (evolvable).(*individual.ActionTreeIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}

			sum := 0.0
			clientId := ""
			for range atfc.testCaseCount {
				fitness, currentClientId, err := atfc.SetupGameAndRun(wi, tree)
				if err != nil {
					logmgr.Error("Failed to setup game and run", logmgr.Field("error", err.Error()))
				} else {
					sum += fitness
					if clientId == "" {
						clientId += currentClientId
					}
				}
			}
			weightsFitnesses = append(weightsFitnesses, Client{
				ID:      clientId,
				Fitness: sum / float64(atfc.testCaseCount),
			})
		}
		sort.Slice(weightsFitnesses, func(i, j int) bool {
			return (weightsFitnesses[i].Fitness > weightsFitnesses[j].Fitness)
		})
		wi.SetFitness(weightsFitnesses[0].Fitness)
		wi.SetClient(weightsFitnesses[0].ID)

		return
	}

	if atok {
		actionTreeFitnesses := make([]Client, len(*atfc.actionTreePopulation))
		for _, evolvable := range *atfc.weightsPopulation {
			weights, ok := (evolvable).(*individual.WeightsIndividual)
			if !ok {
				panic("Not action tree in action tree population")
			}

			sum := 0.0
			clientId := ""
			for range atfc.testCaseCount {
				fitness, currentClientId, err := atfc.SetupGameAndRun(weights, at)
				if err != nil {
					logmgr.Error("Failed to setup game and run", logmgr.Field("error", err.Error()))
				} else {
					sum += fitness
					if clientId == "" {
						clientId += currentClientId
					}
				}
			}
			actionTreeFitnesses = append(actionTreeFitnesses, Client{
				ID:      clientId,
				Fitness: sum / float64(atfc.testCaseCount),
			})
		}
		sort.Slice(actionTreeFitnesses, func(i, j int) bool {
			return actionTreeFitnesses[i].Fitness > actionTreeFitnesses[j].Fitness
		})
		at.SetFitness(actionTreeFitnesses[0].Fitness)
		at.SetClient(actionTreeFitnesses[0].ID)
		return
	}
}

func (atfc *ActionTreeFitnessCalculator) SetupGameAndRun(weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) (float64, string, error) {
	// Get connection from pool
	client, err := atfc.connectionPool.GetConnection()
	if err != nil {
		logmgr.Error("Failed to get connection from pool", logmgr.Field("error", err.Error()))
		return 0.0, "", fmt.Errorf("connection pool error: %w", err)
	}
	clientId := atfc.getClientId()
	logmgr.Debug("Got connection for game evaluation",
		logmgr.Field("client_id", clientId),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)),
		logmgr.Field("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	// Ensure connection is returned to pool
	defer func() {
		if returnErr := atfc.connectionPool.ReturnConnection(client); returnErr != nil {
			logmgr.Error("Failed to return connection to pool", logmgr.Field("error", returnErr.Error()))
		} else {
			logmgr.Debug("Connection returned to pool successfully",
				logmgr.Field("client_id", clientId))
		}
	}()

	// Small delay to ensure connection is stable
	time.Sleep(100 * time.Millisecond)

	// Connect to game
	connectedResp, err := client.ConnectToGame(clientId, atfc.opponentType)
	if err != nil {
		logmgr.Error("Failed to connect to game", logmgr.Field("error", err.Error()))
		return 0.0, "", fmt.Errorf("game connection error: %w", err)
	}

	logmgr.Debug("Connected to game",
		logmgr.Field("agent_id", connectedResp.AgentID),
		logmgr.Field("opponent_id", connectedResp.OpponentID))

	// Play game
	fitness := atfc.playGame(client, weightsInd, actionTreeInd)

	logmgr.Debug("Fitness calculated",
		logmgr.Field("fitness", fitness),
		logmgr.Field("agent_id", clientId))
	return fitness, clientId, nil
}

// playGame plays a single game and returns the fitness score
func (atfc *ActionTreeFitnessCalculator) playGame(client *TCPClient, weightsInd *individual.WeightsIndividual, actionTreeInd *individual.ActionTreeIndividual) float64 {
	totalReward := 0.0
	logmgr.Debug("Starting game evaluation",
		logmgr.Field("max_steps", atfc.maxSteps),
		logmgr.Field("weights_id", fmt.Sprintf("%p", weightsInd)),
		logmgr.Field("action_tree_id", fmt.Sprintf("%p", actionTreeInd)))

	// Get observation from conncetion (Reset)
	obs, err := client.ReceiveObservation()
	if err != nil {
		logmgr.Error("Failed to receive observation", logmgr.Field("error", err.Error()))
	}
	actionExecutor := NewActionExecutor(atfc.actions)
	actionExecutor.validator.SetMountains(obs.Info)

	// Send action to server
	err = client.SendAction([]int{1, 0, 0, 0, 0})
	if err != nil {
		logmgr.Error("Failed to send action", logmgr.Field("error", err))
	}
	for step := range atfc.maxSteps {
		totalReward += obs.Reward
		obs, err = client.ReceiveObservation()
		if err != nil {
			logmgr.Error("Failed to receive observation", logmgr.Field("error", err.Error()))
			break
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
			logmgr.Debug("Game ended",
				logmgr.Field("step", step),
				logmgr.Field("final_reward", obs.Reward),
				logmgr.Field("total_reward", totalReward+obs.Reward),
				logmgr.Field("terminated", obs.Terminated),
				logmgr.Field("truncated", obs.Truncated))
			// Extract final game state
			totalReward += obs.Reward
			break
		}

		// Execute action trees to get action
		action, err := actionExecutor.ExecuteActionTreesWithSoftmax(actionTreeInd, weightsInd, obs.Observation, obs.Info)

		if err != nil {
			logmgr.Error("Failed to execute action trees", logmgr.Field("error", err.Error()))
			// Send a default action (pass) instead of panicking

		}

		// Send action to server
		err = client.SendAction(action)
		if err != nil {
			logmgr.Error("Failed to send action", logmgr.Field("error", err))
			break
		}

		totalReward += obs.Reward
	}
	err = client.SendDisconnect()
	if err != nil {
		logmgr.Error("Failed to disconncet", logmgr.Field("error", err.Error()))
	}

	logmgr.Debug("Final fitness calculation",
		logmgr.Field("total_reward", totalReward))

	return totalReward
}

// Close closes the connection pool and cleans up resources
func (atfc *ActionTreeFitnessCalculator) Close() error {
	if atfc.connectionPool != nil {
		return atfc.connectionPool.Close()
	}
	return nil
}
