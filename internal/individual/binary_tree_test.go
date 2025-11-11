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
