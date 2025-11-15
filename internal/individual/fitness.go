package individual

import "github.com/bxrne/darwin/internal/cfg"

type FitnessCalculator interface {
	CalculateFitness(evolvable *Evolvable)
}

type FitnessSetupInformation struct {
	EvalFunction  string
	VariableSet   []string
	GenomeType    GenomeType
	TestCaseCount int
	Grammar       map[string]Node
}

func GenerateFitnessInfoFromConfig(config *cfg.Config, genomeType GenomeType, grammar map[string]Node) FitnessSetupInformation {
	fitnessInfo := FitnessSetupInformation{}
	fitnessInfo.EvalFunction = config.Fitness.TargetFunction
	fitnessInfo.GenomeType = genomeType
	fitnessInfo.Grammar = grammar
	fitnessInfo.VariableSet = config.Tree.VariableSet
	fitnessInfo.TestCaseCount = config.Fitness.TestCaseCount
	return fitnessInfo
}

func FitnessCalculatorFactory(info FitnessSetupInformation) FitnessCalculator {
	switch info.GenomeType {
	case TreeGenome:
		calc := &TreeFitnessCalculator{}
		calc.SetupEvalFunction(info.EvalFunction, info.VariableSet, info.TestCaseCount)
		return calc
	case BitStringGenome:
		return &BinaryFitnessCalculator{}
	case GrammarTreeGenome:
		calc := &GrammarTreeFitnessCalculator{Grammar: info.Grammar}
		calc.SetupEvalFunction(info.EvalFunction, info.VariableSet, info.TestCaseCount)
		return calc
	default:
		return nil
	}
}
