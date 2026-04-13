package steam

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
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

	compressed, err := c.CompressThumbnail(data)
	if err == nil {
		data = compressed
	}

	ext := detectImageExtension(data)
	if ext == ".img" {
		ext = extensionForURLPath(parsedURL.Path)
	}
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

// CompressThumbnail decodes, resizes (max 128px), and encodes to low-quality JPEG.
func (c *ImageCache) CompressThumbnail(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	const maxDim = 128
	if w > maxDim || h > maxDim {
		newW, newH := w, h
		if w > h {
			newH = h * maxDim / w
			newW = maxDim
		} else {
			newW = w * maxDim / h
			newH = maxDim
		}

		newImg := image.NewRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				srcX := x * w / newW
				srcY := y * h / newH
				newImg.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
			}
		}
		img = newImg
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 40}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

func detectImageExtension(data []byte) string {
	if len(data) >= 8 {
		pngSig := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
		if bytes.Equal(data[:8], pngSig) {
			return ".png"
		}
	}
	if len(data) >= 3 {
		if data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
			return ".jpg"
		}
	}
	if len(data) >= 6 {
		if bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a")) {
			return ".gif"
		}
	}
	if len(data) >= 12 {
		if bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
			return ".webp"
		}
	}

	return ".img"
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
			if filepath.Ext(path) != ".img" {
				return path
			}

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return ""
			}
			if guardErr := guardImage(data); guardErr != nil {
				if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					return ""
				}
				return ""
			}

			detectedExt := detectImageExtension(data)
			if detectedExt == ".img" {
				return path
			}

			normalizedPath := filepath.Join(c.dirPath, trimmedID+detectedExt)
			if removeErr := os.Remove(normalizedPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return path
			}
			if renameErr := os.Rename(path, normalizedPath); renameErr != nil {
				return path
			}
			return normalizedPath
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
	for i := range toRemove {
		if err := os.Remove(files[i].path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
	}
}
