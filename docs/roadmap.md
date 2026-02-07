# Roadmap

This roadmap is intended to keep the initial build focused.

## Phase 1: NEAT Core (library)
- Define core data structures (genome, node, connection, species).
- Implement compatibility distance and speciation.
- Implement mutation operators.
- Implement crossover.
- Implement reproduction and population management.
- Basic activation function library.
- Deterministic RNG plumbing.
- Unit tests and property tests.

## Phase 2: Inference Engine
- Choose representation (topo schedule vs VM program).
- Implement compile step from genome to executable form.
- Add benchmarks and correctness tests.
- Ensure wasm-friendly performance.

## Phase 3: CPPN Helpers
- Standard CPPN inputs (x, y, r, bias, etc.).
- Image sampling and rendering utilities.
- Default fitness heuristics (to be specified).

## Phase 4: Image Zoo App (wasm)
- Minimal UI to visualize evolution.
- Define heuristics and user controls.
- Integrate with wasm build pipeline.

## Phase 5: Polishing
- Documentation, examples, and stability improvements.
