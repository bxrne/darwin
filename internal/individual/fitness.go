package individual

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Pramod-Devireddy/go-exprtk"
)

type FitnessCalculator interface {
	CalculateFitness(evolvable *Evolvable)
}

type FitnessSetupInformation struct {
	evalFunction   string
	parameterCount int
}

type TreeFitnessCalculator struct {
	EvalFunction   exprtk.GoExprtk
	ParameterCount int
	TestCases      []map[string]float64
}

func fitnessCalculatorFactory(evolvable Evolvable, info FitnessSetupInformation) FitnessCalculator {
	switch evolvable.(type) {
	case *BinaryIndividual:
		calc := &TreeFitnessCalculator{}
		calc.setupEvalFunction(info.evalFunction, info.parameterCount)
		return calc
	default:
		return nil
	}
}
func (fitnessCalc *TreeFitnessCalculator) setupEvalFunction(evalFunction string, parameterCount int) {
	exprtkObj := exprtk.NewExprtk()
	exprtkObj.SetExpression(evalFunction)

	runes := []rune("xyzabcdefghijklmnopqrstuvw")

	for i := 0; i < parameterCount && i < len(runes); i++ {
		varName := string(runes[i])
		exprtkObj.AddDoubleVariable(varName)
	}

	err := exprtkObj.CompileExpression()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(exprtkObj.GetEvaluatedValue())
	fitnessCalc.EvalFunction = exprtkObj
	fitnessCalc.ParameterCount = parameterCount

	numCases := 10

	minVal, maxVal := -5.0, 5.0

	// Generate test cases
	testCases := make([]map[string]float64, numCases)
	for i := range numCases {
		caseVars := make(map[string]float64)
		for i := 0; i < parameterCount && i < len(runes); i++ {
			caseVars[string(runes[i])] = minVal + rand.Float64()*(maxVal-minVal)
		}
		testCases[i] = caseVars
	}

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
	tree.SetFitness(math.Sqrt(error / float64(len(actualResults))))
}
