package main

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/BolleA7X/GenetiGo/ga"
)

const NTASKS uint32 = 15
const MAXTARDINESS uint32 = 1000

type job struct {
	id            uint32
	executionTime uint32
	dueTime       uint32
}

type schedule struct {
	ga.MemberData
	sequence [NTASKS]uint32
}

var JOBS = [NTASKS]job{
	{0, 4, 10},
	{1, 3, 8},
	{2, 2, 7},
	{3, 5, 12},
	{4, 7, 15},
	{5, 3, 14},
	{6, 6, 13},
	{7, 2, 9},
	{8, 4, 11},
	{9, 5, 18},
	{10, 6, 20},
	{11, 2, 10},
	{12, 3, 17},
	{13, 5, 16},
	{14, 4, 19},
}

func computeTardiness(sequence [NTASKS]uint32) uint32 {
	var tardiness uint32 = 0
	var startTime uint32 = 0
	for _, jobId := range sequence {
		var endTime = startTime + JOBS[jobId].executionTime - 1
		var dueTime = JOBS[jobId].dueTime
		if endTime > dueTime {
			tardiness += endTime - dueTime
			if tardiness > MAXTARDINESS {
				tardiness = MAXTARDINESS
			}
		}
		startTime = endTime + 1
	}

	return tardiness
}

func (s *schedule) randomSequence() {
	var permutation = rand.Perm(int(NTASKS))
	for i := range NTASKS {
		s.sequence[i] = uint32(permutation[i])
	}
}

func (s *schedule) GetFitnessScore() uint32 {
	return s.FitnessScore
}

func (s *schedule) GetSurvivalChance() float32 {
	return s.SurvivalChance
}

func (s *schedule) SetSurvivalChance(chance float32) {
	s.SurvivalChance = chance
}

func (s *schedule) IsSolution() bool {
	return false
}

func (s *schedule) ComputeAndSetFitnessScore() {
	var tardiness = computeTardiness(s.sequence)
	// Reward low tardiness
	s.FitnessScore = (MAXTARDINESS + tardiness) * (MAXTARDINESS - tardiness)
}

func (s *schedule) Crossover(other ga.Member) ga.Member {
	var otherSchedule = other.(*schedule)
	var newSchedule = schedule{}
	newSchedule.FitnessScore = 0
	newSchedule.SurvivalChance = 0
	if s.FitnessScore >= otherSchedule.FitnessScore {
		newSchedule.sequence = s.sequence
	} else {
		newSchedule.sequence = otherSchedule.sequence
	}
	return &newSchedule
}

func (s *schedule) Mutate() {
	s.randomSequence()
}

func main() {
	// Set solver options

	var params = ga.SolverOptions{
		PopulationSize: 1000,
		MaxGenerations: 300,
		MutationChance: 0.05,
		NBatches:       10,
		Verbose:        true,
	}

	// Randomly generate initial population

	var initialPopulation = make([]*schedule, 0, params.PopulationSize)
	for range params.PopulationSize {
		var newSchedule = &schedule{}
		newSchedule.FitnessScore = 0
		newSchedule.SurvivalChance = 0
		newSchedule.randomSequence()
		initialPopulation = append(initialPopulation, newSchedule)
	}

	// Solve

	var solver = ga.NewSolver(initialPopulation, params)
	var result = solver.Solve()
	var bestTardiness = uint32(math.Sqrt(float64(MAXTARDINESS*MAXTARDINESS - result.GetFitnessScore())))
	fmt.Println("Best sequence: ", result.sequence)
	fmt.Println("Best tardiness: ", bestTardiness)
}
