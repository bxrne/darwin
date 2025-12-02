package population

import (
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type ActionTreeAndWeightsPopulation struct {
	actionTrees        []individual.Evolvable
	Weights            []individual.Evolvable
	CombinedPopulation []individual.WeightsAndActionIndividual
	isTrainingWeights  bool
}

func NewActionTreeAndWeightsPopulation(size int, creator func() individual.Evolvable) *ActionTreeAndWeightsPopulation {
	actionTreePopulation := make([]individual.Evolvable, size)
	weightsPopulation := make([]individual.Evolvable, size)
	combinedPopulation := make([]individual.WeightsAndActionIndividual, size)
	for i := range size {
		trees := creator()
		realTree := trees.(*individual.ActionTreeIndividual)
		actionTreePopulation[i] = realTree
		weightsPopulation[i] = individual.NewWeightsIndividual(10, 10)

	}

	return &ActionTreeAndWeightsPopulation{actionTrees: actionTreePopulation,
		Weights: weightsPopulation, CombinedPopulation: combinedPopulation}
}

func (at *ActionTreeAndWeightsPopulation) Get(index int) individual.Evolvable {
	if at.isTrainingWeights {
		return at.Weights[index]
	}
	return at.actionTrees[index]
}

func (at *ActionTreeAndWeightsPopulation) Update(generation int) {
}

func (at *ActionTreeAndWeightsPopulation) SetPopulation(population []individual.Evolvable) {
	if at.isTrainingWeights {

		_, ok := population[0].(*individual.WeightsIndividual)
		if !ok {
			// element is not a WeightsIndividual
			panic("population is not a WeightsIndividual")
		}
		at.Weights = population
		return
	}
	actionTrees := make([]individual.Evolvable, len(population))
	_, ok := population[0].(*individual.ActionTreeIndividual)
	if !ok {
		// element is not a l
		panic("population is not a ActiontreeIndividual")
	}
	at.actionTrees = actionTrees
}

func (at *ActionTreeAndWeightsPopulation) Count() int {
	if at.isTrainingWeights {
		return len(at.Weights)
	}
	return len(at.actionTrees)
}

func (at *ActionTreeAndWeightsPopulation) GetPopulation() []individual.Evolvable {
	result := make([]individual.Evolvable, len(at.CombinedPopulation))

	for i := range at.CombinedPopulation {
		result[i] = &at.CombinedPopulation[i] // MUST be pointer!
	}

	return result
}

func (at *ActionTreeAndWeightsPopulation) GetPopulations() []*[]individual.Evolvable {
	population := make([]*[]individual.Evolvable, 2)
	population = append(population, &at.Weights)
	population = append(population, &at.actionTrees)

	return population
}

func (at *ActionTreeAndWeightsPopulation) CalculateFitnesses(fitnessCalc fitness.FitnessCalculator) {
	for _, ind := range at.actionTrees {
		fitnessCalc.CalculateFitness(ind)
	}

	for _, ind := range at.Weights {
		fitnessCalc.CalculateFitness(ind)
	}
}
