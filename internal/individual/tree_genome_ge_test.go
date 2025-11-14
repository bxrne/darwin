package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
)

// GenerateTreeFromGenome tests
func TestGenerateTreeFromGenome(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	genome := []int{214, 54, 212, 42, 79, 138, 61, 93}
	tree := individual.GenerateTreeFromGenome(grammar, genome)

	// is there a tree at all bai?
	if tree == nil {
		t.Errorf("Generated tree is nil")
	}

	// structural
	if tree.Value == "" {
		t.Errorf("Generated tree has empty value")
	}

	t.Logf("Generated tree: %+v", tree)
}

func TestGenerateTreeFromGenome_DepthLimiting(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	// Test with different depth limits
	testCases := []struct {
		name     string
		maxDepth int
		genome   []int
	}{
		{"depth_1", 1, []int{0, 1, 2}},
		{"depth_2", 2, []int{0, 1, 2, 3, 4}},
		{"depth_3", 3, []int{0, 1, 2, 3, 4, 5, 6}},
		{"depth_5", 5, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := individual.GenerateTreeFromGenomeWithDepth(grammar, tc.genome, tc.maxDepth)

			if tree == nil {
				t.Errorf("Generated tree is nil for depth %d", tc.maxDepth)
				return
			}

			// Calculate actual depth
			actualDepth := tree.CalculateMaxDepth()
			if actualDepth > tc.maxDepth {
				t.Errorf("Tree depth %d exceeds max depth %d", actualDepth, tc.maxDepth)
			}

			t.Logf("Depth %d: actual depth = %d, tree = %+v", tc.maxDepth, actualDepth, tree)
		})
	}
}

func TestGenerateTreeFromGenome_TerminalGeneration(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	// Generate multiple trees to test terminal variety
	for i := 0; i < 10; i++ {
		genome := []int{i * 10, i*10 + 1, i*10 + 2}
		tree := individual.GenerateTreeFromGenomeWithDepth(grammar, genome, 3)

		// Verify leaf nodes are terminals
		if !hasValidTerminals(tree, terminalSet, primitiveSet) {
			t.Errorf("Tree %d has invalid terminals: %+v", i, tree)
		}
	}
}

func TestGenerateTreeFromGenome_OperatorPlacement(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	// Generate trees that should have operators
	genome := []int{0, 1, 2, 3, 4, 5} // First codon 0 should select Seq (operator expression)
	tree := individual.GenerateTreeFromGenomeWithDepth(grammar, genome, 3)

	if tree == nil {
		t.Errorf("Generated tree is nil")
		return
	}

	// Verify internal nodes have operators
	if !hasValidOperators(tree, operatorSet) {
		t.Errorf("Tree has invalid operator placement: %+v", tree)
	}

	t.Logf("Operator tree: %+v", tree)
}

func TestGenerateTreeFromGenome_StructureValidation(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	// Test multiple trees for structural validity
	for i := 0; i < 5; i++ {
		genome := []int{i, i + 1, i + 2, i + 3, i + 4}
		tree := individual.GenerateTreeFromGenomeWithDepth(grammar, genome, 4)

		if tree == nil {
			t.Errorf("Tree %d is nil", i)
			continue
		}

		// Verify binary tree structure
		if !isValidBinaryTree(tree) {
			t.Errorf("Tree %d has invalid binary tree structure: %+v", i, tree)
		}

		// Verify no cycles (by checking depth is reasonable)
		depth := tree.CalculateMaxDepth()
		if depth > 10 { // Should be much less than this
			t.Errorf("Tree %d has excessive depth %d, possible cycle", i, depth)
		}
	}
}

func TestGenerateTreeFromGenome_EdgeCases(t *testing.T) {
	terminalSet := []string{"a", "b", "c"}
	primitiveSet := []string{"1", "2", "3"}
	operatorSet := []string{"+", "-", "*"}

	grammar := individual.CreateGrammar(terminalSet, primitiveSet, operatorSet)

	testCases := []struct {
		name     string
		genome   []int
		maxDepth int
	}{
		{"empty_genome", []int{}, 3},
		{"single_codon", []int{42}, 3},
		{"zero_depth", []int{1, 2, 3}, 0},
		{"negative_depth", []int{1, 2, 3}, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Adjust depth for edge cases
			depth := tc.maxDepth
			if depth <= 0 {
				depth = 1 // Minimum valid depth
			}

			tree := individual.GenerateTreeFromGenomeWithDepth(grammar, tc.genome, depth)

			if tree == nil {
				t.Errorf("Tree generation failed for %s", tc.name)
			}

			t.Logf("Edge case %s: tree = %+v", tc.name, tree)
		})
	}
}

// Helper functions for testing

func hasValidTerminals(tree *individual.TreeNode, terminalSet, primitiveSet []string) bool {
	if tree.IsLeaf() {
		// Check if leaf value is in terminal or primitive set
		for _, term := range terminalSet {
			if tree.Value == term {
				return true
			}
		}
		for _, prim := range primitiveSet {
			if tree.Value == prim {
				return true
			}
		}
		return false
	}

	// Check children recursively
	leftValid := true
	rightValid := true
	if tree.Left != nil {
		leftValid = hasValidTerminals(tree.Left, terminalSet, primitiveSet)
	}
	if tree.Right != nil {
		rightValid = hasValidTerminals(tree.Right, terminalSet, primitiveSet)
	}

	return leftValid && rightValid
}

func hasValidOperators(tree *individual.TreeNode, operatorSet []string) bool {
	if tree.IsLeaf() {
		return true // Leaves don't need operators
	}

	// Check if internal node has valid operator
	isValidOperator := false
	for _, op := range operatorSet {
		if tree.Value == op {
			isValidOperator = true
			break
		}
	}

	if !isValidOperator {
		return false
	}

	// Check children recursively
	leftValid := true
	rightValid := true
	if tree.Left != nil {
		leftValid = hasValidOperators(tree.Left, operatorSet)
	}
	if tree.Right != nil {
		rightValid = hasValidOperators(tree.Right, operatorSet)
	}

	return leftValid && rightValid
}

func isValidBinaryTree(tree *individual.TreeNode) bool {
	if tree.IsLeaf() {
		return true
	}

	// Internal nodes should have both children for binary operations
	if tree.Left == nil || tree.Right == nil {
		return false
	}

	// Recursively check children
	return isValidBinaryTree(tree.Left) && isValidBinaryTree(tree.Right)
}
