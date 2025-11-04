package cfg

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Evolution EvolutionConfig `toml:"evolution"`
}

func (c *Config) validate() error {
	if err := c.Evolution.validate(); err != nil {
		return fmt.Errorf("evolution config validation failed: %w", err)
	}

	return nil
}

type EvolutionConfig struct {
	PopulationSize      int     `toml:"population_size"`
	GenomeSize          int     `toml:"genome_size"`
	CrossoverPointCount int     `toml:"crossover_point_count"`
	CrossoverRate       float64 `toml:"crossover_rate"`
	MutationRate        float64 `toml:"mutation_rate"`
	Generations         int     `toml:"generations"`
	ElitismPercentage   float64 `toml:"elitism_percentage"`
	Seed                int64   `toml:"seed"`
}

func (ec *EvolutionConfig) validate() error {
	if ec.PopulationSize <= 0 {
		return fmt.Errorf("population_size must be greater than 0")
	}
	if ec.GenomeSize <= 0 {
		return fmt.Errorf("genome_size must be greater than 0")
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
		// Set default seed if not provided
		if ec.Seed == 0 {
			ec.Seed = 42
		}
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
