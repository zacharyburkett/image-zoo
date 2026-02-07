package neat

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrNoConnectionCandidates = errors.New("no valid connection candidates")
	ErrNoEnabledConnections   = errors.New("no enabled connections to split")
)

// MutationConfig controls mutation probabilities and ranges.
type MutationConfig struct {
	AddConnectionProb    float64
	AddNodeProb          float64
	WeightMutateProb     float64
	WeightPerturbProb    float64
	WeightPerturbScale   float64
	WeightResetScale     float64
	WeightInitRange      float64
	BiasMutateProb       float64
	BiasPerturbProb      float64
	BiasPerturbScale     float64
	BiasResetScale       float64
	ToggleEnableProb     float64
	ActivationMutateProb float64
	AllowedActivations   []ActivationType
	MaxAttempts          int
}

// DefaultMutationConfig returns a conservative baseline.
func DefaultMutationConfig() MutationConfig {
	return MutationConfig{
		AddConnectionProb:    0.05,
		AddNodeProb:          0.03,
		WeightMutateProb:     0.8,
		WeightPerturbProb:    0.9,
		WeightPerturbScale:   0.1,
		WeightResetScale:     1.0,
		WeightInitRange:      1.0,
		BiasMutateProb:       0.7,
		BiasPerturbProb:      0.9,
		BiasPerturbScale:     0.1,
		BiasResetScale:       1.0,
		ToggleEnableProb:     0.01,
		ActivationMutateProb: 0.02,
		AllowedActivations: []ActivationType{
			ActivationLinear,
			ActivationSigmoid,
			ActivationTanh,
			ActivationRelu,
			ActivationSin,
			ActivationCos,
			ActivationGaussian,
		},
		MaxAttempts: 30,
	}
}

// Mutate applies NEAT mutations to the genome.
func (m MutationConfig) Mutate(rng RNG, g *Genome, tracker *InnovationTracker) error {
	if g == nil {
		return fmt.Errorf("genome is nil")
	}
	if rng == nil {
		return fmt.Errorf("rng is nil")
	}
	if tracker == nil {
		return fmt.Errorf("innovation tracker is nil")
	}
	if randBool(rng, m.AddConnectionProb) {
		if err := MutateAddConnection(rng, g, tracker, m.WeightInitRange, m.MaxAttempts); err != nil && !errors.Is(err, ErrNoConnectionCandidates) {
			return err
		}
	}
	if randBool(rng, m.AddNodeProb) {
		if err := MutateAddNode(rng, g, tracker, m.AllowedActivations); err != nil && !errors.Is(err, ErrNoEnabledConnections) {
			return err
		}
	}

	MutateWeights(rng, g, m.WeightMutateProb, m.WeightPerturbProb, m.WeightPerturbScale, m.WeightResetScale)
	MutateBiases(rng, g, m.BiasMutateProb, m.BiasPerturbProb, m.BiasPerturbScale, m.BiasResetScale)
	MutateToggleConnections(rng, g, m.ToggleEnableProb)
	MutateActivations(rng, g, m.ActivationMutateProb, m.AllowedActivations)

	return nil
}

