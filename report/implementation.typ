#import "@preview/dashy-todo:0.1.3": todo

== Design Principles

The architecture adheres to several key design principles:

+ *Interface Segregation*: The `Evolvable` interface defines only essential operations, allowing diverse implementations without unnecessary coupling.

+ *Dependency Inversion*: High-level modules (evolution engine) depend on abstractions (Evolvable interface, Selector interface) rather than concrete implementations.

+ *Single Responsibility*: Each package has a clearly defined responsibility, promoting maintainability and testability.

+ *Concurrency Safety*: The channel based architecture ensures thread safe operation while enabling parallel offspring generation.

== Language Choice Rationale

=== Evolution engine

Go was selected for its interface based generics enabling type-safe generic evolution, goroutines and channels for efficient concurrency, native code performance, and excellent tooling. The generic type system eliminates code duplication while maintaining compile-time type safety.

=== Bridge 

The GP→RL bridge is implemented in Python due to the mature RL ecosystem (PettingZoo), direct environment integration without language bindings, multiprocessing for true parallelism (bypassing GIL).

== Overview 

=== Bridge

The GP→RL bridge (`game/bridge.py`) provides a universal translation layer between Genetic Programming systems and Reinforcement Learning environments, addressing the paradigm mismatch where GP produces functions while RL requires interactive agents.

=== Architecture and Design

The bridge follows a multi-process server architecture where each client connection spawns an independent worker process, enabling true parallelism and bypassing Python's GIL (Global Interpreter Lock) which ensures only one bytecode set executes at a time, so it cannot utilize multiple CPU cores with threads. 

The bridge exposes a standard RL API via TCP using JSON encoded messages supporting connection establishment, observation requests, action submission, health checks, and game termination.

Worker processes receive socket file descriptors, manage complete game session lifecycles, handle JSON serialization, and communicate game state back to clients. This isolation prevents interference between concurrent evaluations.

=== Connection Pooling and Health Checking

The Go side implements a TCP connection pool that maintains configurable connections (default: 100), performs health checks before reuse, creates connections on demand, and handles failures gracefully. Health checks use a lightweight `HEALTH` message that the bridge responds to. 

=== Protocol Abstraction

The bridge translates GP tree evaluation outputs into RL action format. Action trees evaluate game state to produce numeric outputs that are combined with weight matrices, mapped to discrete action spaces, serialized as JSON, and converted to PettingZoo action format. This abstraction keeps the evolution engine game agnostic.

=== Action Evaluation 

Each game session involves multiple action and observation cycles. After initialization, games proceed through multiple steps (up to `max_steps`, default: 1000). At each step, the client evaluates all action trees *multiple times* once for each row in the weight matrix. Each evaluation uses different weight values (`w0, w1, ..., wN`) as inputs to the action trees, producing multiple outputs per action tree. These outputs are then combined using softmax selection to produce a single action vector `[pass, cell_i, cell_j, direction, split]` that is submitted via TCP. The bridge processes one action vector per timestep, but the action trees are evaluated multiple times per step (across weight rows) to produce that single action. For fitness evaluation, the engine plays `test_case_count` games (default: 3) per individual, ensuring robust evaluation across diverse scenarios.

=== Valid Action Mapping

The system maintains synchronized mappings of owned cells and mountains on both the bridge (Python) and Go sides to ensure only valid actions are selected. This prevents evolved strategies from attempting invalid moves (e.g., moving from unowned cells or into mountains).

*Bridge Side (Python)*: The bridge sends the mountains grid (from `observations[client_id]["mountains"]`) at reset and a boolean owned cells map (from `valid_start_point_map()`) at each step in the observation `info` field.

*Go Side*: The client stores the mountains map and, at each step, uses the received owned cells map to mask invalid actions. `ActionValidator` ensures actions are only chosen from owned cells (with enough armies) and do not move into mountains or out-of-bounds locations.

*Action Selection Process*: Action tree outputs are converted to probabilities via softmax, then masked by the valid action maps. This ensures that only valid actions (from owned cells, not into mountains) can be selected, even if action trees produce outputs for invalid coordinates. The masking process multiplies invalid action probabilities by 0, effectively removing them from consideration before sampling.

