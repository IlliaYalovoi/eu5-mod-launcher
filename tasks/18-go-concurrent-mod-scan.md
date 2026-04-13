# Task 18 — Go Performance: Concurrent Mod Scan Pipeline

## Goal

Parallelize mod scanning/parsing to speed up startup and refresh on large mod sets.

## Context

Depends on: task 01 scanner baseline.

## Deliverables

### `internal/mods/scanner.go`

Refactor scan pipeline to bounded worker pool:
- producer enumerates candidate mod dirs
- workers parse descriptors concurrently
- collector merges, dedupes, sorts stable output

Configurable worker count:
- default based on `runtime.NumCPU()`
- upper bound to avoid IO thrash

### Error handling

- preserve current tolerant behavior (skip broken mods)
- aggregate/log parse failures succinctly (no noisy floods)

### Tests/bench

- keep existing scanner tests passing
- add benchmark comparing sequential vs concurrent scan for synthetic corpus

## Acceptance criteria

- Same logical result set as baseline scanner.
- Faster on large synthetic workload.
- No data races (`go test -race` clean for scanner package).

## Notes for agent

- Keep deterministic ordering in final output.
- Avoid unbounded goroutine fan-out.

