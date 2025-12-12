#import "@preview/arkheion:0.1.1": arkheion, arkheion-appendices
#import "@preview/fletcher:0.5.8" as fletcher: diagram, node, edge
#import "@preview/fletcher:0.5.8": shapes
#import "@preview/dashy-todo:0.1.3": todo

#show: arkheion.with(
  title: "Towards game playing via Genetic Programming and Backpropagation emulation",
  authors: (
    (name: "Adam Byrne", email: "22338004@studentmail.ul.ie", affiliation: "University of Limerick"),
    (name: "Art O'Liathain", email: "22363092@studentmail.ul.ie", affiliation: "University of Limerick"),
  ),
  keywords: ("Evolutionary Algorithms", "Genetic Programming", "Grammar Evolution", "Symbolic Regression", "Concurrent Computing"),
  date: "December 2025",
)

#show link: underline

#outline(
  title: "List of Figures",
  target: figure.where(kind: image)
)
#outline(
  title: "List of Tables", 
  target: figure.where(kind: table)
)
#pagebreak()

= Introduction - The Problem

Evolutionary Algorithms (EAs) represent a class of population-based metaheuristic optimization algorithms inspired by biological evolution. These algorithms have proven effective for solving complex optimization problems where traditional methods struggle, including function approximation, symbolic regression, and game strategy optimization @Goldberg1989Genetic.

The focus of this paper will be on the application of EA to the complex game environment generals.io. The standard EA approach to solving this could be something like NEAT @stanley_evolving_2002 which creates an artificial neural network, or decision trees @Gold2023Genetic which select predetermined actions.
Instead of those approaches Darwin takes a novel approach leveraging the efficacy that EA have for function approximation and optimization to create an EA that co evolves functions and weights to determine optimal actions within a game.

= Rationale
The core idea for this approach comes from both reinforcement learning. Where a function along with weights get modified and altered to best interact with the environment. Using this idea as a base Darwin evolves two distinct populations, a population of functions and weights, each with their own EA.

There is inspiration taken from multi-tree individual EA approaches as well, where one individual would have multiple trees per potential action allowing each tree to capture heuristics relevant to each action option. This while promising has an issue with ballooning individuals, since as the state space increases the number of trees increase increasing the amount of compute needed per individual. While this approach is impractical, the inspiration taken from this is to use the weights' interactions with the trees to emulate as if there were multiple trees and the learned values of the weights would become the meta heuristics.

== Genetic Algorithms - The Weights
Research by Petroski et al. @Petroski2018Deep demonstrates that Genetic Algorithms can effectively emulate backpropagation for training deep neural networks in reinforcement learning environments. Their work shows that GA-based weight optimization provides a competitive alternative to gradient-based methods, avoiding local minima while maintaining performance.
This made it a simple choice to select GA as the EA for the weights

== Genetic Programming - The Functions
Research by Koza et al. @Koza1992Genetic demonstrates how efficiently GP can create functions from simple operators to mimic the outputs of complex operators and functions. Their work shows that GP can be an effective tool to emulate the behaviour of unknown functions using EA. 
This serves as the basis for the decision logic as the functions can evolve to learn the correct patterns alongside the weights. To create a strategy that can consistently win the game.

== How action selection works
The method for action selection begins with one weights individual(WI) and one tree individual(TI). The TI contains a function that contains constants, variables from the environment and weights. The WI has a row of weights for every potential action allowed, with distinct values. 
The TI function at point T is given the variables from the environment and for every row in weights the function is evaluated. As every row of the weights contain distinct values for every potential action option each action output is varied and can be interpreted as the "Score" for that action.
Softmax is then preformed on the scores to determine the proportion of each value to the whole, from which a roulette wheel is spun to determine the action chosen with weighing based on the action "score".

