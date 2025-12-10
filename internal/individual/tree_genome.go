package individual

import (
	"sort"
	"strconv"

	"github.com/bxrne/darwin/internal/rng"
)

type GrammarTree struct {
	Root    *TreeNode
	Genome  []int
	Fitness float64
	Depth   int
}

// NewGrammarTree creates a new binary individual with random genome
func NewGrammarTree(genomeSize int) *GrammarTree {
	genome := make([]int, genomeSize)
	for i := range genome {
		genome[i] = rng.Intn(255)
	}

	b := GrammarTree{Genome: genome}
	return &b
}

// GetFitness returns the fitness value
func (i *GrammarTree) GetFitness() float64 {
	return i.Fitness
}

// Describe returns a string representation of the genome (truncated for logs)
func (i *GrammarTree) Describe() string {
	if len(i.Genome) == 0 {
		return ""
	}

	const maxShown = 40
	limit := len(i.Genome)
	if limit > maxShown {
		limit = maxShown
	}

	response := ""
	for idx := 0; idx < limit; idx++ {
		response += strconv.Itoa(i.Genome[idx]) + ", "
	}
	if len(i.Genome) > maxShown {
		response += "…"
	}
	return response
}

// Max returns the individual with higher fitness
func (i *GrammarTree) Max(i2 Evolvable) Evolvable {
	o, ok := i2.(*GrammarTree)
	if !ok {
		panic("Max requires GrammarTree")
	}
	if i.Fitness > o.Fitness {
		return i
	}
	return i2
}

// Mutate performs mutation on the genome at specified points
func (i *GrammarTree) Mutate(mutationRate float64, _ *MutateInformation) {
	if mutationRate < rng.Float64() {
		return
	}
	for j := range len(i.Genome) {
		if mutationRate > rng.Float64() {
			i.Genome[j] = rng.Intn(255) // Flip '0' <-> '1'
		}
	}
}

// MultiPointCrossover performs multi-point crossover with another individual
func (i *GrammarTree) MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable) {
	o, ok := i2.(*GrammarTree)
	if !ok {
		panic("MultiPointCrossover requires GrammarTree")
	}

	crossoverPointArray := make([]int, 0)
	newI1Genome := make([]int, 0, len(i.Genome))
	newI2Genome := make([]int, 0, len(i.Genome))

	for range crossoverInformation.CrossoverPoints {
		crossoverPointArray = append(crossoverPointArray, rng.Intn(len(i.Genome)))
	}
	sort.Ints(crossoverPointArray)

	swap := true
	currentPointIndex := 0
	for j := range len(i.Genome) {
		if currentPointIndex < len(crossoverPointArray) && j >= crossoverPointArray[currentPointIndex] {
			swap = !swap
			currentPointIndex++
		}
		if swap {
			newI1Genome = append(newI1Genome, i.Genome[j])
			newI2Genome = append(newI2Genome, o.Genome[j])
		} else {
			newI1Genome = append(newI1Genome, o.Genome[j])
			newI2Genome = append(newI2Genome, i.Genome[j])
		}
	}

	newI1 := GrammarTree{Genome: newI1Genome}
	newI2 := GrammarTree{Genome: newI2Genome}
	return &newI1, &newI2
}

func (i *GrammarTree) SetFitness(fitness float64) {
	i.Fitness = fitness
}

// Clone creates a deep copy of the binary individual
func (i *GrammarTree) Clone() Evolvable {
	genomeCopy := make([]int, len(i.Genome))
	copy(genomeCopy, i.Genome)
	return &GrammarTree{
		Genome:  genomeCopy,
		Fitness: i.Fitness,
	}
}
