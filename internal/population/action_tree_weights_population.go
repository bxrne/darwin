package population

import (
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type ActionTreeAndWeightsPopulation struct {
	actionTrees       []individual.Evolvable
	Weights           []individual.Evolvable
	isTrainingWeights bool
}

func NewActionTreeAndWeightsPopulation(populationInfo *PopulationInfo, creator func() individual.Evolvable) *ActionTreeAndWeightsPopulation {
	actionTreePopulation := make([]individual.Evolvable, populationInfo.Size)
	weightsPopulation := make([]individual.Evolvable, populationInfo.weightsCount)
	for i := range populationInfo.Size {
		actionTreePopulation[i] = creator()
	}

	for i := range populationInfo.weightsCount {
		weightsPopulation[i] = individual.NewWeightsIndividual(populationInfo.maxNumInputs, populationInfo.numColumns)
	}

	return &ActionTreeAndWeightsPopulation{actionTrees: actionTreePopulation,
		Weights: weightsPopulation, isTrainingWeights: populationInfo.trainWeightsFirst}
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
	if at.isTrainingWeights {
		return at.Weights
	}
	return at.actionTrees

}

func (at *ActionTreeAndWeightsPopulation) GetPopulations() []*[]individual.Evolvable {
	population := make([]*[]individual.Evolvable, 2)
	population[0] = &at.Weights
	population[1] = &at.actionTrees

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
