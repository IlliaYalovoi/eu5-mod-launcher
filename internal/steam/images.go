package steam

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // decode guard supports GIF previews
	_ "image/jpeg" // decode guard supports JPEG previews
	_ "image/png"  // decode guard supports PNG previews
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultImageMaxEntries = 1000
	maxImageBytes          = 10 * 1024 * 1024
	maxImageDimension      = 8192
)

var (
	errImageCacheRootEmpty     = errors.New("image cache root is empty")
	errPreviewURLMissing       = errors.New("preview url is missing")
	errPreviewURLSchemeInvalid = errors.New("preview url scheme is unsupported")
	errPreviewStatusNonOK      = errors.New("preview image response status is not ok")
	errImageTooLarge           = errors.New("image payload exceeds size limit")
	errImageDecodeFailed       = errors.New("image decode guard failed")
)

// ImageCache stores workshop preview thumbnails in a bounded local directory.
type ImageCache struct {
	dirPath    string
	maxEntries int
	httpClient *http.Client
	mu         sync.Mutex
}

// NewImageCache creates an image cache at cacheRoot.
func NewImageCache(cacheRoot string, maxEntries int, httpClient *http.Client) (*ImageCache, error) {
	root := strings.TrimSpace(cacheRoot)
	if root == "" {
		return nil, fmt.Errorf("create image cache: %w", errImageCacheRootEmpty)
	}
	if maxEntries <= 0 {
		maxEntries = defaultImageMaxEntries
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	dirPath := filepath.Join(root, "steam", "images")
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return nil, fmt.Errorf("create image cache dir %q: %w", dirPath, err)
	}

	return &ImageCache{dirPath: dirPath, maxEntries: maxEntries, httpClient: httpClient}, nil
}

// CachedPath returns local thumbnail path for itemID if any cached file exists.
func (c *ImageCache) CachedPath(itemID string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cachedPathLocked(itemID)
}

// EnsureStored returns existing image path or downloads it when missing.
func (c *ImageCache) EnsureStored(item WorkshopItem) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if path := c.cachedPathLocked(item.ItemID); path != "" {
		return path, nil
	}
	return c.storeDownloadedLocked(item)
}

// RefreshStored always re-downloads preview image.
func (c *ImageCache) RefreshStored(item WorkshopItem) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.storeDownloadedLocked(item)
}

func (c *ImageCache) storeDownloadedLocked(item WorkshopItem) (string, error) {
	itemID := strings.TrimSpace(item.ItemID)
	if itemID == "" {
		return "", fmt.Errorf("store preview image: %w", errMetadataCacheItemIDEmpty)
	}

	previewURL := strings.TrimSpace(item.PreviewURL)
	if previewURL == "" {
		if path := c.cachedPathLocked(itemID); path != "" {
			return path, nil
		}
		return "", fmt.Errorf("store preview image for %q: %w", itemID, errPreviewURLMissing)
	}

	parsedURL, err := parsePreviewURL(previewURL)
	if err != nil {
		return "", fmt.Errorf("store preview image for %q: %w", itemID, err)
	}

	data, err := c.downloadImage(previewURL)
	if err != nil {
		return "", fmt.Errorf("store preview image for %q: %w", itemID, err)
	}
	if err := guardImage(data); err != nil {
		return "", fmt.Errorf("store preview image for %q: %w", itemID, err)
	}

	ext := extensionForURLPath(parsedURL.Path)
	if err := c.removeItemVariantsLocked(itemID); err != nil {
		return "", err
	}

	finalPath := filepath.Join(c.dirPath, itemID+ext)
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return "", fmt.Errorf("write preview image tmp %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return "", fmt.Errorf(
				"replace preview image %q: %w; cleanup tmp %q: %s",
				finalPath,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return "", fmt.Errorf("replace preview image %q: %w", finalPath, err)
	}

	c.cleanupLocked()
	return finalPath, nil
}

func parsePreviewURL(previewURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(previewURL)
	if err != nil {
		return nil, fmt.Errorf("parse preview url %q: %w", previewURL, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("%w: %q", errPreviewURLSchemeInvalid, parsedURL.Scheme)
	}
	return parsedURL, nil
}

func (c *ImageCache) downloadImage(previewURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, previewURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build preview image request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send preview image request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errPreviewStatusNonOK, resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxImageBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read preview image response: %w", err)
	}
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("%w: %d bytes", errImageTooLarge, len(data))
	}

	return data, nil
}

func guardImage(data []byte) error {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: %s", errImageDecodeFailed, err.Error())
	}
	if invalidDimensions(cfg.Width, cfg.Height) {
		return fmt.Errorf("%w: dimensions %dx%d", errImageDecodeFailed, cfg.Width, cfg.Height)
	}
	return nil
}

func invalidDimensions(width, height int) bool {
	return width <= 0 || height <= 0 || width > maxImageDimension || height > maxImageDimension
}

func extensionForURLPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return ext
	default:
		return ".img"
	}
}

func (c *ImageCache) cachedPathLocked(itemID string) string {
	trimmedID := strings.TrimSpace(itemID)
	if trimmedID == "" {
		return ""
	}
	patterns := []string{
		trimmedID + ".jpg",
		trimmedID + ".jpeg",
		trimmedID + ".png",
		trimmedID + ".gif",
		trimmedID + ".webp",
		trimmedID + ".img",
	}
	for _, name := range patterns {
		path := filepath.Join(c.dirPath, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func (c *ImageCache) removeItemVariantsLocked(itemID string) error {
	matches, err := filepath.Glob(filepath.Join(c.dirPath, itemID+".*"))
	if err != nil {
		return fmt.Errorf("glob image cache variants for %q: %w", itemID, err)
	}
	for _, path := range matches {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("remove stale image cache file %q: %w", path, removeErr)
		}
	}
	return nil
}

func (c *ImageCache) cleanupLocked() {
	entries, err := os.ReadDir(c.dirPath)
	if err != nil {
		return
	}
	if len(entries) <= c.maxEntries {
		return
	}

	type fileEntry struct {
		path    string
		modTime time.Time
	}
	files := make([]fileEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		files = append(files, fileEntry{path: filepath.Join(c.dirPath, entry.Name()), modTime: info.ModTime()})
	}
	if len(files) <= c.maxEntries {
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})
	toRemove := len(files) - c.maxEntries
	for i := 0; i < toRemove; i++ {
		if err := os.Remove(files[i].path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
	}
}
