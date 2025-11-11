package individual

type FitnessCalculator interface {
	CalculateFitness(evolvable *Evolvable)
}

type FitnessSetupInformation struct {
	EvalFunction  string
	TerminalSet   []string
	GenomeType    GenomeType
	TestCaseCount int
}

func FitnessCalculatorFactory(info FitnessSetupInformation) FitnessCalculator {
	switch info.GenomeType {
	case TreeGenome:
		calc := &TreeFitnessCalculator{}
		calc.SetupEvalFunction(info.EvalFunction, info.TerminalSet)
		return calc
	case BitStringGenome:
		return &BinaryFitnessCalculator{}
	default:
		return nil
	}
}
