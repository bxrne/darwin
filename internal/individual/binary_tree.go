package individual

import (
	"math/rand"
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
func NewRandomTree(depth int) *Tree {
	if depth == 0 {
		return &Tree{Root: &TreeNode{Value: strconv.Itoa(rand.Intn(10))}}
	}

	ops := []Operand{Add, Subtract, Multiply, Divide}
	op := ops[rand.Intn(len(ops))]

	return &Tree{
		Root: &TreeNode{
			Value: string(op),
			Left:  NewRandomTree(depth - 1).Root,
			Right: NewRandomTree(depth - 1).Root,
		},
	}
}

// Max returns the tree with the higher fitness
func (t *Tree) Max(t2 Evolvable) Evolvable {
	return t2
}

// MultiPointCrossover performs multi-point crossover between two trees
func (t *Tree) MultiPointCrossover(t2 Evolvable, crossoverPointCount int) (Evolvable, Evolvable) {
	// Placeholder for actual crossover logic
	// For simplicity, we return copies of the original trees
	return t, t2
}

// Mutate mutates the tree based on the given mutation rate
func (t *Tree) Mutate(rate float64) {
	// Placeholder for actual mutation logic
	// For simplicity, we randomly change the value of the root node
	if rand.Float64() < rate {
		t.Root.Value = strconv.Itoa(rand.Intn(10))
	}
}

// CalculateFitness calculates the fitness of the tree
func (t *Tree) CalculateFitness() {
	// Placeholder for actual fitness calculation
	// For simplicity, we assign a random fitness value
	t.Fitness = rand.Float64() * 100
}

func (t *Tree) EvaluateTree(vars *map[string]float64) float64 {

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
