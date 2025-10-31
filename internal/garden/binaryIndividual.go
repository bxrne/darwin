package garden

import (
	"math/rand"
	"sort"
)

type BinaryIndividual struct {
	Genome  []byte
	Fitness float64
}

func newBinaryIndividual(genomeSize int) *BinaryIndividual {
	genome := make([]byte, genomeSize)
	for i := range genome {
		genome[i] = '0' + byte(rand.Intn(2))
	}

	b := BinaryIndividual{Genome: genome}
	b.CalculateFitness()
	return &b
}

func (i *BinaryIndividual) GetFitness() float64 {
	return i.Fitness
}
func (i *BinaryIndividual) CalculateFitness() {
	count := 0
	for _, gene := range i.Genome {
		if gene == '1' {
			count++
		}
	}
	i.Fitness = float64(count) / float64(len(i.Genome))
}

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

func (i *BinaryIndividual) Mutate(points []int, mutatationRate float64) {
	if mutatationRate < rand.Float64() {
		return
	}
	for _, point := range points {
		i.Genome[point] ^= 1 // Flip '0' <-> '1'
	}
}

func (i *BinaryIndividual) MultiPointCrossover(i2 Evolvable, crossoverPoints int) (Evolvable, Evolvable) {
	o, ok := i2.(*BinaryIndividual)
	if !ok {
		panic("MultiPointCrossover requires BinaryIndividual")
	}
	crossoverPointArray := make([]int, 0)
	newI1Genome := make([]byte, 0, len(i.Genome))
	newI2Genome := make([]byte, 0, len(i.Genome))
	for range crossoverPoints {
		crossoverPointArray = append(crossoverPointArray, rand.Intn(len(i.Genome)))
	}
	sort.Ints(crossoverPointArray)
	swap := true
	currentPointIndex := 0
	for j := range len(i.Genome) {
		if currentPointIndex < len(crossoverPointArray) && j >= crossoverPointArray[currentPointIndex] {
			swap = !swap
			currentPointIndex += 1
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
	newI1.CalculateFitness()
	newI2.CalculateFitness()
	return &newI1, &newI2
}
