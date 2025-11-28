package fitness

import (
	"fmt"
	"math"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
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
		output, err := ae.executeTree(tree, inputs)
		if err != nil {
			return "", fmt.Errorf("failed to execute tree for action %s: %w", actionName, err)
		}
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
func (ae *ActionExecutor) ExecuteActionTreesWithSoftmax(actionTreeIndividual *individual.ActionTreeIndividual, inputs []float64) ([]float64, []float64, error) {
	if len(inputs) != ae.numInputs {
		return nil, nil, fmt.Errorf("expected %d inputs, got %d", ae.numInputs, len(inputs))
	}

	// Calculate outputs for each action tree
	actionOutputs := make([]float64, len(ae.actions))

	for i, actionName := range ae.actions {
		tree, exists := actionTreeIndividual.Trees[actionName]
		if !exists {
			return nil, nil, fmt.Errorf("action tree not found: %s", actionName)
		}

		// Execute tree with inputs
		output, err := ae.executeTree(tree, inputs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute tree for action %s: %w", actionName, err)
		}
		actionOutputs[i] = output
	}

	// Apply weights matrix to get final action scores
	weights := actionTreeIndividual.Weights
	r, c := weights.Dims()
	if r != len(ae.actions) || c != len(inputs) {
		return nil, nil, fmt.Errorf("weights matrix dimensions mismatch: expected %dx%d, got %dx%d",
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

	// Apply softmax to convert scores to probabilities
	probabilities := ae.calculateSoftmax(finalScores)

	// Generate selected action array
	// For this test, we'll create action arrays where each tree outputs one component
	selectedAction := make([]float64, 5)

	// Execute all trees to get complete action components
	for i, actionName := range ae.actions {
		tree := actionTreeIndividual.Trees[actionName]
		component, err := ae.executeTree(tree, inputs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute tree for component %s: %w", actionName, err)
		}

		// Map tree output to action component based on action index
		switch i {
		case 0: // pass (0 or 1)
			selectedAction[0] = math.Max(0.0, math.Min(1.0, component)) // Clamp to [0,1]
		case 1: // cell_i (0-17 for 18 height)
			selectedAction[1] = math.Max(0.0, math.Min(17.0, math.Mod(component, 18.0))) // Clamp to [0,17]
		case 2: // cell_j (0-21 for 22 width)
			selectedAction[2] = math.Max(0.0, math.Min(21.0, math.Mod(component, 22.0))) // Clamp to [0,21]
		case 3: // direction (0-3 for up, down, left, right)
			selectedAction[3] = math.Max(0.0, math.Min(3.0, math.Mod(component, 4.0))) // Clamp to [0,3]
		case 4: // split (0 or 1)
			selectedAction[4] = math.Max(0.0, math.Min(1.0, component)) // Clamp to [0,1]
		}
	}

	return selectedAction, probabilities, nil
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
