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
	ID         NodeID         `json:"id"`
	Kind       NodeKind       `json:"kind"`
	Activation ActivationType `json:"activation"`
	Bias       float64        `json:"bias"`
}

// ConnectionGene represents a directed weighted edge between nodes.
type ConnectionGene struct {
	Innovation InnovID `json:"innovation"`
	In         NodeID  `json:"in"`
	Out        NodeID  `json:"out"`
	Weight     float64 `json:"weight"`
	Enabled    bool    `json:"enabled"`
}

// Genome is a collection of node and connection genes.
type Genome struct {
	Nodes       []NodeGene       `json:"nodes"`
	Connections []ConnectionGene `json:"connections"`
	Fitness     float64          `json:"fitness"`
}
