# Task 21 — Go: Steam Metadata Cache & Thumbnail Fetch

## Goal

Cache Steam workshop metadata and preview images locally to avoid repeated network calls and speed up UI rendering.

## Context

Depends on: task 20.

## Deliverables

### Cache layer

Add file-backed cache under user config/cache dir:
- metadata cache with TTL
- image cache for preview thumbnails

Suggested files:
- `internal/steam/cache.go`
- `internal/steam/images.go`

### App methods

Expose methods that return effective metadata with cache semantics:
- stale-while-revalidate optional behavior
- explicit refresh action optional

### Scanner/UI model integration

Ensure mods can expose a resolved thumbnail path (local cached file) for frontend image components.

### Tests

- cache hit/miss/expiry
- image download + decode guard
- invalid URL / network failures

## Acceptance criteria

- Reopening app does not re-fetch unchanged workshop metadata immediately.
- Preview thumbnails render from local cache when available.
- Failures degrade gracefully.

## Notes for agent

- Keep cache schema versionable.
- Bound disk usage (max entries / cleanup policy).

