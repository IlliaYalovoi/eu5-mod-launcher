# Workshop Thumbnail Sync and Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Background download of all mod thumbnails with heavy compression, long-lived cache, and periodic cleanup.

**Architecture:** Extend `steam.ImageCache` with compression support. Add `steam.ThumbnailSync` service for background processing with concurrency limiting. Implement cleanup logic for stale cache files.

**Tech Stack:** Go (Standard Library, `nfnt/resize` for compression if available, or standard `image` package).

---

### Task 1: Extend ImageCache with Compression

**Files:**
- Modify: `internal/steam/images.go`

- [ ] **Step 1: Add thumbnail compression logic**
Add a method to `ImageCache` that resizes and re-encodes images to JPEG with low quality.

- [ ] **Step 2: Update storeDownloadedLocked to use compression**
Modify the storage logic to compress images before saving if they are intended for thumbnails.

- [ ] **Step 3: Commit**
`git commit -m "feat: add image compression to ImageCache"`

### Task 2: Implement Background Thumbnail Sync

**Files:**
- Create: `internal/steam/sync.go`

- [ ] **Step 1: Create ThumbnailSync struct**
Implement a worker pool that processes item IDs, fetches metadata if needed, and downloads/compresses thumbnails. Use a semaphore to limit concurrency.

- [ ] **Step 2: Implement SyncAll method**
Method to trigger sync for all provided mod IDs.

- [ ] **Step 3: Commit**
`git commit -m "feat: implement ThumbnailSync service"`

### Task 4: Implement Cache Cleanup

**Files:**
- Modify: `internal/steam/images.go`
- Modify: `internal/steam/sync.go`

- [ ] **Step 1: Add cleanup logic to ImageCache**
Method to delete files older than a specific TTL.

- [ ] **Step 2: Add periodic cleanup to ThumbnailSync**
Start a ticker that runs cleanup every 15 minutes.

- [ ] **Step 3: Commit**
`git commit -m "feat: add periodic cache cleanup"`

### Task 5: Wire into App Startup

**Files:**
- Modify: `app.go`

- [ ] **Step 1: Initialize ThumbnailSync in startup**
Create the sync service and trigger an initial sync and cleanup.

- [ ] **Step 2: Commit**
`git commit -m "feat: wire background sync into app startup"`

### Task 6: Verification

- [ ] **Step 1: Build the project**
Run `go build ./...`

- [ ] **Step 2: Check logs**
Verify no errors in startup logs.
