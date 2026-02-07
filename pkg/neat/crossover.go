package neat

import (
	"fmt"
	"sort"
)

const disabledInheritProb = 0.75

// Crossover produces a child genome from two parents using NEAT alignment rules.
func Crossover(rng RNG, a, b Genome) (Genome, error) {
	if rng == nil {
		return Genome{}, fmt.Errorf("rng is nil")
	}
	fitterA, equalFitness := fitnessOrder(a, b)
	if !equalFitness {
		if fitterA {
			return crossoverFrom(rng, a, b, false)
		}
		return crossoverFrom(rng, b, a, false)
	}
	return crossoverFrom(rng, a, b, true)
}

func crossoverFrom(rng RNG, primary, secondary Genome, equalFitness bool) (Genome, error) {
	pGenes := sortedConnections(primary.Connections)
	sGenes := sortedConnections(secondary.Connections)

	childConns := make([]ConnectionGene, 0, len(pGenes))
	i, j := 0, 0
	for i < len(pGenes) && j < len(sGenes) {
		p := pGenes[i]
		s := sGenes[j]
		switch {
		case p.Innovation == s.Innovation:
			chosen := p
			if randBool(rng, 0.5) {
				chosen = s
			}
			if !p.Enabled || !s.Enabled {
				if randBool(rng, disabledInheritProb) {
					chosen.Enabled = false
				} else {
					chosen.Enabled = true
				}
			}
			childConns = append(childConns, chosen)
			i++
			j++
		case p.Innovation < s.Innovation:
			if !equalFitness || randBool(rng, 0.5) {
				childConns = append(childConns, p)
			}
			i++
		default:
			if equalFitness && randBool(rng, 0.5) {
				childConns = append(childConns, s)
			}
			j++
		}
	}

	if equalFitness {
		for _, gene := range pGenes[i:] {
			if randBool(rng, 0.5) {
				childConns = append(childConns, gene)
			}
		}
		for _, gene := range sGenes[j:] {
			if randBool(rng, 0.5) {
				childConns = append(childConns, gene)
			}
		}
	} else {
		childConns = append(childConns, pGenes[i:]...)
	}

	childNodes, err := buildChildNodes(rng, primary, secondary, childConns, equalFitness)
	if err != nil {
		return Genome{}, err
	}

	sort.Slice(childConns, func(i, j int) bool { return childConns[i].Innovation < childConns[j].Innovation })
	return Genome{Nodes: childNodes, Connections: childConns, Fitness: 0}, nil
}

func buildChildNodes(rng RNG, primary, secondary Genome, conns []ConnectionGene, equalFitness bool) ([]NodeGene, error) {
	primaryNodes := nodesByID(primary.Nodes)
	secondaryNodes := nodesByID(secondary.Nodes)

	required := make(map[NodeID]struct{})
	for _, n := range primary.Nodes {
		if n.Kind == NodeInput || n.Kind == NodeOutput {
			required[n.ID] = struct{}{}
		}
	}
	for _, n := range secondary.Nodes {
		if n.Kind == NodeInput || n.Kind == NodeOutput {
			required[n.ID] = struct{}{}
		}
	}
	for _, c := range conns {
		required[c.In] = struct{}{}
		required[c.Out] = struct{}{}
	}

	child := make([]NodeGene, 0, len(required))
	for id := range required {
		pn, pOk := primaryNodes[id]
		sn, sOk := secondaryNodes[id]
		if !pOk && !sOk {
			return nil, fmt.Errorf("node %d missing from both parents", id)
		}
		if pOk && sOk {
			if equalFitness {
				if randBool(rng, 0.5) {
					child = append(child, pn)
				} else {
					child = append(child, sn)
				}
			} else {
				child = append(child, pn)
			}
			continue
		}
		if pOk {
			child = append(child, pn)
			continue
		}
		child = append(child, sn)
	}

	sort.Slice(child, func(i, j int) bool { return child[i].ID < child[j].ID })
	return child, nil
}

func nodesByID(nodes []NodeGene) map[NodeID]NodeGene {
	out := make(map[NodeID]NodeGene, len(nodes))
	for _, n := range nodes {
		out[n.ID] = n
	}
	return out
}

func fitnessOrder(a, b Genome) (aFitter bool, equal bool) {
	if a.Fitness == b.Fitness {
		return false, true
	}
	return a.Fitness > b.Fitness, false
}
