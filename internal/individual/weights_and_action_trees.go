package individual

type WeightsAndActionIndividual struct {
	Weight     *WeightsIndividual
	ActionTree *ActionTreeIndividual
	fitness    float64
}

func (w *WeightsAndActionIndividual) Mutate(rate float64, mutateInformation *MutateInformation) {
	// TODO: implement actual behavior
}

func (w *WeightsAndActionIndividual) Max(i2 Evolvable) Evolvable {
	other, ok := i2.(*WeightsAndActionIndividual)
	if !ok {
		// TODO: handle mixed-type comparison or panic
		panic("WeightsAndActionIndividual not given for Max")
	}

	if other.fitness > w.fitness {
		return other
	}
	return w
}

func (w *WeightsAndActionIndividual) MultiPointCrossover(
	i2 Evolvable,
	crossoverInformation *CrossoverInformation,
) (Evolvable, Evolvable) {

	// TODO: implement crossover logic

	// Return simple clones for now
	return w.Clone(), w.Clone()
}

func (w *WeightsAndActionIndividual) GetFitness() float64 {
	return w.fitness
}

func (w *WeightsAndActionIndividual) SetFitness(f float64) {
	w.fitness = f
}

func (w *WeightsAndActionIndividual) Clone() Evolvable {
	// TODO: implement deep clone
	return &WeightsAndActionIndividual{
		Weight:     w.Weight,     // TODO: clone
		ActionTree: w.ActionTree, // TODO: clone
		fitness:    w.fitness,
	}
}

func (w *WeightsAndActionIndividual) Describe() string {
	// TODO: implement a proper description
	return "WeightsAndActionIndividual"
}
