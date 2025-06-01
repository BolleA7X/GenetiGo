# GenetiGo
A framework for genetic algorithms in Go.

## INTRODUCTION

A Genetic Algorithm (GA) is a metaheuristic for solving optimization problems,
inspired by natural selection.

The idea is to have multiple tentative solutions to the problem. At each iteration,
these solutions are evaluated and their chance of surival depends on how "good"
they are. Solutions are picked randomly (depending on their chance of survival)
and bred to create children. These children compose the next generation, upon
which the same operations are executed at the next iteration of the algorithm.

Doing so allows the tentative solutions to improve with each iteration of the
algorithm.

## KEY TERMS

**Individual** or **Member**: A single tentative solution to the
problem.

**Population**: Group of solutions, representing a _generation_.

**Generation**: An iteration of the algorithm.

**Genes**: Components of a tentative solution or decision variables of the problem
(usually binary encoded in some ways).

**Fitness**: Metric or score that represents how "good" a tentative solution is.

**Selection**: Process by which the best performing (highest fitness) solutions
are assigned a higher probability of being used for breeding the next _generation_.

**Crossover**: Process by which the _genes_ of two tentative solutions are mixed
to create a new tentative solution, that becomes part of the next _generation_.

**Mutation**: Process by which one or more _genes_ of a tentative solution randomly
change.

## BASIC FLOW

1. Randomly generate the population of the first generation
2. For N generations:

    1. Compute the fitness score of each individual
    2. Select individuals, depending on their fitness, to be the parents
    3. Crossover each couple of parents to create a child
    4. Possibly, mutate the child
    5. Group all children to create the population of the next generation

## USAGE

### PREREQUISITES

This framework requires Go version 1.22 or above. Tested on Go version 1.23.4.

### DEFINING THE PROBLEM

First, import the ```ga``` package of this module, which provides the genetic
algorithm generic implementation.

```
import "github.com/BolleA7X/GenetiGo/ga"
```

Then:

1. Define a struct to represent an individual/member. You can embed the provided
```ga.MemberData``` type to your struct to make sure it has the correct
attributes, or define them by yourself
2. Make your struct implement the ```ga.Member``` interface so that the solver
knows how crossover and mutation work for your specific problem
3. Randomly create a list of individuals to use as the population of the first
generation
4. Create an instance of ```ga.Solver``` by calling the ```ga.NewSolver```
function, passing the first generation and some options as arguments
5. Call the ```Solve``` method of your ```ga.Solver``` instance. This method
returns the individual with the highest fitness at the last generation

The ```ga.NewSolver``` function expects an object of type ```ga.SolverOptions```
as its second argument, allowing you to customize the parameters and behaviour
of the Genetic Algorithm. These options are:

- **PopulationSize**: Number of members at each generation
- **MaxGenerations**: Maximum number of generations to simulate
- **MutationChance**: Chance that a member of the population randomly mutates
(0 <= chance <= 1)
- **NBatches**: Number of batches. Population is divided into batches, where each
batch is managed by a separate goroutine. If <= 1, the solver works in single-threaded
mode. The number of goroutines is limited to the population size.
- **Verbose**: Enable verbose output on stdout

### EXAMPLES

Some examples are provided in the ```examples``` folder. Each example has its
own folder and main function.

To execute an example, run the following command:

```
go run examples/<example_name>/main.go
```

You can also check for race conditions while running the program:

```
go run -race examples/<example_name>/main.go
```
