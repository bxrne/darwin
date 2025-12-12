#import "@preview/arkheion:0.1.1": arkheion, arkheion-appendices
#import "@preview/fletcher:0.5.8" as fletcher: diagram, node, edge
#import "@preview/fletcher:0.5.8": shapes
#import "@preview/dashy-todo:0.1.3": todo

#show: arkheion.with(
  title: "Darwin: Towards game playing via Genetic Programming and Backpropagation emulation",
  authors: (
    (name: "Adam Byrne", email: "22338004@studentmail.ul.ie", affiliation: "University of Limerick"),
    (name: "Art O'Liathain", email: "22363092@studentmail.ul.ie", affiliation: "University of Limerick"),
  ),
  abstract: [
    This report presents the design, implementation, and evaluation of Darwin, an evolutionary engine written in Go that can evolve individuals against a Reinforcement Learning environment using Genetic Programming (GP) and Genetic Algorithms (GA) over TCP. The engine is designed to be generic and extensible, supporting multiple evolutionary paradigms and individual representations. The system supports multiple evolutionary paradigms including GA with bitstring genomes, GP with tree-based representations and Grammatical Evolution (GE) for symbolic regression. The system leverages Go's concurrency primitives to achieve high performance through parallel offspring generation. A key innovation is the generic GP→RL bridge that manages parallel environments for individuals and generalises the interaction between GP individuals and RL environments.  
  ],
  keywords: ("Evolutionary Algorithms", "Genetic Programming", "Grammar Evolution", "Symbolic Regression", "Concurrent Computing"),
  date: "December 2025",
)

#show link: underline

#pagebreak()

#outline(
  title: "List of Figures",
  target: figure.where(kind: image)
)
#outline(
  title: "List of Tables", 
  target: figure.where(kind: table)
)
#pagebreak()

= Introduction

Evolutionary Algorithms (EAs) represent a class of population-based metaheuristic optimization algorithms inspired by biological evolution. These algorithms have proven effective for solving complex optimization problems where traditional methods struggle, including function approximation, symbolic regression, and game strategy optimization @Goldberg1989Genetic.

Darwin addresses the challenge of evolving game-playing strategies for complex environments like Generals IO, a turn-based strategy game with a very large state space. The system's innovative approach combines Genetic Algorithms (GA) and Genetic Programming (GP) in a dual population evolution scheme, achieving dynamic function creation for action selection and emulating backpropagation purely through evolutionary algorithms.

== Objectives
- Design and implement a modular EA framework supporting multiple individual representations
- Provide extensible architecture enabling custom problem-specific implementations
- Achieve high performance through concurrent processing
- Demonstrate effectiveness through comprehensive benchmarking and testing
- Wrap Reinforcement Learning environments with a generic GP→RL bridge to enable any game to be played via evolved GP individuals

== Key Features

+ *Generic Evolution Engine*: Type-safe, performant Go implementation using generics and interfaces, enabling evolution of any representation that implements the `Evolvable` interface

+ *Multi-Paradigm Support*: Simultaneous support for GA (bitstring genomes) and GP (tree-based genomes)

+ *Generic GP→RL Bridge*: Universal game wrapper that translates Genetic Programming individuals into Reinforcement Learning environments via TCP, enabling evolution of game-playing strategies

+ *Concurrent Evolution Engine*: Channel-based architecture for parallel offspring generation with minimal overhead

+ *Comprehensive Metrics*: Real-time streaming of evolution progress with CSV export

+ *Configuration-Driven Design*: TOML-based configuration system for flexible parameter specification without code changes

= Background Research

The design of Darwin's dual population evolution approach is grounded in established research demonstrating the effectiveness of evolutionary algorithms for game-playing and neural network training.

Research by Petroski et al. @Petroski2018Deep demonstrates that Genetic Algorithms can effectively emulate backpropagation for training deep neural networks in reinforcement learning environments. Their work shows that GA-based weight optimization provides a competitive alternative to gradient-based methods, avoiding local minima while maintaining performance. This validates the use of GA for evolving weight matrices that modulate action selection in game environments.

Genetic Programming has been successfully applied to game-playing scenarios, as demonstrated by Gold et al. @Gold2023Genetic in their work on evolving decision trees for Bomberman using adversarial GP. However, the application of GP to complex RL environments like Generals IO remains underexplored, particularly when combined with GA for weight optimization.

Darwin's approach combines these established techniques: GP creates dynamic decision-making functions that respond to game state, while GA emulates backpropagation by evolving weight matrices that guide action selection. This dual population evolution scheme leverages the strengths of both paradigms, with GP providing expressiveness for function creation and GA providing efficient weight optimization.


= Design and Planning

== Architecture Overview

Darwin follows a modular, interface-driven architecture that promotes extensibility and maintainability. The system is organized into distinct packages, each responsible for a specific aspect of the evolutionary process. As is standard in Go, communication is done via message-passing through channels to ensure concurrency safety and to avoid mutex contention.

