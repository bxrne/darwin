package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestNewActionTreeIndividual_GIVEN_actions_and_trees_WHEN_created_THEN_has_correct_structure(t *testing.T) {
	actions := []string{"move_east", "move_west"}
	numInputs := 2

	// Create test trees
	operands := []string{"+", "-"}
	variables := []string{"x", "y"}
	terminals := []string{"1", "2"}

	tree1 := individual.NewRandomTree(2, operands, variables, terminals)
	tree2 := individual.NewRandomTree(2, operands, variables, terminals)

	initialTrees := map[string]*individual.Tree{
		"move_east": tree1,
		"move_west": tree2,
	}

	// Create ActionTreeIndividual
	ati := individual.NewActionTreeIndividual(actions, numInputs, initialTrees)

	// Verify structure
	assert.NotNil(t, ati)
	assert.Equal(t, 2, len(ati.Trees))
	assert.Contains(t, ati.Trees, "move_east")
	assert.Contains(t, ati.Trees, "move_west")
	assert.Equal(t, tree1, ati.Trees["move_east"])
	assert.Equal(t, tree2, ati.Trees["move_west"])

}

func TestNewRandomActionTreeIndividual_GIVEN_parameters_WHEN_created_THEN_has_random_trees(t *testing.T) {
	variables := []string{"action1", "action2"}
	numInputs := 3
	maxDepth := 2
	operands := []string{"+", "-"}
	terminals := []string{"1", "2"}

	// Create ActionTreeIndividual with random trees
	ati := individual.NewRandomActionTreeIndividual(numInputs, maxDepth, operands, variables, terminals)

	// Verify structure
	assert.NotNil(t, ati)
	assert.Equal(t, 2, len(ati.Trees))
	assert.Contains(t, ati.Trees, "action1")
	assert.Contains(t, ati.Trees, "action2")

	// Verify trees are not nil
	assert.NotNil(t, ati.Trees["action1"])
	assert.NotNil(t, ati.Trees["action2"])
	assert.NotNil(t, ati.Trees["action1"].Root)
	assert.NotNil(t, ati.Trees["action2"].Root)

}

func TestNewRandomActionTreeIndividual_GIVEN_same_parameters_WHEN_created_multiple_THEN_different_trees(t *testing.T) {
	variables := []string{"action1"}
	numInputs := 2
	maxDepth := 2
	operands := []string{"+", "-"}
	terminals := []string{"1", "2"}

	// Create two individuals with same parameters
	ati1 := individual.NewRandomActionTreeIndividual(numInputs, maxDepth, operands, variables, terminals)
	ati2 := individual.NewRandomActionTreeIndividual(numInputs, maxDepth, operands, variables, terminals)

	// They should have different trees (randomness)
	tree1Desc := ati1.Trees["action1"].Describe()
	tree2Desc := ati2.Trees["action1"].Describe()

	// Note: This test might occasionally fail due to random chance, but should usually pass
	// In practice, you might want to seed the RNG for deterministic testing
	assert.NotEqual(t, tree1Desc, tree2Desc)
}

func TestActionTreeIndividual_Clone_GIVEN_individual_WHEN_cloned_THEN_deep_copy(t *testing.T) {
	variables := []string{"move_east", "move_west"}
	numInputs := 2
	operands := []string{"+"}
	terminals := []string{"1"}

	// Create original
	original := individual.NewRandomActionTreeIndividual(numInputs, 2, operands, variables, terminals)
	original.SetFitness(42.0)

	// Clone it
	cloned := original.Clone().(*individual.ActionTreeIndividual)

	// Verify they're equal but not the same object
	assert.Equal(t, original.GetFitness(), cloned.GetFitness())
	assert.Equal(t, original.Trees["move_east"].Describe(), cloned.Trees["move_east"].Describe())
	assert.Equal(t, original.Trees["move_west"].Describe(), cloned.Trees["move_west"].Describe())

	// Verify they're different objects
	assert.NotSame(t, original, cloned)
	assert.NotSame(t, original.Trees["move_east"], cloned.Trees["move_east"])
	assert.NotSame(t, original.Trees["move_west"], cloned.Trees["move_west"])
}