This bidirectional mapping ensures that evolved strategies operate within game constraints without requiring the action trees themselves to learn these constraints, simplifying the evolutionary search space.

== Engine

The evolution engine (`internal/evolution/engine.go`) is implemented as a generic, type-safe system using Go's interfaces and message passing via channels, enabling evolution of any representation type implementing the `Evolvable` interface without code duplication.

=== Extensible Architecture

The `Evolvable` interface provides `Mutate()`, `Max()`, `MultiPointCrossover()`, `GetFitness()`, `SetFitness()`, `Clone()`, and `Describe()` methods. This enables zero code duplication (same engine evolves bitstrings, trees, grammar trees, and action trees), compile-time type safety, and extensibility (new types only need to implement the interface).

The engine receives commands through a command channel and processes generations asynchronously, enabling clean separation, graceful shutdown, and non-blocking submission.

=== CPU Utilization and Concurrent Offspring Generation

The engine maximizes CPU utilization by spawning one goroutine per offspring needed, utilizing all available CPU cores. This scales linearly with CPU count through channel-based communication without mutex contention. The engine doesn't limit goroutine count to `runtime.NumCPU()` because Go's scheduler efficiently manages mapping, offspring generation is I/O-bound with the bridge, and channel buffering prevents blocking. Fitness calculation for dual populations uses `runtime.NumCPU()` to chunk work across cores.

=== Coevolution Support

The engine supports coevolution through the `Population` interface, which can represent single or dual populations. For dual population evolution, `ActionTreeAndWeightsPopulation` manages two separate populations that alternate evolution. The generic design enables coevolution without modification, the `Population` interface abstracts management, and `Population.Update()` enables population specific logic.

=== Fitness Calculation Integration

Fitness calculation is integrated through the `FitnessCalculator` interface. The engine calls `CalculateFitness()` on new individuals during offspring generation and initial population fitness calculation. The interface enables pluggable fitness functions: symbolic regression (MSE based), binary fitness (OneMax, trap functions), and action tree game playing fitness (via GP→RL bridge).

*Multi Game Evaluation for Action Trees*: ActionTreeIndividuals are evaluated across `test_case_count` games (default: 3), not a single game. For each individual, `SetupGameAndRun()` is called multiple times, fitness scores are averaged, and within each game the individual makes multiple action decisions (up to `max_steps`, default: 1000). At each step, all action trees are evaluated *multiple times* (once per weight matrix row), with each evaluation using different weight values as tree inputs. These multiple evaluations produce outputs that are combined using softmax to select a single action vector, which is then submitted to the bridge. This multi level evaluation (multiple games, multiple steps per game, multiple action tree evaluations per step across weight rows) ensures robust fitness assessment.

=== Generation Processing

Each generation: (1) calculates fitness for initial population (generation 1) or inherits from parents, (2) sorts population by fitness, (3) preserves elite individuals ($k = ceil("population_size" * "elitism_percentage")$), (4) generates remaining offspring through selection, crossover ($p_c = 0.7$), mutation ($p_m = 0.3$), and fitness evaluation, (5) calls `Population.Update()` for population specific logic, (6) calculates and streams metrics.

== Library

The library provides foundational interfaces, operators, individual representations, and configuration system enabling the generic evolution engine.

=== Interface Design

The `Evolvable` interface defines the contract for all individual types with minimal, composable, type-safe methods: `Mutate()`, `MultiPointCrossover()`, `Max()`, `GetFitness()`/`SetFitness()`, `Clone()`, and `Describe()`.

=== Genetic Operators

*Crossover*: Multipoint crossover, configurable through `crossover_point_count`. For bitstrings, divides genome into segments; for trees, exchanges subtrees. Crossover rate $p_c = 0.7$ (default).

*Mutation*: Introduces diversity. For bitstrings, flips bits with probability $p_m$. For trees, can replace subtrees, modify node values, or swap subtrees. Mutation rate $p_m = 0.3$ (default).

*Selection*: Two mechanisms—Roulette Wheel (fitness proportional) and Tournament (selects $k$ individuals, returns fittest). Tournament provides better control over selection pressure.

=== Individual Representations

