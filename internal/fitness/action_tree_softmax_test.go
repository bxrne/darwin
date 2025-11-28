package fitness_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/mat"
)

func TestExecuteActionTreesWithSoftmax_GameActions(t *testing.T) {
	rng.Seed(42)

	// Define game actions: [pass, cell_i, cell_j, direction, split]
	actions := []string{"action_0", "action_1", "action_2", "action_3", "action_4"}
	numInputs := 4 // owned_army_count, opponent_army_count, owned_land_count, timestep

	// Create component trees for each action dimension
	// Each tree will output one component of 5-element action array
	trees := make(map[string]*individual.Tree)

	// Tree 0: pass decision (0 or 1)
	trees["action_0"] = &individual.Tree{
		Root: &individual.TreeNode{
			Value: "-",
			Left:  &individual.TreeNode{Value: "owned_army_count"},
			Right: &individual.TreeNode{Value: "10.0"},
		},
	}

	// Tree 1: cell_i (0-17 for 18 height grid)
	trees["action_1"] = &individual.Tree{
		Root: &individual.TreeNode{
			Value: "%",
			Left:  &individual.TreeNode{Value: "timestep"},
			Right: &individual.TreeNode{Value: "18.0"},
		},
	}

	// Tree 2: cell_j (0-21 for 22 width grid)
	trees["action_2"] = &individual.Tree{
		Root: &individual.TreeNode{
			Value: "%",
			Left:  &individual.TreeNode{Value: "timestep"},
			Right: &individual.TreeNode{Value: "22.0"},
		},
	}

	// Tree 3: direction (0-3 for up, down, left, right)
	trees["action_3"] = &individual.Tree{
		Root: &individual.TreeNode{
			Value: "+",
			Left:  &individual.TreeNode{Value: "owned_army_count"},
			Right: &individual.TreeNode{Value: "opponent_army_count"},
		},
	}

	// Tree 4: split decision (0 or 1)
	trees["action_4"] = &individual.Tree{
		Root: &individual.TreeNode{
			Value: ">",
			Left:  &individual.TreeNode{Value: "owned_land_count"},
			Right: &individual.TreeNode{Value: "5.0"},
		},
	}

	// init W
	weights := mat.NewDense(5, numInputs, nil)
	for i := range 5 {
		for j := range numInputs {
			weights.Set(i, j, 1.0) // All weights set to 1.0 for easy maths for adam
		}
	}

	actionIndividual := &individual.ActionTreeIndividual{
		Trees:   trees,
		Weights: weights,
	}
	actionIndividual.SetFitness(0.0)

	executor := fitness.NewActionExecutor(actions, numInputs)

	// Basic softmax calculation
	t.Run("BasicSoftmaxCalculation", func(t *testing.T) {
		inputs := []float64{15.0, 12.0, 8.0, 5.0} // owned_army, opponent_army, owned_land, timestep

		selectedAction, probabilities, err := executor.ExecuteActionTreesWithSoftmax(actionIndividual, inputs)
		assert.NoError(t, err)

		assert.Len(t, selectedAction, 5, "Action should be 5-element array")
		assert.InDelta(t, selectedAction[0], 0.0, 1.0, "pass should be 0 or 1")   // pass
		assert.InDelta(t, selectedAction[1], 0.0, 18.0, "cell_i should be 0-17")  // cell_i
		assert.InDelta(t, selectedAction[2], 0.0, 22.0, "cell_j should be 0-21")  // cell_j
		assert.InDelta(t, selectedAction[3], 0.0, 4.0, "direction should be 0-3") // direction
		assert.InDelta(t, selectedAction[4], 0.0, 1.0, "split should be 0 or 1")  // split

		// Verify probabilities
		assert.Len(t, probabilities, 5, "Should have 5 probabilities")
		assert.InDelta(t, sum(probabilities), 1.0, 1e-10, "Probabilities should sum to 1.0")

		// Verify all probabilities are positive
		for i, prob := range probabilities {
			assert.Greater(t, prob, 0.0, "Probability %d should be positive", i)
			assert.LessOrEqual(t, prob, 1.0, "Probability %d should be <= 1.0", i)
		}
	})

	// Higher scores get higher probabilities
	t.Run("HigherScoresHigherProbabilities", func(t *testing.T) {
		highScoreTrees := make(map[string]*individual.Tree)

		// Action 0 gets very high score
		highScoreTrees["action_0"] = &individual.Tree{
			Root: &individual.TreeNode{Value: "100.0"},
		}

		// Other actions get low scores
		for i := 1; i < 5; i++ {
			actionName := actions[i]
			highScoreTrees[actionName] = &individual.Tree{
				Root: &individual.TreeNode{Value: "1.0"},
			}
		}

		highScoreIndividual := &individual.ActionTreeIndividual{
			Trees:   highScoreTrees,
			Weights: weights,
		}
		highScoreIndividual.SetFitness(0.0)

		inputs := []float64{1.0, 1.0, 1.0, 1.0}
		_, probabilities, err := executor.ExecuteActionTreesWithSoftmax(highScoreIndividual, inputs)
		assert.NoError(t, err)

		assert.Greater(t, probabilities[0], probabilities[1], "Action 0 should have higher probability than action 1")
		assert.Greater(t, probabilities[0], probabilities[2], "Action 0 should have higher probability than action 2")
		assert.Greater(t, probabilities[0], probabilities[3], "Action 0 should have higher probability than action 3")
		assert.Greater(t, probabilities[0], probabilities[4], "Action 0 should have higher probability than action 4")

		// With such a high score, action 0 should be selected most of the time
		// Test multiple selections to verify probability distribution
		actionCounts := make([]int, 5)
		numTrials := 100

		for range numTrials {
			selectedAction, _, err := executor.ExecuteActionTreesWithSoftmax(highScoreIndividual, inputs)
			assert.NoError(t, err)

			if selectedAction[0] > 0.5 { // act 0  picked
				actionCounts[0]++
			} else {
				// Count other actions (simplified for test)
				actionCounts[1]++
			}
		}

		// Action 0 should be selected significantly more often
		action0Ratio := float64(actionCounts[0]) / float64(numTrials)
		assert.Greater(t, action0Ratio, 0.6, fmt.Sprintf("Action 0 should be selected more than 60%% of the time, got %.2f%%", action0Ratio*100))
	})

	// Test case 3: Numerical stability with extreme values
	t.Run("NumericalStability", func(t *testing.T) {
		// Create trees with very large positive and negative values
		extremeTrees := make(map[string]*individual.Tree)
		for i, actionName := range actions {
			if i == 0 {
				extremeTrees[actionName] = &individual.Tree{Root: &individual.TreeNode{Value: "1000.0"}} // Very large
			} else if i == 4 {
				extremeTrees[actionName] = &individual.Tree{Root: &individual.TreeNode{Value: "-1000.0"}} // Very negative
			} else {
				extremeTrees[actionName] = &individual.Tree{Root: &individual.TreeNode{Value: "0.0"}}
			}
		}

		extremeIndividual := &individual.ActionTreeIndividual{
			Trees:   extremeTrees,
			Weights: weights,
		}
		extremeIndividual.SetFitness(0.0)

		inputs := []float64{1.0, 1.0, 1.0, 1.0}
		_, probabilities, err := executor.ExecuteActionTreesWithSoftmax(extremeIndividual, inputs)
		assert.NoError(t, err)

		// Should not overflow or produce NaN/Inf
		for i, prob := range probabilities {
			assert.False(t, math.IsNaN(prob), "Probability %d should not be NaN", i)
			assert.False(t, math.IsInf(prob, 0), "Probability %d should not be Inf", i)
			assert.GreaterOrEqual(t, prob, 0.0, "Probability %d should be >= 0", i)
			assert.LessOrEqual(t, prob, 1.0, "Probability %d should be <= 1", i)
		}

		assert.InDelta(t, sum(probabilities), 1.0, 1e-10, "Probabilities should sum to 1.0")

		// Action 0 should dominate due to very high score
		assert.Greater(t, probabilities[0], 0.99, "Action 0 should have >99% probability")
	})

	// Test case 4: Reproducibility with fixed seed
	t.Run("Reproducibility", func(t *testing.T) {
		inputs := []float64{5.0, 3.0, 7.0, 2.0}

		// Reset seed for reproducibility
		rng.Seed(123)

		selectedAction1, _, err1 := executor.ExecuteActionTreesWithSoftmax(actionIndividual, inputs)
		assert.NoError(t, err1)

		// Reset seed again
		rng.Seed(123)

		selectedAction2, _, err2 := executor.ExecuteActionTreesWithSoftmax(actionIndividual, inputs)
		assert.NoError(t, err2)

		// Results should be identical with same seed
		assert.Equal(t, selectedAction1, selectedAction2, "Results should be reproducible with same seed")
	})

	// Test case 5: All-ones weights matrix verification
	t.Run("AllOnesWeightsMatrix", func(t *testing.T) {
		inputs := []float64{2.0, 3.0, 1.0, 4.0}

		selectedAction, probabilities, err := executor.ExecuteActionTreesWithSoftmax(actionIndividual, inputs)
		assert.NoError(t, err)

		// With all-ones weights, the weighted input contribution should be equal for all actions
		// Weighted sum = 1*2 + 1*3 + 1*1 + 1*4 = 10 for all actions
		// So probabilities should be primarily determined by tree outputs

		// Verify that weights are indeed all ones
		r, c := actionIndividual.Weights.Dims()
		assert.Equal(t, r, 5, "Should have 5 rows (actions)")
		assert.Equal(t, c, 4, "Should have 4 columns (inputs)")

		for i := 0; i < r; i++ {
			for j := 0; j < c; j++ {
				assert.InDelta(t, actionIndividual.Weights.At(i, j), 1.0, 1e-10,
					"Weight at (%d,%d) should be 1.0", i, j)
			}
		}

		// Verify action format consistency
		assert.Len(t, selectedAction, 5, "Action should always be 5-element array")
		assert.InDelta(t, sum(probabilities), 1.0, 1e-10, "Probabilities should sum to 1.0")
	})
}

// Helper function to sum slice of floats
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}
