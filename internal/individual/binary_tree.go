package individual

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

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
	Root      *TreeNode
	Fitness   float64
	fitnessMu sync.RWMutex  // Protects Fitness field from concurrent access
	depth     int
}

// Operand represents the type of operation in the tree nodes
type Operand string

const (
	Add      Operand = "+"
	Subtract Operand = "-"
	Multiply Operand = "*"
	Divide   Operand = "/"
	Modulo   Operand = "%"
	Pow      Operand = "pow"
	Abs      Operand = "abs"
	Min      Operand = "min"
	Max      Operand = "max"
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
			// Return a meaningful penalty value instead of 0 to avoid cascading errors
			return -1000.0
		}
		return left / right
	case Modulo:
		if right == 0 {
			*dividedByZero = true
			return -1000.0
		}
		return math.Mod(left, right)
	case Pow:
		// Handle edge cases for pow
		if left == 0 && right < 0 {
			*dividedByZero = true
			return -1000.0
		}
		return math.Pow(left, right)
	case Abs:
		// abs only uses left operand
		return math.Abs(left)
	case Min:
		return math.Min(left, right)
	case Max:
		return math.Max(left, right)
	default:
		panic(fmt.Sprintf("unknown operator: %s", op))
	}
}

// newFullTree generates a tree where all non-leaf nodes are functions and all leaves are at max Depth
func NewFullTree(depth int, operandSet []string, variableSet []string, terminalSet []string) *Tree {
	functionSet := make([]Operand, 0, len(operandSet))
	for _, prim := range operandSet {
		functionSet = append(functionSet, Operand(prim))
	}
	overallTerminalSet := append(terminalSet, variableSet...)

	return &Tree{
		Root:  newFullTreeNode(depth, overallTerminalSet, functionSet),
		depth: depth,
	}
}

// newFullTreeNode generates a full tree node (functions at all non-zero Depths)
func newFullTreeNode(depth int, terminalSet []string, functionSet []Operand) *TreeNode {
	if depth == 0 {
		return &TreeNode{Value: terminalSet[rng.Intn(len(terminalSet))]}
	}

	op := functionSet[rng.Intn(len(functionSet))]
	return &TreeNode{
		Value: string(op),
		Left:  newFullTreeNode(depth-1, terminalSet, functionSet),
		Right: newFullTreeNode(depth-1, terminalSet, functionSet),
	}
}

// newGrowTree generates a tree where nodes can be functions or terminals at any Depth
func newGrowTree(depth int, operandSet []string, variableSet []string, terminalSet []string) *Tree {
	functionSet := make([]Operand, 0, len(operandSet))
	for _, prim := range operandSet {
		functionSet = append(functionSet, Operand(prim))
	}
	overallTerminalSet := append(terminalSet, variableSet...)

	return &Tree{
		Root:  newGrowTreeNode(depth, overallTerminalSet, functionSet),
		depth: depth,
	}
}

// newGrowTreeNode generates a grow tree node (can choose between function and terminal)
func newGrowTreeNode(depth int, terminalSet []string, functionSet []Operand) *TreeNode {
	if depth == 0 {
		return &TreeNode{Value: terminalSet[rng.Intn(len(terminalSet))]}
	}

	// At non-zero Depth, randomly choose between function and terminal
	if rng.Float64() < 0.3 {
		// Choose terminal
		return &TreeNode{Value: terminalSet[rng.Intn(len(terminalSet))]}
	}

	// Choose function
	op := functionSet[rng.Intn(len(functionSet))]
	return &TreeNode{
		Value: string(op),
		Left:  newGrowTreeNode(depth-1, terminalSet, functionSet),
		Right: newGrowTreeNode(depth-1, terminalSet, functionSet),
	}
}

// NewRandomTree generates a random expression tree using ramped half-and-half method
func NewRandomTree(maxDepth int, operandSet []string, variableSet []string, terminalSet []string) *Tree {
	// For single tree creation, use random depth between 0 and maxDepth
	// This maintains compatibility with existing usage
	depth := rng.Intn(maxDepth + 1)

	// Randomly choose between grow (50%) and full (50%) methods
	if rng.Float64() < 0.5 {
		return newGrowTree(depth, operandSet, variableSet, terminalSet)
	}
	return NewFullTree(depth, operandSet, variableSet, terminalSet)
}

