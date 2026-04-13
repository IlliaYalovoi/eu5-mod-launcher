package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	metadataCacheSchemaVersion = 2
	defaultMetadataTTL         = 24 * time.Hour
	defaultMetadataMaxEntries  = 5000
)

var (
	errMetadataCacheItemIDEmpty = errors.New("metadata cache item id is empty")
	errMetadataCacheRootEmpty   = errors.New("metadata cache root is empty")
)

type metadataCacheFile struct {
	Version int                            `json:"version"`
	Entries map[string]metadataCacheRecord `json:"entries"`
}

type metadataCacheRecord struct {
	Item          WorkshopItem `json:"item"`
	UpdatedAtUTC  time.Time    `json:"updatedAtUtc"`
	AccessedAtUTC time.Time    `json:"accessedAtUtc"`
}

// MetadataLookup is one cache read result.
type MetadataLookup struct {
	Item  WorkshopItem
	Hit   bool
	Stale bool
}

// MetadataResolveResult partitions cache state by freshness.
type MetadataResolveResult struct {
	Fresh   map[string]WorkshopItem
	Stale   map[string]WorkshopItem
	Missing []string
}

// MetadataCache provides a file-backed Steam metadata cache with TTL.
type MetadataCache struct {
	filePath   string
	ttl        time.Duration
	maxEntries int
	now        func() time.Time
	mu         sync.Mutex
}

// NewMetadataCache creates a file-backed metadata cache in cacheRoot.
func NewMetadataCache(cacheRoot string, ttl time.Duration, maxEntries int) (*MetadataCache, error) {
	root := strings.TrimSpace(cacheRoot)
	if root == "" {
		return nil, fmt.Errorf("create metadata cache: %w", errMetadataCacheRootEmpty)
	}
	if ttl <= 0 {
		ttl = defaultMetadataTTL
	}
	if maxEntries <= 0 {
		maxEntries = defaultMetadataMaxEntries
	}

	cacheDir := filepath.Join(root, "steam")
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("create metadata cache dir %q: %w", cacheDir, err)
	}

	return &MetadataCache{
		filePath:   filepath.Join(cacheDir, "metadata_cache_v2.json"),
		ttl:        ttl,
		maxEntries: maxEntries,
		now:        time.Now,
	}, nil
}

// Get returns one cached metadata record and whether it is stale.
func (c *MetadataCache) Get(itemID string) (MetadataLookup, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, err := c.loadLocked()
	if err != nil {
		return MetadataLookup{}, err
	}

	normalizedID := strings.TrimSpace(itemID)
	if normalizedID == "" {
		return MetadataLookup{}, fmt.Errorf("metadata cache get: %w", errMetadataCacheItemIDEmpty)
	}

	record, ok := state.Entries[normalizedID]
	if !ok {
		return MetadataLookup{}, nil
	}

	isStale := c.now().UTC().Sub(record.UpdatedAtUTC) > c.ttl
	return MetadataLookup{Item: record.Item, Hit: true, Stale: isStale}, nil
}

// ResolveMany partitions ids into fresh, stale, and missing sets.
func (c *MetadataCache) ResolveMany(ids []string) (MetadataResolveResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, err := c.loadLocked()
	if err != nil {
		return MetadataResolveResult{}, err
	}

	normalized, err := normalizeWorkshopIDs(ids)
	if err != nil {
		return MetadataResolveResult{}, fmt.Errorf("resolve metadata ids: %w", err)
	}

	result := MetadataResolveResult{
		Fresh:   make(map[string]WorkshopItem),
		Stale:   make(map[string]WorkshopItem),
		Missing: make([]string, 0),
	}
	nowUTC := c.now().UTC()

	for _, id := range normalized {
		record, ok := state.Entries[id]
		if !ok {
			result.Missing = append(result.Missing, id)
			continue
		}
		if nowUTC.Sub(record.UpdatedAtUTC) > c.ttl {
			result.Stale[id] = record.Item
			continue
		}
		result.Fresh[id] = record.Item
	}

	return result, nil
}

// SetMany writes metadata records into cache and runs eviction when needed.
func (c *MetadataCache) SetMany(items map[string]WorkshopItem) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, err := c.loadLocked()
	if err != nil {
		return err
	}

	nowUTC := c.now().UTC()
	for id := range items {
		item := items[id]
		trimmedID := strings.TrimSpace(id)
		if trimmedID == "" {
			continue
		}
		if !isNumericID(trimmedID) {
			return fmt.Errorf("%w: %q", errInvalidWorkshopItemID, id)
		}
		item.ItemID = trimmedID
		state.Entries[trimmedID] = metadataCacheRecord{
			Item:          item,
			UpdatedAtUTC:  nowUTC,
			AccessedAtUTC: nowUTC,
		}
	}

	c.evictLocked(state)
	return c.saveLocked(state)
}

func (c *MetadataCache) loadLocked() (metadataCacheFile, error) {
	content, err := os.ReadFile(c.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			empty := metadataCacheFile{Version: metadataCacheSchemaVersion, Entries: map[string]metadataCacheRecord{}}
			return empty, nil
		}
		return metadataCacheFile{}, fmt.Errorf("read metadata cache %q: %w", c.filePath, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		empty := metadataCacheFile{Version: metadataCacheSchemaVersion, Entries: map[string]metadataCacheRecord{}}
		return empty, nil
	}

	var state metadataCacheFile
	if err := json.Unmarshal(content, &state); err != nil {
		return metadataCacheFile{}, fmt.Errorf("decode metadata cache %q: %w", c.filePath, err)
	}
	if state.Version != metadataCacheSchemaVersion {
		empty := metadataCacheFile{Version: metadataCacheSchemaVersion, Entries: map[string]metadataCacheRecord{}}
		return empty, nil
	}
	if state.Entries == nil {
		state.Entries = map[string]metadataCacheRecord{}
	}
	return state, nil
}

func (c *MetadataCache) saveLocked(state metadataCacheFile) error {
	state.Version = metadataCacheSchemaVersion
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata cache %q: %w", c.filePath, err)
	}
	payload = append(payload, '\n')

	tmpPath := c.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write metadata cache tmp %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, c.filePath); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf(
				"replace metadata cache %q: %w; cleanup tmp %q: %s",
				c.filePath,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return fmt.Errorf("replace metadata cache %q: %w", c.filePath, err)
	}
	return nil
}

func (c *MetadataCache) evictLocked(state metadataCacheFile) {
	if len(state.Entries) <= c.maxEntries {
		return
	}

	type candidate struct {
		id       string
		accessed time.Time
		updated  time.Time
	}
	list := make([]candidate, 0, len(state.Entries))
	for id := range state.Entries {
		record := state.Entries[id]
		list = append(list, candidate{id: id, accessed: record.AccessedAtUTC, updated: record.UpdatedAtUTC})
	}
	sort.Slice(list, func(i, j int) bool {
		if !list[i].accessed.Equal(list[j].accessed) {
			return list[i].accessed.Before(list[j].accessed)
		}
		return list[i].updated.Before(list[j].updated)
	})

	removeCount := len(state.Entries) - c.maxEntries
	for i := 0; i < removeCount; i++ {
		delete(state.Entries, list[i].id)
	}
}
