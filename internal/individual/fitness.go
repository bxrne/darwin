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
	EvalFunction   string
	ParameterCount int
	GenomeType     GenomeType
}

type BinaryFitnessCalculator struct{}

func (binaryFitness *BinaryFitnessCalculator) CalculateFitness(evolvable *Evolvable) {
	binaryIndividual, ok := (*evolvable).(*BinaryIndividual)
	if !ok {
		panic("Binary Indiviual Fitness Needs Binary INdividual")
	}
	count := 0
	for _, gene := range binaryIndividual.Genome {
		if gene == '1' {
			count++
		}
	}
	binaryIndividual.SetFitness(float64(count) / float64(len(binaryIndividual.Genome)))

}

type TreeFitnessCalculator struct {
	EvalFunction   exprtk.GoExprtk
	ParameterCount int
	TestCases      []map[string]float64
}

func FitnessCalculatorFactory(info FitnessSetupInformation) FitnessCalculator {
	switch info.GenomeType {
	case TreeGenome:
		calc := &TreeFitnessCalculator{}
		calc.setupEvalFunction(info.EvalFunction, info.ParameterCount)
		return calc
	case BitStringGenome:
		return &BinaryFitnessCalculator{}
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
	fitnessCalc.TestCases = testCases

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
