package individual

// ActionTreeIndividual implements an individual composed of action trees and a weights matrix for action selection
type ActionTreeIndividual struct {
	Trees    map[string]*Tree // action name -> action tree
	fitness  float64
	clientId string
}

type ActionTuple struct {
	Name  string `toml:"name"`
	Value int    `toml:"value"`
}

// Describe provides a string description of the ActionTreeIndividual
func (ati *ActionTreeIndividual) Describe() string {
	description := "ActionTreeIndividual:" + ati.clientId + " :\n"
	for action, tree := range ati.Trees {
		description += "Action: " + action + "\n"
		description += "Tree: " + tree.Describe() + "\n"
	}
	description += "Weights:\n"
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

	return &ActionTreeIndividual{
		Trees:   clonedTrees,
		fitness: ati.fitness,
	}
}

// SetFitness sets the fitness of the ActionTreeIndividual
func (ati *ActionTreeIndividual) SetFitness(fitness float64) {
	ati.fitness = fitness
}

func (ati *ActionTreeIndividual) SetClient(clientId string) {
	ati.clientId = clientId
}

// Mutate applies mutation to the ActionTreeIndividual
func (ati *ActionTreeIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
	// Mutate each tree based on the mutation rate
	for _, tree := range ati.Trees {
		tree.Mutate(rate, mutateInformation)
		// Safety check: ensure no tree has depth 0 (Tree.Mutate should handle this, but double-check)
		// If depth is 0, regenerate using NewRampedHalfAndHalfTree with depth 1
		if tree.GetDepth() == 0 {
			// Use NewRampedHalfAndHalfTree to regenerate with depth 1
			regeneratedTree := NewRampedHalfAndHalfTree(1, true, mutateInformation.OperandSet, mutateInformation.VariableSet, mutateInformation.TerminalSet)
			*tree = *regeneratedTree
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

	for action := range ati.Trees {
		child1Tree, child2Tree := ati.Trees[action].MultiPointCrossover(other.Trees[action], crossoverInformation)
		cTree1, ok := child1Tree.(*Tree)
		if !ok {
			panic("MultiPointCrossover called with non-ActionTreeIndividual type")
		}
		cTree2, ok := child2Tree.(*Tree)
		if !ok {
			panic("MultiPointCrossover called with non-ActionTreeIndividual type")
		}
		ati.Trees[action] = cTree1
		other.Trees[action] = cTree2
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

func (ati *ActionTreeIndividual) GetMetrics() map[string]float64 {
	totalDepth := 0
	totalNodes := 0

	// Accumulate parameter frequencies over all trees
	totalParamFreq := make(map[string]int)

	for _, tree := range ati.Trees {
		// Depth
		depth := tree.Root.CalculateMaxDepth()
		totalDepth += depth

		// Nodes + variable frequencies
		numNodes, freq := tree.CountNodesAndVars()
		totalNodes += numNodes

		// Merge frequencies into global param frequency map
		for param, count := range freq {
			totalParamFreq[param] += count
		}
	}

	n := float64(len(ati.Trees))

	metrics := map[string]float64{
		"fit":   ati.GetFitness(),
		"depth": float64(totalDepth) / n,
		"nodes": float64(totalNodes) / n,
	}

	// Add each parameter as its own metric key
	for param, count := range totalParamFreq {
		metrics[param] = float64(count) / n // average across trees
	}

	return metrics
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
