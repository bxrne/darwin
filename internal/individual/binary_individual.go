package individual

import (
	"github.com/bxrne/darwin/internal/rng"
	"sort"
)

// BinaryIndividual represents an individual with a binary genome
type BinaryIndividual struct {
	Genome  []byte
	Fitness float64
}

// NewBinaryIndividual creates a new binary individual with random genome
func NewBinaryIndividual(genomeSize int) *BinaryIndividual {
	genome := make([]byte, genomeSize)
	for i := range genome {
		genome[i] = '0' + byte(rng.Intn(2))
	}

	b := BinaryIndividual{Genome: genome}
	return &b
}

// GetFitness returns the fitness value
func (i *BinaryIndividual) GetFitness() float64 {
	return i.Fitness
}

// Max returns the individual with higher fitness
func (i *BinaryIndividual) Max(i2 Evolvable) Evolvable {
	o, ok := i2.(*BinaryIndividual)
	if !ok {
		panic("Max requires BinaryIndividual")
	}
	if i.Fitness > o.Fitness {
		return i
	}
	return i2
}

// Mutate performs mutation on the genome at specified points
func (i *BinaryIndividual) Mutate(mutationRate float64) {
	if mutationRate < rng.Float64() {
		return
	}
	for j := range len(i.Genome) {
		if mutationRate > rng.Float64() {
			i.Genome[j] ^= 1 // Flip '0' <-> '1'
		}
	}
}

// MultiPointCrossover performs multi-point crossover with another individual
func (i *BinaryIndividual) MultiPointCrossover(i2 Evolvable, crossoverPoints int) (Evolvable, Evolvable) {
	o, ok := i2.(*BinaryIndividual)
	if !ok {
		panic("MultiPointCrossover requires BinaryIndividual")
	}

	crossoverPointArray := make([]int, 0)
	newI1Genome := make([]byte, 0, len(i.Genome))
	newI2Genome := make([]byte, 0, len(i.Genome))

	for range crossoverPoints {
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

	newI1 := BinaryIndividual{Genome: newI1Genome}
	newI2 := BinaryIndividual{Genome: newI2Genome}
	return &newI1, &newI2
}

func (i *BinaryIndividual) SetFitness(fitness float64) {
	i.Fitness = fitness
}
