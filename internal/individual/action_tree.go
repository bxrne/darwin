package individual

import (
	"fmt"

	"github.com/bxrne/darwin/internal/rng"
	"gonum.org/v1/gonum/mat"
)

// ActionTreeIndividual implements an individual composed of action trees and a weights matrix for action selection
type ActionTreeIndividual struct {
	Trees   map[string]*Tree // action name -> action tree
	Weights *mat.Dense       // weights matrix (numActions x numInputs)
	fitness float64
}

// Describe provides a string description of the ActionTreeIndividual
func (ati *ActionTreeIndividual) Describe() string {
	description := "ActionTreeIndividual:\n"
	for action, tree := range ati.Trees {
		description += "Action: " + action + "\n"
		description += "Tree: " + tree.Describe() + "\n"
	}
	description += "Weights:\n"
	r, c := ati.Weights.Dims()
	for i := range r {
		for j := range c {
			description += fmt.Sprintf("%f ", ati.Weights.At(i, j))
		}
		description += "\n"
	}
	description += "Fitness: " + fmt.Sprintf("%f", ati.fitness) + "\n"
	return description
}

// GetFitness returns the fitness of the ActionTreeIndividual
func (ati *ActionTreeIndividual) GetFitness() float64 {
	return ati.fitness
}

// Clone creates a deep copy of the ActionTreeIndividual
func (ati *ActionTreeIndividual) Clone() Evolvable {
	// Clone trees
	clonedTrees := make(map[string]*Tree)
	for action, tree := range ati.Trees {
		var ok bool
		clonedTrees[action], ok = tree.Clone().(*Tree)
		if !ok {
			panic("Failed to clone non-Tree type in ActionTreeIndividual")
		}
	}

	// Clone weights
	r, c := ati.Weights.Dims()
	clonedWeights := mat.NewDense(r, c, nil)
	for i := range r {
		for j := range c {
			clonedWeights.Set(i, j, ati.Weights.At(i, j))
		}
	}

	return &ActionTreeIndividual{
		Trees:   clonedTrees,
		Weights: clonedWeights,
		fitness: ati.fitness,
	}
}

// SetFitness sets the fitness of the ActionTreeIndividual
func (ati *ActionTreeIndividual) SetFitness(fitness float64) {
	ati.fitness = fitness
}

// Mutate applies mutation to the ActionTreeIndividual
func (ati *ActionTreeIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
	// Mutate each tree based on the mutation rate
	for _, tree := range ati.Trees {
		tree.Mutate(rate, mutateInformation)
	}

	// Mutate weights by adding small random values
	r, c := ati.Weights.Dims()
	for i := range r {
		for j := range c {
			if rng.Float64() < rate {
				delta := (rng.Float64() - 0.5) * 0.1 // small change between -0.05 and 0.05
				ati.Weights.Set(i, j, ati.Weights.At(i, j)+delta)
			}
		}
	}
}

// Max returns the ActionTreeIndividual with the higher fitness
func (ati *ActionTreeIndividual) Max(i2 Evolvable) Evolvable {
	other, ok := i2.(*ActionTreeIndividual)
	if !ok {
		panic("Max called with non-ActionTreeIndividual type")
	}
	if ati.fitness >= other.fitness {
		return ati
	}
	return other
}

// MultiPointCrossover performs multi-point crossover between two ActionTreeIndividuals
func (ati *ActionTreeIndividual) MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable) {
	other, ok := i2.(*ActionTreeIndividual)
	if !ok {
		panic("MultiPointCrossover called with non-ActionTreeIndividual type")
	}

	// Crossover trees
	child1Trees := make(map[string]*Tree)
	child2Trees := make(map[string]*Tree)
	for action := range ati.Trees {
		if rng.Float64() < 0.5 {
			child1Trees[action] = ati.Trees[action]
			child2Trees[action] = other.Trees[action]
		} else {
			child1Trees[action] = other.Trees[action]
			child2Trees[action] = ati.Trees[action]
		}
	}

	// Crossover Weights
	r, c := ati.Weights.Dims()
	child1Weights := mat.NewDense(r, c, nil)
	child2Weights := mat.NewDense(r, c, nil)
	for i := range r {
		for j := range c {
			if rng.Float64() < 0.5 {
				child1Weights.Set(i, j, ati.Weights.At(i, j))
				child2Weights.Set(i, j, other.Weights.At(i, j))
			} else {
				child1Weights.Set(i, j, other.Weights.At(i, j))
				child2Weights.Set(i, j, ati.Weights.At(i, j))
			}
		}
	}

	// new kids
	child1 := &ActionTreeIndividual{
		Trees:   child1Trees,
		Weights: child1Weights,
		fitness: 0.0,
	}
	child2 := &ActionTreeIndividual{
		Trees:   child2Trees,
		Weights: child2Weights,
		fitness: 0.0,
	}
	return child1, child2
}

// NewActionTreeIndividual creates a new ActionTreeIndividual
func NewActionTreeIndividual(actions []string, numInputs int, initialTrees map[string]*Tree) *ActionTreeIndividual {
	weights := mat.NewDense(len(actions), numInputs, nil)
	// Initialize weights to zero
	for i := range actions {
		for j := range make([]int, numInputs) {
			weights.Set(i, j, 0.0)
		}
	}
	return &ActionTreeIndividual{
		Trees:   initialTrees,
		Weights: weights,
		fitness: 0.0,
	}
}
