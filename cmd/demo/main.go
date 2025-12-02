package main

import (
	"fmt"

	"github.com/bxrne/darwin/internal/individual"
)

func main() {
	fmt.Println("ActionTreeIndividual Demo")

	// Game: Navigate to position 5, can move east or west
	actions := []string{"move_east", "move_west"}
	observations := []string{"current_pos", "target_pos"}

	fmt.Printf("Actions=%v\n", actions)
	fmt.Printf("Observations=%v\n", observations)

	// Create ActionTreeIndividual with trees that use observations as variables
	operands := []string{"+", "-", "*", "/"}
	variables := observations // trees use all observations
	terminals := []string{"0", "1", "2", "3", "4", "5"}

	ati := individual.NewRandomActionTreeIndividual(len(observations), 3, operands, variables, terminals)

	fmt.Println("\nActionTreeIndividual")
	fmt.Printf("move_east_tree=%s\n", ati.Trees["move_east"].Describe())
	fmt.Printf("move_west_tree=%s\n", ati.Trees["move_west"].Describe())

	// Test tree evaluation with observations
	vars := map[string]float64{
		"current_pos": 2.0,
		"target_pos":  5.0,
	}

	eastValue, _ := ati.Trees["move_east"].Root.EvaluateTree(&vars)
	westValue, _ := ati.Trees["move_west"].Root.EvaluateTree(&vars)

	fmt.Printf("\nAt position=2.0, target=5.0:\n")
	fmt.Printf("move_east_evaluates=%.2f\n", eastValue)
	fmt.Printf("move_west_evaluates=%.2f\n", westValue)

	if eastValue > westValue {
		fmt.Println("Decision=move_east (higher value)")
	} else {
		fmt.Println("Decision=move_west (higher value)")
	}

	// Demonstrate mutation
	fmt.Println("\nMutation Demo")
	fmt.Printf("Before_mutation_move_east=%s\n", ati.Trees["move_east"].Describe())

	mutateInfo := &individual.MutateInformation{
		VariableSet: variables,
		TerminalSet: terminals,
		OperandSet:  operands,
	}

	ati.Mutate(0.3, mutateInfo)
	fmt.Printf("After_mutation_move_east=%s\n", ati.Trees["move_east"].Describe())

	fmt.Println("\nDemo Complete")
}
