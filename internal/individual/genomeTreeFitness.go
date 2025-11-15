package individual

type GrammarTreeFitnessCalculator struct {
	TestCases     []map[string]float64
	TargetResults []float64
	Grammar       map[string]Node
}

func (gtreeCalculator *GrammarTreeFitnessCalculator) CalculateFitness(evolvable *Evolvable) {
	tree, ok := (*evolvable).(*GrammarTree)
	if !ok {
		panic("Tree fitness Needs tree structure")
	}
	tree.Root = GenerateTreeFromGenome(gtreeCalculator.Grammar, tree.Genome)
	tree.Depth = tree.Root.CalculateMaxDepth()
	tree.SetFitness(CalculateTreeFitness(tree.Root, gtreeCalculator.TargetResults, gtreeCalculator.TestCases))

}

func (fitnessCalc *GrammarTreeFitnessCalculator) SetupEvalFunction(evalFunction string, variableSet []string, testCaseCount int) {
	testCases, targetResults := SetupEvalFunction(evalFunction, variableSet, testCaseCount)
	fitnessCalc.TestCases = testCases
	fitnessCalc.TargetResults = targetResults
}