=== User Activity 

The user activity diagram illustrates how users interact with the Darwin system:

#figure(
  image("user-activity.drawio.png", width: 100%),
  caption: [User activity flow showing configuration and bridge startup occurring in parallel before the evolution engine begins.]
) <user-arch>


=== System architecture

The system architecture diagram illustrates the high-level components and their interactions:

#figure(
  image("system-architecture.drawio.png", width: 100%),
  caption: [System architecture diagram showing multi-engine and multi-client ability of the Bridge RL environment wrapper and isolated evolution per client.]
) <sys-arch>

=== Core Components

The architecture consists of the following primary components:

==== Generic Evolution Engine
(`internal/evolution/`): A performant, type-safe evolution engine built on Go's interface-based generics. The engine operates generically on any type implementing the `Evolvable` interface. It orchestrates the evolutionary process using a channel-based command pattern, processing generation commands asynchronously with concurrent offspring generation that scales linearly with CPU cores. The generic design enables the same engine code to evolve bitstrings, trees, grammar trees, and action trees without modification.

==== Individual Representations
(`internal/individual/`): Implements the `Evolvable` interface, defining the contract for all individual types. Four implementations are provided:
- `BinaryIndividual`: Bitstring genomes for traditional GA problems
- `Tree`: Standard GP tree structures for symbolic regression
- `TreeGenomeGE`: Grammar Evolution implementation mapping integer genomes to trees via grammar rules
- `ActionTreeIndividual`: Specialized representation for game-playing with weighted action selection

==== Selection Mechanisms
(`internal/selection/`): Provides pluggable selection strategies:
- `RouletteSelector`: Fitness-proportional selection
- `TournamentSelector`: Tournament-based selection with configurable tournament size

==== Fitness Calculation
(`internal/fitness/`): Modular fitness evaluation system supporting:
- Tree-based symbolic regression fitness
- Binary fitness functions
- Action tree game-playing fitness (via the generic GP→RL bridge connecting to game servers)

==== Population Management
(`internal/population/`): Handles population initialization, updates, and fitness calculation coordination.

==== Metrics Collection
(`internal/metrics/`): Asynchronous metrics streaming with CSV export capabilities.

= Implementation 
#include("./implementation.typ")

== Design
==== Rationale for GA
Genetic Algorithms emulate backpropagation by evolving weight matrices that modulate action selection. Research supports using GA to emulate backpropagation for game environments @Petroski2018Deep, providing a performant solution that avoids local minima common in gradient-based methods.

==== Rationale for GP
Genetic Programming creates dynamic functions that make decisions based on the dynamic game state. This underexplored area of GP enables the evolution of adaptive strategies that respond to changing game conditions, keeping the entire system purely evolutionary without requiring neural network architectures.

==== Genetic Programming (GP)
Creates dynamic decision functions that make decisions based on the dynamic game state. Action trees encode game state evaluation functions, with separate trees for different action types (pass, move direction, split).

==== Genetic Algorithms (GA)
Emulates backpropagation by evolving weight matrices that guide action selection. These weights modulate the outputs of action trees, similar to how neural network weights modulate neuron outputs.

*Dual Population Architecture*: The system maintains two parallel populations:
- Action tree population (GP): Evolves the decision-making functions
- Weights population (GA): Evolves the action selection weights

The populations alternate evolution every `switch_training_target_step` generations (default: 10) using an *evolutionary islands* approach.

== Fitness Functions

=== Symbolic Regression

For tree-based individuals, fitness is calculated using mean squared error (MSE) over test cases:

$"fitness" = 1 - ("MSE")/("MSE"_max)$

where $"MSE" = (1)/(n) sum_(i=1)^(n) (y_i - hat(y)_i)^2$ and $"MSE"_max$ normalizes fitness to $[0, 1]$.

Test cases are generated by sampling the target function over a specified domain. The target function is specified in the configuration via `target_function` (default: `"(x^3)*y+y^3"`), with `test_case_count` controlling the number of test cases (default: 3).

=== Binary Fitness

For bitstring individuals, fitness functions can be defined based on problem requirements. Common examples include:
- OneMax: Count of 1-bits
- Trap functions: Deceptive fitness landscapes
- Problem-specific encodings

=== Game-Playing Fitness

Action tree individuals are evaluated through game simulations via the GP→RL bridge. Fitness is determined by:
- Win/loss ratio
- Score achieved
- Performance metrics specific to the game

A critical aspect of Action Tree fitness evaluation is the use of `test_case_count` (default: 3) to ensure robust strategy evolution. Each Action Tree individual plays multiple games (`test_case_count` games) rather than a single game. This multi-game evaluation approach prevents evolution from converging on tactics that exploit specific game states or opponent behaviors in a single game instance. Instead, by evaluating fitness across multiple games with varying initial conditions and opponent behaviors, evolution is guided toward discovering general strategies that perform consistently across diverse game states. This ensures that evolved strategies are robust and transferable, rather than overfitting to particular game scenarios.