// NewRampedHalfAndHalfTree generates a tree with specified Depth using ramped half-and-half
// This is useful for population initialization where specific Depths are needed
func NewRampedHalfAndHalfTree(depth int, useGrow bool, operandSet []string, variableSet []string, terminalSet []string) *Tree {
	if useGrow {
		return newGrowTree(depth, operandSet, variableSet, terminalSet)
	}
	return NewFullTree(depth, operandSet, variableSet, terminalSet)
}

// GetDepth returns the Depth of the tree
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

func (t *Tree) CalculateCrossoverPoint(otherTreeDepth int, maxDepth int) (*TreeNode, *TreeNode, bool) {
	maxTreeDepth := t.Root.CalculateMaxDepth()
	if maxTreeDepth <= 0 {
		return nil, nil, false
	}
	treeDepth := max(rng.Intn(maxTreeDepth+1), 1)
	leftNodeSelected := true

	treeNode := t.Root

	prevTreeNode := t.Root
	for i := range maxTreeDepth {

		// Only break if we've reached desired Depth AND current node is a function (has children)
		if ((otherTreeDepth+treeNode.CalculateMaxDepth()) <= maxDepth && i >= treeDepth) || treeNode.IsLeaf() {
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
func (t *Tree) MultiPointCrossover(t2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable) {
	tree2, ok := t2.(*Tree)
	if !ok {
		panic("Need Tree for Crossover")
	}

	// Handle case where either tree has Depth 0 (no crossover possible)
	if t.depth <= 0 || tree2.depth <= 0 {
		return t, tree2
	}

	prevFirstTreeNode, firstTreeNode, leftFirstNodeSelected := t.CalculateCrossoverPoint(tree2.depth, crossoverInformation.MaxDepth)
	prevSecondTreeNode, secondTreeNode, leftSecondNodeSelected := tree2.CalculateCrossoverPoint(t.depth, crossoverInformation.MaxDepth)

	// Check if crossover points are valid
	if prevFirstTreeNode == nil || prevSecondTreeNode == nil || firstTreeNode == nil || secondTreeNode == nil {
		return t, tree2
	}

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
func (t *Tree) Mutate(rate float64, mutateInformation *MutateInformation) {
	newSet := append(mutateInformation.TerminalSet, mutateInformation.VariableSet...)
	t.Root = t.Root.mutateRecursive(rate, mutateInformation.OperandSet, newSet)
}

func (t *TreeNode) EvaluateTree(vars *map[string]float64) (float64, bool) {
	dividedByZero := false
	return t.NavigateTreeNode(vars, &dividedByZero), dividedByZero
}

func (tn *TreeNode) NavigateTreeNode(vars *map[string]float64, dividedByZero *bool) float64 {
	if tn == nil {
		panic("attempted to navigate nil tree node")
	}

	// Check if this is a terminal node (variable or constant)
	if tn.IsLeaf() {
		if val, ok := (*vars)[tn.Value]; ok {
			return val
		}
		if num, err := strconv.ParseFloat(tn.Value, 64); err == nil {
			return num
		}
		panic(fmt.Sprintf("invalid terminal value: %s", tn.Value))
	}

	// This is a function node, evaluate children
	if tn.Left == nil || tn.Right == nil {
		panic(fmt.Sprintf("function node %s has nil child", tn.Value))
	}

	leftVal := tn.Left.NavigateTreeNode(vars, dividedByZero)
	rightVal := tn.Right.NavigateTreeNode(vars, dividedByZero)

	return applyOperator(string(tn.Value), leftVal, rightVal, dividedByZero)
}

func (t *Tree) SetFitness(fitness float64) {
	t.fitnessMu.Lock()
	defer t.fitnessMu.Unlock()
	t.Fitness = fitness
}

// GetFitness returns the fitness of the tree
func (t *Tree) GetFitness() float64 {
	t.fitnessMu.RLock()
	defer t.fitnessMu.RUnlock()
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
	t.fitnessMu.RLock()
	fitness := t.Fitness
	t.fitnessMu.RUnlock()
	return &Tree{
		Root:    clonedRoot,
		Fitness: fitness,
		depth:   t.depth,
	}
}

// cloneNode creates a deep copy of a tree node
func (tn *TreeNode) cloneNode() *TreeNode {
	if tn == nil {
		return nil
	}
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
func TreeToJSON(t *Tree) string {
	b, _ := json.MarshalIndent(t, "", "  ")
	return string(b)
}
