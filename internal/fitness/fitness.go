package fitness

import (
	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/individual"
)

type FitnessCalculator interface {
	CalculateFitness(evolvable *individual.Evolvable)
}

type FitnessSetupInformation struct {
	EvalFunction  string
	VariableSet   []string
	GenomeType    individual.GenomeType
	TestCaseCount int
	Grammar       map[string]individual.Node
}

func GenerateFitnessInfoFromConfig(config *cfg.Config, genomeType individual.GenomeType, grammar map[string]individual.Node) FitnessSetupInformation {
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
	case individual.TreeGenome:
		calc := &TreeFitnessCalculator{}
		calc.SetupEvalFunction(info.EvalFunction, info.VariableSet, info.TestCaseCount)
		return calc
	case individual.BitStringGenome:
		return &BinaryFitnessCalculator{}
	case individual.GrammarTreeGenome:
		calc := &GrammarTreeFitnessCalculator{Grammar: info.Grammar}
		calc.SetupEvalFunction(info.EvalFunction, info.VariableSet, info.TestCaseCount)
		return calc
	default:
		return nil
	}
}
