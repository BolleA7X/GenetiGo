package neat

import (
	"math"
	"math/rand/v2"
	"slices"
	"sync"

	"github.com/BolleA7X/GenetiGo/ga"
	"github.com/BolleA7X/GenetiGo/internal/activation"
)

var innovationNoTracker uint64
var mutationHistory map[connectionNodes]mutation
var mutHistoryMutex sync.Mutex
var testSet DataSet

type node struct {
	id         uint32
	layer      uint32
	activation activation.ActivationFunction
}

type connectionNodes struct {
	senderId   uint32
	receiverId uint32
}

type connection struct {
	endpoints    connectionNodes
	weight       float64
	enabled      bool
	innovationNo uint64
}

type mutationType = uint32

const (
	addNode mutationType = iota
	addConn
)

type mutation struct {
	typeId mutationType
	conn   connection
}

type DataEntry struct {
	input  []float64
	output []float64
}

type DataSet = []DataEntry

type Network struct {
	ga.MemberData
	nodes       []node
	connections []connection
	nInputs     uint32
	nOutputs    uint32
	nSimilar    uint32
	sorted      bool
}

const maxDepth uint32 = 20

func getNewInnovationNumber() uint64 {
	innovationNoTracker++
	return innovationNoTracker - 1
}

func randomWeight() float64 {
	return rand.Float64()*2 - 1
}

func clearMutationHistory() {
	for cn := range mutationHistory {
		delete(mutationHistory, cn)
	}
}

func NewNetwork(nInputs uint32, nOutputs uint32) *Network {
	var newNetwork = &Network{}
	newNetwork.nInputs = nInputs
	newNetwork.nOutputs = nOutputs
	newNetwork.FitnessScore = 0
	newNetwork.SurvivalChance = 0
	newNetwork.nSimilar = 0
	newNetwork.sorted = false

	for i := range nInputs {
		var newNode = node{i, 0, activation.IdentityActivation{}}
		newNetwork.nodes = append(newNetwork.nodes, newNode)
	}

	for i := range nOutputs {
		var newNode = node{i + nInputs, maxDepth, activation.SigmoidActivation{}}
		newNetwork.nodes = append(newNetwork.nodes, newNode)
	}

	return newNetwork
}

func (net *Network) addConnection() {
	// Select two unconnected nodes

	var senderNodeId = rand.IntN(len(net.nodes))
	var receiverNodeId = rand.IntN(len(net.nodes))
	var newEndpoints = connectionNodes{uint32(senderNodeId), uint32(receiverNodeId)}

	// Check if a connection can be made:
	// 1. Not the same node
	// 2. Connection is not backwards
	// 3. Connection not already present

	if senderNodeId == receiverNodeId || net.nodes[senderNodeId].layer >= net.nodes[receiverNodeId].layer {
		return
	}
	for _, conn := range net.connections {
		if conn.endpoints == newEndpoints {
			return
		}
	}

	// Check if such connection was already added in another network
	// If so, copy it
	// Otherwise, create a new connection, add it to the network and add it to the mutation history

	mutHistoryMutex.Lock()

	var mut, present = mutationHistory[newEndpoints]
	if present && mut.typeId == addConn {
		net.connections = append(net.connections, mut.conn)
	} else {
		var newConnection = connection{newEndpoints, randomWeight(), true, getNewInnovationNumber()}
		net.connections = append(net.connections, newConnection)
		mutationHistory[newEndpoints] = mutation{addConn, newConnection}
	}

	mutHistoryMutex.Unlock()
}

