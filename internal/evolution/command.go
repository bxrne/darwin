package evolution

// EvolutionCommand represents commands sent to the evolution engine
type EvolutionCommand struct {
	Type            CommandType
	Generation      int
	CrossoverPoints int
	CrossoverRate   float64
	MutationRate    float64
	ElitismPct      float64
}

// CommandType defines the type of evolution command
type CommandType int

const (
	CmdStartGeneration CommandType = iota
	CmdStop
)
