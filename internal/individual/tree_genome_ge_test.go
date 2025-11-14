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

	genome := []int{10, 20, 33, 51}
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