func (net *Network) addNode() {
	// Get a random enabled connection

	var connIndex = rand.IntN(len(net.connections))
	var connRef = &net.connections[connIndex]
	if !connRef.enabled {
		return
	}

	// Check if there is space for a new layer

	var newLayer = net.nodes[connRef.endpoints.senderId].layer + 1
	if newLayer >= maxDepth {
		return
	}

	// Split in two and add a node in between

	var newNode = node{uint32(len(net.nodes)), newLayer, activation.TanhActivation{}}
	var newEndpoints1 = connectionNodes{connRef.endpoints.senderId, newNode.id}
	var newEndpoints2 = connectionNodes{newNode.id, connRef.endpoints.receiverId}
	net.nodes = append(net.nodes, newNode)

	// Check if the two new connections were already added in another network
	// If so, copy them
	// Otherwise, create the new connections, add them to the network and add them to the mutation history

	mutHistoryMutex.Lock()

	var mut1, present1 = mutationHistory[newEndpoints1]
	var mut2, present2 = mutationHistory[newEndpoints2]
	if present1 && present2 && mut1.typeId == addNode && mut2.typeId == addNode {
		net.connections = append(net.connections, []connection{mut1.conn, mut2.conn}...)
	} else {
		var newConnection1 = connection{newEndpoints1, randomWeight(), true, getNewInnovationNumber()}
		var newConnection2 = connection{newEndpoints2, randomWeight(), true, getNewInnovationNumber()}
		net.connections = append(net.connections, []connection{newConnection1, newConnection2}...)
		mutationHistory[newEndpoints1] = mutation{addNode, newConnection1}
		mutationHistory[newEndpoints2] = mutation{addNode, newConnection2}
	}

	mutHistoryMutex.Unlock()
}

func (net *Network) mutateWeights() {
	const mutationChance = 0.10
	for i := range net.connections {
		if rand.Float32() < mutationChance {
			net.connections[i].weight = randomWeight()
		}
	}
}

func (net *Network) GetFitnessScore() uint32 {
	return net.FitnessScore
}

func (net *Network) GetSurvivalChance() float32 {
	return net.SurvivalChance
}

func (net *Network) SetSurvivalChance(chance float32) {
	net.SurvivalChance = chance
}

func (net *Network) Distance(other ga.Member) float64 {
	var otherNewtork = other.(*Network)

	const c1 = 1.0
	const c2 = 1.0
	const c3 = 0.4

	var M uint32 = 0                                                 // Matching connections
	var E uint32 = 0                                                 // Exceeding connections
	var D uint32 = 0                                                 // Disjoint connections
	var W float64 = 0                                                // Weighted sum of distances between matching connections
	var N = max(len(net.connections), len(otherNewtork.connections)) // Max number of connections between the two networks

	// To not penalize small networks and handle cases where N = 0
	if N < 20 {
		N = 1
	}

	// Sort the connections of the two networks by innovation number

	if !net.sorted {
		slices.SortFunc(net.connections, func(a, b connection) int {
			return int(int64(a.innovationNo) - int64(b.innovationNo))
		})
		net.sorted = true
	}

	if !otherNewtork.sorted {
		slices.SortFunc(otherNewtork.connections, func(a, b connection) int {
			return int(int64(a.innovationNo) - int64(b.innovationNo))
		})
		otherNewtork.sorted = true
	}

	// Compute M, E, D and the sum of distances between matching connections

	var weightsDiff float64 = 0
	for i := range N {
		if i >= len(net.connections) || i >= len(otherNewtork.connections) {
			E++
			continue
		}
		if net.connections[i].innovationNo == otherNewtork.connections[i].innovationNo {
			M++
			weightsDiff += math.Abs(net.connections[i].weight - otherNewtork.connections[i].weight)
		} else {
			D++
		}
	}

	// Compute W

	if M > 0 {
		W = weightsDiff / float64(M)
	}

	// Compute the distance, which is a linear combination of three components

	var distance = (c1 * float64(E) / float64(N)) + (c2 * float64(D) / float64(N)) + (c3 * W)

	// Increment the counter of "similar" networks
	const threshold = 0.3
	if distance < threshold {
		net.nSimilar++
	}

	return distance
}

func (net *Network) ComputeAndSetFitnessScore() {
	if len(mutationHistory) > 0 {
		clearMutationHistory()
	}

	const K = 10000
	var fitness float64 = 0
	for _, entry := range testSet {
		if len(entry.input) != int(net.nInputs) || len(entry.output) != int(net.nOutputs) {
			panic("incorrect data entry")
		}

		var result = net.Feed(entry.input)

		var err float64 = 0
		for i := range net.nOutputs {
			err += math.Abs(result[i] - entry.output[i])
		}

		fitness += math.Pow((float64(net.nOutputs) - err), 2)
	}

	net.FitnessScore = uint32(K * fitness / float64(net.nSimilar)) // Adjusted fitness
}

