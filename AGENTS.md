# Agent Notes (Image Zoo)

This file captures tricky or easy-to-miss details for agentic development on this repo. Keep it short and update as decisions are made.

## Decisions Made
- Go module path: `github.com/zacharyburkett/image-zoo`.
- Inference representation: deterministic topological schedule (acyclic only).
- Networks: strictly acyclic (no recurrent connections).

## Decisions Pending
- CPPN fitness heuristics for the Image Zoo app (start with entropy).

## Tricky Bits / Invariants
- Innovation numbers must be consistent across the population; track globally.
- Crossover must align by innovation id and preserve enable/disable semantics.
- Speciation depends on compatibility distance; ensure it is deterministic.
- CPPNs should be acyclic; enforce DAG creation (no recurrent edges).
- Mutation ops must avoid creating duplicate connections.
- Deterministic RNG: all randomness must route through an injected RNG.
- Add-connection mutation must preserve acyclic order (use topo order).
- Crossover may introduce cycles; we disable highest-innovation enabled edges until acyclic.

## Testing Expectations
- Unit tests for each mutation operator and distance metric.
- Property tests for genome invariants (unique innovation ids, valid node ids).
- Small end-to-end tests with fixed seeds.

## Perf Notes (for later)
- Prefer contiguous slices and stable ordering to reduce allocations.
- Compile genomes into linear execution plans for fast inference.
