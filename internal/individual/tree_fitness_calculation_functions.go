package individual

import (
	"math"

	"github.com/Pramod-Devireddy/go-exprtk"
	"github.com/bxrne/darwin/internal/rng"
)

func Round(x float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(x*factor) / factor
}
func CalculateTreeFitness(tree *TreeNode, targetResults []float64, testCases []map[string]float64) float64 {
	actualResults := make([]float64, 0)
	error := 0.0
	dividedByZero := false
	for index, vars := range testCases {

		actualResult, hasDividedByZero := tree.EvaluateTree(&vars)
		dividedByZero = hasDividedByZero
		actualResult = Round(actualResult, 6)
		actualResults = append(actualResults, actualResult)
		error += math.Pow(actualResults[index]-targetResults[index], 2)
	}
	//r error
	if dividedByZero {
		return math.Inf(-20)
	} else {
		return ((math.Sqrt(error/float64(len(actualResults))) + 0.01*float64(tree.CalculateMaxDepth())) * -1)
	}
}

func SetupEvalFunction(evalFunction string, variableSet []string, testCaseCount int) ([]map[string]float64, []float64) {
	exprtkObj := exprtk.NewExprtk()
	exprtkObj.SetExpression(evalFunction)

	// Extract variables from terminal set (exclude numeric constants)
	for _, varName := range variableSet {
		exprtkObj.AddDoubleVariable(varName)
	}

	err := exprtkObj.CompileExpression()
	if err != nil {
		panic("Expression will not compile")
	}

	minVal, maxVal := -5.0, 5.0

	// Generate test cases using only variables from terminal set
	testCases := make([]map[string]float64, testCaseCount)
	targetResults := make([]float64, 0)
	for i := range testCaseCount {
		caseVars := make(map[string]float64)
		for _, varName := range variableSet {
			caseVars[varName] = minVal + rng.Float64()*(maxVal-minVal)
		}
		for name, val := range caseVars {
			exprtkObj.SetDoubleVariableValue(name, val)
		}
		targetResults = append(targetResults, Round(exprtkObj.GetEvaluatedValue(), 6))
		testCases[i] = caseVars

	}
	return testCases, targetResults
}
