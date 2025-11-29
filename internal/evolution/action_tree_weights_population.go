package evolution

import "github.com/bxrne/darwin/internal/individual"

type ActionTreeAndWeightsPopulation struct {
	actionTrees        []*individual.ActionTreeIndividual
	Weights            []*individual.WeightsIndividual
	CombinedPopulation []individual.WeightsAndActionIndividual
	isTrainingWeights  bool
}

func newActionTreeAndWeightsPopulation(size int, creator func() individual.Evolvable) *ActionTreeAndWeightsPopulation {
	actionTreePopulation := make([]individual.ActionTreeIndividual, size)
	weightsPopulation := make([]individual.WeightsIndividual, size)
	combinedPopulation := make([]individual.WeightsAndActionIndividual, size)
	for i := range size {
		trees := creator()
		realTree := trees.(*individual.ActionTreeIndividual)
		actionTreePopulation[i] = *realTree
		weightsPopulation[i] = *individual.NewWeightsIndividual(10, 10)
		combinedPopulation[i] = individual.WeightsAndActionIndividual{ActionTree: &actionTreePopulation[i], Weight: &weightsPopulation[i]}

	}

	actionTreePtrs := make([]*individual.ActionTreeIndividual, size)
	weightsPtrs := make([]*individual.WeightsIndividual, size)

	for i := 0; i < size; i++ {
		actionTreePtrs[i] = &actionTreePopulation[i]
		weightsPtrs[i] = &weightsPopulation[i]
	}

	return &ActionTreeAndWeightsPopulation{actionTrees: actionTreePtrs,
		Weights: weightsPtrs, CombinedPopulation: combinedPopulation}
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

		weights := make([]*individual.WeightsIndividual, len(population))

		for i, e := range population {
			wi, ok := e.(*individual.WeightsIndividual)
			if !ok {
				// element is not a WeightsIndividual
				panic("population is not a WeightsIndividual")
			}
			weights[i] = wi
		}
		at.Weights = weights
		return
	}
	actionTrees := make([]*individual.ActionTreeIndividual, len(population))

	for i, e := range population {
		wi, ok := e.(*individual.ActionTreeIndividual)
		if !ok {
			// element is not a l
			panic("population is not a ActiontreeIndividual")
		}
		actionTrees[i] = wi
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
