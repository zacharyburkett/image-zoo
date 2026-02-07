# Roadmap

This roadmap is intended to keep the initial build focused.

## Phase 1: NEAT Core (library)
- Define core data structures (genome, node, connection, species). (done)
- Implement compatibility distance and speciation. (done)
- Implement mutation operators. (done)
- Implement crossover. (done)
- Implement reproduction and population management. (done)
- Basic activation function library. (done)
- Deterministic RNG plumbing. (done)
- Unit tests and property tests. (in progress)

## Phase 2: Inference Engine
- Choose representation (topo schedule vs VM program). (done)
- Implement compile step from genome to executable form. (done)
- Add benchmarks and correctness tests. (in progress)
- Ensure wasm-friendly performance. (in progress)

## Phase 3: CPPN Helpers
- Standard CPPN inputs (x, y, r, bias, etc.). (done)
- Image sampling and rendering utilities. (done)
- Default fitness heuristics (metrics + novelty). (done)

## Phase 4: Image Zoo App (wasm)
- Minimal UI to visualize evolution. (done)
- Define heuristics and user controls. (done)
- Integrate with wasm build pipeline. (done)
- Improve UX (presets, modal detail, tuning). (in progress)

## Phase 5: Polishing
- Documentation, examples, and stability improvements.
