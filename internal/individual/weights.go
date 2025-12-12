package individual

import (
	"fmt"
	"sort"

	"github.com/bxrne/darwin/internal/rng"
	"gonum.org/v1/gonum/mat"
)

type WeightsIndividual struct {
	Weights  *mat.Dense
	fitness  float64
	minVal   float64
	maxVal   float64
	clientId string
}

func NewWeightsIndividual(height int, width int) *WeightsIndividual {
	Weights := mat.NewDense(height, width, nil)

	minVal, maxVal := -5.0, 5.0
	for i := range height {
		for j := range width {
			Weights.Set(i, j, minVal+rng.Float64()*(maxVal-minVal))
		}
	}
	return &WeightsIndividual{Weights: Weights, minVal: minVal, maxVal: maxVal}
}

func (wi *WeightsIndividual) Max(i2 Evolvable) Evolvable {
	if wi.fitness > i2.GetFitness() {
		return wi
	}
	return i2
}

func (wi *WeightsIndividual) Describe() string {
	r, c := wi.Weights.Dims()
	description := wi.clientId + ": "
	for i := range r {
		for j := range c {
			description += fmt.Sprintf("%f ", wi.Weights.At(i, j))
		}
		description += "\n"
	}
	description += "Fitness: " + fmt.Sprintf("%f", wi.fitness) + "\n"
	return description
}

func (wi *WeightsIndividual) SetClient(clientId string) {
	wi.clientId = clientId
}

func (wi *WeightsIndividual) GetFitness() float64 {
	return wi.fitness
}

func (wi WeightsIndividual) Clone() Evolvable {

	// Clone Weights
	r, c := wi.Weights.Dims()
	clonedWeights := mat.NewDense(r, c, nil)
	for i := range r {
		for j := range c {
			clonedWeights.Set(i, j, wi.Weights.At(i, j))
		}
	}
	return &WeightsIndividual{
		Weights: clonedWeights,
		fitness: wi.fitness,
		minVal:  wi.minVal,
		maxVal:  wi.maxVal,
	}
}

func (wi *WeightsIndividual) MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable) {
	r1, c1 := wi.Weights.Dims()
	wi2, ok := i2.(*WeightsIndividual)
	if !ok {
		panic("Trying to cross wrong Evolvable with weights individual")
	}
	crossoverPointArray := make([]int, crossoverInformation.CrossoverPoints)

	for range crossoverInformation.CrossoverPoints {
		crossoverPointArray = append(crossoverPointArray, rng.Intn(c1))
	}
	sort.Ints(crossoverPointArray)

	swap := true
	currentPointIndex := 0
	for i := range r1 {
		swap = !swap
		for j := range c1 {
			if currentPointIndex < len(crossoverPointArray) && j >= crossoverPointArray[currentPointIndex] {
				swap = !swap
				currentPointIndex++
			}
			if swap {
				temp := wi.Weights.At(i, j)
				wi.Weights.Set(i, j, wi2.Weights.At(i, j))
				wi2.Weights.Set(i, j, temp)
			}
		}
	}
	return wi, i2
}

func (wi *WeightsIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
	r, c := wi.Weights.Dims()
	for i := range r {
		for j := range c {
			if rng.Float64() < rate {
				// Small modulation rather than wholly new weights
				newVal := wi.Weights.At(i, j) + (wi.minVal+rng.Float64()*(wi.maxVal-wi.minVal))*0.3
				wi.Weights.Set(i, j, float64(newVal))
			}
		}
	}
}

func (wi *WeightsIndividual) SetFitness(fitness float64) {
	wi.fitness = fitness
}

func (wi *WeightsIndividual) GetMetrics() map[string]float64 {
	return map[string]float64{
		"fit": wi.GetFitness(),
	}
}
