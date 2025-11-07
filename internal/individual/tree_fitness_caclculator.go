package individual

import (
	"github.com/Pramod-Devireddy/go-exprtk"
	"math"
	"math/rand"
	"strconv"
)

type TreeFitnessCalculator struct {
	EvalFunction exprtk.GoExprtk
	TerminalSet  []string
	TestCases    []map[string]float64
}

func (fitnessCalc *TreeFitnessCalculator) SetupEvalFunction(evalFunction string, terminalSet []string) {
	exprtkObj := exprtk.NewExprtk()
	exprtkObj.SetExpression(evalFunction)

	// Extract variables from terminal set (exclude numeric constants)
	variables := extractVariables(terminalSet)
	for _, varName := range variables {
		exprtkObj.AddDoubleVariable(varName)
	}

	err := exprtkObj.CompileExpression()
	if err != nil {
		panic("Expression will not compile")
	}

	fitnessCalc.EvalFunction = exprtkObj
	fitnessCalc.TerminalSet = terminalSet

	numCases := 10
	minVal, maxVal := -5.0, 5.0

	// Generate test cases using only variables from terminal set
	testCases := make([]map[string]float64, numCases)
	for i := range numCases {
		caseVars := make(map[string]float64)
		for _, varName := range variables {
			caseVars[varName] = minVal + rand.Float64()*(maxVal-minVal)
		}
		testCases[i] = caseVars
	}
	fitnessCalc.TestCases = testCases
}

// extractVariables filters terminal set to return only variable names (excluding numeric constants)
func extractVariables(terminalSet []string) []string {
	var variables []string
	for _, terminal := range terminalSet {
		if isVariable(terminal) {
			variables = append(variables, terminal)
		}
	}
	return variables
}

// isVariable checks if a string is a variable (not a numeric constant)
func isVariable(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err != nil // If it can't be parsed as float, it's a variable
}

func (fitnessCalc *TreeFitnessCalculator) CalculateFitness(evolvable *Evolvable) {
	tree, ok := (*evolvable).(*Tree)
	if !ok {
		panic("Tree fitness Needs tree structure")
	}
	targetResults := make([]float64, 0, 10)
	actualResults := make([]float64, 0, 10)
	error := 0.0
	for index, vars := range fitnessCalc.TestCases {

		for name, val := range vars {
			fitnessCalc.EvalFunction.SetDoubleVariableValue(name, val)
		}
		targetResults = append(targetResults, fitnessCalc.EvalFunction.GetEvaluatedValue())
		actualResults = append(actualResults, tree.EvaluateTree(&vars))
		error += math.Pow(actualResults[index]-targetResults[index], 2)
	}
	//root mean sqr error
	tree.SetFitness(math.Sqrt(error/float64(len(actualResults))) * -1)
}
