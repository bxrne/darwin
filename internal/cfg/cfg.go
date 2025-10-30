package cfg

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Evolution EvolutionConfig `toml:"evolution"`
}

func (c *Config) Validate() error {
	if err := c.Evolution.Validate(); err != nil {
		return fmt.Errorf("evolution config validation failed: %w", err)
	}

	return nil
}

type EvolutionConfig struct {
	PopulationSize      int     `toml:"population_size"`
	GenomeSize          int     `toml:"genome_size"`
	CrossoverPointCount int     `toml:"crossover_point_count"`
	MutationRate        float64 `toml:"mutation_rate"`
	MutationPoints      []int   `toml:"mutation_points"`
	Generations         int     `toml:"generations"`
	ElitismPercentage   float64 `toml:"elitism_percentage"`
}

func (ec *EvolutionConfig) Validate() error {
	if ec.PopulationSize <= 0 {
		return fmt.Errorf("population_size must be greater than 0")
	}
	if ec.GenomeSize <= 0 {
		return fmt.Errorf("genome_size must be greater than 0")
	}
	if ec.CrossoverPointCount <= 0 {
		return fmt.Errorf("crossover_point_count must be above 0")
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
	for _, point := range ec.MutationPoints {
		if point < 0 || point >= ec.GenomeSize {
			return fmt.Errorf("mutation_points must be within the range of genome_size")
		}
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &config, nil
}
