package individual

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

type WeightsIndividual struct {
	Weights *mat.Dense
	fitness float64
}

func NewWeightsIndividual(height int, width int) *WeightsIndividual {
	Weights := mat.NewDense(height, width, nil)
	for i := range height {
		for j := range make([]int, width) {
			Weights.Set(i, j, 0.0)
		}
	}
	return &WeightsIndividual{Weights: Weights}
}

func (wi *WeightsIndividual) Max(i2 Evolvable) Evolvable {
	if wi.fitness > i2.GetFitness() {
		return wi
	}
	return i2
}

func (wi *WeightsIndividual) Describe() string {
	r, c := wi.Weights.Dims()
	description := ""
	for i := range r {
		for j := range c {
			description += fmt.Sprintf("%f ", wi.Weights.At(i, j))
		}
		description += "\n"
	}
	description += "Fitness: " + fmt.Sprintf("%f", wi.fitness) + "\n"
	return description
}

func (wi *WeightsIndividual) GetFitness() float64 {
	return wi.fitness
}

func (wi *WeightsIndividual) Clone() Evolvable {

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
	}
}

func (wi *WeightsIndividual) MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable) {
	return wi, i2
}

func (wi *WeightsIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
}

func (wi *WeightsIndividual) SetFitness(fitness float64) {
	wi.fitness = fitness
}
