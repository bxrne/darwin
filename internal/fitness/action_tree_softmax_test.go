package fitness_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/stretchr/testify/assert"
)

// softmaxTestCase defines a deterministic test case for softmax functionality
type softmaxTestCase struct {
	name          string
	description   string // INFO: What this test demonstrates
	inputs        []float64
	trees         map[string]*individual.Tree
	weights       *individual.WeightsIndividual
	expectedError string
	seed          int64
}

func TestExecuteActionTreesWithSoftmax_Deterministic(t *testing.T) {
	testCases := []softmaxTestCase{
		{
			name:        "UniformDistribution",
			description: "INFO: Test Case: Uniform Distribution with Equal Scores - All trees output 0.0, all-ones weights, inputs [1,2,3,4]",
			inputs:      []float64{1.0, 2.0, 3.0, 4.0},
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_1": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_2": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_3": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_4": {Root: &individual.TreeNode{Value: "0.0"}},
			},
			weights: createAllOnesWeights(5, 4),
			seed:    42, // deterministic seed
		},
		{
			name:        "LinearProgression",
			description: "INFO: Test Case: Linear Score Progression - Trees output [1,2,3,4,5], all-ones weights, inputs [1,1,1,1]",
			inputs:      []float64{1.0, 1.0, 1.0, 1.0},
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "1.0"}},
				"action_1": {Root: &individual.TreeNode{Value: "2.0"}},
				"action_2": {Root: &individual.TreeNode{Value: "3.0"}},
				"action_3": {Root: &individual.TreeNode{Value: "4.0"}},
				"action_4": {Root: &individual.TreeNode{Value: "5.0"}},
			},
			weights: createAllOnesWeights(5, 4),
			seed:    1, // deterministic seed
		},
		{
			name:        "ExtremeDominance",
			description: "INFO: Test Case: Extreme Score Dominance - One tree outputs 100.0, others output 0.0, all-ones weights",
			inputs:      []float64{1.0, 1.0, 1.0, 1.0},
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "100.0"}},
				"action_1": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_2": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_3": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_4": {Root: &individual.TreeNode{Value: "0.0"}},
			},
			weights: createAllOnesWeights(5, 4),
			seed:    42, // deterministic seed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.description)

			// Set seed for deterministic behavior
			rng.Seed(tc.seed)

			// Create action individual
			actionIndividual := &individual.ActionTreeIndividual{
				Trees: tc.trees,
			}
			actionIndividual.SetFitness(0.0)

			// Create executor
			actions := []string{"action_0", "action_1", "action_2", "action_3", "action_4"}
			executor := fitness.NewActionExecutor(actions, len(tc.inputs))

			// Execute softmax
			selectedAction, probabilities, err := executor.ExecuteActionTreesWithSoftmax(actionIndividual, tc.weights, tc.inputs)

			// Check for expected error
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}

			// Check no unexpected error
			assert.NoError(t, err)

			// Verify probabilities are valid
			assert.Len(t, probabilities, 5, "Should have 5 probabilities")
			sumProbs := sum(probabilities)
			assert.InDelta(t, sumProbs, 1.0, 1e-10, "Probabilities should sum to 1.0")

			// Verify all probabilities are positive and <= 1
			for i, prob := range probabilities {
				assert.Greater(t, prob, 0.0, "Probability %d should be positive", i)
				assert.LessOrEqual(t, prob, 1.0, "Probability %d should be <= 1.0", i)
			}

			// Verify action format
			assert.Len(t, selectedAction, 5, "Action should be 5-element array")

			// Verify action component bounds
			verifyActionBounds(t, selectedAction)
		})
	}
}

// TestExecuteActionTreesWithSoftmax_ErrorCases tests error conditions
func TestExecuteActionTreesWithSoftmax_ErrorCases(t *testing.T) {
	testCases := []softmaxTestCase{
		{
			name:        "WrongInputLength",
			description: "INFO: Test Case: Input Length Mismatch - 3 inputs provided but executor expects 4",
			inputs:      []float64{1.0, 2.0, 3.0}, // wrong length
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "0.0"}},
			},
			weights:       createAllOnesWeights(1, 4), // expects 4 inputs
			expectedError: "action tree not found: action_1",
			seed:          42,
		},
		{
			name:        "MissingTree",
			description: "INFO: Test Case: Missing Action Tree - Action tree for action_1 is missing",
			inputs:      []float64{1.0, 2.0, 3.0, 4.0},
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "0.0"}},
				// action_1 is missing
				"action_2": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_3": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_4": {Root: &individual.TreeNode{Value: "0.0"}},
			},
			weights:       createAllOnesWeights(5, 4),
			expectedError: "action tree not found: action_1",
			seed:          42,
		},
		{
			name:        "WeightsDimensionMismatch",
			description: "INFO: Test Case: Weights Matrix Dimension Mismatch - Weights matrix is 3x4 but should be 5x4",
			inputs:      []float64{1.0, 2.0, 3.0, 4.0},
			trees: map[string]*individual.Tree{
				"action_0": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_1": {Root: &individual.TreeNode{Value: "0.0"}},
				"action_2": {Root: &individual.TreeNode{Value: "0.0"}},
			},
			weights:       createAllOnesWeights(3, 4), // wrong dimensions
			expectedError: "action tree not found: action_3",
			seed:          42,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.description)

			rng.Seed(tc.seed)

			actionIndividual := &individual.ActionTreeIndividual{
				Trees: tc.trees,
			}
			actionIndividual.SetFitness(0.0)

			actions := []string{"action_0", "action_1", "action_2", "action_3", "action_4"}
			executor := fitness.NewActionExecutor(actions, len(tc.inputs))

			_, _, err := executor.ExecuteActionTreesWithSoftmax(actionIndividual, tc.weights, tc.inputs)

			assert.Error(t, err, "Should return an error")
			assert.Contains(t, err.Error(), tc.expectedError, "Error message should match expected")
		})
	}
}

// Helper functions

// createAllOnesWeights creates a weights matrix filled with 1.0
func createAllOnesWeights(rows, cols int) *individual.WeightsIndividual {
	return individual.NewWeightsIndividual(rows, cols)
}

// sum calculates the sum of a float slice
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

// verifyActionBounds checks that all action components are within valid ranges
func verifyActionBounds(t *testing.T, action []float64) {
	// pass: should be 0 or 1
	assert.GreaterOrEqual(t, action[0], 0.0, "pass should be >= 0")
	assert.LessOrEqual(t, action[0], 1.0, "pass should be <= 1")

	// cell_i: should be 0-17 (grid height)
	assert.GreaterOrEqual(t, action[1], 0.0, "cell_i should be >= 0")
	assert.LessOrEqual(t, action[1], 17.0, "cell_i should be <= 17")

	// cell_j: should be 0-21 (grid width)
	assert.GreaterOrEqual(t, action[2], 0.0, "cell_j should be >= 0")
	assert.LessOrEqual(t, action[2], 21.0, "cell_j should be <= 21")

	// direction: should be 0-3 (up, down, left, right)
	assert.GreaterOrEqual(t, action[3], 0.0, "direction should be >= 0")
	assert.LessOrEqual(t, action[3], 3.0, "direction should be <= 3")

	// split: should be 0 or 1
	assert.GreaterOrEqual(t, action[4], 0.0, "split should be >= 0")
	assert.LessOrEqual(t, action[4], 1.0, "split should be <= 1")
}
