package fitness

import "github.com/bxrne/darwin/internal/individual"

type GrammarTreeFitnessCalculator struct {
	TestCases     []map[string]float64
	TargetResults []float64
	Grammar       map[string]individual.Node
}

func (gtreeCalculator *GrammarTreeFitnessCalculator) CalculateFitness(evolvable individual.Evolvable) {
	tree, ok := evolvable.(*individual.GrammarTree)
	if !ok {
		panic("Tree fitness Needs tree structure")
	}
	tree.Root = individual.GenerateTreeFromGenome(gtreeCalculator.Grammar, tree.Genome)
	tree.Depth = tree.Root.CalculateMaxDepth()
	tree.SetFitness(CalculateTreeFitness(tree.Root, gtreeCalculator.TargetResults, gtreeCalculator.TestCases))

}

func (fitnessCalc *GrammarTreeFitnessCalculator) SetupEvalFunction(evalFunction string, variableSet []string, testCaseCount int) {
	testCases, targetResults := SetupEvalFunction(evalFunction, variableSet, testCaseCount)
	fitnessCalc.TestCases = testCases
	fitnessCalc.TargetResults = targetResults
}
