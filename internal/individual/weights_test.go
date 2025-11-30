package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"gonum.org/v1/gonum/mat"
)

// helper
func matEqual(a, b *mat.Dense) bool {
	ra, ca := a.Dims()
	rb, cb := b.Dims()
	if ra != rb || ca != cb {
		return false
	}
	for i := range ra {
		for j := range ca {
			if a.At(i, j) != b.At(i, j) {
				return false
			}
		}
	}
	return true
}

func TestNewWeightsIndividual(t *testing.T) {
	tests := []struct {
		height int
		width  int
	}{
		{1, 1},
		{3, 5},
		{10, 10},
	}

	for _, tc := range tests {
		wi := individual.NewWeightsIndividual(tc.height, tc.width)

		if wi.Weights == nil {
			t.Fatalf("Weights should not be nil")
		}

		r, c := wi.Weights.Dims()
		if r != tc.height || c != tc.width {
			t.Fatalf("Expected dims %dx%d, got %dx%d", tc.height, tc.width, r, c)
		}

		// Values must be between minVal and maxVal
		for i := range r {
			for j := range c {
				v := wi.Weights.At(i, j)
				if v < -5.0 || v > 5.0 {
					t.Fatalf("Value out of bounds: %f", v)
				}
			}
		}
	}
}

func TestClone(t *testing.T) {
	wi := individual.NewWeightsIndividual(3, 3)
	wi.SetFitness(12.5)

	clone := wi.Clone().(*individual.WeightsIndividual)

	if clone == wi {
		t.Fatalf("Clone should be a different pointer")
	}
	if !matEqual(wi.Weights, clone.Weights) {
		t.Fatalf("Cloned matrix does not match original")
	}
	if clone.GetFitness() != wi.GetFitness() {
		t.Fatalf("Clone did not copy fitness")
	}

	// ensure mutation does not affect clone
	wi.Weights.Set(0, 0, 999)
	if clone.Weights.At(0, 0) == 999 {
		t.Fatalf("Clone shares matrix memory; must be deep copy")
	}
}

func TestMax(t *testing.T) {
	wi1 := individual.NewWeightsIndividual(2, 2)
	wi2 := individual.NewWeightsIndividual(2, 2)

	wi1.SetFitness(10)
	wi2.SetFitness(20)

	max := wi1.Max(wi2)
	if max != wi2 {
		t.Fatalf("Expected wi2 as max, got wi1")
	}

	wi1.SetFitness(30)
	max = wi1.Max(wi2)
	if max != wi1 {
		t.Fatalf("Expected wi1 as max, got wi2")
	}
}

func TestMutate(t *testing.T) {
	wi := individual.NewWeightsIndividual(4, 4)

	// clone original for comparison
	before := wi.Clone().(*individual.WeightsIndividual)

	// force mutation
	wi.Mutate(1.0, nil)

	if matEqual(before.Weights, wi.Weights) {
		t.Fatalf("Mutation with rate 1.0 should change values")
	}
}

func TestMultiPointCrossoverSmoketest(t *testing.T) {
	parent1 := individual.NewWeightsIndividual(3, 3)
	parent2 := individual.NewWeightsIndividual(3, 3)

	// Force different starting matrices
	parent1.Weights.Set(0, 0, -999)
	parent2.Weights.Set(0, 0, 999)

	ci := &individual.CrossoverInformation{CrossoverPoints: 2}

	child1Raw, child2Raw := parent1.MultiPointCrossover(parent2, ci)
	child1 := child1Raw.(*individual.WeightsIndividual)
	child2 := child2Raw.(*individual.WeightsIndividual)

	// Check at least one element differs
	differ := false
	r, c := child1.Weights.Dims()
	for i := range r {
		for j := range c {
			if child1.Weights.At(i, j) != child2.Weights.At(i, j) {
				differ = true
				break
			}
		}
	}

	if !differ {
		t.Fatalf("expected crossover to produce at least some difference between children")
	}
}
