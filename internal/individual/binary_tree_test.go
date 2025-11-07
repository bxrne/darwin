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
