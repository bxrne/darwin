#import "@preview/dashy-todo:0.1.3": todo

=== Design Principles

The architecture adheres to several key design principles:

+ *Interface Segregation*: The `Evolvable` interface defines only essential operations, allowing diverse implementations without unnecessary coupling.

+ *Dependency Inversion*: High-level modules (evolution engine) depend on abstractions (Evolvable interface, Selector interface) rather than concrete implementations.

+ *Single Responsibility*: Each package has a clearly defined responsibility, promoting maintainability and testability.

+ *Concurrency Safety*: The channel-based architecture ensures thread-safe operation while enabling parallel offspring generation.

== Language Choice Rationale

=== Evolution engine

Go was selected as the implementation language for several reasons:

*Generic Type System*: Go's interface-based generics enable a truly generic evolution engine. The `Evolvable` interface allows the evolution engine to operate on any type that implements genetic operations, providing type safety at compile time while maintaining runtime flexibility. This eliminates the need for code generation or runtime type assertions, resulting in both safety and performance.

*Concurrency Support*: Go's goroutines and channels provide elegant primitives for concurrent evolutionary operations, enabling efficient parallel offspring generation without complex thread management. The channel-based architecture ensures thread-safe operation with minimal synchronization overhead.

*Performance*: Go compiles to native code, providing performance comparable to C/C++ while maintaining higher-level abstractions. This is crucial for EA systems where fitness evaluation may be computationally expensive. The zero-cost abstractions of interfaces and goroutines ensure minimal runtime overhead.

*Type Safety*: Go's interface system enables compile-time type checking while maintaining flexibility through structural typing. The generic evolution engine benefits from this, ensuring that all operations are type-safe without sacrificing extensibility.

*Tooling*: Excellent tooling including `go test` for benchmarking, `go vet` for static analysis, and built-in profiling support.

=== Bridge 

The GP→RL bridge is implemented in Python due to the following reasons:

*RL Ecosystem*: The target game environment (Generals IO) uses PettingZoo, a Python-based multi-agent RL library. Python has the most mature ecosystem for reinforcement learning environments, with extensive support for game environments and RL frameworks.

*Environment Integration*: The bridge needs to interface directly with PettingZoo environments, which are Python-native. Implementing the bridge in Python eliminates the need for language bindings or complex inter-process communication layers, simplifying the integration.

*Multiprocessing*: Python's multiprocessing module enables true parallelism for concurrent game simulations, bypassing the Global Interpreter Lock (GIL) that limits threading. This allows the bridge to run multiple game instances in parallel, essential for efficient fitness evaluation.

*Scope*: The bridge is a means to proving the concept of evolving game-playing strategies via GP. Python's ease of use and rapid development capabilities make it suitable for this purpose. Future iterations could consider re-implementing the bridge in Go if scale or performance demands increase.


= Implementation Details

== Generic Evolution Engine

The evolution engine is implemented as a *generic*, type-safe system using Go's interface-based generics. This design enables the engine to evolve any representation type that implements the `Evolvable` interface, without requiring code duplication or runtime type assertions. The engine receives commands through a command channel and processes generations asynchronously, leveraging Go's concurrency primitives for optimal performance.

The `Evolvable` interface serves as the foundation for the generic evolution engine:

```go
type Evolvable interface {
    Mutate(rate float64, mutateInformation *MutateInformation)
    Max(i2 Evolvable) Evolvable
    MultiPointCrossover(i2 Evolvable, crossoverInformation *CrossoverInformation) (Evolvable, Evolvable)
    GetFitness() float64
    SetFitness(fitness float64)
    Clone() Evolvable
    Describe() string
}
```

This generic interface enables:
- *Zero Code Duplication*: The same evolution engine code evolves bitstrings, trees, grammar trees, and action trees without modification
- *Compile-Time Type Safety*: Go's interface system ensures all operations are type-safe at compile time, eliminating runtime type assertions
- *Extensibility*: Adding new individual types requires only implementing the `Evolvable` interface—no changes to core evolution logic


=== Generation Processing

Each generation follows this sequence:

1. *Fitness Calculation*: For generation 1, calculate fitness for the initial population. Subsequent generations inherit fitness from parent selection.

