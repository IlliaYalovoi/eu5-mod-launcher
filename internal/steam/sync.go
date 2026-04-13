package steam

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"eu5-mod-launcher/internal/logging"
)

// ThumbnailSync unifies background synchronization of workshop metadata and thumbnails.
type ThumbnailSync struct {
	client        *Client
	metadataCache *MetadataCache
	imageCache    *ImageCache
	semaphore     chan struct{}
	isDirty       atomic.Bool
}

// NewThumbnailSync creates a new sync service with a worker limit.
func NewThumbnailSync(client *Client, metadataCache *MetadataCache, imageCache *ImageCache, maxWorkers int) *ThumbnailSync {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	return &ThumbnailSync{
		client:        client,
		metadataCache: metadataCache,
		imageCache:    imageCache,
		semaphore:     make(chan struct{}, maxWorkers),
	}
}

// StartPeriodicCleanup starts a background ticker for cache maintenance.
func (s *ThumbnailSync) StartPeriodicCleanup(ctx context.Context, interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				logging.Infof("Starting periodic image cache cleanup (TTL: %v)", ttl)
				s.imageCache.CleanupOlderThan(ttl)
				logging.Infof("Completed periodic image cache cleanup")
			case <-ctx.Done():
				return
			}
		}
	}()
}

// SyncAll resolve metadata and downloads missing thumbnails for the given workshop IDs.
func (s *ThumbnailSync) SyncAll(ctx context.Context, modIDs []string) {
	if len(modIDs) == 0 {
		return
	}

	logging.Infof("Starting background thumbnail sync for %d items", len(modIDs))

	// 1. Resolve metadata cache status
	resolved, err := s.metadataCache.ResolveMany(modIDs)
	if err != nil {
		logging.Errorf("Failed to resolve metadata from cache: %v", err)
		return
	}

	// 2. Fetch missing or stale metadata
	toFetch := append(resolved.Missing, getMapKeys(resolved.Stale)...)
	if len(toFetch) > 0 {
		fetched, err := s.client.FetchWorkshopMetadata(toFetch)
		if err != nil {
			logging.Errorf("Failed to fetch workshop metadata: %v", err)
		} else {
			if err := s.metadataCache.SetMany(fetched); err != nil {
				logging.Errorf("Failed to update metadata cache: %v", err)
			}
			// Update fresh map with newly fetched items for thumbnail processing
			for id, item := range fetched {
				resolved.Fresh[id] = item
			}
		}
	}

	// 3. Ensure all fresh thumbnails are stored
	var wg sync.WaitGroup
	for id, item := range resolved.Fresh {
		// Quick check if already in image cache to avoid spawning goroutine if not needed
		if s.imageCache.CachedPath(id) != "" {
			continue
		}

		wg.Add(1)
		go func(it WorkshopItem) {
			defer wg.Done()

			select {
			case s.semaphore <- struct{}{}:
				defer func() { <-s.semaphore }()
			case <-ctx.Done():
				return
			}

			if _, err := s.imageCache.EnsureStored(it); err != nil {
				logging.Errorf("Failed to sync thumbnail for %s: %v", it.ItemID, err)
			} else {
				s.isDirty.Store(true)
			}
		}(item)
	}

	wg.Wait()
	logging.Infof("Completed background thumbnail sync")
}

// HasNewThumbnails returns true if new thumbnails were downloaded since last call.
func (s *ThumbnailSync) HasNewThumbnails() bool {
	return s.isDirty.Swap(false)
}

func getMapKeys(m map[string]WorkshopItem) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
