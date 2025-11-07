package individual

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bxrne/darwin/internal/rng"
)

// TreeNode represents a node in the expression tree
type TreeNode struct {
	Value string
	Left  *TreeNode
	Right *TreeNode
}

// Tree represents the entire expression tree
type Tree struct {
	Root    *TreeNode
	Fitness float64
	depth   int
}

// Operand represents the type of operation in the tree nodes
type Operand string

const (
	Add      Operand = "+"
	Subtract Operand = "-"
	Multiply Operand = "*"
	Divide   Operand = "/"
)

func applyOperator(opStr string, left, right float64) float64 {
	op := Operand(opStr) // cast string to Operand

	switch op {
	case Add:
		return left + right
	case Subtract:
		return left - right
	case Multiply:
		return left * right
	case Divide:
		return left / right
	default:
		panic(fmt.Sprintf("unknown operator: %s", op))
	}
}

// NewRandomTree generates a random expression tree
func NewRandomTree(depth int, primitiveSet []string, terminalSet []string) *Tree {
	if depth == 0 {
		return &Tree{Root: &TreeNode{Value: terminalSet[rng.Intn(len(terminalSet))]}}
	}

	// Convert primitive strings to Operand types
	functionSet := make([]Operand, 0, len(primitiveSet))
	for _, prim := range primitiveSet {
		functionSet = append(functionSet, Operand(prim))
	}

	op := functionSet[rng.Intn(len(functionSet))]

	return &Tree{
		Root: &TreeNode{
			Value: string(op),
			Left:  NewRandomTreeNode(depth-1, terminalSet, functionSet),
			Right: NewRandomTreeNode(depth-1, terminalSet, functionSet),
		},
		depth: depth,
	}
}

// NewRandomTreeNode generates a random expression treenode
func NewRandomTreeNode(depth int, terminalSet []string, functionSet []Operand) *TreeNode {
	if depth == 0 {
		return &TreeNode{Value: terminalSet[rng.Intn(len(terminalSet))]}
	}

	op := functionSet[rng.Intn(len(functionSet))]

	return &TreeNode{
		Value: string(op),
		Left:  NewRandomTreeNode(depth-1, terminalSet, functionSet),
		Right: NewRandomTreeNode(depth-1, terminalSet, functionSet),
	}
}

// GetDepth returns the depth of the tree
func (t *Tree) GetDepth() int {
	return t.depth
}

// Describe returns a human-readable string representation of the tree
func (t *Tree) Describe() string {
	if t.Root == nil {
		return "empty tree"
	}
	return t.Root.describeNode()
}

// describeNode recursively creates a human-readable expression
func (tn *TreeNode) describeNode() string {
	if tn.IsLeaf() {
		return tn.Value
	}

	leftExpr := tn.Left.describeNode()
	rightExpr := tn.Right.describeNode()

	return fmt.Sprintf("(%s %s %s)", leftExpr, tn.Value, rightExpr)
}

// Max returns the tree with the higher fitness
func (t *Tree) Max(t2 Evolvable) Evolvable {
	return t2
}

func (t *TreeNode) CalculateMaxDepth() int {
	leftDepth := -1
	rightDepth := -1
	if t.Left != nil {
		leftDepth = t.Left.CalculateMaxDepth()
	}
	if t.Right != nil {
		rightDepth = t.Right.CalculateMaxDepth()
	}
	return max(leftDepth, rightDepth) + 1

}

