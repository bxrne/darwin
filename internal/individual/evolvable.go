package individual

// Evolvable represents an individual that can evolve through genetic operations
type Evolvable interface {
	Mutate(rate float64)
	Max(i2 Evolvable) Evolvable
	MultiPointCrossover(i2 Evolvable, crossoverPointCount int) (Evolvable, Evolvable)
	GetFitness() float64
	SetFitness(fitness float64)
	Clone() Evolvable
	Describe() string
}

type GenomeType int

const (
	BitStringGenome GenomeType = iota
	TreeGenome
)