While the system supports bitstring individuals, standard tree individuals, and grammar evolution, the primary focus was on `ActionTreeIndividual`, the core representation for game playing strategy evolution.

==== Action Tree Individual

The `ActionTreeIndividual` contains a collection of expression trees, one for each action type (pass, move direction, split, cell coordinates). These trees evaluate game state variables to produce numeric outputs combined with weight matrices to select actions.

*Structure*: Contains `Trees` (map from action names to `Tree` structures), `fitness`, and `clientId` for tracking.

*Initialization with Ramped Half-and-Half*: Population is divided into `initialDepth` depth groups (depths 1 to `initialDepth`). Within each depth group, individuals are split: first half uses *grow* method (variable depth), second half uses *full* method (complete binary trees). All trees within a single individual use the same depth and method.

*Grow vs Full Methods*: Grow creates trees by randomly selecting from function and terminal sets at each level, producing variable shapes. Full creates complete binary trees with all nodes at depth < target as functions, all at target depth as terminals.

*Variable Set Construction*: Combines weight variables (`w0, w1, ..., wN`) and game state variables (e.g., `army_diff`, `land_diff`, `distance_to_enemy_general`), enabling trees to access both during evaluation.

*Genetic Operations*:

*Mutation*: Applied to each tree independently using recursive traversal with three mutation types: Value Mutation (60%) replaces terminal/function values, Shrink Mutation (20%) replaces non-terminal subtrees with terminals (reducing depth), Grow Mutation (20%) replaces terminals with function nodes and children (increasing depth). Shrink only occurs if resulting tree depth ≥ 1; grow only if current depth < `maxDepth`. Safety check regenerates trees that reach depth 0.

*Crossover*: Multipoint crossover performed independently for each action tree. Exchanges subtrees between corresponding trees from two parents, ensuring combined depth doesn't exceed `maxDepth`, then recalculates tree depths.

*Clone*: Creates deep copy of all trees ensuring independent operations.

*Max*: Returns ActionTreeIndividual with higher fitness for offspring selection.

==== Weights Individual

The `WeightsIndividual` (`internal/individual/weights.go`) represents a weight matrix that biasesaction tree outputs during action selection. Each weights individual contains a dense matrix of floating-point values that serve as inputs to action trees during evaluation.

*Structure*: Contains `Weights` (dense matrix from gonum), `fitness`, `minVal`/`maxVal` (bounds for mutation, default: -5.0 to 5.0), and `clientId` for tracking which game client produced the best fitness.

*Initialization*: Weights individuals are initialized with random values uniformly distributed between -5.0 and 5.0. Matrix dimensions are determined by configuration:
- Height (`maxNumInputs`): Maximum value across all action types, determining how many weight rows are available
- Width (`weights_column_count`): Number of weight columns, corresponding to the number of weight variables (`w0, w1, ..., wN`) that action trees can access (max depth of the action trees)

The weights population size is configured separately via `weights_count` (default: 5), which is typically smaller than the action tree population size.

*Usage in Action Evaluation*: During action tree evaluation, each row of the weight matrix is used as inputs to the action trees. For each weight row, the values (`w0, w1, ..., wN`) are set as variables in the tree evaluation context, and all action trees are evaluated with those weight values. This produces multiple outputs per action tree (one per weight row), which are then combined using softmax selection to produce the final action vector.


*Genetic Operations*:

*Mutation*: Each weight value in the matrix is independently mutated with probability equal to the mutation rate. When mutated, the value is reset to a random value uniformly distributed between `minVal` and `maxVal` (default: -5.0 to 5.0). This allows weights to explore the full range of possible values.

*Crossover*: Multipoint crossover is performed along the column dimension. Crossover points are randomly selected along columns, and weight values are swapped between parents at those points. The crossover alternates which parent's values are inherited, creating offspring that combine weight patterns from both parents.

*Clone*: Creates a deep copy of the weight matrix, ensuring independent mutation and crossover operations without side effects.

*Max*: Returns WeightsIndividual with higher fitness, used during offspring generation to select the better of two children.

*Fitness Evaluation*: Weights individuals are evaluated by testing them against all action trees in the population. Each weight-tree combination is evaluated across `test_case_count` games, and the weight's fitness is the maximum fitness achieved across all tree combinations. This ensures weights are optimized for the best-performing action trees.