// MultiPointCrossover performs multi-point crossover between two trees
func (t *Tree) MultiPointCrossover(t2 Evolvable, crossoverPointCount int) (Evolvable, Evolvable) {
	tree2, ok := t2.(*Tree)
	if !ok {
		panic("Need Tree for Crossover")
	}
	firstTreeDepth := rand.Intn(t.depth)
	secondTreeDepth := rand.Intn(tree2.depth)
	firstTreeNode := &TreeNode{}
	secondTreeNode := &TreeNode{}
	leftFirstNodeSelected := true
	leftSecondNodeSelected := true
	if rand.Intn(2) == 1 {

		firstTreeNode = t.Root.Left
		leftSecondNodeSelected = false
		secondTreeNode = tree2.Root.Right
	} else {
		leftFirstNodeSelected = true
		firstTreeNode = t.Root.Right
		secondTreeNode = tree2.Root.Left
	}

	firstTreeSet := false
	secondTreeSet := false
	newFirstTreeDepth := t.depth
	newSecondTreeDepth := tree2.depth
	var prevFirstTreeNode *TreeNode = nil
	var prevSecondTreeNode *TreeNode = nil

	for i := range max(firstTreeDepth, secondTreeDepth) {
		if !firstTreeSet && i >= firstTreeDepth || (firstTreeNode.Left == nil && firstTreeNode.Right == nil) {
			firstTreeSet = true
			newFirstTreeDepth = max(newFirstTreeDepth, i+firstTreeNode.CalculateMaxDepth())
		}
		if !secondTreeSet && i >= secondTreeDepth || (secondTreeNode.Left == nil && secondTreeNode.Right == nil) {
			secondTreeSet = true
			newSecondTreeDepth = max(newSecondTreeDepth, i+secondTreeNode.CalculateMaxDepth())
		}
		if !firstTreeSet && rand.Intn(2) == 1 && firstTreeNode.Left != nil {
			leftFirstNodeSelected = true
			prevFirstTreeNode = firstTreeNode
			firstTreeNode = firstTreeNode.Left

		} else {
			leftFirstNodeSelected = true
			prevFirstTreeNode = firstTreeNode
			firstTreeNode = firstTreeNode.Right
		}

		if secondTreeSet && rand.Intn(2) == 1 && secondTreeNode.Left != nil {
			leftSecondNodeSelected = true
			prevSecondTreeNode = secondTreeNode
			secondTreeNode = secondTreeNode.Left
		} else {
			leftSecondNodeSelected = false
			prevSecondTreeNode = secondTreeNode
			secondTreeNode = secondTreeNode.Right
		}
	}
	if prevFirstTreeNode == nil {
		if leftFirstNodeSelected {
			t.Root.Left = secondTreeNode
		} else {
			t.Root.Right = secondTreeNode
		}
	} else {
		if leftFirstNodeSelected {
			prevFirstTreeNode.Left = secondTreeNode
		} else {
			prevFirstTreeNode.Right = secondTreeNode
		}
	}
	if prevSecondTreeNode == nil {
		if leftSecondNodeSelected {
			tree2.Root.Left = firstTreeNode
		} else {
			tree2.Root.Right = firstTreeNode
		}
	} else {
		if leftSecondNodeSelected {
			prevSecondTreeNode.Left = firstTreeNode
		} else {
			prevSecondTreeNode.Right = firstTreeNode
		}
	}
	t.depth = newFirstTreeDepth
	tree2.depth = newSecondTreeDepth

	return t, tree2
}

// Mutate mutates the tree based on the given mutation rate (interface compatibility)
func (t *Tree) Mutate(rate float64) {
	// This method maintains interface compatibility
	// The actual mutation logic should be called via MutateWithSets
	// This is a fallback that does nothing
}

// MutateWithSets mutates the tree using provided primitive and terminal sets
func (t *Tree) MutateWithSets(rate float64, primitiveSet []string, terminalSet []string) {
	t.Root = t.Root.mutateRecursive(rate, primitiveSet, terminalSet)
}

func (t *Tree) EvaluateTree(vars *map[string]float64) float64 {
	// If root is a leaf, just return its value
	if t.Root.IsLeaf() {
		return t.Root.NavigateTreeNode(vars)
	}

	leftVal := t.Root.Left.NavigateTreeNode(vars)
	rightVal := t.Root.Right.NavigateTreeNode(vars)

	// Either use tn.Operator directly if filled in, or tn.Value
	return applyOperator(string(t.Root.Value), leftVal, rightVal)

}

