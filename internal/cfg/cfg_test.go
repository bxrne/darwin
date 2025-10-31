package cfg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvolutionConfigValidate_GIVEN_valid_config_WHEN_validate_THEN_no_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.NoError(t, err)
}

func TestEvolutionConfigValidate_GIVEN_population_size_zero_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      0,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "population_size must be greater than 0")
}

func TestEvolutionConfigValidate_GIVEN_genome_size_zero_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          0,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "genome_size must be greater than 0")
}

func TestEvolutionConfigValidate_GIVEN_crossover_point_count_zero_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 0,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "crossover_point_count must be above 0")
}

func TestEvolutionConfigValidate_GIVEN_mutation_rate_negative_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        -0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutation_rate must be between 0 and 1")
}

func TestEvolutionConfigValidate_GIVEN_mutation_rate_above_one_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        1.5,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutation_rate must be between 0 and 1")
}

func TestEvolutionConfigValidate_GIVEN_generations_zero_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         0,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generations must be greater than 0")
}

func TestEvolutionConfigValidate_GIVEN_elitism_percentage_zero_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   0.0,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "elitism_percentage must be greater than 0 and less than 1")
}

func TestEvolutionConfigValidate_GIVEN_elitism_percentage_above_one_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 2},
		Generations:         100,
		ElitismPercentage:   1.5,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "elitism_percentage must be greater than 0 and less than 1")
}

func TestEvolutionConfigValidate_GIVEN_mutation_points_out_of_range_WHEN_validate_THEN_error(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize:      10,
		GenomeSize:          5,
		CrossoverPointCount: 2,
		MutationRate:        0.1,
		MutationPoints:      []int{0, 5}, // 5 is out of range for genome size 5
		Generations:         100,
		ElitismPercentage:   0.2,
	}

	err := config.validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutation_points must be within the range of genome_size")
}

func TestConfigValidate_GIVEN_valid_evolution_config_WHEN_validate_THEN_no_error(t *testing.T) {
	config := Config{
		Evolution: EvolutionConfig{
			PopulationSize:      10,
			GenomeSize:          5,
			CrossoverPointCount: 2,
			MutationRate:        0.1,
			MutationPoints:      []int{0, 2},
			Generations:         100,
			ElitismPercentage:   0.2,
		},
	}

	err := config.validate()

	assert.NoError(t, err)
}

func TestConfigValidate_GIVEN_invalid_evolution_config_WHEN_validate_THEN_error(t *testing.T) {
	config := Config{
		Evolution: EvolutionConfig{
			PopulationSize:      0, // invalid
			GenomeSize:          5,
			CrossoverPointCount: 2,
			MutationRate:        0.1,
			MutationPoints:      []int{0, 2},
			Generations:         100,
			ElitismPercentage:   0.2,
		},
	}

	err := config.validate()

	assert.Error(t, err)
}

func TestLoadConfig_GIVEN_valid_toml_file_WHEN_load_THEN_config_returned(t *testing.T) {
	tomlContent := `
[evolution]
population_size = 10
genome_size = 5
crossover_point_count = 2
mutation_rate = 0.1
mutation_points = [0, 2]
generations = 100
elitism_percentage = 0.2
`

	// Create temp file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
	assert.NoError(t, err)

	config, err := LoadConfig(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 10, config.Evolution.PopulationSize)
}

func TestLoadConfig_GIVEN_invalid_toml_file_WHEN_load_THEN_error(t *testing.T) {
	invalidToml := `
[evolution
population_size = 10
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")
	err := os.WriteFile(configPath, []byte(invalidToml), 0644)
	assert.NoError(t, err)

	config, err := LoadConfig(configPath)

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_GIVEN_invalid_config_WHEN_load_THEN_validation_error(t *testing.T) {
	tomlContent := `
[evolution]
population_size = 0
genome_size = 5
crossover_point_count = 2
mutation_rate = 0.1
mutation_points = [0, 2]
generations = 100
elitism_percentage = 0.2
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
	assert.NoError(t, err)

	config, err := LoadConfig(configPath)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "population_size must be greater than 0")
}
