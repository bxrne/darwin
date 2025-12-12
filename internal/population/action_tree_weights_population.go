package population

import (
	"runtime"
	"sync"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type ActionTreeAndWeightsPopulation struct {
	actionTrees          []individual.Evolvable
	Weights              []individual.Evolvable
	isTrainingWeights    bool
	switchPopulationStep int
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
		Weights: weightsPopulation, isTrainingWeights: populationInfo.trainWeightsFirst, switchPopulationStep: populationInfo.SwitchPopulationStep}
}

func (at *ActionTreeAndWeightsPopulation) Get(index int) individual.Evolvable {
	if at.isTrainingWeights {
		return at.Weights[index]
	}
	return at.actionTrees[index]
}

func (at *ActionTreeAndWeightsPopulation) Update(generation int, fitnessCalc fitness.FitnessCalculator) {
	if generation%at.switchPopulationStep == 0 {
		at.isTrainingWeights = !at.isTrainingWeights
		at.CalculateFitnesses(fitnessCalc)
	}

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
	_, ok := population[0].(*individual.ActionTreeIndividual)
	if !ok {
		// element is not a l
		panic("population is not a ActiontreeIndividual")
	}
	at.actionTrees = population
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
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	if !at.isTrainingWeights {

		treeChunkSize := (len(at.actionTrees) + numWorkers - 1) / numWorkers
		for i := range numWorkers {
			start := i * treeChunkSize
			end := start + treeChunkSize
			end = min(end, len(at.actionTrees))

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for j := start; j < end; j++ {
					fitnessCalc.CalculateFitness(at.actionTrees[j])
					at.actionTrees[j].Describe()
				}
			}(start, end) // chunk to use
		}
	} else {
		weightChunkSize := (len(at.Weights) + numWorkers - 1) / numWorkers
		for i := range numWorkers {
			start := i * weightChunkSize
			end := start + weightChunkSize
			end = min(end, len(at.Weights))

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for j := start; j < end; j++ {
					fitnessCalc.CalculateFitness(at.Weights[j])
				}
			}(start, end) // chunk to use
		}
	}
	wg.Wait()
}