2. *Population Sorting*: Sort population by fitness (descending) to identify elite individuals.

3. *Elitism*: Preserve the top $"k"$ individuals, where $"k" = ceil("population_size" times "elitism_percentage")$.

4. *Offspring Generation*: Generate remaining individuals through:
   - Parent selection (roulette or tournament)
   - Crossover with probability $"p"_c$ (default 0.7)
   - Mutation with probability $"p"_m$ (default 0.3)
   - Fitness evaluation

5. *Metrics Collection*: Calculate and stream generation metrics including best/average/min fitness, population statistics, and execution time.

=== Generic Concurrent Offspring Generation

The evolution engine's generic design enables concurrent offspring generation for any `Evolvable` type:

```go
for i := 0; i < offspringNeeded; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        ee.generateOffspring(cmd, offspringChan)
    }()
}
```

This design maximizes CPU utilization across all available cores, with the generic interface ensuring type-safe operations regardless of the underlying representation. The channel-based communication provides efficient synchronization without mutex contention, resulting in near-linear scaling with CPU count. 

The generic evolution engine's innovation lies in achieving both type safety and performance through Go's interface-based generics, enabling a single codebase to evolve diverse representations without code duplication or runtime overhead. The generic design eliminates the traditional trade-off between flexibility and performance. By leveraging Go's interface-based generics, the engine achieves compile-time type safety while maintaining runtime efficiency comparable to specialized implementations. The concurrent architecture scales linearly with CPU count, demonstrating that generic design need not sacrifice performance. This validates that interface-based generics in Go provide both type safety and performance, enabling a single evolution engine to handle bitstrings, trees, grammar trees, and action trees. 

== Genetic Operators

=== Crossover

Darwin implements multi-point crossover, configurable through `crossover_point_count`. For bitstring individuals, crossover points divide the genome into segments that are alternately inherited from parents. For tree individuals, crossover exchanges subtrees between parents.

The crossover rate $"p"_c$ determines the probability of performing crossover versus direct mutation. Default: $"p"_c = 0.7$.

=== Mutation

Mutation introduces diversity into the population. For bitstring individuals, mutation flips bits with probability $p_m$. For tree individuals, mutation can:
- Replace a subtree with a randomly generated one
- Modify node values within depth constraints
- Swap subtrees

The mutation rate $p_m$ controls the probability of mutation per individual. Default: $p_m = 0.3$.

=== Selection

Two selection mechanisms are implemented:

+ *Roulette Wheel Selection*: Fitness-proportional selection where each individual's selection probability is proportional to its fitness relative to total population fitness.

+ *Tournament Selection*: Randomly selects $k$ individuals (tournament size) and returns the fittest. Tournament size is configurable (default: 3-7).

Tournament selection provides better control over selection pressure and is less sensitive to fitness scaling issues.

== Individual Representations

=== Bitstring Individuals

Traditional GA representation using binary genomes. Suitable for problems where solutions can be encoded as fixed-length bitstrings (e.g., knapsack, subset selection).

Key parameters:
- `genome_size`: Length of binary genome

=== Tree Individuals (GP)

Standard Genetic Programming representation using expression trees. Nodes represent functions (operators) or terminals (variables/constants).

Key parameters:
- `max_depth`: Maximum tree depth (default: 8)
- `initial_depth`: Minimum initial tree depth (default: 4)
- `operand_set`: Available operators (e.g., `["+", "-", "*", "/", "^"]`)
- `terminal_set`: Terminal values (constants)
- `variable_set`: Variable names (e.g., game state variables for action trees)

=== Grammar Evolution

Grammar Evolution (GE) maps integer genomes to expression trees via a context-free grammar. This approach combines the search efficiency of integer genomes with the expressiveness of tree structures.

The GE implementation:
1. Defines a grammar mapping non-terminals to production rules
2. Uses integer codons (genome values) to select production rules
3. Expands the grammar starting from the start symbol to generate trees
4. Handles depth limits by forcing terminal selection at maximum depth

This enables evolution of programs that conform to specific syntactic constraints, useful for symbolic regression where certain function combinations may be invalid.

