package neat

// NodeID identifies a node within a genome.
type NodeID int

// InnovID identifies a connection innovation number.
type InnovID int

// NodeKind indicates the role of a node in the network.
type NodeKind uint8

const (
	NodeInput NodeKind = iota
	NodeHidden
	NodeOutput
)

// NodeGene represents a node in a genome.
type NodeGene struct {
	ID         NodeID
	Kind       NodeKind
	Activation ActivationType
	Bias       float64
}

// ConnectionGene represents a directed weighted edge between nodes.
type ConnectionGene struct {
	Innovation InnovID
	In         NodeID
	Out        NodeID
	Weight     float64
	Enabled    bool
}

// Genome is a collection of node and connection genes.
type Genome struct {
	Nodes       []NodeGene
	Connections []ConnectionGene
	Fitness     float64
}
