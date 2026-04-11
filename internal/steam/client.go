package steam

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultUserAgent  = "eu5-mod-launcher/1.0 (+https://github.com/IlliaYalovoi/eu5-mod-launcher)"
	defaultRetryCount = 3
)

var errInvalidWorkshopItemID = fmt.Errorf("invalid workshop item id")

type WorkshopItem struct {
	ItemID      string `json:"itemId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PreviewURL  string `json:"previewUrl"`
}

type Client struct {
	rc      *resty.Client
	retries int
}

func NewClient() *Client {
	rc := resty.New().
		SetBaseURL("https://api.steampowered.com").
		SetHeader("User-Agent", defaultUserAgent).
		SetTimeout(10 * time.Second).
		SetRetryCount(defaultRetryCount).
		SetRetryWaitTime(300 * time.Millisecond).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == 429 || r.StatusCode() == 502 || r.StatusCode() == 503 || r.StatusCode() >= 500
		})
	return &Client{rc: rc, retries: defaultRetryCount}
}

func (c *Client) FetchWorkshopMetadata(ids []string) (map[string]WorkshopItem, error) {
	normalizedIDs, err := normalizeWorkshopIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("normalize workshop ids: %w", err)
	}
	if len(normalizedIDs) == 0 {
		return map[string]WorkshopItem{}, nil
	}

	requestBody := buildRequestBody(normalizedIDs)
	requestSet := make(map[string]struct{}, len(normalizedIDs))
	for _, id := range normalizedIDs {
		requestSet[id] = struct{}{}
	}

	resp, err := c.rc.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(requestBody).
		Post("/ISteamRemoteStorage/GetPublishedFileDetails/v1/")
	if err != nil {
		return nil, fmt.Errorf("fetch workshop metadata: %w", err)
	}

	if resp.StatusCode() >= 500 {
		return nil, fmt.Errorf("steam server error: status %d", resp.StatusCode())
	}

	items, err := parseWorkshopResponse(resp.Body(), requestSet)
	if err != nil {
		return nil, fmt.Errorf("parse steam metadata response: %w", err)
	}

	return items, nil
}

func normalizeWorkshopIDs(ids []string) ([]string, error) {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))

	for _, rawID := range ids {
		id := strings.TrimSpace(rawID)
		if id == "" {
			continue
		}
		if !isNumericID(id) {
			return nil, fmt.Errorf("%w: %q", errInvalidWorkshopItemID, rawID)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	sort.Strings(out)
	return out, nil
}

func buildRequestBody(ids []string) string {
	var sb strings.Builder
	sb.WriteString("itemcount=")
	sb.WriteString(fmt.Sprintf("%d", len(ids)))
	for i, id := range ids {
		sb.WriteString(fmt.Sprintf("&publishedfileids[%d]=%s", i, id))
	}
	return sb.String()
}

func isNumericID(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
