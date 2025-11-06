package individual

type FitnessCalculator interface {
	CalculateFitness(evolvable *Evolvable)
}

type FitnessSetupInformation struct {
	EvalFunction   string
	ParameterCount int
	GenomeType     GenomeType
}

func FitnessCalculatorFactory(info FitnessSetupInformation) FitnessCalculator {
	switch info.GenomeType {
	case TreeGenome:
		calc := &TreeFitnessCalculator{}
		calc.setupEvalFunction(info.EvalFunction, info.ParameterCount)
		return calc
	case BitStringGenome:
		return &BinaryFitnessCalculator{}
	default:
		return nil
	}
}
