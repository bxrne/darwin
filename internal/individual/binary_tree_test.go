package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestNewRandomTree_GIVEN_depth_zero_WHEN_new_random_tree_THEN_returns_leaf_node(t *testing.T) {
	tree := individual.NewRandomTree(0, 0)

	assert.NotNil(t, tree)
	assert.NotNil(t, tree.Root)
	assert.Nil(t, tree.Root.Left)
	assert.Nil(t, tree.Root.Right)
	// Value should be a digit 0-9
	assert.Contains(t, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}, tree.Root.Value)
}
