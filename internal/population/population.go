package population

import (
	"runtime"
	"sync"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type Population interface {
	Get(index int) individual.Evolvable
	Count() int
	Update(generation int)
	SetPopulation(Population []individual.Evolvable)
	GetPopulation() []individual.Evolvable
	GetPopulations() []*[]individual.Evolvable
	CalculateFitnesses(fitnessCalc fitness.FitnessCalculator)
}

type PopulationInfo struct {
	Size                 int
	weightsCount         int
	maxNumInputs         int
	numColumns           int
	trainWeightsFirst    bool
	GenomeType           individual.GenomeType
	SwitchPopulationStep int
}

func NewPopulationInfo(config *cfg.Config, genomeType individual.GenomeType) PopulationInfo {
	maxValue := config.ActionTree.Actions[0].Value // start with first value
	for _, a := range config.ActionTree.Actions[1:] {
		if a.Value > maxValue {
			maxValue = a.Value
		}
	}
	return PopulationInfo{
		Size:                 config.Evolution.PopulationSize,
		weightsCount:         config.ActionTree.WeightsCount,
		numColumns:           config.ActionTree.WeightsColumnCount,
		trainWeightsFirst:    config.ActionTree.TrainWeightsFirst,
		SwitchPopulationStep: config.ActionTree.SwitchTrainingTargetStep,
		maxNumInputs:         maxValue,
		GenomeType:           genomeType,
	}
}

// PopulationBuilder creates initial populations
type PopulationBuilder struct{}

// NewPopulationBuilder creates a new population builder
func NewPopulationBuilder() *PopulationBuilder {
	return &PopulationBuilder{}
}

// BuildPopulation creates a population of binary individuals
func (pb *PopulationBuilder) BuildPopulation(popInfo *PopulationInfo, creator func() individual.Evolvable) Population {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()

	chunkSize := (popInfo.Size + numWorkers - 1) / numWorkers
	switch popInfo.GenomeType {
	case individual.ActionTreeGenome:
		return NewActionTreeAndWeightsPopulation(popInfo, creator)
	default:
		population := make([]individual.Evolvable, popInfo.Size)
		for i := range numWorkers {
			start := i * chunkSize
			end := start + chunkSize
			end = min(end, popInfo.Size)

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for j := start; j < end; j++ {
					population[j] = creator()
				}
			}(start, end) // chunk to use
		}

		wg.Wait()
		newPop := newGenericPopulation(popInfo.Size)
		newPop.SetPopulation(population)
		return newPop
	}
}
