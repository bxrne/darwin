package evolution

import "github.com/bxrne/darwin/internal/individual"

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

func (gp *GenericPopulation) Update(generation int) {
	return
}

func (gp *GenericPopulation) SetPopulation(population []individual.Evolvable) {
	gp.population = population
}

func (gp *GenericPopulation) GetPopulation() []individual.Evolvable {
	return gp.population
}