Darwin's Grammar Evolution implementation represents an innovation in constraint-guided program evolution. Unlike standard GP which may generate syntactically invalid programs, GE ensures all evolved programs conform to domain-specific constraints through grammar rules. This approach enables:

- *Constraint satisfaction*: Evolved programs automatically satisfy syntactic constraints, eliminating invalid solutions from the search space
- *Domain knowledge encoding*: Grammar rules can encode domain expertise, focusing evolution on meaningful solutions
- *Reduced search space*: Grammar constraints dramatically reduce the search space, accelerating convergence

This approach is particularly valuable for symbolic regression where certain function combinations are mathematically invalid, demonstrating how grammar-guided evolution can outperform standard GP for constrained problems.

=== Action Tree Individuals and Dual Population Evolution

#todo(position: right)[Add gameplay GIF/video demonstrating evolved strategies in action against Generals IO opponents.]

Action Tree individuals represent a novel approach to evolving game-playing strategies for complex game environments like Generals IO. The system employs a *dual population evolution* scheme that combines Genetic Algorithms and Genetic Programming:

 This modulation mechanism creates distinct evolutionary phases:

- *Action Tree Evolution Phase*: During this phase, the action tree population (GP) evolves while the weights population remains static. This allows action trees to explore new decision-making functions and adapt to the current weight configuration, discovering novel strategies that leverage the existing weight structure.

- *Weight Evolution Phase*: During this phase, the weights population (GA) evolves while the action trees remain static. This allows weights to optimize action selection policies for the current set of action trees, fine-tuning how different action types are prioritized and combined.

- *Modulation Effect*: By alternating between these phases, the system modulates the evolutionary pressure on each component. When action trees evolve, they must work with the current weights, preventing them from overfitting to a specific weight configuration. When weights evolve, they must optimize for the current action trees, preventing them from exploiting weaknesses in a single tree structure. This modulation creates a stabilizing effect that prevents premature convergence and encourages robust co-adaptation.

- *Co-Adaptation*: Over multiple alternation cycles, both populations co-evolve toward complementary solutions. Action trees evolve to produce outputs that work well with the evolving weight matrices, while weights evolve to effectively combine the evolving tree outputs. This co-adaptation enables the discovery of sophisticated strategies that neither component could achieve independently.

This evolutionary islands approach enables the system to evolve complex game strategies that adapt to opponent behavior, demonstrating that evolutionary algorithms can effectively combine different paradigms for superior performance. The modulation mechanism ensures that both components evolve in harmony rather than competing or diverging, resulting in robust, generalizable strategies. The approach avoids local minima that plague pure GP approaches while maintaining the expressiveness of evolved programs.

== Generic GP→RL Bridge

The fitness evaluation leverages a *generic GP→RL bridge* (`game/bridge.py`), a universal wrapper that translates GP individuals into Reinforcement Learning environments. This bridge provides:

=== Standard RL Interface
The bridge exposes a standard RL-like API via TCP, accepting observations and returning actions, rewards, and termination signals. This enables any GP system to interact with RL environments without modification.

=== Protocol Abstraction
The bridge handles the translation between GP tree evaluation (which produces action vectors) and RL environment requirements (which expect structured actions). This abstraction allows the evolution engine to remain agnostic to the specific game being played.

=== Concurrent Game Execution
The bridge supports multiple concurrent game simulations through multiprocessing, enabling parallel fitness evaluation of multiple individuals simultaneously. This maximizes throughput when evaluating game-playing strategies.

=== Environment Agnostic
The bridge design is game-agnostic—it can wrap any PettingZoo-style environment, making it a truly generic GP→RL translation layer. The current implementation uses the Generals game environment, but the bridge can be adapted to any game following the standard interface.

The *generic GP→RL bridge* represents a novel solution to the fundamental challenge of connecting Genetic Programming systems with Reinforcement Learning environments. While GP traditionally operates on static fitness functions, the bridge enables GP to evolve strategies for interactive, dynamic environments—a largely underexplored application area. The bridge solves the paradigm mismatch between GP (which produces functions) and RL (which requires interactive agents). By providing a universal translation layer, the bridge enables any GP system to evolve game-playing strategies without modification. This generic design means the bridge can wrap any PettingZoo-style environment, making it a truly universal GP→RL translation layer.


