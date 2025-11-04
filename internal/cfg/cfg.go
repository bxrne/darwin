package cfg

import (
	"fmt"

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
	MinDepth    int      `toml:"min_depth"`
	FunctionSet []string `toml:"function_set"`
	TerminalSet []string `toml:"terminal_set"`
}

// validate validates the TreeIndividualConfig.
func (tic *TreeIndividualConfig) validate() error {
	if tic.MaxDepth <= 0 {
		return fmt.Errorf("max_depth must be greater than 0")
	}
	if tic.MinDepth < 0 || tic.MinDepth > tic.MaxDepth {
		return fmt.Errorf("min_depth must be between 0 and max_depth")
	}
	if len(tic.FunctionSet) == 0 {
		return fmt.Errorf("function_set must not be empty")
	}
	if len(tic.TerminalSet) == 0 {
		return fmt.Errorf("terminal_set must not be empty")
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
	Evolution EvolutionConfig           `toml:"evolution"`
	BitString BitStringIndividualConfig `toml:"bitstring_individual"`
	Tree      TreeIndividualConfig      `toml:"tree_individual"`
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

	// Mutual exclusivity
	if c.Tree.Enabled && c.BitString.Enabled {
		return fmt.Errorf("only one individual type can be enabled at a time")
	}

	return nil
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
