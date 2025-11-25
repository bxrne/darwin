package fitness

import (
	"fmt"
	"math"

	"github.com/bxrne/darwin/internal/individual"
)

// ActionExecutor evaluates action trees and converts outputs to game actions
type ActionExecutor struct {
	actions   []string
	numInputs int
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(actions []string, numInputs int) *ActionExecutor {
	return &ActionExecutor{
		actions:   actions,
		numInputs: numInputs,
	}
}

// ExecuteActionTrees evaluates all action trees with given inputs and returns selected action
func (ae *ActionExecutor) ExecuteActionTrees(actionTreeIndividual *individual.ActionTreeIndividual, inputs []float64) (string, error) {
	if len(inputs) != ae.numInputs {
		return "", fmt.Errorf("expected %d inputs, got %d", ae.numInputs, len(inputs))
	}

	// Calculate outputs for each action tree
	actionOutputs := make([]float64, len(ae.actions))

	for i, actionName := range ae.actions {
		tree, exists := actionTreeIndividual.Trees[actionName]
		if !exists {
			return "", fmt.Errorf("action tree not found: %s", actionName)
		}

		// Execute tree with inputs
		output := ae.executeTree(tree, inputs)
		actionOutputs[i] = output
	}

	// Apply weights matrix to get final action scores
	weights := actionTreeIndividual.Weights
	r, c := weights.Dims()
	if r != len(ae.actions) || c != len(inputs) {
		return "", fmt.Errorf("weights matrix dimensions mismatch: expected %dx%d, got %dx%d",
			len(ae.actions), len(inputs), r, c)
	}

	// Calculate weighted scores for each action
	finalScores := make([]float64, len(ae.actions))
	for actionIdx := range ae.actions {
		score := 0.0
		for inputIdx := range inputs {
			score += weights.At(actionIdx, inputIdx) * inputs[inputIdx]
		}
		finalScores[actionIdx] = score + actionOutputs[actionIdx]
	}

	// Select action with highest score
	bestActionIdx := 0
	bestScore := finalScores[0]
	for i := 1; i < len(finalScores); i++ {
		if finalScores[i] > bestScore {
			bestScore = finalScores[i]
			bestActionIdx = i
		}
	}

	return ae.actions[bestActionIdx], nil
}

// executeTree recursively evaluates a tree with given inputs
func (ae *ActionExecutor) executeTree(tree *individual.Tree, inputs []float64) float64 {
	if tree == nil || tree.Root == nil {
		return 0.0
	}

	return ae.executeTreeNode(tree.Root, inputs)
}

// executeTreeNode recursively evaluates a tree node with given inputs
func (ae *ActionExecutor) executeTreeNode(node *individual.TreeNode, inputs []float64) float64 {
	if node == nil {
		return 0.0
	}

	// If it's a leaf node (terminal)
	if node.Left == nil && node.Right == nil {
		return ae.evaluateTerminal(node.Value, inputs)
	}

	// If it's an internal node (operator)
	leftValue := ae.executeTreeNode(node.Left, inputs)
	rightValue := ae.executeTreeNode(node.Right, inputs)

	return ae.applyOperator(node.Value, leftValue, rightValue)
}

// evaluateTerminal evaluates a terminal value
func (ae *ActionExecutor) evaluateTerminal(value string, inputs []float64) float64 {
	// Check if it's a variable
	for i, inputName := range []string{"x", "y", "z", "w"} {
		if i < len(inputs) && value == inputName {
			return inputs[i]
		}
	}

	// Try to parse as constant
	if val, err := parseFloat(value); err == nil {
		return val
	}

	// Default to 0.0 for unknown terminals
	return 0.0
}

// applyOperator applies an operator to two values
func (ae *ActionExecutor) applyOperator(operator string, left, right float64) float64 {
	switch operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right != 0 {
			return left / right
		}
		return 0.0
	case "^":
		return math.Pow(left, right)
	case "max":
		if left > right {
			return left
		}
		return right
	case "min":
		if left < right {
			return left
		}
		return right
	default:
		return 0.0
	}
}

// parseFloat safely parses a string to float64
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
