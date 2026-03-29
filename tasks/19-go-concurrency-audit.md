# Task 19 — Go Performance: Concurrency Audit & Additional Parallelism

## Goal

Identify and parallelize other CPU-bound or high-latency operations beyond mod scanning.

## Context

Depends on: task 18.
This is a profiling-driven task.

## Deliverables

### Profiling pass

Measure hotspots for representative scenarios:
- initial startup
- refresh mods
- autosort
- loading large playsets/layouts

Produce short profiling notes in:
- `tasks/examples/perf-notes.md` (new)

### Candidate optimizations

Implement 1-2 safe optimizations from profiling data, for example:
- parallel thumbnail metadata probing
- batched normalization passes
- concurrent enrichment post-processing

### Safety checks

- race-free (`go test -race ./...` or targeted package subset)
- deterministic externally visible order

## Acceptance criteria

- profiling evidence captured before and after
- measurable improvement in selected hotspots
- no behavior regression

## Notes for agent

- Do not parallelize tiny operations just for style.
- Prefer measurable wins with bounded complexity.

