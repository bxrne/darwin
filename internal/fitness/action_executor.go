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

// ExecuteActionTrees evaluates all action trees with given inputs and returns selected action
func (ae *ActionExecutor) ExecuteActionTrees(actionTreeIndividual *individual.ActionTreeIndividual, inputs []float64, weights *individual.WeightsIndividual) (string, error) {

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

	r, c := weights.Weights.Dims()
	if r != len(ae.actions) || c != len(inputs) {
		return "", fmt.Errorf("weights matrix dimensions mismatch: expected %dx%d, got %dx%d",
			len(ae.actions), len(inputs), r, c)
	}

	// Calculate weighted scores for each action
	finalScores := make([]float64, len(ae.actions))
	for actionIdx := range ae.actions {
		score := 0.0
		for inputIdx := range inputs {
			score += weights.Weights.At(actionIdx, inputIdx) * inputs[inputIdx]
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
func (ae *ActionExecutor) ExecuteActionTreesWithSoftmax(actionTreeIndividual *individual.ActionTreeIndividual, weights *individual.WeightsIndividual, inputs []float64) ([]float64, []float64, error) {
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

	r, c := weights.Weights.Dims()
	if r != len(ae.actions) || c != len(inputs) {
		return nil, nil, fmt.Errorf("weights matrix dimensions mismatch: expected %dx%d, got %dx%d",
			len(ae.actions), len(inputs), r, c)
	}

	// Calculate weighted scores for each action
	finalScores := make([]float64, len(ae.actions))
	for actionIdx := range ae.actions {
		score := 0.0
		for inputIdx := range inputs {
			score += weights.Weights.At(actionIdx, inputIdx) * inputs[inputIdx]
		}
		finalScores[actionIdx] = score + actionOutputs[actionIdx]
	}

	// Apply softmax to convert scores to probabilities
	probabilities := ae.calculateSoftmax(finalScores)

	// Sample which action tree to use based on probabilities
	selectedActionIdx := ae.sampleAction(probabilities)

	// Execute the selected action tree to get the action components
	selectedActionName := ae.actions[selectedActionIdx]
	selectedTree := actionTreeIndividual.Trees[selectedActionName]

	// Execute the selected tree to get raw component value
	rawComponent, err := ae.executeTree(selectedTree, inputs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute selected tree %s: %w", selectedActionName, err)
	}

	// Create the 5-element action array based on which tree was selected
	// The action array always has 5 components regardless of number of actions
	selectedAction := make([]float64, 5)

	// Define clamping ranges for each of the 5 action components
	componentRanges := [][2]float64{
		{0.0, 1.0},  // pass: [0,1]
		{0.0, 17.0}, // cell_i: [0,17]
		{0.0, 21.0}, // cell_j: [0,21]
		{0.0, 3.0},  // direction: [0,3]
		{0.0, 1.0},  // split: [0,1]
	}

	// Generate action components where selected tree determines the primary action
	// and other components are set to reasonable defaults
	// Map the selected action index to one of the 5 components using modulo
	componentIdx := selectedActionIdx % 5

	for i := 0; i < 5; i++ {
		if i == componentIdx {
			// Apply clamping for the selected component
			minVal, maxVal := componentRanges[i][0], componentRanges[i][1]

			// Apply appropriate clamping based on component type
			if i == 0 || i == 4 { // pass and split: simple clamping
				selectedAction[i] = math.Max(minVal, math.Min(maxVal, rawComponent))
			} else { // cell_i, cell_j, direction: clamping with modulo for range wrapping
				selectedAction[i] = math.Max(minVal, math.Min(maxVal, math.Mod(rawComponent, maxVal+1.0)))
			}
		} else {
			// Set non-selected components to defaults
			selectedAction[i] = 0.0
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
