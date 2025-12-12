package fitness

import (
	"fmt"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
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
func (ae *ActionExecutor) ExecuteActionTreesWithSoftmax(actionTreeIndividual *individual.ActionTreeIndividual, weights *individual.WeightsIndividual, inputs map[string]float64, owned_cells [][]bool, checkConstantActions *[]bool) ([]int, error) {
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
	selectedActions, err := ae.validator.SelectValidAction(actionOutputs, *checkConstantActions, owned_cells)
	if err != nil {
		//Pass if no vlaid acitons(need more troops)
		return []int{1, 0, 0, 0, 0}, nil
	}
	return selectedActions, nil
}

// calculateSoftmax converts scores to probabilities using numerically stable softmax
func CalculateSoftmax(scores []float64) []float64 {
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

func ArgMax(values []float64) int {
	if len(values) == 0 {
		return -1
	}

	maxIndex := 0
	maxValue := values[0]

	for i := 1; i < len(values); i++ {
		if values[i] > maxValue {
			maxValue = values[i]
			maxIndex = i
		}
	}

	return maxIndex
}

func SampleAction(probabilties []float64) int {
	if len(probabilties) == 0 {
		return -1 // or panic, depending on your use case
	}

	sum := 0.0
	for _, prob := range probabilties {
		sum += prob
	}

	cum_prob := 0.0
	randVal := rng.Float64() * sum
	for i, prob := range probabilties {
		cum_prob += prob
		if randVal <= cum_prob {
			return i
		}
	}

	return len(probabilties) - 1

}
