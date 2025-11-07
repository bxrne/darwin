package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestNewRandomTree_GIVEN_depth_zero_WHEN_new_random_tree_THEN_returns_leaf_node(t *testing.T) {
	primitiveSet := []string{"+", "-", "*", "/"}
	terminalSet := []string{"x", "y", "1.0", "2.0", "3.0"}
	tree := individual.NewRandomTree(0, primitiveSet, terminalSet)

	assert.NotNil(t, tree)
	assert.NotNil(t, tree.Root)
	assert.Nil(t, tree.Root.Left)
	assert.Nil(t, tree.Root.Right)
	// Value should be from terminal set
	assert.Contains(t, terminalSet, tree.Root.Value)
}

func TestTreeNode_IsLeaf_GIVEN_leaf_node_WHEN_check_THEN_returns_true(t *testing.T) {
	node := &individual.TreeNode{Value: "x"}
	assert.True(t, node.IsLeaf())
}

func TestTreeNode_IsLeaf_GIVEN_internal_node_WHEN_check_THEN_returns_false(t *testing.T) {
	node := &individual.TreeNode{
		Value: "+",
		Left:  &individual.TreeNode{Value: "x"},
		Right: &individual.TreeNode{Value: "y"},
	}
	assert.False(t, node.IsLeaf())
}

func TestTreeNode_mutateTerminal_GIVEN_terminal_WHEN_mutate_THEN_changes_to_different_terminal(t *testing.T) {
	terminalSet := []string{"x", "y", "1.0", "2.0"}
	node := &individual.TreeNode{Value: "x"}

	originalValue := node.Value
	node.MutateTerminal(terminalSet)

	assert.NotEqual(t, originalValue, node.Value)
	assert.Contains(t, terminalSet, node.Value)
}

func TestTreeNode_mutateFunction_GIVEN_function_WHEN_mutate_THEN_changes_to_different_function(t *testing.T) {
	primitiveSet := []string{"+", "-", "*", "/"}
	node := &individual.TreeNode{Value: "+"}

	originalValue := node.Value
	node.MutateFunction(primitiveSet)

	assert.NotEqual(t, originalValue, node.Value)
	assert.Contains(t, primitiveSet, node.Value)
}

func TestTree_MutateWithSets_GIVEN_mutation_rate_zero_WHEN_mutate_THEN_no_change(t *testing.T) {
	primitiveSet := []string{"+", "-", "*", "/"}
	terminalSet := []string{"x", "y", "1.0", "2.0"}
	tree := individual.NewRandomTree(2, primitiveSet, terminalSet)

	originalValues := extractNodeValues(tree)

	tree.MutateWithSets(0.0, primitiveSet, terminalSet) // 0% mutation rate

	newValues := extractNodeValues(tree)
	assert.Equal(t, originalValues, newValues)
}

func TestTree_MutateWithSets_GIVEN_mutation_rate_one_WHEN_mutate_THEN_all_nodes_change(t *testing.T) {
	primitiveSet := []string{"+", "-"} // Small sets for predictable behavior
	terminalSet := []string{"x", "y"}  // Small sets for predictable behavior
	tree := individual.NewRandomTree(2, primitiveSet, terminalSet)

	originalValues := extractNodeValues(tree)

	tree.MutateWithSets(1.0, primitiveSet, terminalSet) // 100% mutation rate

	newValues := extractNodeValues(tree)
	// All nodes should have different values (except when sets have only 2 elements)
	for i, original := range originalValues {
		if len(primitiveSet) > 2 || len(terminalSet) > 2 {
			assert.NotEqual(t, original, newValues[i], "Node at position %d should have changed", i)
		}
	}
}

func TestTree_Mutate_GIVEN_interface_call_WHEN_mutate_THEN_no_change(t *testing.T) {
	primitiveSet := []string{"+", "-", "*", "/"}
	terminalSet := []string{"x", "y", "1.0", "2.0"}
	tree := individual.NewRandomTree(2, primitiveSet, terminalSet)

	originalValues := extractNodeValues(tree)

	tree.Mutate(1.0) // Interface method - should do nothing

	newValues := extractNodeValues(tree)
	assert.Equal(t, originalValues, newValues)
}

// Helper function to extract all node values from tree
func extractNodeValues(tree *individual.Tree) []string {
	var values []string
	extractValuesRecursive(tree.Root, &values)
	return values
}

// Helper function to recursively extract node values
func extractValuesRecursive(node *individual.TreeNode, values *[]string) {
	*values = append(*values, node.Value)
	if node.Left != nil {
		extractValuesRecursive(node.Left, values)
	}
	if node.Right != nil {
		extractValuesRecursive(node.Right, values)
	}
}
