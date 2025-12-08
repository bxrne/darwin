package population

import (
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type GenericPopulation struct {
	population []individual.Evolvable
	count      int
}

func newGenericPopulation(populationCount int) *GenericPopulation {
	return &GenericPopulation{
		count: populationCount}
}

func (gp *GenericPopulation) Get(index int) individual.Evolvable {
	return gp.population[index]
}

func (gp *GenericPopulation) Count() int {
	return gp.count
}

func (gp *GenericPopulation) Update(generation int, fitnessCalc fitness.FitnessCalculator) {
}

func (gp *GenericPopulation) SetPopulation(population []individual.Evolvable) {
	gp.population = population
}

func (gp *GenericPopulation) GetPopulation() []individual.Evolvable {
	return gp.population
}

func (gp *GenericPopulation) GetPopulations() []*[]individual.Evolvable {
	return nil
}

func (gp *GenericPopulation) CalculateFitnesses(fitnessCalc fitness.FitnessCalculator) {
	for _, ind := range gp.population {
		fitnessCalc.CalculateFitness(ind)
	}
}
