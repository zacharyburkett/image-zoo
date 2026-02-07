package neat

import (
	"fmt"
	"sort"
)

// CompiledConn is a connection referencing source value index.
type CompiledConn struct {
	Src    int
	Weight float64
}

// CompiledNode is a node scheduled for evaluation.
type CompiledNode struct {
	ID         NodeID
	ValueIndex int
	Bias       float64
	Activation ActivationType
	Incoming   []CompiledConn
}

// Plan is a compiled, acyclic execution plan.
type Plan struct {
	Inputs     []NodeID
	Outputs    []NodeID
	nodes      []CompiledNode
	valueIndex map[NodeID]int
	outIndex   []int
}

// BuildAcyclicPlan compiles a genome into a deterministic, acyclic execution plan.
// If inputs or outputs are nil/empty, they are inferred from node kinds.
func BuildAcyclicPlan(g Genome, inputs []NodeID, outputs []NodeID) (*Plan, error) {
	if len(g.Nodes) == 0 {
		return nil, fmt.Errorf("genome has no nodes")
	}

	nodeByID := make(map[NodeID]NodeGene, len(g.Nodes))
	for _, n := range g.Nodes {
		if _, exists := nodeByID[n.ID]; exists {
			return nil, fmt.Errorf("duplicate node id %d", n.ID)
		}
		nodeByID[n.ID] = n
	}

	if len(inputs) == 0 {
		inputs = nodesByKind(nodeByID, NodeInput)
	}
	if len(outputs) == 0 {
		outputs = nodesByKind(nodeByID, NodeOutput)
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no input nodes")
	}
	if len(outputs) == 0 {
		return nil, fmt.Errorf("no output nodes")
	}

	for _, id := range inputs {
		n, ok := nodeByID[id]
		if !ok {
			return nil, fmt.Errorf("input node %d not found", id)
		}
		if n.Kind != NodeInput {
			return nil, fmt.Errorf("node %d is not input", id)
		}
	}
	for _, id := range outputs {
		n, ok := nodeByID[id]
		if !ok {
			return nil, fmt.Errorf("output node %d not found", id)
		}
		if n.Kind != NodeOutput {
			return nil, fmt.Errorf("node %d is not output", id)
		}
	}

	order, err := topoOrder(nodeByID, g.Connections)
	if err != nil {
		return nil, err
	}

	valueIndex := make(map[NodeID]int, len(g.Nodes))
	for i, id := range inputs {
		valueIndex[id] = i
	}
	idx := len(inputs)

	compiledNodes := make([]CompiledNode, 0, len(order))
	for _, id := range order {
		if nodeByID[id].Kind == NodeInput {
			continue
		}
		valueIndex[id] = idx
		idx++
	}

	incoming := make(map[NodeID][]CompiledConn, len(g.Nodes))
	for _, c := range g.Connections {
		if !c.Enabled {
			continue
		}
		inNode, ok := nodeByID[c.In]
		if !ok {
			return nil, fmt.Errorf("connection %d has unknown in node %d", c.Innovation, c.In)
		}
		outNode, ok := nodeByID[c.Out]
		if !ok {
			return nil, fmt.Errorf("connection %d has unknown out node %d", c.Innovation, c.Out)
		}
		if outNode.Kind == NodeInput {
			return nil, fmt.Errorf("connection %d targets input node %d", c.Innovation, c.Out)
		}
		srcIdx, ok := valueIndex[inNode.ID]
		if !ok {
			return nil, fmt.Errorf("connection %d references in node %d not in value index", c.Innovation, c.In)
		}
		incoming[c.Out] = append(incoming[c.Out], CompiledConn{Src: srcIdx, Weight: c.Weight})
	}

	for _, id := range order {
		n := nodeByID[id]
		if n.Kind == NodeInput {
			continue
		}
		compiledNodes = append(compiledNodes, CompiledNode{
			ID:         n.ID,
			ValueIndex: valueIndex[n.ID],
			Bias:       n.Bias,
			Activation: n.Activation,
			Incoming:   incoming[n.ID],
		})
	}

	outIndex := make([]int, 0, len(outputs))
	for _, id := range outputs {
		idx, ok := valueIndex[id]
		if !ok {
			return nil, fmt.Errorf("output node %d not found in value index", id)
		}
		outIndex = append(outIndex, idx)
	}

	return &Plan{
		Inputs:     inputs,
		Outputs:    outputs,
		nodes:      compiledNodes,
		valueIndex: valueIndex,
		outIndex:   outIndex,
	}, nil
}

// Eval executes the plan with the provided inputs and returns output values.
func (p *Plan) Eval(inputs []float64) ([]float64, error) {
	if len(inputs) != len(p.Inputs) {
		return nil, fmt.Errorf("expected %d inputs, got %d", len(p.Inputs), len(inputs))
	}

	values := make([]float64, len(p.valueIndex))
	copy(values, inputs)

	for _, n := range p.nodes {
		sum := n.Bias
		for _, c := range n.Incoming {
			sum += values[c.Src] * c.Weight
		}
		values[n.ValueIndex] = n.Activation.Apply(sum)
	}

	out := make([]float64, len(p.outIndex))
	for i, idx := range p.outIndex {
		out[i] = values[idx]
	}
	return out, nil
}

func nodesByKind(nodes map[NodeID]NodeGene, kind NodeKind) []NodeID {
	ids := make([]NodeID, 0, len(nodes))
	for id, n := range nodes {
		if n.Kind == kind {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func topoOrder(nodes map[NodeID]NodeGene, conns []ConnectionGene) ([]NodeID, error) {
	inDegree := make(map[NodeID]int, len(nodes))
	outgoing := make(map[NodeID][]NodeID, len(nodes))
	for id := range nodes {
		inDegree[id] = 0
	}

	for _, c := range conns {
		if !c.Enabled {
			continue
		}
		if _, ok := nodes[c.In]; !ok {
			return nil, fmt.Errorf("connection %d has unknown in node %d", c.Innovation, c.In)
		}
		if _, ok := nodes[c.Out]; !ok {
			return nil, fmt.Errorf("connection %d has unknown out node %d", c.Innovation, c.Out)
		}
		outgoing[c.In] = append(outgoing[c.In], c.Out)
		inDegree[c.Out]++
	}

	queue := make([]NodeID, 0, len(nodes))
	for id, deg := range inDegree {
		if deg == 0 {
			queue = insertSorted(queue, id)
		}
	}

	order := make([]NodeID, 0, len(nodes))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, out := range outgoing[id] {
			inDegree[out]--
			if inDegree[out] == 0 {
				queue = insertSorted(queue, out)
			}
		}
	}

	for id, deg := range inDegree {
		if deg != 0 {
			return nil, fmt.Errorf("cycle detected at node %d", id)
		}
	}
	return order, nil
}

func insertSorted(queue []NodeID, id NodeID) []NodeID {
	idx := sort.Search(len(queue), func(i int) bool { return queue[i] > id })
	queue = append(queue, 0)
	copy(queue[idx+1:], queue[idx:])
	queue[idx] = id
	return queue
}
