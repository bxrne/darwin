package cfg

import (
	"fmt"
	"strconv"

	"github.com/BurntSushi/toml"
)

// BitStringIndividualConfig holds configuration for bitstring individuals.
type BitStringIndividualConfig struct {
	Enabled    bool `toml:"enabled"`
	GenomeSize int  `toml:"genome_size"`
}

// validate validates the BitStringIndividualConfig.
func (bic *BitStringIndividualConfig) validate() error {
	if bic.GenomeSize <= 0 {
		return fmt.Errorf("genome_size must be greater than 0")
	}
	return nil
}

// TreeIndividualConfig holds configuration for tree individuals.
type TreeIndividualConfig struct {
	Enabled     bool     `toml:"enabled"`
	MaxDepth    int      `toml:"max_depth"`
	InitalDepth int      `toml:"initial_depth"`
	VariableSet []string `toml:"variable_set"`
	OperandSet  []string `toml:"operand_set"`
	TerminalSet []string `toml:"terminal_set"`
}

// validate validates the TreeIndividualConfig.
func (tic *TreeIndividualConfig) validate() error {
	if tic.MaxDepth <= 0 {
		return fmt.Errorf("max_depth must be greater than 0")
	}
	if tic.InitalDepth < 0 || tic.InitalDepth > tic.MaxDepth {
		return fmt.Errorf("initial_depth must be between 0 and max_depth")
	}
	if len(tic.VariableSet) == 0 {
		return fmt.Errorf("variable_set must not be empty")
	}
	if len(tic.OperandSet) == 0 {
		return fmt.Errorf("operand_set must not be empty")
	}

	if len(tic.TerminalSet) == 0 {
		return fmt.Errorf("terminal_set must not be empty")
	}

	// Validate primitive set contains only valid operators
	if err := validateTerminalSet(tic.VariableSet); err != nil {
		return fmt.Errorf("Variable_set validation failed: %w", err)
	}

	// Validate terminal set contains valid variables and constants
	if err := validateTerminalSet(tic.TerminalSet); err != nil {
		return fmt.Errorf("terminal_set validation failed: %w", err)
	}
	if err := validateOperandSet(tic.OperandSet); err != nil {
		return fmt.Errorf("operand set validation failed: %w", err)
	}

	return nil
}

// MetricsConfig holds configuration for metrics output.
type MetricsConfig struct {
	CSVEnabled bool   `toml:"csv_enabled"`
	CSVFile    string `toml:"csv_file"`
}

// validate validates the MetricsConfig.
func (mc *MetricsConfig) validate() error {
	if mc.CSVEnabled && mc.CSVFile == "" {
		return fmt.Errorf("csv_file must be specified when csv_enabled is true")
	}
	return nil
}

type FitnessConfig struct {
	TestCaseCount  int    `toml:"test_case_count"`
	TargetFunction string `toml:"target_function"`
}

func (fc *FitnessConfig) validate() error {
	if fc.TestCaseCount <= 0 {
		return fmt.Errorf("test_case_count must be positive greater than 0")
	}
	return nil
}

type GrammarTreeConfig struct {
	GenomeSize int  `toml:"genome_size"`
	Enabled    bool `toml:"enabled"`
}

func (gtc *GrammarTreeConfig) validate() error {
	if gtc.GenomeSize <= 0 {
		return fmt.Errorf("genome_size must be postive int")
	}
	return nil
}

// EvolutionConfig holds configuration for the evolutionary algorithm.
type EvolutionConfig struct {
	PopulationSize      int     `toml:"population_size"`
	CrossoverPointCount int     `toml:"crossover_point_count"`
	CrossoverRate       float64 `toml:"crossover_rate"`
	MutationRate        float64 `toml:"mutation_rate"`
	Generations         int     `toml:"generations"`
	ElitismPercentage   float64 `toml:"elitism_percentage"`
	SelectionSize       int     `toml:"selection_size"`
	SelectionType       string  `toml:"selection_type"`
	Seed                int64   `toml:"seed"`
}

