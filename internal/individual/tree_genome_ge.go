package individual

type Node any

type Choice struct{ Options []Node }
type Seq struct{ Items []Node }
type Term struct{ Value string }
type NonTerm struct{ Name string }

func expandArrayToTermArray(arr []string, isTerminal bool) []Node {
	nodeArray := make([]Node, 0, 0)
	for index, element := range arr {
		if isTerminal {
			nodeArray = append(nodeArray, Term{element})
		} else {
			nodeArray = append(nodeArray, NonTerm{element})
		}
	}
	return nodeArray
}

func createMap(terminalSet []string, primitiveSet []string, operatorSet []string) map[string]Node {
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

func generateTreeFromGenome(grammar map[string]Node, genome []int8) *TreeNode {
	idx := 0
	return
}

func (tn * TreeNode) ExpandTree(n Node, codons []int, idx *int, grammar map[string]Node) *TreeNode {
	switch v := n.(type) {

	case Term:
		return v.Value

	case NonTerm:
		// Lookup its production rule and expand it
		if v.Name == "Operand" {
			tn.Left = ExpandTree(grammar[v.Name], codons, idx, grammar) 
			tn.Right = ExpandTree(grammar[v.Name], codons, idx, grammar)}
		}
		return ExpandTree(grammar[v.Name], codons, idx, grammar)

	case Seq:
		for _, item := range v.Items {
			ExpandTree(item, codons, idx, grammar)
		}
		return out

	case Choice:
		// Consume codon with wrap-around
		if len(codons) == 0 {
			panic("no codons available")
		}

		c := codons[*idx%len(codons)]
		*idx++

		option := v.Options[c%len(v.Options)]
		return ExpandTree(option, codons, idx, grammar)
	}

	return &TreeNode{}
}
