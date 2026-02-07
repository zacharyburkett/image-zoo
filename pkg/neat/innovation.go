package neat

import "fmt"

type connKey struct {
	in  NodeID
	out NodeID
}

// InnovationTracker maintains global innovation numbers and node ids.
type InnovationTracker struct {
	nextInnov InnovID
	nextNode  NodeID
	conns     map[connKey]InnovID
}

// NewInnovationTracker initializes a tracker with next ids derived from genomes.
func NewInnovationTracker(genomes []Genome) (*InnovationTracker, error) {
	maxInnov := InnovID(0)
	maxNode := NodeID(0)
	conns := make(map[connKey]InnovID)
	for _, g := range genomes {
		for _, n := range g.Nodes {
			if n.ID > maxNode {
				maxNode = n.ID
			}
		}
		for _, c := range g.Connections {
			key := connKey{in: c.In, out: c.Out}
			if existing, ok := conns[key]; ok && existing != c.Innovation {
				return nil, fmt.Errorf("conflicting innovation for connection %d->%d: %d vs %d", c.In, c.Out, existing, c.Innovation)
			}
			conns[key] = c.Innovation
			if c.Innovation > maxInnov {
				maxInnov = c.Innovation
			}
		}
	}

	return &InnovationTracker{
		nextInnov: maxInnov + 1,
		nextNode:  maxNode + 1,
		conns:     conns,
	}, nil
}

// NextNodeID returns a new node id.
func (t *InnovationTracker) NextNodeID() NodeID {
	id := t.nextNode
	t.nextNode++
	return id
}

// Innovation returns the innovation number for a connection, creating one if needed.
func (t *InnovationTracker) Innovation(in, out NodeID) InnovID {
	key := connKey{in: in, out: out}
	if innov, ok := t.conns[key]; ok {
		return innov
	}
	innov := t.nextInnov
	t.nextInnov++
	t.conns[key] = innov
	return innov
}

// SeedConnectionInnovation sets a specific innovation id for a connection.
func (t *InnovationTracker) SeedConnectionInnovation(in, out NodeID, innov InnovID) error {
	key := connKey{in: in, out: out}
	if existing, ok := t.conns[key]; ok && existing != innov {
		return fmt.Errorf("connection %d->%d already mapped to %d", in, out, existing)
	}
	t.conns[key] = innov
	if innov >= t.nextInnov {
		t.nextInnov = innov + 1
	}
	return nil
}