*Backpropagation Emulation*: The WeightsIndividual representation enables the system to emulate backpropagation purely through evolutionary algorithms. In neural networks, backpropagation optimizes weight matrices using gradient descent to minimize loss. Here, Genetic Algorithms evolve weight matrices using selection, crossover, and mutation to maximize fitness. The weights modulate action tree outputs in the same way neural network weights modulate neuron outputs each weight row provides inputs that influence how action trees evaluate game state. Research demonstrates that GA-based weight optimization provides a competitive alternative to gradient-based methods, avoiding local minima while maintaining performance @Petroski2018Deep. By evolving weights through GA rather than backpropagation, the system achieves weight optimization without requiring differentiable functions or gradient computation, enabling pure evolutionary optimization of game-playing strategies.

==== Other Individual Types

*Bitstring Individuals*: Traditional GA with fixed length binary genomes.

*Tree Individuals*: Standard GP expression trees for symbolic regression.

*Grammar Evolution*: Maps integer genomes to expression trees via context free grammar rules.

== Configuration System

Darwin uses TOML (Tom's Obvious Minimal Language) configuration files for flexible parameter specification. This approach separates configuration from code, enabling experimentation without recompilation. The configuration system supports:

- Evolution parameters (population size, generations, rates)
- Individual-specific settings (genome size, tree depth, function sets)
- Fitness function configuration (target functions, test cases)
- Metrics and logging settings
- Bridge connection parameters (host, port, timeouts)

This design follows the 12-Factor App methodology's configuration principle @Wiggins2012Twelve, storing configuration in the environment (files) rather than hardcoding values.

== Metrics

The metrics system provides realtime streaming of evolution progress with CSV export capabilities.

=== Metrics Handler

Uses channel based streaming architecture. The `MetricsStreamer` broadcasts metrics to all subscribers asynchronously, supports multiple concurrent handlers, and uses non-blocking broadcast to prevent backpressure.

=== CSV Export

The CSV writer provides structured export with header rows, one row per generation, flush-after-write for persistence, and mutex-protected concurrent access. Columns include generation, duration, fitness statistics, best description, tree depth statistics, population size, and timestamp.

=== Metrics Collection

The engine calculates metrics after each generation: fitness statistics, population size, tree depth statistics, generation execution time, and best individual description. Metrics are calculated synchronously and sent asynchronously through channels.

== Dual Population Evolution

The dual population evolution scheme combines GA and GP in a coevolutionary framework, maintaining two parallel populations that alternate evolution.

=== Architecture

The system maintains an Action Tree Population (GP) that evolves decision making functions, and a Weights Population (GA) that evolves weight matrices. `ActionTreeAndWeightsPopulation` manages both populations and implements the `Population` interface, enabling transparent operation by the generic engine.

=== Modulation Mechanism

Populations alternate evolution every `switch_training_target_step` generations (default: 10) using an evolutionary islands approach. During Action Tree Evolution Phase, action trees evolve while weights remain static. During Weight Evolution Phase, weights evolve while action trees remain static. When populations switch, fitness is recalculated for both to reflect the new coevolutionary context.

=== Modulation Effect

Alternating phases modulate evolutionary pressure: action trees must work with current weights (preventing overfitting), weights must optimize for current trees (preventing exploitation). This creates stabilizing effects preventing premature convergence and encouraging robust coadaptation. Over multiple cycles, both populations coevolve toward complementary solutions.

=== Test Case Count and Robust Evaluation

Each Action Tree individual plays `test_case_count` games (default: 3) rather than one. The fitness calculator evaluates each individual across multiple games, averaging results. This prevents convergence on tactics exploiting specific game states. The `test_case_count` parameter controls the robustness vs. computational cost tradeoff.

For weights evaluation, each weight individual is tested against all action trees, with each combination evaluated across `test_case_count` games. Weight fitness is the maximum across all tree combinations. Similarly, for action tree evaluation, each tree is tested against all weights, with each combination evaluated across `test_case_count` games. Tree fitness is the maximum across all weight combinations. This dual evaluation strategy ensures both components evolve toward complementary solutions performing robustly across diverse game scenarios.
