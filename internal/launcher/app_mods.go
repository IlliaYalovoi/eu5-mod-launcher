package launcher

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/mods"
	"fmt"
)

func (a *App) GetAllMods() ([]mods.Mod, error) {
	if err := a.mustBeReady(); err != nil {
		return nil, err
	}
	roots := make([]string, 0, 1+len(a.gamePaths.WorkshopModDirs))
	roots = append(roots, a.effectiveModsDir())
	roots = append(roots, a.gamePaths.WorkshopModDirs...)

	allMods, nextPaths, err := a.svc.modsService.Discover(roots, a.loadOrder.OrderedIDs, a.modPathByID)
	if err != nil {
		logging.Errorf("mods scan failed for roots %q: %v", roots, err)
		return nil, fmt.Errorf("get all mods: %w", err)
	}
	a.modPathByID = nextPaths

	for i := range allMods {
		itemID := a.workshopItemIDForMod(allMods[i].ID)
		if itemID == "" || a.svc.steamImage == nil {
			continue
		}
		if cachedPath := a.svc.steamImage.CachedPath(itemID); cachedPath != "" {
			if src := a.resolveImageSource(cachedPath); src != "" {
				allMods[i].ThumbnailPath = src
			} else {
				allMods[i].ThumbnailPath = cachedPath
			}
		}
	}
	return allMods, nil
}

func (a *App) GetLoadOrder() []string {
	if err := a.mustBeReady(); err != nil {
		return []string{}
	}
	return append([]string(nil), a.loadOrder.OrderedIDs...)
}

func (a *App) SetLoadOrder(ids []string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	next, err := a.svc.loadorderSvc.ValidateAndNormalize(ids)
	if err != nil {
		return fmt.Errorf("set load order: %w", err)
	}
	newOrder := domain.LoadOrder{GameID: a.activeGameID, PlaysetIdx: domain.PlaysetIndex(a.launcherIdx), OrderedIDs: next}
	if err := a.svc.loadOrderRepo.Save(newOrder); err != nil {
		return fmt.Errorf("save fallback load order: %w", err)
	}
	if a.gamePaths.PlaysetsPath != "" {
		if err := a.svc.gameSvc.ExportModList(a.activeGameID, a.gamePaths.PlaysetsPath, a.launcherIdx, newOrder, a.modPathByID); err != nil {
			return fmt.Errorf("save load order to playsets %q: %w", a.gamePaths.PlaysetsPath, err)
		}
	}
	a.loadOrder = newOrder
	nextLayout := a.launcherLayout
	if err := a.svc.layoutSvc.Persist(&nextLayout, a.loadOrder.OrderedIDs); err != nil {
		logging.Warnf("set load order: failed to save launcher layout: %v", err)
	} else {
		a.launcherLayout = nextLayout
	}
	return nil
}

func (a *App) EnableMod(id string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	next, err := a.svc.loadorderSvc.Enable(a.loadOrder.OrderedIDs, id)
	if err != nil {
		return fmt.Errorf("enable mod %q: %w", id, err)
	}
	return a.SetLoadOrder(next)
}

func (a *App) DisableMod(id string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	next, err := a.svc.loadorderSvc.Disable(a.loadOrder.OrderedIDs, id)
	if err != nil {
		return fmt.Errorf("disable mod %q: %w", id, err)
	}
	return a.SetLoadOrder(next)
}
