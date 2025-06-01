# GA EXAMPLE: SCHEDULING

This example shows how to solve a scheduling problem with a genetic algorithm.

## PROBLEM DEFINITION

Given the following set of jobs:

| Job ID | Exec time | Due time |
| :---: | :---: | :---: |
| 0 | 4 | 10 |
| 1 | 3 | 8 |
| 2 | 2 | 7 |
| 3 | 5 | 12 |
| 4 | 7 | 15 |
| 5 | 3 | 14 |
| 6 | 6 | 13 |
| 7 | 2 | 9 |
| 8 | 4 | 11 |
| 9 | 5 | 18 |
| 10 | 6 | 20 |
| 11 | 2 | 10 |
| 12 | 3 | 17 |
| 13 | 5 | 16 |
| 14 | 4 | 19 |

find the sequence of jobs that minimizes the total tardiness $T$.

The tardiness of the i-eth job $T_i$ is the difference between job's end time
$E_i$ and due time $D_i$ if this difference is positive, 0 otherwise.

$$
T_i = max(0, E_i - D_i)
$$

The total tardiness $T$ is the sum of all jobs tardiness.

$$
T = \sum_{i=0}^{N}{T_i}
$$

In this example time is considered discrete: the start time of a task is given
by the end time of the previously executed task plus 1.

## MEMBER

A tentative solution is a temporal sequence of jobs, representing the proposed
scheduling.

## GENETIC OPERATORS

### FITNESS

The fitness score of the i-eth job $f_i$ is computed such that a low tardiness
is rewarded:

$$
f_i = T_{MAX}^2 - T^2 = (T_{MAX} + T)(T_{MAX} - T)
$$

where $T_{MAX}$ is the maximum tardiness allowed.

### CROSSOVER

The crossover between two members simply consists in inheriting the sequence of
jobs of the fittest member.

### MUTATION

The mutation simply consists in generating a new sequence of jobs randomly.