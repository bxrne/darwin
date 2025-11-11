package individual

import (
	"encoding/json"
	"fmt"
	"github.com/bxrne/darwin/internal/rng"
	"strconv"
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

func applyOperator(opStr string, left, right float64, dividedByZero *bool) float64 {
	op := Operand(opStr) // cast string to Operand

	switch op {
	case Add:
		return left + right
	case Subtract:
		return left - right
	case Multiply:
		return left * right
	case Divide:
		if right == 0 {
			*dividedByZero = true
			return 0
		}
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

// Max returns the individual with higher fitness
func (i *Tree) Max(i2 Evolvable) Evolvable {
	if i.GetFitness() > i2.GetFitness() {
		return i
	}
	return i2
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

func (t *Tree) CalculateCrossoverPoint() (*TreeNode, *TreeNode, bool) {
	treeDepth := max(rng.Intn(t.depth), 1)
	leftNodeSelected := true

	treeNode := t.Root

	prevTreeNode := t.Root

	for i := range treeDepth {

		if i >= treeDepth || (treeNode.Left == nil && treeNode.Right == nil) {
			break
		}
		if rng.Intn(2) == 1 && treeNode.Left != nil {
			leftNodeSelected = true
			prevTreeNode = treeNode
			treeNode = treeNode.Left

		} else {
			leftNodeSelected = false
			prevTreeNode = treeNode
			treeNode = treeNode.Right
		}
	}
	return prevTreeNode, treeNode, leftNodeSelected

}

// MultiPointCrossover performs multi-point crossover between two trees
func (t *Tree) MultiPointCrossover(t2 Evolvable, crossoverPointCount int) (Evolvable, Evolvable) {
	tree2, ok := t2.(*Tree)
	if !ok {
		panic("Need Tree for Crossover")
	}
	prevFirstTreeNode, firstTreeNode, leftFirstNodeSelected := t.CalculateCrossoverPoint()
	prevSecondTreeNode, secondTreeNode, leftSecondNodeSelected := tree2.CalculateCrossoverPoint()
	if leftFirstNodeSelected {
		prevFirstTreeNode.Left = secondTreeNode.cloneNode()
	} else {
		prevFirstTreeNode.Right = secondTreeNode.cloneNode()
	}
	if leftSecondNodeSelected {
		prevSecondTreeNode.Left = firstTreeNode.cloneNode()
	} else {
		prevSecondTreeNode.Right = firstTreeNode.cloneNode()
	}

	t.depth = t.Root.CalculateMaxDepth()
	tree2.depth = tree2.Root.CalculateMaxDepth()

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

func (t *Tree) EvaluateTree(vars *map[string]float64) (float64, bool) {
	dividedByZero := false
	leftVal := t.Root.Left.NavigateTreeNode(vars, &dividedByZero)
	rightVal := t.Root.Right.NavigateTreeNode(vars, &dividedByZero)

	// Either use tn.Operator directly if filled in, or tn.Value
	return applyOperator(string(t.Root.Value), leftVal, rightVal, &dividedByZero), dividedByZero
}

func (tn *TreeNode) NavigateTreeNode(vars *map[string]float64, dividedByZero *bool) float64 {
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

	leftVal := tn.Left.NavigateTreeNode(vars, dividedByZero)
	rightVal := tn.Right.NavigateTreeNode(vars, dividedByZero)

	// Either use tn.Operator directly if filled in, or tn.Value
	return applyOperator(string(tn.Value), leftVal, rightVal, dividedByZero)
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
