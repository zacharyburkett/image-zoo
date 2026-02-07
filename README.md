# Image Zoo

Image Zoo is a Go implementation of NEAT (NeuroEvolution of Augmenting Topologies) with a focus on evolving CPPNs to generate images, plus general-purpose NEAT use.

This repository is intentionally starting small and will grow in phases. The immediate goal is a correct, well-tested NEAT core with fast inference and clear documentation.

## Goals
- NEAT core: genomes, speciation, mutation, crossover, innovation tracking, reproduction.
- Activation functions: multiple types and per-node assignment.
- Efficient inference: compiled, acyclic execution for CPPNs and general NEAT networks.
- Determinism: reproducible runs via explicit RNG management.
- Testing: unit tests for core operations, property tests for invariants, and end-to-end tests for small evolutions.
- Documentation: developer-oriented docs plus agent-friendly guidance in AGENTS.md.

## Non-Goals (for now)
- GPU inference.
- Large-scale distributed evolution.

## Status
- Repository scaffolded.
- Architecture and roadmap docs drafted in docs/.

## Next Steps (Pending Decisions)
- Specify the initial CPPN fitness heuristics for the wasm app (start with entropy).

## Suggested Layout (will evolve)
- cmd/            Entry points (e.g., tools, local CLI tests)
- pkg/neat/       Core NEAT types and operations
- pkg/cppn/       CPPN-specific helpers (inputs, coordinate mapping)
- internal/       App-specific logic for Image Zoo
- docs/           Architecture, roadmap, and design notes

## License
See LICENSE.
