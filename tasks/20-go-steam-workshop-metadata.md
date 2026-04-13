# Task 20 — Go: Steam Workshop Metadata Client

## Goal

Add backend Steam integration to resolve workshop mod metadata (title, description, preview image URL) by workshop item ID.

## Context

Depends on: task 18 (scanner provides workshop IDs/path context).

## Deliverables

### New package `internal/steam`

Implement client methods:

```go
// Given one or more workshop item IDs, returns metadata map.
func FetchWorkshopMetadata(ids []string) (map[string]WorkshopItem, error)

type WorkshopItem struct {
    ItemID      string
    Title       string
    Description string
    PreviewURL  string
}
```

Data source options:
- Steam Web API endpoint(s) for workshop details
- robust timeouts + retries + user agent

### App bridge methods

Expose Wails-callable methods:
- fetch for single mod
- optional fetch batch for visible list

### Tests

- unit tests for parsing/ID mapping
- HTTP client tests with mocked responses

## Acceptance criteria

- Selecting a workshop mod can fetch title/description/preview URL from backend.
- Non-workshop mods return empty/no-op result without error.
- Network errors are surfaced cleanly and do not crash app.

## Notes for agent

- Keep this task metadata-only; image download/cache is next task.
- Respect rate limits (batch where possible).

