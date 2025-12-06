package population

import (
	"runtime"
	"sync"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type ActionTreeAndWeightsPopulation struct {
	actionTrees       []individual.Evolvable
	Weights           []individual.Evolvable
	isTrainingWeights bool
}

func NewActionTreeAndWeightsPopulation(populationInfo *PopulationInfo, creator func() individual.Evolvable) *ActionTreeAndWeightsPopulation {
	return NewActionTreeAndWeightsPopulationWithWeightRange(populationInfo, creator, -5.0, 5.0, false)
}

// NewActionTreeAndWeightsPopulationWithWeightRange creates population with configurable weight initialization
func NewActionTreeAndWeightsPopulationWithWeightRange(populationInfo *PopulationInfo, creator func() individual.Evolvable, weightsMinVal float64, weightsMaxVal float64, useRampedRange bool) *ActionTreeAndWeightsPopulation {
	actionTreePopulation := make([]individual.Evolvable, populationInfo.Size)
	weightsPopulation := make([]individual.Evolvable, populationInfo.weightsCount)
	for i := range populationInfo.Size {
		actionTreePopulation[i] = creator()
	}

	// Initialize weights with ramped ranges if enabled
	if useRampedRange {
		// Distribute weights across different ranges (ramped)
		// Create ranges from small to large
		numRanges := populationInfo.weightsCount
		if numRanges > 10 {
			numRanges = 10 // Limit to 10 different ranges
		}
		
		rangeStep := (weightsMaxVal - weightsMinVal) / float64(numRanges)
		weightsPerRange := populationInfo.weightsCount / numRanges
		remaining := populationInfo.weightsCount % numRanges
		
		weightIndex := 0
		for r := 0; r < numRanges; r++ {
			// Calculate range for this group
			rangeMin := weightsMinVal + float64(r)*rangeStep
			rangeMax := weightsMinVal + float64(r+1)*rangeStep
			
			count := weightsPerRange
			if r < remaining {
				count++
			}
			
			for i := 0; i < count && weightIndex < populationInfo.weightsCount; i++ {
				weightsPopulation[weightIndex] = individual.NewWeightsIndividualWithRange(
					populationInfo.maxNumInputs,
					populationInfo.numColumns,
					rangeMin,
					rangeMax,
				)
				weightIndex++
			}
		}
	} else {
		// Use same range for all weights
		for i := range populationInfo.weightsCount {
			weightsPopulation[i] = individual.NewWeightsIndividualWithRange(
				populationInfo.maxNumInputs,
				populationInfo.numColumns,
				weightsMinVal,
				weightsMaxVal,
			)
		}
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
