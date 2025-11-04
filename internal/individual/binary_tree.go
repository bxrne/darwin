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
	o, ok := t2.(*Tree)
	if !ok {
		panic("Max requires Tree")
	}
	if t.Fitness > o.Fitness {
		return t
	}
	return t2
}

// MultiPointCrossover performs multi-point crossover between two Evolvables
func (t *Tree) MultiPointCrossover(partner Evolvable, points int) (Evolvable, Evolvable) {
	o, ok := partner.(*Tree)
	if !ok {
		panic("MultiPointCrossover requires Tree")
	}

	for i := 0; i < points; i++ {
		// For simplicity, we swap the root nodes
		t.Root, o.Root = o.Root, t.Root
	}

	return t, o
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

// GetFitness returns the fitness of the tree
func (t *Tree) GetFitness() float64 {
	return t.Fitness
}
