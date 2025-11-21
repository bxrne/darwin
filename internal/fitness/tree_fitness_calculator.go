package fitness

import "github.com/bxrne/darwin/internal/individual"

type TreeFitnessCalculator struct {
	TestCases     []map[string]float64
	TargetResults []float64
}

func (fitnessCalc *TreeFitnessCalculator) SetupEvalFunction(evalFunction string, variableSet []string, testCaseCount int) {
	testCases, targetResults := SetupEvalFunction(evalFunction, variableSet, testCaseCount)
	fitnessCalc.TestCases = testCases
	fitnessCalc.TargetResults = targetResults
}

func (fitnessCalc *TreeFitnessCalculator) CalculateFitness(evolvable *individual.Evolvable) {
	tree, ok := (*evolvable).(*individual.Tree)
	if !ok {
		panic("Tree fitness Needs tree structure")
	}
	tree.SetFitness(CalculateTreeFitness(tree.Root, fitnessCalc.TargetResults, fitnessCalc.TestCases))

}
