package garden

import (
	"fmt"
	"math/rand"
	"sort"
)

type Individual struct {
	Genome  string
	Fitness int
}

func makeIndividual(genomeSize int) Individual {
	individual := Individual{Genome: string(func() []byte {
		b := make([]byte, genomeSize)
		for i := range b {
			b[i] = '0' + byte(rand.Intn(2))
		}
		return b
	}())}
	individual.CalculateFitness()
	return individual

}

func (i *Individual) CalculateFitness() {
	count := 0
	for _, gene := range i.Genome {
		if gene == '1' {
			count++
		}
	}
	i.Fitness = count
}

func (i *Individual) Max(i2 Individual) Individual {
	if i.Fitness > i2.Fitness {
		return *i
	}
	return i2
}

func (i *Individual) Mutate(points []int, mutatationRate float64) {
	if mutatationRate < rand.Float64() {
		return
	}
	for _, point := range points {
		if i.Genome[point] == '1' {
			i.Genome = i.Genome[:point] + "0" + i.Genome[point+1:]
		} else {
			i.Genome = i.Genome[:point] + "1" + i.Genome[point+1:]
		}
	}
}

func (i *Individual) MultiPointCrossover(i2 Individual, crossoverPoints int) (Individual, Individual) {
	crossoverPointArray := make([]int, 0)
	newI1 := Individual{Genome: ""}
	newI2 := Individual{Genome: ""}
	for range crossoverPoints {
		crossoverPointArray = append(crossoverPointArray, rand.Intn(len(i.Genome)))
	}
	sort.Ints(crossoverPointArray)
	swap := true
	currentPointIndex := 0
	for j := range len(i.Genome) {
		if currentPointIndex < len(crossoverPointArray) && j >= crossoverPointArray[currentPointIndex] {
			swap = false
			currentPointIndex += 1
		}
		if swap {
			newI1.Genome += string(i.Genome[j])
			newI2.Genome += string(i2.Genome[j])
		} else {
			newI1.Genome += string(i2.Genome[j])
			newI2.Genome += string(i.Genome[j])
		}
	}
	newI1.CalculateFitness()
	newI2.CalculateFitness()
	return newI1, newI2
}
