package main

import (
	"errors"
	"eu5-mod-launcher/internal/steam"
	"path/filepath"
	"strings"
	"time"
)

const (
	steamMetadataTTL        = 24 * time.Hour
	steamMetadataMaxEntries = 5000
	steamImageMaxEntries    = 1000
	steamDescImageMaxEntry  = 3000
)

var errSteamCacheRootEmpty = errors.New("steam cache root is empty")

func (a *App) ensureSteamCaches() error {
	if a.svc.steamMeta != nil && a.svc.steamImage != nil && a.svc.steamDesc != nil {
		return nil
	}
	cacheRoot := filepath.Dir(a.settingsPath)
	if strings.TrimSpace(cacheRoot) == "" && a.svc.loadOrderRepo != nil {
		cacheRoot = filepath.Dir(a.svc.loadOrderRepo.Path())
	}
	if strings.TrimSpace(cacheRoot) == "" {
		return errSteamCacheRootEmpty
	}
	metaCache, err := steam.NewMetadataCache(cacheRoot, steamMetadataTTL, steamMetadataMaxEntries)
	if err != nil {
		return err
	}
	imageCache, err := steam.NewImageCache(cacheRoot, steamImageMaxEntries, nil)
	if err != nil {
		return err
	}
	descCache, err := steam.NewDescriptionImageCache(cacheRoot, steamDescImageMaxEntry, nil)
	if err != nil {
		return err
	}
	a.svc.steamMeta = metaCache
	a.svc.steamImage = imageCache
	a.svc.steamDesc = descCache
	return nil
}

func (a *App) openURLInApp(rawURL string) error {
	return nil
}

func (a *App) workshopItemIDForMod(modID string) string {
	return ""
}

func (a *App) resolveImageSource(cachedPath string) string {
	if cachedPath == "" {
		return ""
	}
	a.imageDataMu.Lock()
	defer a.imageDataMu.Unlock()
	if v, ok := a.imageData[cachedPath]; ok {
		return v
	}
	return ""
}
