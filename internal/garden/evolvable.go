package garden

type Evolvable interface {
	CalculateFitness()
	Mutate(points []int, rate float64)
	GetFitness() float64
	Max(i2 Evolvable) Evolvable
	MultiPointCrossover(i2 Evolvable, crossoverPoints int) (Evolvable, Evolvable)
}
