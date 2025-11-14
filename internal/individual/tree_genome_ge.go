package individual

type Node any

type Choice struct{ Options []Node }
type Seq struct{ Items []Node }
type Term struct{ Value string }
type NonTerm struct{ Name string }

func expandArrayToTermArray(arr []string, isTerminal bool) []Node {
	nodeArray := make([]Node, 0, 0)
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
				Seq{Items: []Node{NonTerm{"Expr"}, Term{"Operand"}, NonTerm{"Expr"}}},
				NonTerm{"Var"},
			},
		},
		"Operand": Choice{
			Options: []Node{
				expandArrayToTermArray(operatorSet, true),
			},
		},
		"Var": Choice{
			Options: []Node{
				NonTerm{"Primitive"},
				NonTerm{"Terminals"},
			},
		},
		"Terminals": Choice{
			Options: []Node{
				expandArrayToTermArray(terminalSet, true),
			},
		},
		"Primitive": Choice{Options: []Node{
			expandArrayToTermArray(primitiveSet, true),
		}},
	}
}

func GenerateTreeFromGenome(grammar map[string]Node, genome []int) *TreeNode {
	idx := 0
	treeNode := &TreeNode{}
	treeNode.ExpandTree(NonTerm{"Expr"}, genome, &idx, grammar)
	return treeNode
}

func (tn *TreeNode) ExpandTree(n Node, codons []int, idx *int, grammar map[string]Node) {
	switch v := n.(type) {

	case Term:
		tn.Value = v.Value
		return

	case NonTerm:
		// Lookup its production rule and expand it
		tn.ExpandTree(grammar[v.Name], codons, idx, grammar)

	case Seq:
		tn.Left = &TreeNode{}
		tn.Left.ExpandTree(NonTerm{"Expr"}, codons, idx, grammar)
		tn.Right = &TreeNode{}
		tn.Right.ExpandTree(NonTerm{"Expr"}, codons, idx, grammar)
		tn.ExpandTree(NonTerm{"Operand"}, codons, idx, grammar)

	case Choice:
		// Consume codon with wrap-around
		if len(codons) == 0 {
			panic("no codons available")
		}

		c := codons[*idx%len(codons)]
		*idx++

		option := v.Options[c%len(v.Options)]
		tn.ExpandTree(option, codons, idx, grammar)
	}
}
