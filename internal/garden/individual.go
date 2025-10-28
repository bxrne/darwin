package garden

type Individual struct {
	Genome  string
	Fitness int
}

func (i *Individual) CalculateFitness() {
	count := 0
	for _, gene := range i.Genome {
		if gene == '1' {
			count++
		}
	}
	i.Fitness = count
}

func (i *Individual) Mutate(points []int) {
	for _, point := range points {
		if i.Genome[point] == '1' {
			i.Genome = i.Genome[:point] + "0" + i.Genome[point+1:]
		} else {
			i.Genome = i.Genome[:point] + "1" + i.Genome[point+1:]
		}
	}
}
