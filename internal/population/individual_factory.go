package population

import (
	"fmt"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/individual"
)

// IndividualFactory creates individuals based on genome type
type IndividualFactory struct {
	config      *cfg.Config
	treeCounter int64
}

// NewIndividualFactory creates a new individual factory
func NewIndividualFactory(config *cfg.Config) *IndividualFactory {
	return &IndividualFactory{
		config: config,
	}
}

// CreateIndividual creates an individual of the specified type
func (f *IndividualFactory) CreateIndividual(populationType individual.GenomeType) individual.Evolvable {
	switch populationType {
	case individual.BitStringGenome:
		return individual.NewBinaryIndividual(f.config.BitString.GenomeSize)
	case individual.TreeGenome:
		return f.createRampedHalfAndHalfTree()
	case individual.GrammarTreeGenome:
		return individual.NewGrammarTree(f.config.GrammarTree.GenomeSize)
	case individual.ActionTreeGenome:
		return f.createActionTreeIndividual()
	default:
		fmt.Printf("Unknown genome type: %v\n", populationType)
		return nil
	}
}

// createRampedHalfAndHalfTree creates a tree using ramped half-and-half initialization
func (f *IndividualFactory) createRampedHalfAndHalfTree() *individual.Tree {
	popSize := f.config.Evolution.PopulationSize
	index := f.getNextTreeCounter()
	initialDepth := f.config.Tree.InitalDepth

	// Calculate depth group: divide population into initialDepth groups (depths 1 to initialDepth)
	// Depth 0 is disallowed
	depthGroups := initialDepth
	if depthGroups <= 0 {
		depthGroups = 1 // Ensure at least one group
	}
	groupSize := popSize / depthGroups
	remainder := popSize % depthGroups

	// Determine which depth group this individual belongs to (1 to initialDepth)
	depth := 1 // Start from depth 1
	groupIndex := index
	for d := 0; d < depthGroups; d++ {
		groupCount := groupSize
		if d < remainder {
			groupCount++ // Distribute remainder across first groups
		}
		if groupIndex < groupCount {
			depth = d + 1 // Depth is 1-indexed (1 to initialDepth)
			break
		}
		groupIndex -= groupCount
	}

	// Within the depth group, determine if we use grow (first half) or full (second half)
	// Recalculate group boundaries for this specific depth
	groupStart := 0
	for d := 0; d < (depth - 1); d++ { // depth - 1 because depth is 1-indexed
		groupCount := groupSize
		if d < remainder {
			groupCount++
		}
		groupStart += groupCount
	}
	groupCount := groupSize
	if (depth - 1) < remainder {
		groupCount++
	}
	localIndex := index - groupStart
	useGrow := localIndex < groupCount/2

	return individual.NewRampedHalfAndHalfTree(depth, useGrow, f.config.Tree.OperandSet, f.config.Tree.VariableSet, f.config.Tree.TerminalSet)
}

// createActionTreeIndividual creates an action tree individual with trees for each action
func (f *IndividualFactory) createActionTreeIndividual() *individual.ActionTreeIndividual {
	// Create random trees for each action using ramped half-and-half
	// All trees in an individual use the same depth and method
	// Depth 0 is disallowed, so we distribute from depth 1 to initialDepth
	index := f.getNextTreeCounter()
	popSize := f.config.Evolution.PopulationSize
	initialDepth := f.config.Tree.InitalDepth

	// Calculate depth and method for this individual
	// Depth 0 is disallowed, so we have initialDepth groups (depths 1 to initialDepth)
	depthGroups := initialDepth
	if depthGroups <= 0 {
		depthGroups = 1 // Ensure at least one group
	}
	groupSize := popSize / depthGroups
	remainder := popSize % depthGroups

	depth := 1 // Start from depth 1
	groupIndex := index
	for d := 0; d < depthGroups; d++ {
		groupCount := groupSize
		if d < remainder {
			groupCount++
		}
		if groupIndex < groupCount {
			depth = d + 1 // Depth is 1-indexed (1 to initialDepth)
			break
		}
		groupIndex -= groupCount
	}

	groupStart := 0
	for d := 0; d < (depth - 1); d++ { // depth - 1 because depth is 1-indexed
		groupCount := groupSize
		if d < remainder {
			groupCount++
		}
		groupStart += groupCount
	}
	groupCount := groupSize
	if (depth - 1) < remainder {
		groupCount++
	}
	localIndex := index - groupStart
	useGrow := localIndex < groupCount/2

	initialTrees := make(map[string]*individual.Tree)
	variableSet := make([]string, f.config.ActionTree.WeightsColumnCount)
	for i := range f.config.ActionTree.WeightsColumnCount {
		key := fmt.Sprintf("w%d", i)
		variableSet[i] = key
	}
	variableSet = append(variableSet, f.config.Tree.VariableSet...)
	for _, action := range f.config.ActionTree.Actions {
		tree := individual.NewRampedHalfAndHalfTree(depth, useGrow, f.config.Tree.OperandSet, variableSet, f.config.Tree.TerminalSet)
		initialTrees[action.Name] = tree
	}
	result := individual.NewActionTreeIndividual(f.config.ActionTree.Actions, initialTrees)
	return result
}

// getNextTreeCounter returns the next tree counter value
func (f *IndividualFactory) getNextTreeCounter() int {
	f.treeCounter++
	return int(f.treeCounter - 1)
}
