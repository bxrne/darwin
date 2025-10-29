package garden

import (
	"math/rand"
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

func (i *Individual) SinglePointCrossover(i2 Individual, crossoverRate float64) Individual {
	if crossoverRate < rand.Float64() {
		return *i
	}
	index := rand.Intn(len(i.Genome))
	newI := Individual{Genome: ""}
	for j := range len(i.Genome) {
		if j < index {
			newI.Genome += string(i.Genome[j])
		} else {
			newI.Genome += string(i2.Genome[j])
		}
	}
	return newI
}
