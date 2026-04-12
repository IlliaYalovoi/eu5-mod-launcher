package launcher

import (
	"errors"
	"fmt"
	"net/url"
	goruntime "runtime"
	"strconv"
	"strings"

	"eu5-mod-launcher/internal/steam"
)

func (a *App) OpenWorkshopItem(itemID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	normalizedID, err := normalizeWorkshopItemID(itemID)
	if err != nil {
		return fmt.Errorf("open workshop item %q: %w", itemID, err)
	}
	httpsURL := "https://steamcommunity.com/sharedfiles/filedetails/?id=" + normalizedID
	if err := a.OpenExternalLink(httpsURL); err != nil {
		return fmt.Errorf("open workshop item %q: %w", normalizedID, err)
	}
	return nil
}

func (a *App) UnsubscribeWorkshopMod(itemID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if !a.IsUnsubscribeEnabled() {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", itemID, errUnsubscribeDisabled)
	}
	trimmedID := strings.TrimSpace(itemID)
	if trimmedID == "" {
		return nil
	}
	unsubscribeURL, err := a.svc.launchSvc.BuildWorkshopUnsubscribeURL(trimmedID)
	if err != nil {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", itemID, err)
	}
	if err := a.OpenExternalLink(unsubscribeURL); err != nil {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", trimmedID, err)
	}
	return nil
}

func (*App) IsUnsubscribeEnabled() bool { return compileEnableUnsubscribe }

func (a *App) OpenExternalLink(rawURL string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	normalizedURL, linkErr := normalizeExternalLink(rawURL)
	if linkErr != nil {
		return fmt.Errorf("open external link %q: %w", rawURL, linkErr)
	}
	parsedURL, parseErr := url.Parse(normalizedURL)
	if parseErr != nil {
		return fmt.Errorf("open external link %q: parse normalized url: %w", normalizedURL, parseErr)
	}
	attempts := make([]error, 0, 3)
	if isSteamLikeLink(parsedURL) {
		steamURL := toSteamClientURL(parsedURL)
		if err := a.openURL(goruntime.GOOS, steamURL); err == nil {
			return nil
		} else {
			attempts = append(attempts, fmt.Errorf("open in steam client: %w", err))
		}
	}
	if err := a.openURL(goruntime.GOOS, normalizedURL); err == nil {
		return nil
	} else {
		attempts = append(attempts, fmt.Errorf("open in default browser: %w", err))
	}
	if err := a.openInAppURL(normalizedURL); err == nil {
		return nil
	} else {
		attempts = append(attempts, fmt.Errorf("open in wails window fallback: %w", err))
	}
	return fmt.Errorf("open external link %q: %w", normalizedURL, errors.Join(attempts...))
}

func normalizeExternalLink(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", errExternalLinkInvalid
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("%w: parse %q: %s", errExternalLinkInvalid, rawURL, err.Error())
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "steam" {
		return "", fmt.Errorf("%w: unsupported scheme %q", errExternalLinkInvalid, scheme)
	}
	if scheme != "steam" && strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("%w: missing host", errExternalLinkInvalid)
	}
	return parsed.String(), nil
}

func isSteamLikeLink(u *url.URL) bool {
	if u == nil {
		return false
	}
	if strings.EqualFold(u.Scheme, "steam") {
		return true
	}
	host := strings.ToLower(u.Hostname())
	return host == "steamcommunity.com" || strings.HasSuffix(host, ".steamcommunity.com") ||
		host == "store.steampowered.com" || strings.HasSuffix(host, ".steampowered.com")
}

func toSteamClientURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	if strings.EqualFold(u.Scheme, "steam") {
		return u.String()
	}
	if itemID := workshopItemIDFromCommunityURL(u); itemID != "" {
		return "steam://url/CommunityFilePage/" + itemID
	}
	return "steam://openurl/" + u.String()
}

func workshopItemIDFromCommunityURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if host != "steamcommunity.com" && !strings.HasSuffix(host, ".steamcommunity.com") {
		return ""
	}
	queryID := strings.TrimSpace(u.Query().Get("id"))
	if queryID == "" || !isWorkshopNumericID(queryID) {
		return ""
	}
	path := strings.ToLower(strings.TrimSpace(u.Path))
	if strings.Contains(path, "/sharedfiles/filedetails") || strings.Contains(path, "/workshop/filedetails") {
		return queryID
	}
	return ""
}

func normalizeWorkshopItemID(itemID string) (string, error) {
	normalizedID := strings.TrimSpace(itemID)
	if normalizedID == "" {
		return "", errWorkshopItemIDInvalid
	}
	parsed, err := strconv.ParseUint(normalizedID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errWorkshopItemIDInvalid, itemID)
	}
	return strconv.FormatUint(parsed, 10), nil
}

func isWorkshopNumericID(id string) bool {
	_, err := strconv.ParseUint(id, 10, 64)
	return err == nil
}

func (a *App) FetchWorkshopMetadataForMod(modID string) (steam.WorkshopItem, error) {
	if err := a.mustBeReady(); err != nil {
		return steam.WorkshopItem{}, err
	}
	workshopID := a.workshopItemIDForMod(modID)
	if workshopID == "" {
		return steam.WorkshopItem{}, fmt.Errorf("no workshop id for mod %q", modID)
	}
	items, err := a.svc.steamClient.FetchWorkshopMetadata([]string{workshopID})
	if err != nil {
		return steam.WorkshopItem{}, fmt.Errorf("fetch workshop metadata for mod %q: %w", modID, err)
	}
	item, ok := items[workshopID]
	if !ok {
		return steam.WorkshopItem{}, fmt.Errorf("workshop item %q not found", workshopID)
	}
	return item, nil
}