// MutateAddConnection adds a new acyclic connection between existing nodes.
func MutateAddConnection(rng RNG, g *Genome, tracker *InnovationTracker, weightRange float64, maxAttempts int) error {
	if g == nil {
		return fmt.Errorf("genome is nil")
	}
	if rng == nil {
		return fmt.Errorf("rng is nil")
	}
	if tracker == nil {
		return fmt.Errorf("innovation tracker is nil")
	}
	if len(g.Nodes) == 0 {
		return fmt.Errorf("genome has no nodes")
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	nodeByID := make(map[NodeID]NodeGene, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeByID[n.ID] = n
	}

	order, err := topoOrder(nodeByID, g.Connections)
	if err != nil {
		return err
	}
	index := make(map[NodeID]int, len(order))
	for i, id := range order {
		index[id] = i
	}

	candidates := make([]connKey, 0)
	for _, in := range g.Nodes {
		if in.Kind == NodeOutput {
			continue
		}
		for _, out := range g.Nodes {
			if out.Kind == NodeInput {
				continue
			}
			if in.ID == out.ID {
				continue
			}
			if index[in.ID] >= index[out.ID] {
				continue
			}
			if connectionExists(g, in.ID, out.ID) {
				continue
			}
			candidates = append(candidates, connKey{in: in.ID, out: out.ID})
		}
	}

	if len(candidates) == 0 {
		return ErrNoConnectionCandidates
	}

	pick := candidates[rng.Intn(len(candidates))]
	innov := tracker.Innovation(pick.in, pick.out)
	weight := randRange(rng, -weightRange, weightRange)

	g.Connections = insertConnectionSorted(g.Connections, ConnectionGene{
		Innovation: innov,
		In:         pick.in,
		Out:        pick.out,
		Weight:     weight,
		Enabled:    true,
	})
	return nil
}

// MutateAddNode splits an existing enabled connection, inserting a new hidden node.
func MutateAddNode(rng RNG, g *Genome, tracker *InnovationTracker, activations []ActivationType) error {
	if g == nil {
		return fmt.Errorf("genome is nil")
	}
	if rng == nil {
		return fmt.Errorf("rng is nil")
	}
	if tracker == nil {
		return fmt.Errorf("innovation tracker is nil")
	}
	idx := enabledConnectionIndex(rng, g)
	if idx < 0 {
		return ErrNoEnabledConnections
	}

	old := g.Connections[idx]
	g.Connections[idx].Enabled = false

	newNode := NodeGene{
		ID:         tracker.NextNodeID(),
		Kind:       NodeHidden,
		Activation: chooseActivation(rng, activations),
		Bias:       0,
	}
	g.Nodes = insertNodeSorted(g.Nodes, newNode)

	innov1 := tracker.Innovation(old.In, newNode.ID)
	innov2 := tracker.Innovation(newNode.ID, old.Out)

	g.Connections = insertConnectionSorted(g.Connections, ConnectionGene{
		Innovation: innov1,
		In:         old.In,
		Out:        newNode.ID,
		Weight:     1.0,
		Enabled:    true,
	})
	g.Connections = insertConnectionSorted(g.Connections, ConnectionGene{
		Innovation: innov2,
		In:         newNode.ID,
		Out:        old.Out,
		Weight:     old.Weight,
		Enabled:    true,
	})
	return nil
}

// MutateWeights mutates connection weights.
func MutateWeights(rng RNG, g *Genome, mutateProb, perturbProb, perturbScale, resetScale float64) {
	if rng == nil {
		return
	}
	for i := range g.Connections {
		if !randBool(rng, mutateProb) {
			continue
		}
		if randBool(rng, perturbProb) {
			g.Connections[i].Weight += randRange(rng, -perturbScale, perturbScale)
		} else {
			g.Connections[i].Weight = randRange(rng, -resetScale, resetScale)
		}
	}
}

// MutateBiases mutates node biases (non-input nodes only).
func MutateBiases(rng RNG, g *Genome, mutateProb, perturbProb, perturbScale, resetScale float64) {
	if rng == nil {
		return
	}
	for i := range g.Nodes {
		if g.Nodes[i].Kind == NodeInput {
			continue
		}
		if !randBool(rng, mutateProb) {
			continue
		}
		if randBool(rng, perturbProb) {
			g.Nodes[i].Bias += randRange(rng, -perturbScale, perturbScale)
		} else {
			g.Nodes[i].Bias = randRange(rng, -resetScale, resetScale)
		}
	}
}

// MutateToggleConnections flips the enabled flag for connections.
func MutateToggleConnections(rng RNG, g *Genome, toggleProb float64) {
	if rng == nil {
		return
	}
	nodeMap := nodeGeneMap(g.Nodes)
	for i := range g.Connections {
		if randBool(rng, toggleProb) {
			if g.Connections[i].Enabled {
				g.Connections[i].Enabled = false
				continue
			}
			g.Connections[i].Enabled = true
			if _, err := topoOrder(nodeMap, g.Connections); err != nil {
				g.Connections[i].Enabled = false
			}
		}
	}
}

// MutateActivations changes activation functions on non-input nodes.
func MutateActivations(rng RNG, g *Genome, mutateProb float64, activations []ActivationType) {
	if rng == nil {
		return
	}
	if len(activations) == 0 {
		return
	}
	for i := range g.Nodes {
		if g.Nodes[i].Kind == NodeInput {
			continue
		}
		if randBool(rng, mutateProb) {
			g.Nodes[i].Activation = activations[rng.Intn(len(activations))]
		}
	}
}

func connectionExists(g *Genome, in, out NodeID) bool {
	for _, c := range g.Connections {
		if c.In == in && c.Out == out {
			return true
		}
	}
	return false
}

func enabledConnectionIndex(rng RNG, g *Genome) int {
	indices := make([]int, 0, len(g.Connections))
	for i, c := range g.Connections {
		if c.Enabled {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return -1
	}
	return indices[rng.Intn(len(indices))]
}

func chooseActivation(rng RNG, activations []ActivationType) ActivationType {
	if len(activations) == 0 {
		return ActivationSigmoid
	}
	return activations[rng.Intn(len(activations))]
}

func insertNodeSorted(nodes []NodeGene, node NodeGene) []NodeGene {
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].ID > node.ID })
	nodes = append(nodes, NodeGene{})
	copy(nodes[idx+1:], nodes[idx:])
	nodes[idx] = node
	return nodes
}

func insertConnectionSorted(conns []ConnectionGene, conn ConnectionGene) []ConnectionGene {
	idx := sort.Search(len(conns), func(i int) bool { return conns[i].Innovation > conn.Innovation })
	conns = append(conns, ConnectionGene{})
	copy(conns[idx+1:], conns[idx:])
	conns[idx] = conn
	return conns
}
