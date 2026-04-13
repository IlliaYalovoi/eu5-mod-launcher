package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func DiscoverSteamLibraryRoots() []string {
	steamRoot := ""
	if runtime.GOOS == "windows" {
		fallbacks := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
			filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
		}
		for _, candidate := range fallbacks {
			if DirExists(candidate) {
				steamRoot = filepath.Clean(candidate)
				break
			}
		}
	}

	if steamRoot == "" {
		return nil
	}

	libraryFoldersPath := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")
	libraryRoots := append([]string{steamRoot}, ParseLibraryFoldersVDF(libraryFoldersPath)...)
	seen := make(map[string]struct{}, len(libraryRoots))
	out := make([]string, 0, len(libraryRoots))

	for _, root := range libraryRoots {
		if strings.TrimSpace(root) == "" {
			continue
		}

		cleanRoot := filepath.Clean(root)
		if _, ok := seen[cleanRoot]; ok {
			continue
		}
		seen[cleanRoot] = struct{}{}
		out = append(out, cleanRoot)
	}

	return out
}

func ParseLibraryFoldersVDF(vdfPath string) []string {
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	matches := regexp.MustCompile(`"path"\s*"([^"]+)"`).FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]string, 0, len(matches))
	for _, match := range matches {
		raw := match[1]
		raw = strings.ReplaceAll(raw, `\\`, `\`)
		if strings.TrimSpace(raw) == "" {
			continue
		}
		out = append(out, filepath.Clean(raw))
	}

	return out
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DiscoverWorkshopModDirs(appID string) []string {
	libraryRoots := DiscoverSteamLibraryRoots()
	if len(libraryRoots) == 0 {
		return nil
	}

	out := make([]string, 0, len(libraryRoots))
	for _, cleanRoot := range libraryRoots {
		workshopDir := filepath.Join(cleanRoot, "steamapps", "workshop", "content", appID)
		if DirExists(workshopDir) {
			out = append(out, workshopDir)
		}
	}

	return out
}
