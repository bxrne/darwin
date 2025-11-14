package individual

import "fmt"

type Node any

type Choice struct{ Options []Node }
type Seq struct{ Items []Node }
type Term struct{ Value string }
type NonTerm struct{ Name string }

func expandArrayToTermArray(arr []string, isTerminal bool) []Node {
	nodeArray := make([]Node, 0)
	for _, element := range arr {
		if isTerminal {
			nodeArray = append(nodeArray, Term{element})
		} else {
			nodeArray = append(nodeArray, NonTerm{element})
		}
	}
	return nodeArray
}

func CreateGrammar(terminalSet []string, primitiveSet []string, operatorSet []string) map[string]Node {
	return map[string]Node{
		"Expr": Choice{
			Options: []Node{
				Seq{Items: []Node{NonTerm{"Expr"}, NonTerm{"Operand"}, NonTerm{"Expr"}}},
				NonTerm{"Var"},
			},
		},
		"Operand": Choice{
			Options: expandArrayToTermArray(operatorSet, true),
		},
		"Var": Choice{
			Options: []Node{
				NonTerm{"Primitive"},
				NonTerm{"Terminals"},
			},
		},
		"Terminals": Choice{
			Options: expandArrayToTermArray(terminalSet, true),
		},
		"Primitive": Choice{
			Options: expandArrayToTermArray(primitiveSet, true),
		},
	}
}

func GenerateTreeFromGenome(grammar map[string]Node, genome []int) *TreeNode {
	return GenerateTreeFromGenomeWithDepth(grammar, genome, 5) // Default max depth
}

func GenerateTreeFromGenomeWithDepth(grammar map[string]Node, genome []int, maxDepth int) *TreeNode {
	idx := 0
	treeNode := &TreeNode{}
	treeNode.ExpandTree(NonTerm{"Expr"}, genome, &idx, grammar, 0, maxDepth)
	return treeNode
}

func (tn *TreeNode) ExpandTree(n Node, codons []int, idx *int, grammar map[string]Node, currentDepth int, maxDepth int) {
	switch v := n.(type) {

	case Term:
		tn.Value = v.Value
		return

	case NonTerm:
		// Lookup its production rule and expand it
		tn.ExpandTree(grammar[v.Name], codons, idx, grammar, currentDepth, maxDepth)

	case Seq:
		// Seq represents: Expr Operand Expr (infix notation)
		// Current node becomes the operator, children become expressions
		tn.Left = &TreeNode{}
		tn.Left.ExpandTree(v.Items[0], codons, idx, grammar, currentDepth+1, maxDepth) // First Expr
		tn.Right = &TreeNode{}
		tn.Right.ExpandTree(v.Items[2], codons, idx, grammar, currentDepth+1, maxDepth) // Second Expr

		// Handle the operator (middle item)
		switch op := v.Items[1].(type) {
		case Term:
			tn.Value = op.Value
		case NonTerm:
			// Expand non-terminal to get the actual operator
			tempNode := &TreeNode{}
			tempNode.ExpandTree(op, codons, idx, grammar, currentDepth, maxDepth)
			tn.Value = tempNode.Value
		default:
			panic(fmt.Sprintf("Unexpected operator type in Seq: %T", v.Items[1]))
		}

	case Choice:
		// When at max depth, force selection of terminal-only options
		if currentDepth >= maxDepth {
			tn.expandChoiceWithDepthLimit(v, codons, idx, grammar, currentDepth, maxDepth)
			return
		}

		// Consume codon with wrap-around
		if len(codons) == 0 {
			// No codons available, select first option as fallback
			option := v.Options[0]
			tn.ExpandTree(option, codons, idx, grammar, currentDepth, maxDepth)
			return
		}

		c := codons[*idx%len(codons)]
		*idx++

		option := v.Options[c%len(v.Options)]
		tn.ExpandTree(option, codons, idx, grammar, currentDepth, maxDepth)
	}
}

// Helper method to expand Choice with depth limiting (forces terminal selection)
func (tn *TreeNode) expandChoiceWithDepthLimit(choice Choice, codons []int, idx *int, grammar map[string]Node, currentDepth int, maxDepth int) {
	// Find terminal-only options
	var terminalOptions []Node
	for _, option := range choice.Options {
		if tn.isTerminalOption(option, grammar) {
			terminalOptions = append(terminalOptions, option)
		}
	}

	// If no terminal options found, pick the first option as fallback
	var selectedOption Node
	if len(terminalOptions) > 0 {
		// Consume codon to select from terminal options
		if len(codons) == 0 {
			selectedOption = terminalOptions[0]
		} else {
			c := codons[*idx%len(codons)]
			*idx++
			selectedOption = terminalOptions[c%len(terminalOptions)]
		}
	} else {
		// Fallback: pick first option
		selectedOption = choice.Options[0]
	}

	tn.ExpandTree(selectedOption, codons, idx, grammar, currentDepth, maxDepth)
}

// Check if an option leads directly to terminals (no Expr recursion)
func (tn *TreeNode) isTerminalOption(option Node, grammar map[string]Node) bool {
	switch v := option.(type) {
	case NonTerm:
		// Check if this non-terminal leads to terminals without Expr recursion
		return v.Name == "Var" || v.Name == "Terminals" || v.Name == "Primitive" || v.Name == "Operand"
	case Term:
		return true
	case Seq:
		// Seq contains Expr, so it's not terminal
		return false
	case Choice:
		// Check if all options in this choice are terminal
		for _, opt := range v.Options {
			if !tn.isTerminalOption(opt, grammar) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
