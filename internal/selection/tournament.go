package selection

import (
	"math/rand"

	"github.com/bxrne/darwin/internal/individual"
)

// TournamentSelector implements tournament selection
type TournamentSelector struct {
	TournamentSize int
}

// NewTournamentSelector creates a new tournament selector
func NewTournamentSelector(tournamentSize int) *TournamentSelector {
	return &TournamentSelector{TournamentSize: tournamentSize}
}

// Select performs tournament selection
func (ts *TournamentSelector) Select(population []individual.Evolvable) individual.Evolvable {
	tournamentPop := make([]individual.Evolvable, 0, ts.TournamentSize)
	for range ts.TournamentSize {
		randIndex := rand.Intn(len(population))
		tournamentPop = append(tournamentPop, population[randIndex])
	}

	max := tournamentPop[0]
	for _, ind := range tournamentPop[1:] {
		max = ind.Max(max)
	}
	return max
}
