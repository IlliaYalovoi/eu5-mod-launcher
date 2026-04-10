package steam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

const defaultDescriptionImageMaxEntries = 3000

var (
	errDescriptionImageCacheRootEmpty = errors.New("description image cache root is empty")
	errDescriptionImageURLMissing     = errors.New("description image url is missing")
	errDescriptionImageURLInvalid     = errors.New("description image url scheme is unsupported")
)

// DescriptionImageCache stores images referenced by workshop descriptions.
type DescriptionImageCache struct {
	dirPath    string
	maxEntries int
	httpClient *http.Client
	mu         sync.Mutex
}

type descriptionImageURLInfo struct {
	normalized string
	ext        string
}

func NewDescriptionImageCache(
	cacheRoot string,
	maxEntries int,
	httpClient *http.Client,
) (*DescriptionImageCache, error) {
	root := strings.TrimSpace(cacheRoot)
	if root == "" {
		return nil, fmt.Errorf("create description image cache: %w", errDescriptionImageCacheRootEmpty)
	}
	if maxEntries <= 0 {
		maxEntries = defaultDescriptionImageMaxEntries
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	dirPath := filepath.Join(root, "steam", "description-images")
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return nil, fmt.Errorf("create description image cache dir %q: %w", dirPath, err)
	}

	return &DescriptionImageCache{dirPath: dirPath, maxEntries: maxEntries, httpClient: httpClient}, nil
}

func (c *DescriptionImageCache) EnsureStored(itemID, imageURL string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	normalizedID := strings.TrimSpace(itemID)
	if !isNumericID(normalizedID) {
		return "", fmt.Errorf("store description image: %w: %q", errInvalidWorkshopItemID, itemID)
	}

	urlInfo, err := normalizeDescriptionImageURL(imageURL)
	if err != nil {
		return "", fmt.Errorf("store description image for %q: %w", normalizedID, err)
	}

	fileBase := descriptionImageFileBase(normalizedID, urlInfo)
	if cachedPath := c.cachedPathByBaseLocked(fileBase); cachedPath != "" {
		existingData, readErr := os.ReadFile(cachedPath)
		if readErr != nil {
			return "", fmt.Errorf("read cached description image %q: %w", cachedPath, readErr)
		}
		if guardErr := guardImage(existingData); guardErr == nil {
			return c.normalizeCachedDescriptionImagePathLocked(cachedPath, fileBase, existingData), nil
		}
		if removeErr := os.Remove(cachedPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return "", fmt.Errorf("remove invalid cached description image %q: %w", cachedPath, removeErr)
		}
	}

	data, err := c.downloadImage(urlInfo.normalized)
	if err != nil {
		return "", fmt.Errorf("store description image for %q: %w", normalizedID, err)
	}
	if err := guardImage(data); err != nil {
		return "", fmt.Errorf("store description image for %q: %w", normalizedID, err)
	}

	ext := detectImageExtension(data)
	if ext == ".img" {
		ext = urlInfo.ext
	}
	if err := c.removeByBaseLocked(fileBase); err != nil {
		return "", err
	}
	finalPath := filepath.Join(c.dirPath, fileBase+ext)
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return "", fmt.Errorf("write description image tmp %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return "", fmt.Errorf(
				"replace description image %q: %w; cleanup tmp %q: %s",
				finalPath,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return "", fmt.Errorf("replace description image %q: %w", finalPath, err)
	}

	c.cleanupLocked()
	return finalPath, nil
}

func normalizeDescriptionImageURL(rawURL string) (descriptionImageURLInfo, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return descriptionImageURLInfo{}, errDescriptionImageURLMissing
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return descriptionImageURLInfo{}, fmt.Errorf("parse description image url %q: %w", rawURL, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return descriptionImageURLInfo{}, fmt.Errorf("%w: %q", errDescriptionImageURLInvalid, parsedURL.Scheme)
	}

	return descriptionImageURLInfo{normalized: parsedURL.String(), ext: extensionForURLPath(parsedURL.Path)}, nil
}

func descriptionImageFileBase(itemID string, urlInfo descriptionImageURLInfo) string {
	h := sha256.Sum256([]byte(urlInfo.normalized))
	hash := hex.EncodeToString(h[:8])
	return itemID + "_" + hash
}

func (c *DescriptionImageCache) cachedPathByBaseLocked(fileBase string) string {
	if strings.TrimSpace(fileBase) == "" {
		return ""
	}
	matches, err := filepath.Glob(filepath.Join(c.dirPath, fileBase+".*"))
	if err != nil {
		return ""
	}
	for _, path := range matches {
		if strings.HasSuffix(path, ".tmp") {
			continue
		}
		info, statErr := os.Stat(path)
		if statErr != nil || info.IsDir() {
			continue
		}
		return path
	}
	return ""
}

func (c *DescriptionImageCache) removeByBaseLocked(fileBase string) error {
	matches, err := filepath.Glob(filepath.Join(c.dirPath, fileBase+".*"))
	if err != nil {
		return fmt.Errorf("glob description image cache variants for %q: %w", fileBase, err)
	}
	for _, path := range matches {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("remove stale description image cache file %q: %w", path, removeErr)
		}
	}
	return nil
}

func (c *DescriptionImageCache) normalizeCachedDescriptionImagePathLocked(path, fileBase string, data []byte) string {
	if filepath.Ext(path) != ".img" {
		return path
	}
	detectedExt := detectImageExtension(data)
	if detectedExt == ".img" {
		return path
	}

	targetPath := filepath.Join(c.dirPath, fileBase+detectedExt)
	if removeErr := os.Remove(targetPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return path
	}
	if renameErr := os.Rename(path, targetPath); renameErr != nil {
		return path
	}

	return targetPath
}

func (c *DescriptionImageCache) downloadImage(imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, imageURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build description image request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send description image request: %w", err)
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
		return nil, fmt.Errorf("read description image response: %w", err)
	}
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("%w: %d bytes", errImageTooLarge, len(data))
	}

	return data, nil
}

func (c *DescriptionImageCache) cleanupLocked() {
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
