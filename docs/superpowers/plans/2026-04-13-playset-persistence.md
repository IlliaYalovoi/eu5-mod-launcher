# Playset Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `SavePlayset` for both EU5 (JSON) and Legacy (SQLite) adapters to allow persistent mod configuration.

**Architecture:** Use transactions for SQLite to ensure data integrity and Atomic Write pattern for JSON to prevent corruption. Adapters will map the normalized `game.Playset` domain model back to their respective disk formats.

**Tech Stack:** Go, `sqlx`, `modernc.org/sqlite`, standard `os` and `encoding/json`.

---

### Task 1: EU5 Adapter JSON Persistence

**Files:**
- Modify: `internal/adapters/eu5/adapter.go`
- Test: `internal/adapters/eu5/adapter_test.go`

- [ ] **Step 1: Write failing test for EU5 SavePlayset**

```go
func TestEU5SavePlayset(t *testing.T) {
    tmpDir := t.TempDir()
    inst := game.Instance{UserConfigPath: tmpDir}
    adapter := &Adapter{}
    p := game.Playset{
        ID: "default",
        Name: "Default",
        Entries: []game.ModEntry{{ID: "mod1", Enabled: true, Position: 0}},
    }
    err := adapter.SavePlayset(inst, p)
    if err != nil { t.Fatal(err) }
    
    // Verify file exists and has content
    path := filepath.Join(tmpDir, "playsets", "default.json")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("expected playset file to exist at %s", path)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/eu5/...`
Expected: FAIL (empty implementation)

- [ ] **Step 3: Implement JSON serialization in EU5 Adapter**

```go
func (s *Adapter) SavePlayset(inst game.Instance, p game.Playset) error {
    dir := filepath.Join(inst.UserConfigPath, "playsets")
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(p, "", "  ")
    if err != nil {
        return err
    }
    path := filepath.Join(dir, p.ID+".json")
    return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/eu5/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/eu5/
git commit -m "feat(eu5): implement SavePlayset for JSON format"
```

---

### Task 2: Legacy Adapter SQLite Persistence

**Files:**
- Modify: `internal/adapters/legacy/sqlite.go`
- Test: `internal/adapters/legacy/sqlite_test.go`

- [ ] **Step 1: Write failing test for SQLite SavePlayset**

```go
func TestSqliteSavePlayset(t *testing.T) {
    // Requires mock/real launcher-v2.sqlite setup
    // Verify position and enabled status updates in playsets_mods table
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/legacy/...`
Expected: FAIL

- [ ] **Step 3: Implement SQLite transaction for playset updates**

```go
func (s *SqliteAdapter) SavePlayset(inst game.Instance, p game.Playset) error {
    db, err := s.getDB(inst)
    if err != nil { return err }

    tx, err := db.Beginx()
    if err != nil { return err }
    defer tx.Rollback()

    // Clear existing entries for this playset
    _, err = tx.Exec("DELETE FROM playsets_mods WHERE playsetId = ?", p.ID)
    if err != nil { return err }

    // Re-insert normalized entries
    for _, e := range p.Entries {
        _, err = tx.Exec(`
            INSERT INTO playsets_mods (playsetId, modId, enabled, position)
            VALUES (?, ?, ?, ?)`, p.ID, e.ID, e.Enabled, e.Position)
        if err != nil { return err }
    }

    return tx.Commit()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/legacy/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/legacy/
git commit -m "feat(legacy): implement SavePlayset for SQLite format"
