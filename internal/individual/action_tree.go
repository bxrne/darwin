package individual

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
)

// ActionTreeIndividual
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
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
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
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
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
	// Do nothing for now
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
	// Do nothing for now
	return ati.Clone(), i2.Clone()
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