Key configuration parameters:
- `test_case_count`: Number of games played per individual for fitness evaluation (default: 3). Higher values promote more robust strategies but increase evaluation time.
- `max_steps`: Maximum game steps per evaluation (default: 1000)
- `opponent_type`: Opponent strategy for evaluation (default: "expander")
- `connection_pool_size`: TCP connection pool size (default: 100)
- `connection_timeout`: Connection timeout duration (default: "30s")
- `health_check_timeout`: Health check timeout (default: "30s")

The system supports connection pooling and health checking for reliable game server communication.

== Metrics and Extensibility

Darwin's metrics system provides real-time streaming with CSV export, enabling detailed analysis of evolution dynamics through rich statistics (best/average/min fitness, population diversity, tree depth, performance metrics). The interface-driven architecture enables extensibility through well-defined interfaces (`Evolvable`, `FitnessCalculator`, `Selector`), allowing customization for diverse problem domains without modifying core system code.

== Parameters and Configuration

Key evolutionary parameters:

#figure(
  table(
    columns: 4,
    align: (left, left, center, left),
    [*Parameter*], [*Description*], [*Default*], [*Range*],
    [`population_size`], [Number of individuals], [20], [10-10000],
    [`generations`], [Evolution iterations], [50], [1-1000],
    [`crossover_rate`], [Crossover probability], [0.7], [0.0-1.0],
    [`mutation_rate`], [Mutation probability], [0.3], [0.0-1.0],
    [`elitism_percentage`], [Elite preservation], [0.01], [0.0-1.0],
    [`crossover_point_count`], [Crossover points], [1], [1-10],
    [`selection_type`], [Selection method], ["tournament"], ["tournament", "roulette"],
    [`selection_size`], [Tournament size], [3], [2-20],
  ),
  caption: [Key evolutionary parameters with their descriptions, default values, and valid ranges.]
) <params-table>

These parameters were selected based on EA literature recommendations @Eiben2003Introduction and empirical tuning.

= Analysis and Evaluation 

== Performance Analysis
#todo(position: right)[Add performance metrics charts: generation time vs. population size, fitness convergence over generations, concurrent speedup vs. core count, memory usage over time.]

The following section will contain detailed performance metrics and analysis:

#figure(
  image("placeholder.png", width: 100%),
  caption: [Placeholder for performance metrics visualization showing evolution engine performance characteristics. Metrics to include: generation time vs. population size, fitness convergence over generations, concurrent speedup vs. core count, memory usage over time.]
) <perf-metrics>


== Resource Metrics

#todo(position: right)[Add resource utilization charts: CPU utilization, memory patterns, goroutine/channel metrics, network bandwidth, GC pause times.]

The following section will contain detailed resource utilization metrics:

#figure(
  image("placeholder.png", width: 100%),
  caption: [Placeholder for resource metrics visualization showing system resource utilization during evolution. Metrics to include: CPU utilization over time, memory allocation patterns, goroutine count and channel usage, network bandwidth (TCP bridge), GC pause times.]
) <resource-metrics>

= Conclusion

Darwin successfully demonstrates that Genetic Programming can evolve effective game-playing strategies for complex environments:

*Primary Achievement*: The system evolves game-playing strategies for Generals IO using GP, demonstrating that evolutionary algorithms can create adaptive decision-making functions that respond to dynamic game state. This represents a significant application of GP to interactive game environments.

*Dual Population Evolution*: The innovative GA+GP hybrid system combines Genetic Algorithms (for weight evolution) and Genetic Programming (for action tree evolution) in a co-evolutionary framework. This dual population approach alternates evolution every 10 generations, enabling both components to co-adapt and improve together, avoiding local minima while maintaining the expressiveness of evolved programs.

*Generic GP→RL Bridge*: The universal GP→RL bridge enables evolution of game-playing strategies by translating GP individuals into RL environments. This generic translation layer is environment-agnostic, allowing application to diverse game domains without modifying the core evolution engine. The bridge provides a framework to map reinforcement learning libraries to pure GP function approximation, connecting Python RL environments to Go evolution via TCP.

*Framework Emergence*: In pursuit of the game-playing goal, a flexible evolutionary computation framework emerged that supports multiple paradigms. While this framework enables symbolic regression and other applications, its primary purpose is to support the evolution of game-playing strategies efficiently.

The system has been validated through comprehensive testing and benchmarking, demonstrating effectiveness in evolving game-playing strategies for Generals IO. The dual population evolution approach successfully combines GA and GP to create adaptive strategies. While the framework that emerged supports symbolic regression and other applications, the primary achievement is demonstrating that GP can effectively evolve game-playing strategies for complex interactive environments.

Future enhancements could include:
- Distributed evolution across multiple machines
- Adversarial (Individual vs. Individual) training modes 
- Integration with additional game environments via the generic GP→RL bridge 

#bibliography("bibliography.bib")
