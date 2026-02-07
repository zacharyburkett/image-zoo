package neat

import "sort"

func cloneGenome(g Genome) Genome {
	clone := Genome{
		Fitness: g.Fitness,
	}
	if len(g.Nodes) > 0 {
		clone.Nodes = make([]NodeGene, len(g.Nodes))
		copy(clone.Nodes, g.Nodes)
	}
	if len(g.Connections) > 0 {
		clone.Connections = make([]ConnectionGene, len(g.Connections))
		copy(clone.Connections, g.Connections)
	}
	return clone
}

func nodeGeneMap(nodes []NodeGene) map[NodeID]NodeGene {
	out := make(map[NodeID]NodeGene, len(nodes))
	for _, n := range nodes {
		out[n.ID] = n
	}
	return out
}

func sortMembersByFitness(genomes []Genome, members []int) []int {
	sorted := make([]int, len(members))
	copy(sorted, members)
	sort.Slice(sorted, func(i, j int) bool {
		fi := genomes[sorted[i]].Fitness
		fj := genomes[sorted[j]].Fitness
		if fi == fj {
			return sorted[i] < sorted[j]
		}
		return fi > fj
	})
	return sorted
}