func (tn *TreeNode) NavigateTreeNode(vars *map[string]float64) float64 {
	if val, ok := (*vars)[tn.Value]; ok {
		return val
	}
	if num, err := strconv.ParseFloat(tn.Value, 64); err == nil {
		return num
	}

	// Check if this is a leaf node - if so, we shouldn't be here
	if tn.IsLeaf() {
		panic(fmt.Sprintf("attempted to navigate leaf node as operator: %s", tn.Value))
	}

	leftVal := tn.Left.NavigateTreeNode(vars)
	rightVal := tn.Right.NavigateTreeNode(vars)

	// Either use tn.Operator directly if filled in, or tn.Value
	return applyOperator(string(tn.Value), leftVal, rightVal)
}

func (t *Tree) SetFitness(fitness float64) {
	t.Fitness = fitness
}

// GetFitness returns the fitness of the tree
func (t *Tree) GetFitness() float64 {
	return t.Fitness
}

// IsLeaf checks if the node is a leaf (terminal)
func (tn *TreeNode) IsLeaf() bool {
	return tn.Left == nil && tn.Right == nil
}

// MutateTerminal replaces a terminal node with a different terminal from the set
func (tn *TreeNode) MutateTerminal(terminalSet []string) {
	currentTerminal := tn.Value
	availableTerminals := make([]string, 0, len(terminalSet))

	// Exclude current terminal to ensure actual change
	for _, terminal := range terminalSet {
		if terminal != currentTerminal {
			availableTerminals = append(availableTerminals, terminal)
		}
	}

	if len(availableTerminals) > 0 {
		newTerminal := availableTerminals[rng.Intn(len(availableTerminals))]
		tn.Value = newTerminal
	}
}

// MutateFunction replaces a function node with a different function from the primitive set
func (tn *TreeNode) MutateFunction(primitiveSet []string) {
	currentFunction := tn.Value
	availableFunctions := make([]string, 0, len(primitiveSet))

	// Exclude current function to ensure actual change
	for _, function := range primitiveSet {
		if function != currentFunction {
			availableFunctions = append(availableFunctions, function)
		}
	}

	if len(availableFunctions) > 0 {
		newFunction := availableFunctions[rng.Intn(len(availableFunctions))]
		tn.Value = newFunction
	}
}

// mutateRecursive traverses the tree and gives each node a chance to mutate
func (tn *TreeNode) mutateRecursive(rate float64, primitiveSet []string, terminalSet []string) *TreeNode {
	// First, recursively mutate children (if any)
	if tn.Left != nil {
		tn.Left = tn.Left.mutateRecursive(rate, primitiveSet, terminalSet)
	}
	if tn.Right != nil {
		tn.Right = tn.Right.mutateRecursive(rate, primitiveSet, terminalSet)
	}

	// Then, decide if this node should mutate
	if rng.Float64() < rate {
		if tn.IsLeaf() {
			// It's a terminal node
			tn.MutateTerminal(terminalSet)
		} else {
			// It's a function node
			tn.MutateFunction(primitiveSet)
		}
	}

	return tn
}

// Clone creates a deep copy of the tree
func (t *Tree) Clone() Evolvable {
	clonedRoot := t.Root.cloneNode()
	return &Tree{
		Root:    clonedRoot,
		Fitness: t.Fitness,
		depth:   t.depth,
	}
}

// cloneNode creates a deep copy of a tree node
func (tn *TreeNode) cloneNode() *TreeNode {
	if tn.IsLeaf() {
		return &TreeNode{
			Value: tn.Value,
			Left:  nil,
			Right: nil,
		}
	}

	return &TreeNode{
		Value: tn.Value,
		Left:  tn.Left.cloneNode(),
		Right: tn.Right.cloneNode(),
	}
}

func PrintTreeJSON(t *Tree) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))
}
