package fitness

import (
	"fmt"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/logmgr"
	"math"
)

// ActionExecutor evaluates action trees and converts outputs to game actions
type ActionExecutor struct {
	actions   []individual.ActionTuple
	validator *ActionValidator
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(actions []individual.ActionTuple) *ActionExecutor {
	return &ActionExecutor{
		actions:   actions,
		validator: NewActionValidator(),
	}
}

// ExecuteActionTreesWithSoftmax evaluates all action trees with given inputs and returns selected action using softmax
func (ae *ActionExecutor) ExecuteActionTreesWithSoftmax(actionTreeIndividual *individual.ActionTreeIndividual, weights *individual.WeightsIndividual, inputs map[string]float64) ([]int, error) {
	// Calculate outputs for each action tree
	actionOutputs := make([][]float64, len(ae.actions))
	r, c := weights.Weights.Dims()
	for row := range r {
		for column := range c {
			key := fmt.Sprintf("w%d", column)
			inputs[key] = weights.Weights.At(row, column)
		}
		for i, actionName := range ae.actions {
			if actionName.Value <= row {
				continue
			}
			tree, exists := actionTreeIndividual.Trees[actionName.Name]
			if !exists {
				return nil, fmt.Errorf("action tree not found: %s", actionName.Name)
			}

			// Execute tree with inputs
			fitness, _ := tree.Root.EvaluateTree(&inputs)
			actionOutputs[i] = append(actionOutputs[i], fitness)
		}
	}

	// Apply softmax to convert scores to probabilities
	selectedActions := make([]int, len(actionOutputs))
	for i, actionArray := range actionOutputs {
		probabilities := ae.calculateSoftmax(actionArray)
		selectedActions[i] = ae.sampleAction(probabilities)
	}

	// Convert inputs to interface{} map for validation
	observationInputs := make(map[string]interface{})
	for k, v := range inputs {
		observationInputs[k] = v
	}

	// Validate the selected action
	if !ae.validator.ValidateAction(selectedActions, observationInputs) {
		// If action is invalid, return a safe default action (pass turn)
		logmgr.Debug("Invalid action detected, using default pass action",
			logmgr.Field("invalid_action", selectedActions))
		return []int{1, 0, 0, 0, 0}, nil // pass_turn=1, others=0
	}

	return selectedActions, nil
}

// calculateSoftmax converts scores to probabilities using numerically stable softmax
func (ae *ActionExecutor) calculateSoftmax(scores []float64) []float64 {
	if len(scores) == 0 {
		return []float64{}
	}

	// Find maximum score for numerical stability
	maxScore := scores[0]
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	// Compute exponentials shifted by max score
	exps := make([]float64, len(scores))
	sum := 0.0
	for i, score := range scores {
		exp := math.Exp(score - maxScore)
		exps[i] = exp
		sum += exp
	}

	// Normalize to get probabilities
	if sum == 0.0 {
		// If sum is zero (shouldn't happen with proper softmax), return uniform distribution
		uniform := 1.0 / float64(len(scores))
		probabilities := make([]float64, len(scores))
		for i := range probabilities {
			probabilities[i] = uniform
		}
		return probabilities
	}

	probabilities := make([]float64, len(scores))
	for i, exp := range exps {
		probabilities[i] = exp / sum
	}

	return probabilities
}

// sampleAction selects an action index based on probabilities
func (ae *ActionExecutor) sampleAction(probabilities []float64) int {
	if len(probabilities) == 0 {
		return 0
	}

	// Generate random number between 0 and 1
	rand := rng.Float64()

	// Find which probability interval the random number falls into
	cumulative := 0.0
	for i, prob := range probabilities {
		cumulative += prob
		if rand < cumulative {
			return i
		}
	}

	// Due to floating point precision, return last index
	return len(probabilities) - 1
}
