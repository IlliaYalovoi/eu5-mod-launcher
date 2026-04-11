package main

import (
	"errors"
	"eu5-mod-launcher/internal/steam"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSteamClient struct {
	itemsByID map[string]steam.WorkshopItem
	calls     [][]string
}

var errSteamUnavailable = errors.New("steam unavailable")

func setCompileUnsubscribeFlag(t *testing.T, enabled bool) {
	t.Helper()
	previous := compileEnableUnsubscribe
	compileEnableUnsubscribe = enabled
	t.Cleanup(func() {
		compileEnableUnsubscribe = previous
	})
}

func (f *fakeSteamClient) FetchWorkshopMetadata(ids []string) (map[string]steam.WorkshopItem, error) {
	copied := append([]string(nil), ids...)
	f.calls = append(f.calls, copied)
	out := make(map[string]steam.WorkshopItem, len(ids))
	for _, id := range ids {
		if item, ok := f.itemsByID[id]; ok {
			out[id] = item
		}
	}
	return out, nil
}

func TestFetchWorkshopMetadataForMod_NonWorkshopNoop(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()
	app.steamClient = &fakeSteamClient{}
	app.modPathByID["local-mod"] = filepath.Join(t.TempDir(), "mod", "local-mod")
	app.gamePaths.WorkshopModDirs = []string{filepath.Join(t.TempDir(), "workshop", "content", eu5SteamAppID)}

	item, err := app.FetchWorkshopMetadataForMod("local-mod")
	require.NoError(t, err)
	assert.Equal(t, steam.WorkshopItem{}, item)
}

func TestFetchWorkshopMetadataBatch_MapsByModID(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()
	workshopRoot := filepath.Join(t.TempDir(), "workshop", "content", eu5SteamAppID)
	app.gamePaths.WorkshopModDirs = []string{workshopRoot}
	app.modPathByID["mod-a"] = filepath.Join(workshopRoot, "111111")
	app.modPathByID["mod-b"] = filepath.Join(workshopRoot, "222222")
	app.modPathByID["mod-local"] = filepath.Join(t.TempDir(), "local", "my-mod")
	app.steamClient = &fakeSteamClient{itemsByID: map[string]steam.WorkshopItem{
		"111111": {ItemID: "111111", Title: "A"},
		"222222": {ItemID: "222222", Title: "B"},
	}}

	items, err := app.FetchWorkshopMetadataBatch([]string{"mod-a", "mod-local", "mod-b"})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "A", items["mod-a"].Title)
	assert.Equal(t, "222222", items["mod-b"].ItemID)
}

func TestWorkshopItemIDFromPath(t *testing.T) {
	workshopRoot := filepath.FromSlash("C:/Program Files (x86)/Steam/steamapps/workshop/content/" + eu5SteamAppID)
	modPath := filepath.Join(workshopRoot, "3691046296")

	itemID := workshopItemIDFromPath(modPath, []string{workshopRoot})
	require.Equal(t, "3691046296", itemID)

	nonWorkshop := workshopItemIDFromPath(filepath.FromSlash("C:/mods/localmod"), []string{workshopRoot})
	require.Equal(t, "", nonWorkshop)
}

func TestOpenWorkshopItem_FallbackOrder(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	attempts := make([]string, 0, 3)
	app.openURL = func(_, rawURL string) error {
		attempts = append(attempts, rawURL)
		if len(attempts) == 1 {
			return errSteamUnavailable
		}
		return nil
	}
	app.openInAppURL = func(rawURL string) error {
		attempts = append(attempts, rawURL)
		return nil
	}

	err := app.OpenWorkshopItem("12345")
	require.NoError(t, err)
	require.Len(t, attempts, 2)
	assert.Equal(t, "steam://url/CommunityFilePage/12345", attempts[0])
	assert.Equal(t, "https://steamcommunity.com/sharedfiles/filedetails/?id=12345", attempts[1])
}

func TestOpenWorkshopItem_InvalidID(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	err := app.OpenWorkshopItem("abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workshop item id is invalid")
}

func TestOpenExternalLink_NonSteamURLBrowserThenInApp(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	attempts := make([]string, 0, 2)
	app.openURL = func(_, rawURL string) error {
		attempts = append(attempts, "browser:"+rawURL)
		return errSteamUnavailable
	}
	app.openInAppURL = func(rawURL string) error {
		attempts = append(attempts, "inapp:"+rawURL)
		return nil
	}

	err := app.OpenExternalLink("https://example.com/docs")
	require.NoError(t, err)
	require.Len(t, attempts, 2)
	assert.Equal(t, "browser:https://example.com/docs", attempts[0])
	assert.Equal(t, "inapp:https://example.com/docs", attempts[1])
}

func TestOpenExternalLink_SteamCommunityURLSteamThenBrowser(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	attempts := make([]string, 0, 3)
	app.openURL = func(_, rawURL string) error {
		attempts = append(attempts, rawURL)
		if len(attempts) == 1 {
			return errSteamUnavailable
		}
		return nil
	}
	app.openInAppURL = func(rawURL string) error {
		attempts = append(attempts, rawURL)
		return nil
	}

	err := app.OpenExternalLink("https://steamcommunity.com/sharedfiles/filedetails/?id=123456")
	require.NoError(t, err)
	require.Len(t, attempts, 2)
	assert.Equal(t, "steam://url/CommunityFilePage/123456", attempts[0])
	assert.Equal(t, "https://steamcommunity.com/sharedfiles/filedetails/?id=123456", attempts[1])
}

func TestUnsubscribeWorkshopMod_NonWorkshopNoop(t *testing.T) {
	setCompileUnsubscribeFlag(t, true)
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	app.openURL = func(_, _ string) error {
		t.Fatalf("openURL should not be called for non-workshop no-op")
		return nil
	}

	err := app.UnsubscribeWorkshopMod("   ")
	require.NoError(t, err)
}

func TestUnsubscribeWorkshopMod_FeatureDisabled(t *testing.T) {
	setCompileUnsubscribeFlag(t, false)
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	err := app.UnsubscribeWorkshopMod("12345")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsubscribe feature is disabled")
}

func TestUnsubscribeWorkshopMod_InvalidID(t *testing.T) {
	setCompileUnsubscribeFlag(t, true)
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	err := app.UnsubscribeWorkshopMod("abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workshop item id")
}

func TestUnsubscribeWorkshopMod_SteamThenBrowserFallback(t *testing.T) {
	setCompileUnsubscribeFlag(t, true)
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()

	attempts := make([]string, 0, 3)
	app.openURL = func(_, rawURL string) error {
		attempts = append(attempts, rawURL)
		if len(attempts) == 1 {
			return errSteamUnavailable
		}
		return nil
	}
	app.openInAppURL = func(rawURL string) error {
		attempts = append(attempts, rawURL)
		return nil
	}

	err := app.UnsubscribeWorkshopMod("12345")
	require.NoError(t, err)
	require.Len(t, attempts, 2)
	assert.Equal(t, "steam://openurl/https://steamcommunity.com/sharedfiles/unsubscribe?id=12345", attempts[0])
	assert.Equal(t, "https://steamcommunity.com/sharedfiles/unsubscribe?id=12345", attempts[1])
}

func TestIsUnsubscribeEnabled(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	setCompileUnsubscribeFlag(t, true)
	assert.True(t, app.IsUnsubscribeEnabled())
	setCompileUnsubscribeFlag(t, false)
	assert.False(t, app.IsUnsubscribeEnabled())
}

