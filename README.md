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
