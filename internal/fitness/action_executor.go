package fitness

import (
	"fmt"
	"math"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
)

// ActionExecutor evaluates action trees and converts outputs to game actions
type ActionExecutor struct {
	actions []string
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(actions []string) *ActionExecutor {
	return &ActionExecutor{
		actions: actions,
	}
}

// executeTree recursively evaluates a tree with given inputs
func (ae *ActionExecutor) executeTree(tree *individual.Tree, inputs []float64) (float64, error) {
	if tree == nil || tree.Root == nil {
		return 0.0, fmt.Errorf("tree or root is nil")
	}

	result, err := ae.executeTreeNode(tree.Root, inputs)
	if err != nil {
		return 0.0, fmt.Errorf("tree execution failed: %w", err)
	}
	return result, nil
}

// executeTreeNode recursively evaluates a tree node with given inputs
func (ae *ActionExecutor) executeTreeNode(node *individual.TreeNode, inputs []float64) (float64, error) {
	if node == nil {
		return 0.0, fmt.Errorf("node is nil")
	}

	// If it's a leaf node (terminal)
	if node.Left == nil && node.Right == nil {
		value, err := ae.evaluateTerminal(node.Value, inputs)
		if err != nil {
			return 0.0, fmt.Errorf("terminal evaluation failed: %w", err)
		}
		return value, nil
	}

	// If it's an internal node (operator)
	leftValue, err := ae.executeTreeNode(node.Left, inputs)
	if err != nil {
		return 0.0, fmt.Errorf("left subtree failed: %w", err)
	}
	rightValue, err := ae.executeTreeNode(node.Right, inputs)
	if err != nil {
		return 0.0, fmt.Errorf("right subtree failed: %w", err)
	}

	result, err := ae.applyOperator(node.Value, leftValue, rightValue)
	if err != nil {
		return 0.0, fmt.Errorf("operator failed: %w", err)
	}
	return result, nil
}

// evaluateTerminal evaluates a terminal value
func (ae *ActionExecutor) evaluateTerminal(value string, inputs []float64) (float64, error) {
	// Check if it's a standard variable
	for i, inputName := range []string{"x", "y", "z", "w"} {
		if i < len(inputs) && value == inputName {
			return inputs[i], nil
		}
	}

	// Map game variables to input
	gameVariables := map[string]int{
		"owned_army_count":    0,
		"opponent_army_count": 1,
		"owned_land_count":    2,
		"timestep":            3,
		"opponent_land_count": 4,
		"owned_cities_count":  4,
	}

	if index, exists := gameVariables[value]; exists && index < len(inputs) {
		return inputs[index], nil
	}

	if val, err := parseFloat(value); err == nil {
		return val, nil
	}

	return 0.0, fmt.Errorf("unknown terminal: %s", value)
}

// applyOperator applies an operator to two values
func (ae *ActionExecutor) applyOperator(operator string, left, right float64) (float64, error) {
	switch operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right != 0 {
			return left / right, nil
		}
		return 0.0, fmt.Errorf("division by zero")
	case "^":
		return math.Pow(left, right), nil
	case "max":
		if left > right {
			return left, nil
		}
		return right, nil
	case "min":
		if left < right {
			return left, nil
		}
		return right, nil
	case ">":
		if left > right {
			return 1.0, nil
		}
		return 0.0, nil
	case "%":
		if right != 0 {
			return math.Mod(left, right), nil
		}
		return 0.0, fmt.Errorf("modulo by zero")
	default:
		return 0.0, fmt.Errorf("unknown operator: %s", operator)
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
			tree, exists := actionTreeIndividual.Trees[actionName]
			if !exists {
				return nil, fmt.Errorf("action tree not found: %s", actionName)
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

// parseFloat safely parses a string to float64
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
