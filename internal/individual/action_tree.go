package individual

import (
	"sync"

	"github.com/bxrne/darwin/internal/rng"
)

// ActionTreeIndividual implements an individual composed of action trees and a weights matrix for action selection
type ActionTreeIndividual struct {
	Trees     map[string]*Tree // action name -> action tree
	fitness   float64
	fitnessMu sync.RWMutex // Protects fitness field from concurrent access
}

type ActionTuple struct {
	Name  string `toml:"name"`
	Value int    `toml:"value"`
}

// Describe provides a string description of the ActionTreeIndividual
func (ati *ActionTreeIndividual) Describe() string {
	description := "ActionTreeIndividual:\n"
	for action, tree := range ati.Trees {
		description += "Action: " + action + "\n"
		description += "Tree: " + tree.Describe() + "\n"
	}
	description += "Weights:\n"
	return description
}

// GetFitness returns the fitness of the ActionTreeIndividual
func (ati *ActionTreeIndividual) GetFitness() float64 {
	ati.fitnessMu.RLock()
	defer ati.fitnessMu.RUnlock()
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

	ati.fitnessMu.RLock()
	fitness := ati.fitness
	ati.fitnessMu.RUnlock()
	return &ActionTreeIndividual{
		Trees:   clonedTrees,
		fitness: fitness,
	}
}

// SetFitness sets the fitness of the ActionTreeIndividual (thread-safe)
func (ati *ActionTreeIndividual) SetFitness(fitness float64) {
	ati.fitnessMu.Lock()
	defer ati.fitnessMu.Unlock()
	ati.fitness = fitness
}

// Mutate applies mutation to the ActionTreeIndividual
func (ati *ActionTreeIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
	// Mutate each tree based on the mutation rate
	for _, tree := range ati.Trees {
		tree.Mutate(rate, mutateInformation)
	}

}

// Max returns the ActionTreeIndividual with the higher fitness
func (ati *ActionTreeIndividual) Max(i2 Evolvable) Evolvable {
	other, ok := i2.(*ActionTreeIndividual)
	if !ok {
		panic("Max called with non-ActionTreeIndividual type")
	}
	ati.fitnessMu.RLock()
	other.fitnessMu.RLock()
	atiFitness := ati.fitness
	otherFitness := other.fitness
	ati.fitnessMu.RUnlock()
	other.fitnessMu.RUnlock()
	if atiFitness >= otherFitness {
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

	return ati, other
}

// NewActionTreeIndividual creates a new ActionTreeIndividual with provided trees
func NewActionTreeIndividual(actions []ActionTuple, initialTrees map[string]*Tree) *ActionTreeIndividual {
	return &ActionTreeIndividual{
		Trees:   initialTrees,
		fitness: 0.0,
	}
}

// NewRandomActionTreeIndividual creates a new ActionTreeIndividual with random trees
func NewRandomActionTreeIndividual(actions []ActionTuple, maxDepth int, operands []string, variables []string, terminals []string) *ActionTreeIndividual {
	trees := make(map[string]*Tree)

	// Create random tree for each action
	for _, action := range actions {
		tree := NewRandomTree(maxDepth, operands, variables, terminals)
		trees[action.Name] = tree
	}

	return NewActionTreeIndividual(actions, trees)
}
