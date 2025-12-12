# Darwin

Darwin is a flexible evolutionary computation framework supporting Genetic Algorithms (GA), Genetic Programming (GP), and Grammar Evolution. It features an extensible architecture with the Evolvable interface, channel-based evolution engine, and async metrics streaming. It emerged out of training action trees and tuning them with genetic algorithm emulation of backpropagation to play the Generals IO game.

## Architecture

- Channel-based evolution engine for concurrent processing
- Thread-safe random number generation
- Async metrics streaming via channels
- Extensible `Evolvable` interface for custom individual types

For detailed configuration options, see the individual config files in `config/`.

## Quick Start

### Build

```bash
git clone https://github.com/bxrne/darwin.git
cd darwin
go mod tidy
go build ./cmd/darwin
```

### Run Evolution

```bash
./darwin -config config/small.toml
```

### Run Tests

```bash
go test ./...
```

### Game Server (for Action Tree Evolution)

```bash
cd game
uv venv && uv sync
uv run main.py
```

### Plot Results

```bash
cd plot
uv venv && uv sync  
uv run main.py --csv path/to/metrics.csv
```

## Available Configurations

All config files are located in the `config/` directory:

| Config File | Purpose | Individual Type | Use Case |
|-------------|---------|-----------------|----------|
| `small.toml` | Quick testing | Bitstring | Fast experiments (100 pop, 10 gen) |
| `medium.toml` | Balanced experiments | Bitstring | Standard runs (500 pop, 50 gen) |
| `large.toml` | Comprehensive evolution | Bitstring | Full experiments (2000 pop, 200 gen) |
| `default.toml` | Default settings | Action Tree | Basic genetic programming |
| `test.toml` | Action tree evolution | Action Tree | Interactive game evolution |
| `ge_problem.toml` | Grammar evolution | Grammar Tree | Symbolic regression problems |


The `default.toml` is what is used for evolving to Generals IO.

## Project Components

### Main Darwin Binary
- Core evolution engine with configurable individual types
- Supports bitstring, tree, grammar, and action-based genomes
- Async metrics streaming to CSV files

### Game Server (`game/`)
- TCP server for interactive action tree evolution
- Used when `action_tree` individual type is enabled
- Config: `server_addr = "127.0.0.1:5000"` (default)

### Plotter (`plot/`)
- Visualizes evolution metrics from CSV output
- Generates fitness progression plots
- Usage: `uv run main.py --csv path/to/metrics.csv`

## Individual Types

Darwin supports different genome representations:

**Bitstring Individuals** (`[bitstring_individual]`)
- Binary genomes for classic GA problems
- Fixed-length bit strings with configurable size

**Tree Individuals** (`[tree_individual]`)  
- Expression trees for genetic programming
- Variable depth with customizable function/terminal sets

**Grammar Tree Individuals** (`[grammar_tree]`)
- Grammar-based evolution for structured problems
- Useful for symbolic regression and language generation

**Action Tree Individuals** (`[action_tree]`)
- Interactive evolution via game server
- Evolves action sequences for game-playing agents

## Development

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/evolution
go test ./internal/fitness
```

### Benchmarking

Darwin includes comprehensive benchmarks for all individual types and scenarios.

#### Benchmark Categories

**Standard Scaling Benchmarks:**
- **BitString**: Binary genome evolution (Small/Medium/Large/Huge)
- **TreeIndividual**: Expression tree evolution (Small/Medium/Large/Huge)  
- **GrammarTree**: Grammar-based evolution (Small/Medium/Large/Huge)
- **ActionTree**: Game-based evolution (requires server)

**Comparative Benchmarks:**
- All individual types solving the same symbolic regression problem (`x^2 + 2*y + 1`)
- Direct performance and solution quality comparisons

#### Running Benchmarks

```bash
# Run all benchmarks (19 total scenarios)
go test -bench=BenchmarkEvolution ./cmd/darwin -benchmem

# Run specific individual types
go test -bench=BenchmarkEvolution_BitString ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Tree ./cmd/darwin -benchmem  
go test -bench=BenchmarkEvolution_GrammarTree ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_ActionTree ./cmd/darwin -benchmem

# Run specific sizes
go test -bench=BenchmarkEvolution_Small ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Medium ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Large ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Huge ./cmd/darwin -benchmem

# Run comparative benchmarks (same problem, different types)
go test -bench=BenchmarkEvolution_Compare ./cmd/darwin -benchmem

# Performance profiling
go test -bench=BenchmarkEvolution ./cmd/darwin -cpuprofile=cpu.prof
go test -bench=BenchmarkEvolution ./cmd/darwin -memprofile=mem.prof
go tool pprof cpu.prof
go tool pprof mem.prof
```

#### ActionTree Benchmarks (Requires Game Server)

Before running ActionTree benchmarks, start the game server:

```bash
cd game
uv venv && uv sync  
uv run main.py
```

Then run ActionTree benchmarks:
```bash
go test -bench=BenchmarkEvolution_ActionTree ./cmd/darwin -benchmem
```

#### Benchmark Descriptions

| Benchmark | Population | Genome/Depth | Generations | Description |
|-----------|------------|--------------|-------------|-------------|
| BitString_Small | 100 | 50 bits | 10 | Quick binary evolution test |
| BitString_Medium | 500 | 200 bits | 50 | Standard binary GA |
| BitString_Large | 1000 | 500 bits | 100 | Large-scale binary evolution |
| BitString_Huge | 500 | 5000 bits | 10 | Large genome stress test |
| Tree_Small | 100 | depth 3 | 10 | Small expression trees |
| Tree_Medium | 500 | depth 5 | 50 | Medium genetic programming |
| Tree_Large | 1000 | depth 7 | 100 | Large GP evolution |
| Tree_Huge | 500 | depth 10 | 10 | Deep tree evolution |
| GrammarTree_Small | 100 | 50 genome | 10 | Small grammar evolution |
| GrammarTree_Medium | 500 | 100 genome | 50 | Medium grammatical evolution |
| GrammarTree_Large | 1000 | 200 genome | 100 | Large grammar-based evolution |
| GrammarTree_Huge | 500 | 500 genome | 10 | Grammar evolution stress test |
| ActionTree_GameServer | 200 | 3 actions | 25 | Game-playing agent evolution |
| Compare_BitString | 300 | 150 bits | 30 | Symbolic regression with binary GA |
| Compare_Tree | 300 | depth 4 | 30 | Symbolic regression with GP |
| Compare_GrammarTree | 300 | 75 genome | 30 | Symbolic regression with grammar evolution |

**Target Function for Comparative Benchmarks:** `x^2 + 2*y + 1` with 20 test cases


