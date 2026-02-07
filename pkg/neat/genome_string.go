package neat

import (
	"fmt"
	"sort"
	"strings"
)

// String returns a detailed, deterministic summary of the genome.
func (g Genome) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Genome\n")
	fmt.Fprintf(&b, "Fitness: %.4f\n", g.Fitness)
	fmt.Fprintf(&b, "Nodes: %d\n", len(g.Nodes))

	nodes := make([]NodeGene, len(g.Nodes))
	copy(nodes, g.Nodes)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	for _, n := range nodes {
		fmt.Fprintf(&b, "  Node %d %s act=%s bias=%.4f\n", n.ID, n.Kind, n.Activation, n.Bias)
	}

	b.WriteString("Connections: ")
	fmt.Fprintf(&b, "%d\n", len(g.Connections))
	conns := sortedConnections(g.Connections)
	for _, c := range conns {
		state := "on"
		if !c.Enabled {
			state = "off"
		}
		fmt.Fprintf(&b, "  Conn %d %d->%d w=%.4f %s\n", c.Innovation, c.In, c.Out, c.Weight, state)
	}

	return b.String()
}
