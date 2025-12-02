package fitness

import (
	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/individual"
)

type FitnessCalculator interface {
	CalculateFitness(evolvable individual.Evolvable)
}

type FitnessSetupInformation struct {
	EvalFunction                  string
	VariableSet                   []string
	GenomeType                    individual.GenomeType
	TestCaseCount                 int
	Grammar                       map[string]individual.Node
	ServerAddr                    string
	OpponentType                  string
	MaxSteps                      int
	Actions                       []string
	Population                    []*[]individual.Evolvable
	ActionTreeSelectionPercentage float64
}

func GenerateFitnessInfoFromConfig(config *cfg.Config, genomeType individual.GenomeType, grammar map[string]individual.Node, populations []*[]individual.Evolvable) FitnessSetupInformation {
	fitnessInfo := FitnessSetupInformation{}
	fitnessInfo.EvalFunction = config.Fitness.TargetFunction
	fitnessInfo.GenomeType = genomeType
	fitnessInfo.Grammar = grammar
	fitnessInfo.VariableSet = config.Tree.VariableSet
	fitnessInfo.TestCaseCount = config.Fitness.TestCaseCount

	// Add ActionTree specific config
	if genomeType == individual.ActionTreeGenome {
		fitnessInfo.ServerAddr = config.ActionTree.ServerAddr
		fitnessInfo.OpponentType = config.ActionTree.OpponentType
		fitnessInfo.MaxSteps = config.ActionTree.MaxSteps
		fitnessInfo.Actions = config.ActionTree.Actions
		fitnessInfo.Population = populations
		fitnessInfo.ActionTreeSelectionPercentage = 0.3
	}

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
	case individual.ActionTreeGenome:
		calc := NewActionTreeFitnessCalculator(
			info.ServerAddr,
			info.OpponentType,
			info.Actions,
			info.MaxSteps,
			info.Population,
			info.ActionTreeSelectionPercentage,
		)
		return calc
	default:
		return nil
	}
}