== Fitness function and Root Squared error
As the goal of Darwin is to use a novel approach to emulate reinforcement learning the reward function given to a RL agent was selected as the fitness function. In the generals environment @straka_strakamgenerals-bots_2025 there are multiple reward functions available to the user two of which are of note, FrequentRewardFunc, LandRewardFunc.
Through testing a limitation of the approach was discovered. FrequentRewardFunc while well suited for reinforcement learning as every action was rewarded to some degree but some actions were rewarded more than others. This is integral to RL as it can learn on a per action basis to optimise whereas with the approach Darwin employs only the final reward was taken into consideration. This lead to suboptimal strategies being rewarded in lieu of a real strategy being developed. This lead to LandRewardFunc being better suited to EA as its final reward is the number of land tiles owned by the agent. This meant that the final state was most important and encouraged the agents to learn a more robust strategy.

Another issue encountered was as the agent improved it would occasionally win, and this while a desirable outcome was usually based on luck rather than skill. Due to how high the reward for winning was it lead to the EA learning from one lucky tree rather than a solid strategy. To combat this Mean Root Error was used. This is an where the sum of all roots of all attempts by the same individual are added together while preserving the sign. This meant that negative and low rewards had an impact on the final result while still rewarding agents who won. This encouraged strategies with consistent rewards rather than one off victories to be evolved and used over time.

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

Darwin's metrics system provides real-time streaming with CSV export, enabling detailed analysis of evolution dynamics through rich statistics (best/average/min of user defined statistics, population diversity, tree depth, performance metrics). The interface-driven architecture enables extensibility through well-defined interfaces (`Evolvable`, `FitnessCalculator`, `Selector`), allowing customization for diverse problem domains without modifying core system code.


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

These parameters were selected based on empirical tuning and EA literature best practices.

= Analysis and Evaluation 
== Case Study
These are metrics collected from 220 generations of the default config with 30 tree population and 15 weight population:

=== Population Diversity 
Pictured in @treeDepthGraph which pictures the depth statistics of trees and @treeNodeGraph picturing the node count of trees. These figures show a strong population diversity among the trees persisting throughout the 100+ generations they were trained on. This shows that while initially full tree were generated and prevalent, smaller grow trees produce better fitness scores and the mutation and crossover allow that to happen.
#figure(image("images/2025-12-12-18:01:12.png", width: 80%), caption: [Tree Depth Graph]) <treeDepthGraph>
#figure(image("images/2025-12-12-18:02:00.png", width: 80%),caption: [Tree Node Graph]) <treeNodeGraph>

The following figures show the count of variables and operands used within the nodes of the trees, @WeightsUsageGraph, @operandUsageGraph, @observationUsageGraph.
Worryingly when looking at these graphs while it is clear that both operands and observations maintain a healthy population spread, weights nearly all get bred out and the model relies on one weight for every decision across all trees. This leads to weight collapse.
This is a terrible sign as this would lead to deterministic results regardless of the other operands as this would be the one differentiating variable between actions. This behaviour leads the individuals to rely on the randomness of softmax rather than deterministic decisions leading to randomness over strategy being rewarded.
To combat this penalties for only having one weight and punishing individuals relying on randomness is the ideal solution to enforce strategy learning over weight collapse.
#figure(image("images/2025-12-12-18:02:13.png", width: 80%), caption: [Weights Usage Graph]) <WeightsUsageGraph>
#figure(image("images/2025-12-12-18:02:30.png", width: 80%), caption: [Operand usage Graph]) <operandUsageGraph>
#figure(image("images/2025-12-12-18:02:49.png", width: 80%), caption: [Observation usage Graph]) <observationUsageGraph>

The fitnesses @fitnessGraph shows that the individuals stagnate in learning after reaching about 15 generations in. This can be explained by the weight collapse observed in @WeightsUsageGraph.
The correlation between the individuals increasing fitness and the weight collapse leads one to believe that if accurate safeguards were in place continued growth instead of stagnation could be observed. This leaves the question that further refinement of parameters and safeguards could take this novel method into a useable EA approach in future.

#figure(image("images/2025-12-12-18:02:06.png", width: 80%), caption: [Fitness Graph]) <fitnessGraph>

== Performance Analysis

This section presents comprehensive benchmark results for the Darwin evolutionary computation framework, covering all 4 supported individual types across various scaling scenarios and comparative problems.