// validate validates the EvolutionConfig.
func (ec *EvolutionConfig) validate() error {
	if ec.PopulationSize <= 0 {
		return fmt.Errorf("population_size must be greater than 0")
	}

	if ec.CrossoverPointCount <= 0 {
		return fmt.Errorf("crossover_point_count must be above 0")
	}

	if ec.SelectionSize <= 0 {
		return fmt.Errorf("selection_size must be above 0")
	}
	if ec.SelectionType != "tournament" && ec.SelectionType != "roulette" {
		return fmt.Errorf("selection_type must be either tournament or roulette")
	}
	if ec.CrossoverRate < 0 || ec.CrossoverRate > 1 {
		return fmt.Errorf("crossover_rate must be above 0")
	}

	if ec.MutationRate < 0 || ec.MutationRate > 1 {
		return fmt.Errorf("mutation_rate must be between 0 and 1")
	}

	if ec.Generations <= 0 {
		return fmt.Errorf("generations must be greater than 0")
	}

	if ec.ElitismPercentage <= 0 || ec.ElitismPercentage > 1 {
		return fmt.Errorf("elitism_percentage must be greater than 0 and less than 1")
	}
	// Set default seed if not provided

	if ec.Seed == 0 {
		ec.Seed = 42
	}

	// Set default seed if not provided
	return nil
}

// Config holds the entire configuration for the evolutionary algorithm.
type Config struct {
	Evolution   EvolutionConfig           `toml:"evolution"`
	BitString   BitStringIndividualConfig `toml:"bitstring_individual"`
	Tree        TreeIndividualConfig      `toml:"tree_individual"`
	Metrics     MetricsConfig             `toml:"metrics"`
	Fitness     FitnessConfig             `toml:"fitness"`
	GrammarTree GrammarTreeConfig         `toml:"grammar_tree"`
}

// validate validates the entire Config.
func (c *Config) validate() error {
	// Subconfigs
	if err := c.Evolution.validate(); err != nil {
		return fmt.Errorf("evolution config validation failed: %w", err)
	}

	if err := c.BitString.validate(); err != nil {
		return fmt.Errorf("bitstring individual config validation failed: %w", err)
	}

	if err := c.Tree.validate(); err != nil {
		return fmt.Errorf("tree individual config validation failed: %w", err)
	}

	if err := c.Metrics.validate(); err != nil {
		return fmt.Errorf("metrics config validation failed: %w", err)
	}

	if err := c.Fitness.validate(); err != nil {
		return fmt.Errorf("fitness config validation failed: %w", err)
	}
	if err := c.GrammarTree.validate(); err != nil {
		return fmt.Errorf("grammar tree config validation failed: %w", err)
	}
	// Mutual exclusivity
	if c.Tree.Enabled && c.BitString.Enabled && c.GrammarTree.Enabled {
		return fmt.Errorf("only one individual type can be enabled at a time")
	}

	return nil
}

// validatePrimitiveSet checks that primitive set contains only valid operators
func validateOperandSet(primitiveSet []string) error {
	validPrimitives := map[string]bool{
		"+": true, "-": true, "*": true, "/": true,
	}

	for _, prim := range primitiveSet {
		if !validPrimitives[prim] {
			return fmt.Errorf("invalid primitive: %s", prim)
		}
	}
	return nil
}

// validateTerminalSet checks that terminal set contains valid variables and numeric constants
func validateTerminalSet(terminalSet []string) error {
	for _, terminal := range terminalSet {
		// Check if it's a valid variable name or numeric constant
		if !isValidVariableName(terminal) && !isValidNumber(terminal) {
			return fmt.Errorf("invalid terminal: %s (must be variable name or numeric constant)", terminal)
		}
	}
	return nil
}

// isValidVariableName checks if string is a valid variable name (alphabetic)
func isValidVariableName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// isValidNumber checks if string is a valid numeric constant
func isValidNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}
