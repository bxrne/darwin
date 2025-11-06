package individual

type BinaryFitnessCalculator struct{}

func (binaryFitness *BinaryFitnessCalculator) CalculateFitness(evolvable *Evolvable) {
	binaryIndividual, ok := (*evolvable).(*BinaryIndividual)
	if !ok {
		panic("Binary Indiviual Fitness Needs Binary INdividual")
	}
	count := 0
	for _, gene := range binaryIndividual.Genome {
		if gene == '1' {
			count++
		}
	}
	binaryIndividual.SetFitness(float64(count) / float64(len(binaryIndividual.Genome)))

}