=== System Information

#figure(
  table(
    columns: 2,
    align: (left, left),
    [*Property*], [*Value*],
    [Operating System], [Linux archbook 6.17.7-arch1-Watanare-T2-1-t2 (x86_64)],
    [CPU], [Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz],
    [Memory], [31 GB total, 25 GB available],
    [Go Version], [go1.25.4 X:nodwarf5 linux/amd64],
    [Architecture], [x86_64 GNU/Linux],
    [Benchmark Date], [December 12, 2025],
  ),
  caption: [System configuration for benchmark execution.]
) <sys-info>

=== Individual Type Performance

==== BitString Individuals (Classic Genetic Algorithm)

#figure(
  table(
    columns: 6,
    align: (left, center, center, center, center, center),
    [*Benchmark*], [*Population*], [*Genome Size (bits)*], [*Generations*], [*Time (ms)*], [*Memory (MB)*],
    [Small], [100], [50], [10], [3.8], [0.71],
    [Medium], [500], [200], [50], [98.0], [19.5],
    [Large], [1000], [500], [100], [578.4], [126.1],
    [Huge], [500], [5000], [10], [464.6], [48.7],
  ),
  caption: [BitString individual resource usage across scaling scenarios. Excellent linear scaling with predictable performance characteristics.]
) <bitstring-perf>

==== Tree Individuals (Genetic Programming)

#figure(
  table(
    columns: 6,
    align: (left, center, center, center, center, center),
    [*Benchmark*], [*Population*], [*Max Depth*], [*Generations*], [*Time (ms)*], [*Memory (MB)*],
    [Small], [100], [3], [10], [6.8], [0.91],
    [Medium], [500], [5], [50], [147.2], [21.3],
    [Large], [1000], [7], [100], [554.1], [85.3],
    [Huge], [500], [10], [10], [34.8], [4.6],
  ),
  caption: [Tree individual resource usage across scaling scenarios. Good performance with variable memory based on tree depth.]
) <tree-perf>

==== GrammarTree Individuals (Grammar Evolution)

#figure(
  table(
    columns: 6,
    align: (left, center, center, center, center, center),
    [*Benchmark*], [*Population*], [*Genome Size (integers)*], [*Generations*], [*Time (ms)*], [*Memory (MB)*],
    [Small], [100], [50], [10], [6.9], [1.4],
    [Medium], [500], [100], [50], [222.8], [56.1],
    [Large], [1000], [200], [100], [798.3], [354.0],
    [Huge], [500], [500], [10], [95.9], [37.7],
  ),
  caption: [GrammarTree individual resource usage across scaling scenarios. Moderate performance with higher memory usage due to grammar mapping overhead.]
) <grammartree-perf>

==== ActionTree Individuals (Game-Based Evolution)

#figure(
  table(
    columns: 6,
    align: (left, center, center, center, center, center),
    [*Benchmark*], [*Trees*], [*Weights*], [*Generations*], [*Time (ms)*], [*Memory (MB)*],
    [ActionTree], [5], [2], [1], [2876.9], [55.7],
  ),
  caption: [ActionTree individual resource usage. Configuration: 5 trees, 2 weights, 2 test cases, 1 generation. Network-bound performance with game server interaction overhead.]
) <actiontree-perf>

=== Scaling Analysis

#figure(
  table(
    columns: 4,
    align: (left, center, center, center),
    [*Population*], [*BitString (ms)*], [*Tree (ms)*], [*GrammarTree (ms)*],
    [100], [3.8], [6.8], [6.9],
    [300], [36.2], [54.4], [52.8],
    [500], [98.0], [147.2], [222.8],
    [1000], [578.4], [554.1], [798.3],
  ),
  caption: [Performance scaling with population size across individual types. BitString shows excellent linear scaling, while Tree and GrammarTree exhibit good scaling characteristics.]
) <scaling>


== Resource Metrics

This section presents detailed resource utilization metrics and memory efficiency analysis across all individual types.

=== Memory Efficiency

