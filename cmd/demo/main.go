package main

import (
	"fmt"

	"github.com/bxrne/darwin/internal/individual"
)

// stubGame simulates a simple navigation game
// Goal: reach position 5 (east), can move east or west
func stubGame() ([]string, []string) {
	actions := []string{"move_east", "move_west"}
	observations := []string{"current_pos", "target_pos"}
	return actions, observations
}

// createInitialTrees creates random trees for each action
func createInitialTrees(actions []string, maxDepth int) map[string]*individual.Tree {
	trees := make(map[string]*individual.Tree)

	operands := []string{"+", "-", "*", "/"}
	variables := []string{"current_pos", "target_pos"}
	terminals := []string{"0", "1", "2", "3", "4", "5"}

	for _, action := range actions {
		tree := individual.NewRandomTree(maxDepth, operands, variables, terminals)
		trees[action] = tree
	}

	return trees
}

// evaluateFitness tests if individual can reach target
func evaluateFitness(ati *individual.ActionTreeIndividual) float64 {
	position := 0.0 // start at west
	target := 5.0   // target is east
	steps := 10

	for step := range steps {
		vars := map[string]float64{
			"current_pos": position,
			"target_pos":  target,
		}

		// Get action values from trees
		moveEastValue, _ := ati.Trees["move_east"].Root.EvaluateTree(&vars)
		moveWestValue, _ := ati.Trees["move_west"].Root.EvaluateTree(&vars)

		// Choose action with higher value
		if moveEastValue > moveWestValue {
			position++ // move east
		} else {
			position-- // move west
		}

		// Check if reached target
		if position == target {
			return float64(steps - step) // bonus for reaching quickly
		}
	}

	// Return distance to target as negative fitness
	return -float64(int(position - target))
}

// demonstrateMutation shows mutation on ActionTreeIndividual
func demonstrateMutation(ati *individual.ActionTreeIndividual) {
	fmt.Println("\n=== Mutation Demo ===")
	fmt.Printf("Before: fitness=%.2f\n", ati.GetFitness())

	mutateInfo := &individual.MutateInformation{
		VariableSet: []string{"current_pos", "target_pos"},
		TerminalSet: []string{"0", "1", "2", "3", "4", "5"},
		OperandSet:  []string{"+", "-", "*", "/"},
	}

	ati.Mutate(0.3, mutateInfo)
	ati.SetFitness(evaluateFitness(ati))
	fmt.Printf("After: fitness=%.2f\n", ati.GetFitness())
}

// demonstrateCrossover shows crossover between ActionTreeIndividuals
func demonstrateCrossover(parent1, parent2 *individual.ActionTreeIndividual) {
	fmt.Println("\n=== Crossover Demo ===")
	fmt.Printf("Parent1: fitness=%.2f\n", parent1.GetFitness())
	fmt.Printf("Parent2: fitness=%.2f\n", parent2.GetFitness())

	crossoverInfo := &individual.CrossoverInformation{
		CrossoverPoints: 2,
		MaxDepth:        3,
	}

	child1, child2 := parent1.MultiPointCrossover(parent2, crossoverInfo)
	child1.SetFitness(evaluateFitness(child1.(*individual.ActionTreeIndividual)))
	child2.SetFitness(evaluateFitness(child2.(*individual.ActionTreeIndividual)))

	fmt.Printf("Child1: fitness=%.2f\n", child1.GetFitness())
	fmt.Printf("Child2: fitness=%.2f\n", child2.GetFitness())
}

func main() {
	fmt.Println("=== ActionTreeIndividual Demo ===")

	actions, observations := stubGame()
	fmt.Printf("Game: Navigate to position 5\n")
	fmt.Printf("Actions: %v\n", actions)
	fmt.Printf("Observations: %v\n", observations)

	// Create initial individual
	initialTrees := createInitialTrees(actions, 3)
	ati := individual.NewActionTreeIndividual(actions, len(observations), initialTrees)
	ati.SetFitness(evaluateFitness(ati))

	fmt.Println("\n=== Initial Individual ===")
	fmt.Printf("Fitness: %.2f\n", ati.GetFitness())
	fmt.Println(ati.Describe())

	// Demonstrate mutation
	mutated := ati.Clone().(*individual.ActionTreeIndividual)
	demonstrateMutation(mutated)

	// Demonstrate crossover
	initialTrees2 := createInitialTrees(actions, 3)
	ati2 := individual.NewActionTreeIndividual(actions, len(observations), initialTrees2)
	ati2.SetFitness(evaluateFitness(ati2))
	demonstrateCrossover(ati, ati2)

	fmt.Println("\n=== Demo Complete ===")
}
