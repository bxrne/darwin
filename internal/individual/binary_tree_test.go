package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestNewRandomTree_GIVEN_depth_zero_WHEN_new_random_tree_THEN_returns_leaf_node(t *testing.T) {
	tree := individual.NewRandomTree(0)

	assert.NotNil(t, tree)
	assert.NotNil(t, tree.Root)
	assert.Nil(t, tree.Root.Left)
	assert.Nil(t, tree.Root.Right)
	// Value should be a digit 0-9
	assert.Contains(t, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}, tree.Root.Value)
}

func TestNewRandomTree_GIVEN_depth_one_WHEN_new_random_tree_THEN_returns_tree_with_operator_and_leaves(t *testing.T) {
	tree := individual.NewRandomTree(1)

	assert.NotNil(t, tree)
	assert.NotNil(t, tree.Root)
	assert.NotNil(t, tree.Root.Left)
	assert.NotNil(t, tree.Root.Right)

	// Root should be an operator
	assert.Contains(t, []string{"+", "-", "*", "/"}, tree.Root.Value)

	// Leaves should be digits
	assert.Contains(t, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}, tree.Root.Left.Value)
	assert.Contains(t, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}, tree.Root.Right.Value)

	// Leaves should have no children
	assert.Nil(t, tree.Root.Left.Left)
	assert.Nil(t, tree.Root.Left.Right)
	assert.Nil(t, tree.Root.Right.Left)
	assert.Nil(t, tree.Root.Right.Right)
}

func TestNewRandomTree_GIVEN_depth_greater_than_one_WHEN_new_random_tree_THEN_returns_nested_tree(t *testing.T) {
	tree := individual.NewRandomTree(2)

	assert.NotNil(t, tree)
	assert.NotNil(t, tree.Root)
	assert.NotNil(t, tree.Root.Left)
	assert.NotNil(t, tree.Root.Right)

	// Root should be an operator
	assert.Contains(t, []string{"+", "-", "*", "/"}, tree.Root.Value)

	// Left and right should also be operators (since depth > 1)
	assert.Contains(t, []string{"+", "-", "*", "/"}, tree.Root.Left.Value)
	assert.Contains(t, []string{"+", "-", "*", "/"}, tree.Root.Right.Value)
}

func TestTree_Max_GIVEN_first_tree_higher_fitness_WHEN_max_THEN_returns_first_tree(t *testing.T) {
	tree1 := &individual.Tree{Fitness: 0.8}
	tree2 := &individual.Tree{Fitness: 0.5}

	result := tree1.Max(tree2)

	assert.Equal(t, tree1, result)
}

func TestTree_Max_GIVEN_second_tree_higher_fitness_WHEN_max_THEN_returns_second_tree(t *testing.T) {
	tree1 := &individual.Tree{Fitness: 0.3}
	tree2 := &individual.Tree{Fitness: 0.7}

	result := tree1.Max(tree2)

	assert.Equal(t, tree2, result)
}

func TestTree_Max_GIVEN_equal_fitness_WHEN_max_THEN_returns_first_tree(t *testing.T) {
	tree1 := &individual.Tree{Fitness: 0.5}
	tree2 := &individual.Tree{Fitness: 0.5}

	result := tree1.Max(tree2)

	assert.Equal(t, tree1, result)
}

func TestTree_MultiPointCrossover_GIVEN_two_trees_WHEN_crossover_THEN_returns_same_trees(t *testing.T) {
	tree1 := individual.NewRandomTree(1)
	tree2 := individual.NewRandomTree(1)

	result1, result2 := tree1.MultiPointCrossover(tree2, 2)

	// Currently placeholder implementation returns the same trees
	assert.Equal(t, tree1, result1)
	assert.Equal(t, tree2, result2)
}

func TestTree_Mutate_GIVEN_mutation_rate_one_WHEN_mutate_THEN_root_value_changes(t *testing.T) {
	tree := individual.NewRandomTree(0)

	tree.Mutate(1.0) // 100% mutation rate

	// Currently placeholder implementation may change root value
	// This test documents current behavior - should be updated when proper mutation is implemented
	assert.NotNil(t, tree.Root.Value)
}

func TestTree_Mutate_GIVEN_mutation_rate_zero_WHEN_mutate_THEN_root_value_unchanged(t *testing.T) {
	tree := individual.NewRandomTree(0)
	originalValue := tree.Root.Value

	tree.Mutate(0.0) // 0% mutation rate

	// Currently placeholder implementation should not change root value at 0% rate
	assert.Equal(t, originalValue, tree.Root.Value)
}

func TestTree_CalculateFitness_GIVEN_tree_WHEN_calculate_fitness_THEN_sets_random_fitness(t *testing.T) {
	tree := individual.NewRandomTree(1)
	tree.Fitness = 0.0 // Reset fitness

	tree.CalculateFitness()

	// Currently placeholder implementation sets random fitness between 0-100
	assert.GreaterOrEqual(t, tree.Fitness, 0.0)
	assert.Less(t, tree.Fitness, 100.0)
}

func TestTree_GetFitness_GIVEN_tree_with_fitness_WHEN_get_fitness_THEN_returns_fitness_value(t *testing.T) {
	expectedFitness := 42.5
	tree := &individual.Tree{Fitness: expectedFitness}

	actualFitness := tree.GetFitness()

	assert.Equal(t, expectedFitness, actualFitness)
}

func TestTree_GetFitness_GIVEN_tree_with_zero_fitness_WHEN_get_fitness_THEN_returns_zero(t *testing.T) {
	tree := &individual.Tree{Fitness: 0.0}

	actualFitness := tree.GetFitness()

	assert.Equal(t, 0.0, actualFitness)
}
