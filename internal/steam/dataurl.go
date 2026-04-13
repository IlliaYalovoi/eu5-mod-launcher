package steam

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// EncodeImageFileAsDataURL reads an image file and returns a data URL usable by WebView.
func EncodeImageFileAsDataURL(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", fmt.Errorf("encode image file as data url: empty path")
	}

	data, err := os.ReadFile(trimmedPath)
	if err != nil {
		return "", fmt.Errorf("encode image file as data url: read %q: %w", trimmedPath, err)
	}
	if err := guardImage(data); err != nil {
		return "", fmt.Errorf("encode image file as data url: guard %q: %w", trimmedPath, err)
	}

	mime := mimeForImageBytes(data)
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:" + mime + ";base64," + encoded, nil
}

func mimeForImageBytes(data []byte) string {
	ext := detectImageExtension(data)
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		detected := http.DetectContentType(data)
		if strings.HasPrefix(detected, "image/") {
			return detected
		}
		return "application/octet-stream"
	}
}
