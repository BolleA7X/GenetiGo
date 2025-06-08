// Package ga provides a framework for solving problems with a genetic algorithm.
package ga

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/BolleA7X/GenetiGo/internal/batching"
)

// A Member represents a single individual of the population.
// This interface must be implemented by a custom data type to be used by this
// package.
type Member interface {
	GetFitnessScore() uint32          // Getter of the FitnessScore attribute
	GetSurvivalChance() float32       // Getter of the SurvivalChance attribute
	SetSurvivalChance(chance float32) // Setter of the SurvivalChance attribute
	Distance(other Member) float64    // Defines how the distance between two Members is computed
	ComputeAndSetFitnessScore()       // Defines how the fitness score is computed for this Member
	Crossover(other Member) Member    // Defines how the crossover operator works for this Member
	Mutate()                          // Defines how the mutation operator works for this Member
}

// MemberData contains the information that every Member must have: the fitness
// score and the survival chance. This struct is meant to be embedded in the
// custom data type that implements the Member interface, but it's not mandatory.
type MemberData struct {
	FitnessScore   uint32
	SurvivalChance float32
}

// SolverOptions is a container for all the user-defined options and parameters
// the user can set to change the behaviour of the genetic algorithm.
type SolverOptions struct {
	PopulationSize uint32  // Number of members at each generation
	MaxGenerations uint32  // Maximum number of generations to simulate
	MutationChance float32 // Chance that a member of the population randomly mutates
	NBatches       uint32  // Population is divided into batches, each batch managed by a separate goroutine
	Speciation     bool    // Enables speciation (usually disabled, tipically used in NEAT)
	Verbose        bool    // Enable verbose output on stdout
}

// Solver contains the options and current state of a single run of the
// genetic algorithm.
type Solver[M Member] struct {
	population   []M
	options      SolverOptions
	generationNo uint32
}

// NewSolver instantiates a new Solver object. The user must provide the desired
// options and a list of Members that represent the first generation. It panics
// if the population size set in the options is zero or if it doesn't match
// the number of members of the first generation.
func NewSolver[M Member](members []M, options SolverOptions) *Solver[M] {
	// Population size check

	if options.PopulationSize == 0 {
		panic("population size must be greater than zero")
	}
	if uint32(len(members)) != options.PopulationSize {
		var panicString = fmt.Sprintf("initial popultation doesn't match population size (expected %d, found %d)",
			options.PopulationSize, len(members))
		panic(panicString)
	}

	// Batches limiting (min: 1; max: 1 for each single element of the population)

	if options.NBatches == 0 {
		options.NBatches = 1
	}
	if options.NBatches > options.PopulationSize {
		options.NBatches = options.PopulationSize
	}

	// Create and return the new Solver

	var solver = &Solver[M]{}
	solver.population = members
	solver.options = options
	solver.generationNo = 1
	return solver
}

// pickRandomMember returns a random Member from the current population. Members
// with a higher survival chance have a higher probability of being picked.
func (solver *Solver[M]) pickRandomMember() M {
	var randValue = rand.Float32()
	for _, member := range solver.population {
		var survivalChance = member.GetSurvivalChance()
		if randValue < survivalChance {
			return member
		}
		randValue -= survivalChance
	}

	return solver.population[solver.options.PopulationSize-1] // Should never be reached
}

// getBestMember returns the Member with the highest fitness score in the current
// population.
func (solver *Solver[M]) getBestMember() M {
	return slices.MaxFunc(solver.population, func(a, b M) int {
		return int(a.GetFitnessScore()) - int(b.GetFitnessScore())
	})
}

// Solve executes the genetic algorithm, up to the last generation. It returns
// the Member of the last generation with the highest fitness score.
func (solver *Solver[M]) Solve() M {
	// Verbose mode
	if solver.options.Verbose {
		defer fmt.Println()
	}

	// Build the list of batches

	var batches = batching.BuildBatchesList(solver.options.PopulationSize, solver.options.NBatches)

	// Verbose mode
	if solver.options.Verbose {
		fmt.Printf("GenetiGo - GA solver\n")
		fmt.Printf(""+
			"\tPopulation size: %d\n"+
			"\tMax generation: %d\n"+
			"\tMutation chance: %.2f\n"+
			"\tNumber of jobs/batches: %d\n\n",
			solver.options.PopulationSize, solver.options.MaxGenerations,
			solver.options.MutationChance, solver.options.NBatches,
		)
	}

	for {
		var wg sync.WaitGroup

		// If speciation is enabled, compute the distances between each couple of members (parallelized)

		if solver.options.Speciation {
			for _, batch := range batches {
				wg.Add(1)
				go func(b batching.BatchInfo) {
					defer wg.Done()
					for i := b.Start; i < b.End; i++ {
						for j := range solver.population {
							solver.population[i].Distance(solver.population[j])
						}
					}
				}(batch)
			}
		}

		// Compute the fitness score of each individual (parallelized)

		var fitnessSum uint32 = 0 // Sum is atomically computed
		for _, batch := range batches {
			wg.Add(1)
			go func(b batching.BatchInfo) {
				defer wg.Done()
				for i := b.Start; i < b.End; i++ {
					solver.population[i].ComputeAndSetFitnessScore()
					atomic.AddUint32(&fitnessSum, solver.population[i].GetFitnessScore())
				}
			}(batch)
		}
		wg.Wait()

		// Verbose mode
		if solver.options.Verbose {
			var bestMember = solver.getBestMember()
			fmt.Printf("\r                                                                                            ")
			fmt.Printf("\r[GENERATION %d] Best fitness score: %d", solver.generationNo, bestMember.GetFitnessScore())
		}

		// Return the best member at the last generation
		if solver.generationNo == solver.options.MaxGenerations {
			var bestMember = solver.getBestMember()
			return bestMember
		}

		// Compute the survival chance of each individual (serial)

		for i := range solver.population {
			var fitnessScore = solver.population[i].GetFitnessScore()
			var survivalChance float32 = 0
			if fitnessSum != 0 {
				survivalChance = float32(fitnessScore) / float32(fitnessSum)
			} else {
				survivalChance = 1.0 / float32(solver.options.PopulationSize)
			}
			solver.population[i].SetSurvivalChance(survivalChance)
		}

		// Generate children by applying the crossover and mutation operators (parallelized)

		var nextPopulation = make([]M, solver.options.PopulationSize)
		for _, batch := range batches {
			wg.Add(1)
			go func(b batching.BatchInfo) {
				defer wg.Done()
				for i := b.Start; i < b.End; i++ {
					var parent1 = solver.pickRandomMember()
					var parent2 = solver.pickRandomMember()
					var child = parent1.Crossover(parent2).(M)
					if rand.Float32() < solver.options.MutationChance {
						child.Mutate()
					}
					nextPopulation[i] = child // no need for mutex because each goroutine accesses a separate subset of the slice
				}
			}(batch)
		}
		wg.Wait()

		// Update the population and generation number

		solver.population = nextPopulation
		solver.generationNo++
	}
}
