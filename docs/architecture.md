# Architecture

This document captures the intended architecture for the NEAT/CPPN system. It is a living document and will change as implementation proceeds.

## Core Concepts
- Genome: nodes + connections (with innovation numbers).
- Node: type (input/hidden/output), activation function, bias.
- Connection: in/out nodes, weight, enabled flag, innovation id.
- Species: groups of similar genomes by compatibility distance.
- Population: collection of genomes, species, global innovation tracking.

## Evolution Loop (high level)
1. Initialize population with minimal topology.
2. Evaluate fitness for each genome (domain-specific).
3. Speciate using compatibility distance.
4. Allocate offspring counts per species.
5. Reproduce via crossover + mutation.
6. Repeat until stopping criteria.

## Mutation Operators
- Add connection.
- Add node (split connection).
- Weight perturbation / reset.
- Bias perturbation / reset.
- Toggle connection enabled.
- Activation function mutation.

## Crossover
- Align genes by innovation number.
- Inherit excess/disjoint genes from more fit parent.
- Handle disabled gene inheritance with a probability.

## Inference Engine
We use a deterministic topological schedule compiled from the genome. Networks are acyclic only, which keeps evaluation simple and wasm-friendly.

## Fitness (CPPN)
Fitness uses a multi-metric score (entropy, edge density, fine edges, variance, symmetry, color variance, noise penalty) plus optional novelty search. Presets in the UI adjust weights.

## Determinism
- Central RNG: all randomness through an injected RNG.
- Explicit seeds in tests and sample runs.

## Testing Strategy
- Unit tests for mutations, crossover, distance metrics.
- Property tests for invariants (no duplicate innovations, DAG for CPPNs).
- Small integration test: evolve XOR or radial pattern for sanity.