func (net *Network) Crossover(other ga.Member) ga.Member {
	var otherNewtork = other.(*Network)
	var newNetwork = NewNetwork(net.nInputs, net.nOutputs)
	var N = max(len(net.connections), len(otherNewtork.connections)) // Max number of connections between the two networks

	// Sort the connections of the two networks by innovation number

	if !net.sorted {
		slices.SortFunc(net.connections, func(a, b connection) int {
			return int(int64(a.innovationNo) - int64(b.innovationNo))
		})
		net.sorted = true
	}

	if !otherNewtork.sorted {
		slices.SortFunc(otherNewtork.connections, func(a, b connection) int {
			return int(int64(a.innovationNo) - int64(b.innovationNo))
		})
		otherNewtork.sorted = true
	}

	// Compare the connections
	// Matching connections are inherited randomly between the two networks
	// Disjoint/excess connections are inherited only from the dominant parent (chosen to be `net`)

	for i := range N {
		// Reached end of dominant parent connections
		if i >= len(net.connections) {
			break
		}

		if i >= len(otherNewtork.connections) {
			// Inherit disjoint/excess connections from the dominant parent
			newNetwork.connections = append(newNetwork.connections, net.connections[i])
		} else {
			// Inherit the connection randomly
			if rand.Float32() < 0.5 {
				newNetwork.connections = append(newNetwork.connections, net.connections[i])
			} else {
				newNetwork.connections = append(newNetwork.connections, otherNewtork.connections[i])
			}
		}
	}

	// Copy the nodes from the dominant parent

	copy(newNetwork.nodes, net.nodes)

	return newNetwork
}

func (net *Network) Mutate() {
	const addConnectionChance = 0.3
	const addNodeChance = addConnectionChance + 0.05

	var mutationType = rand.Float32()
	if mutationType < addConnectionChance {
		net.addConnection()
	} else if mutationType < addNodeChance {
		net.addNode()
	} else {
		net.mutateWeights()
	}
}

func (net *Network) Feed(input []float64) []float64 {
	// Initialize the accumulators for each node

	var accumulators = make([]float64, len(net.nodes))
	copy(accumulators[:net.nInputs], input)

	// Group connections by layer (the layer of a connection is set as the layer of its sender node)

	var connectionsByLayer = make([][]*connection, maxDepth)
	for i := range net.connections {
		var connRef = &net.connections[i]
		var connLayer = net.nodes[connRef.endpoints.senderId].layer
		connectionsByLayer[connLayer] = append(connectionsByLayer[connLayer], connRef)
	}

	// Feed the network layer by layer

	for i := range maxDepth {
		for _, conn := range connectionsByLayer[i] {
			if !conn.enabled {
				continue
			}
			var inNodeRef = &net.nodes[conn.endpoints.senderId]
			var inNodeAccumulator = accumulators[conn.endpoints.senderId]
			var connInput = inNodeRef.activation.Compute(inNodeAccumulator)
			var connOutput = connInput * conn.weight
			accumulators[conn.endpoints.receiverId] += connOutput
		}
	}

	// Call the activation function of the output layer

	var output = make([]float64, 0, net.nOutputs)
	for i := net.nInputs; i < net.nInputs+net.nOutputs; i++ {
		var nodeRef = &net.nodes[i]
		output = append(output, nodeRef.activation.Compute(accumulators[i]))
	}

	return output
}

type SolverOptions struct {
	PopulationSize uint32  // Number of members at each generation
	MaxGenerations uint32  // Maximum number of generations to simulate
	MutationChance float32 // Chance that a member of the population randomly mutates
	NBatches       uint32  // Population is divided into batches, each batch managed by a separate goroutine
	Verbose        bool    // Enable verbose output on stdout
}

type Solver struct {
	options  SolverOptions
	gaSolver *ga.Solver[*Network]
}

func NewSolver(nInputs uint32, nOutputs uint32, options SolverOptions) *Solver {
	// Creation of the first generation

	var members = make([]*Network, 0, options.PopulationSize)
	for range options.PopulationSize {
		members = append(members, NewNetwork(nInputs, nOutputs))
	}

	// Instantiation of the GA solver

	var params = ga.SolverOptions{
		PopulationSize: options.PopulationSize,
		MaxGenerations: options.MaxGenerations,
		MutationChance: options.MutationChance,
		NBatches:       options.NBatches,
		Speciation:     true,
		Verbose:        options.Verbose,
	}

	var gaSolver = ga.NewSolver(members, params)

	// Create and return the new NEAT solver

	var solver = &Solver{}
	solver.options = options
	solver.gaSolver = gaSolver
	return solver
}

func (solver *Solver) Solve() *Network {
	return solver.gaSolver.Solve()
}