#figure(
  table(
    columns: 4,
    align: (left, center, center, center),
    [*Individual Type*], [*Small (MB/1000)*], [*Medium (MB/1000)*], [*Large (MB/1000)*],
    [BitString], [7.1], [39.0], [126.1],
    [Tree], [9.1], [42.6], [85.3],
    [GrammarTree], [14.3], [112.2], [354.0],
  ),
  caption: [Memory efficiency per 1000 individuals across different scale scenarios. BitString demonstrates excellent memory efficiency with fixed allocation patterns, while Tree and GrammarTree show variable memory usage.]
) <memory-efficiency>

=== Performance Characteristics

*Time Complexity:*
- *BitString*: O(n×g) time complexity with linear scaling, where n is population size and g is generations
- *Tree*: O(n×d×g) where d is tree depth, with good performance on shallow trees
- *GrammarTree*: O(n×k×g) where k is genome size, with moderate overhead from grammar mapping
- *ActionTree*: Network-bound performance with game server interaction overhead

*Memory Patterns:*
- *BitString*: Fixed allocation with predictable memory usage (~130KB per individual)
- *Tree*: Dynamic tree allocation with variable memory based on tree structure
- *GrammarTree*: Grammar mapping overhead results in higher memory consumption
- *ActionTree*: Matrix operations and game state management result in higher memory usage

=== Fitness Performance

#figure(
  table(
    columns: 4,
    align: (left, center, center, center),
    [*Individual Type*], [*Best Fitness*], [*Avg Fitness*], [*Min Fitness*],
    [BitString], [0.940], [0.899], [0.820],
    [Tree], [1.000], [0.954], [0.737],
    [GrammarTree], [1.000], [0.998], [0.898],
  ),
  caption: [Fitness performance across individual types on symbolic regression problem (`x + y`). Tree and GrammarTree achieve perfect fitness (1.000), demonstrating excellent convergence. BitString shows strong performance with fitness approaching 1.0.]
) <fitness-perf>

== Future Work
Future improvements on the project would be:
- An adversarial training mode on the project where individuals train against each other
- More fine tuned parameter testing to find optimal parameters
- Implementing safeguards against weight collapse
- Modifying fitness functions to give a per tree fitness
- Adapting the code to work with continuous RL environments
- Investigate backpropagation for weights over GA
These are features that could bring this approach to the table when compared to other EA approaches to games in the same manner as @Petroski2018Deep. Laying the groundwork for a more blended approach to AI and EA where EA emulates AI.

= Conclusion

Darwin successfully demonstrates that Genetic Programming can evolve effective game-playing strategies for complex environments:

*Primary Achievement*: The system evolves game-playing strategies for Generals IO using a co-evolution between GP and GA, demonstrating that evolutionary algorithms can create adaptive decision-making functions that respond to dynamic game state. This shows a novel approach to a well explored problem in the field.

*Dual Population Evolution*: The innovative GA+GP hybrid system combines Genetic Algorithms (for weight evolution) and Genetic Programming (for action tree evolution) in a co-evolutionary framework. Both components to co-adapt and improve together, avoiding local minima while maintaining the expressiveness of evolved programs.

*Generic GP→RL Bridge*: The universal GP→RL bridge enables evolution of game-playing strategies by translating GP individuals into RL environments. This generic translation layer is environment-agnostic, allowing application to diverse game domains without modifying the core evolution engine. The bridge provides a framework to map reinforcement learning libraries to pure GP function approximation, connecting Python RL environments to Go evolution via TCP.

*Framework Emergence*: In pursuit of the game-playing goal, a flexible evolutionary computation framework emerged that supports multiple paradigms. While this framework enables symbolic regression and other applications, its primary purpose is to support the evolution of game-playing strategies efficiently.

The system has been validated through comprehensive testing and benchmarking, demonstrating effectiveness in evolving game-playing strategies for Generals IO. The dual population evolution approach successfully combines GA and GP to create adaptive strategies. While the framework that emerged supports symbolic regression and other applications, the primary achievement is demonstrating that GP can effectively evolve game-playing strategies for complex interactive environments.

#bibliography("bibliography.bib")
