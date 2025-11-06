package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTreeFitness_GIVEN_various_trees_WHEN_calculate_THEN_sets_expected_fitness(t *testing.T) {
	tests := []struct {
		name            string
		tree            *individual.Tree
		expectedFitness float64
	}{
		{"simple binary +", &individual.Tree{Root: &individual.TreeNode{Value: "+", Left: &individual.TreeNode{Value: "1"}, Right: &individual.TreeNode{Value: "x"}}}, -6},
		{"two-layer binary *", &individual.Tree{Root: &individual.TreeNode{Value: "*", Left: &individual.TreeNode{Value: "+", Left: &individual.TreeNode{Value: "x"}, Right: &individual.TreeNode{Value: "1"}}, Right: &individual.TreeNode{Value: "2"}}}, -4},
		{"two-layer nested mix", &individual.Tree{Root: &individual.TreeNode{Value: "-", Left: &individual.TreeNode{Value: "*", Left: &individual.TreeNode{Value: "x"}, Right: &individual.TreeNode{Value: "0"}}, Right: &individual.TreeNode{Value: "+", Left: &individual.TreeNode{Value: "1"}, Right: &individual.TreeNode{Value: "1"}}}}, -10}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fitnessCalc := &individual.TreeFitnessCalculator{}
			fitnessCalc.SetupEvalFunction("x*2+3*2", 1)
			vars := make([]map[string]float64, 1)
			vars[0] = map[string]float64{"x": 1}
			fitnessCalc.TestCases = vars
			var ind individual.Evolvable = tt.tree
			fitnessCalc.CalculateFitness(&ind)
			// Assert fitness is as expected
			assert.Equal(t, tt.expectedFitness, tt.tree.Fitness)
		})
	}
}
