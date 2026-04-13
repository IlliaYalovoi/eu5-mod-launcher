# Async Thumbnail Sync with Frontend Polling Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make thumbnail synchronization fully asynchronous and add a pollable flag (`isNewThumbnailsAvailable`) for the frontend to detect when to refresh.

**Architecture:** Add an atomic dirty flag to `ThumbnailSync`. Expose it via a new Wails method. Reset the flag after it's read or when the frontend re-fetches.

---

### Task 1: Update ThumbnailSync Service

**Files:**
- Modify: `internal/steam/sync.go`

- [ ] **Step 1: Add atomic dirty flag**
Add `isDirty atomic.Bool` to `ThumbnailSync` struct.

- [ ] **Step 2: Set dirty flag on successful thumbnail storage**
Update `SyncAll` or individual sync logic to set `isDirty` to `true` whenever a new thumbnail is saved.

- [ ] **Step 3: Add GetAndResetDirty method**
Add a thread-safe method to return the current value and reset it to `false`.

- [ ] **Step 4: Commit**
`git commit -m "feat: add dirty flag to ThumbnailSync"`

### Task 2: Expose Polling Method in App

**Files:**
- Modify: `app.go`

- [ ] **Step 1: Implement HasNewThumbnails Wails method**
Add `HasNewThumbnails() bool` to `App` that calls `thumbSync.GetAndResetDirty()`.

- [ ] **Step 2: Ensure background sync is non-blocking**
Verify `SyncAll` is already called in a goroutine in `startup`.

- [ ] **Step 3: Commit**
`git commit -m "feat: expose HasNewThumbnails to frontend"`

### Task 3: Background Cleanup and Wiring

- [ ] **Step 1: Verify build**
Run `go build ./...`

- [ ] **Step 2: Commit**
`git commit -m "fix: final wiring for async thumbnail sync"`
