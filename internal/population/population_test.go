package population_test

import (
	"fmt"
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/population"
	"github.com/stretchr/testify/assert"
)

// --- TEST DOUBLE TYPES --------------------------------------------------------

// mock fitness calculator
type mockFitnessCalc struct{}

func (f *mockFitnessCalc) CalculateFitness(e individual.Evolvable) {
	// Assign deterministic fitness for tests
	(e).SetFitness(42.0)
}

func (f *mockFitnessCalc) SetupEvalFunction(expr string, vars []string, c int) {}

// --- PARAMETERIZED TEST --------------------------------------------------------

func TestPopulationBuilder_BuildPopulation_GIVEN_various_genome_types_WHEN_build_THEN_populations_correct(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		genomeType individual.GenomeType
		initFunc   func() individual.Evolvable
	}{
		{
			name:       "GenericGenome small",
			size:       5,
			genomeType: individual.BitStringGenome,
			initFunc:   func() individual.Evolvable { return individual.NewBinaryIndividual(5) },
		},
		{
			name:       "GenericGenome large",
			size:       200,
			genomeType: individual.TreeGenome,
			initFunc: func() individual.Evolvable {
				return individual.NewRandomTree(5, []string{"+", "-", "*"}, []string{"a"}, []string{"1.0"})
			},
		},
		{
			name:       "ActionTree genome type triggers ActionTreePopulation",
			size:       10,
			genomeType: individual.ActionTreeGenome,
			initFunc: func() individual.Evolvable {
				return individual.NewActionTreeIndividual(
					[]string{"move", "jump", "turn"}, // actions
					map[string]*individual.Tree{ // initialTrees
						"move": individual.NewRandomTree(3,
							[]string{"+", "-", "*"},
							[]string{"x", "y"},
							[]string{"1", "2"},
						),
						"jump": individual.NewRandomTree(3,
							[]string{"+", "*"},
							[]string{"x"},
							[]string{"1"},
						),
						"turn": individual.NewRandomTree(3,
							[]string{"-"},
							[]string{"x"},
							[]string{"1"},
						),
					},
				)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // captured for parallel runs
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pb := population.NewPopulationBuilder()

			fit := &mockFitnessCalc{}
			popInfo := &population.PopulationInfo{Size: tt.size, GenomeType: tt.genomeType}
			pop := pb.BuildPopulation(popInfo, tt.initFunc)
			pop.CalculateFitnesses(fit)
			// -- Validate population exists --
			assert.NotNil(t, pop, "Returned population should not be nil")
			fmt.Println(pop.GetPopulation())
			assert.Equal(t, tt.size, pop.Count(), "Population size mismatch")

			// -- Validate type correctness --
			if tt.genomeType == individual.ActionTreeGenome {
				_, ok := pop.(*population.ActionTreeAndWeightsPopulation)
				assert.True(t, ok, "Expected ActionTreeAndWeightsPopulation for ActionTreeGenome")
				return // Other checks don't apply to action-tree path
			} else {
				_, ok := pop.(*population.GenericPopulation)
				assert.True(t, ok, "Expected GenericPopulation for GenericGenome")
			}

			// -- Validate all individuals created --
			all := pop.GetPopulation()
			assert.Len(t, all, tt.size)

			for i, ind := range all {
				assert.NotNil(t, ind, "individual at index %d is nil", i)

				assert.Equal(t, float64(42), ind.GetFitness(), "Fitness incorrectly assigned")
			}
		})
	}
}
