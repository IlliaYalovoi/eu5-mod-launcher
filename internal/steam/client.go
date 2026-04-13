package steam

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	defaultEndpoint   = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	defaultUserAgent  = "eu5-mod-launcher/1.0 (+https://github.com/illia/eu5-mod-launcher)"
	defaultTimeout    = 10 * time.Second
	defaultRetryCount = 2
	defaultRetryDelay = 300 * time.Millisecond
)

var (
	errInvalidWorkshopItemID = errors.New("invalid workshop item id")
	errRetryableStatus       = errors.New("steam metadata request returned retryable status")
	errNonOKStatus           = errors.New("steam metadata request failed with non-ok status")
)

// WorkshopItem represents metadata returned by Steam for a workshop item.
type WorkshopItem struct {
	ItemID      string `json:"itemId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PreviewURL  string `json:"previewUrl"`
}

// Client fetches workshop metadata from Steam web endpoints.
type Client struct {
	httpClient *http.Client
	endpoint   string
	userAgent  string
	retries    int
	retryDelay time.Duration
}

// NewClient creates a Steam workshop metadata client with default resiliency settings.
func NewClient() *Client {
	return NewClientWithOptions(
		&http.Client{Timeout: defaultTimeout},
		defaultEndpoint,
		defaultUserAgent,
		defaultRetryCount,
		defaultRetryDelay,
	)
}

// NewClientWithOptions creates a client with custom transport options.
func NewClientWithOptions(
	httpClient *http.Client,
	endpoint, userAgent string,
	retries int,
	retryDelay time.Duration,
) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if strings.TrimSpace(endpoint) == "" {
		endpoint = defaultEndpoint
	}
	if strings.TrimSpace(userAgent) == "" {
		userAgent = defaultUserAgent
	}
	if retries < 0 {
		retries = 0
	}
	if retryDelay < 0 {
		retryDelay = 0
	}
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		userAgent:  userAgent,
		retries:    retries,
		retryDelay: retryDelay,
	}
}

// FetchWorkshopMetadata fetches metadata for one or more workshop item IDs using a default client.
func FetchWorkshopMetadata(ids []string) (map[string]WorkshopItem, error) {
	return NewClient().FetchWorkshopMetadata(ids)
}

// FetchWorkshopMetadata fetches metadata for one or more workshop item IDs.
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

	attempts := max(c.retries+1, 1)

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		items, fetchErr := c.fetchOnce(requestBody, requestSet)
		if fetchErr == nil {
			return items, nil
		}
		lastErr = fetchErr
		if attempt+1 >= attempts {
			break
		}
		time.Sleep(c.retryDelay)
	}

	return nil, fmt.Errorf("fetch workshop metadata after %d attempt(s): %w", attempts, lastErr)
}

func (c *Client) fetchOnce(requestBody string, requestSet map[string]struct{}) (map[string]WorkshopItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("build steam metadata request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send steam metadata request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	if shouldRetryStatusCode(resp.StatusCode) {
		return nil, fmt.Errorf("%w: status %d", errRetryableStatus, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		bodyPreview, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if readErr != nil {
			return nil, fmt.Errorf(
				"%w: status %d and body read error: %w",
				errNonOKStatus,
				resp.StatusCode,
				readErr,
			)
		}
		return nil, fmt.Errorf(
			"%w: status %d: %s",
			errNonOKStatus,
			resp.StatusCode,
			strings.TrimSpace(string(bodyPreview)),
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read steam metadata response body: %w", err)
	}

	items, err := parseWorkshopResponse(body, requestSet)
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
	form := url.Values{}
	form.Set("itemcount", fmt.Sprintf("%d", len(ids)))
	for i, id := range ids {
		form.Set(fmt.Sprintf("publishedfileids[%d]", i), id)
	}
	return form.Encode()
}

func shouldRetryStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
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
