package selection

import (
	"math"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/rng"
)

// Selector defines the interface for selection strategies
type Selector interface {
	Select(population []individual.Evolvable) individual.Evolvable
}

// RouletteSelector implements roulette wheel selection
type RouletteSelector struct {
	SampleSize int
}

// NewRouletteSelector creates a new roulette selector
func NewRouletteSelector(sampleSize int) *RouletteSelector {
	return &RouletteSelector{SampleSize: sampleSize}
}

// Select performs roulette wheel selection
func (rs *RouletteSelector) Select(population []individual.Evolvable) individual.Evolvable {
	rouletteTable := make([]individual.Evolvable, 0, rs.SampleSize)
	total := 0.0

	for range rs.SampleSize {
		randIndex := rng.Intn(len(population))
		rouletteTable = append(rouletteTable, population[randIndex])
		total += math.Abs(population[randIndex].GetFitness())
	}

	runningTotal := 0.0
	randomValue := rng.Float64() * total
	for i := range rs.SampleSize {
		runningTotal += math.Abs(rouletteTable[i].GetFitness())
		if runningTotal > randomValue {
			return rouletteTable[i]
		}
	}
	return rouletteTable[len(rouletteTable)-1]
}
