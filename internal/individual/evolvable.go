package individual

// Evolvable represents an individual that can evolve through genetic operations
type Evolvable interface {
	Mutate(rate float64, mutateInformation *MutateInformation)
	Max(i2 Evolvable) Evolvable
	MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable)
	GetFitness() float64
	SetFitness(fitness float64)
	Clone() Evolvable
	Describe() string
}

type GenomeType int

const (
	BitStringGenome GenomeType = iota
	TreeGenome
	GrammarTreeGenome
)

type CrossoverInformation struct {
	CrossoverPoints int
	MaxDepth        int
}

type MutateInformation struct {
	VariableSet []string
	TerminalSet []string
	OperandSet  []string
}
