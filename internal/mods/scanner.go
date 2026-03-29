package mods

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"eu5-mod-launcher/internal/logging"
)

const maxScanWorkers = 8

type scanCandidate struct {
	index          int
	id             string
	modDir         string
	descriptorPath string
}

type scanResult struct {
	index int
	mod   Mod
	err   error
}

// ScanDir walks dirPath and returns one Mod per valid mod subdirectory.
// Errors reading individual mods are logged and skipped, not fatal.
func ScanDir(dirPath string) ([]Mod, error) {
	return ScanDirs([]string{dirPath})
}

// ScanDirs walks multiple roots and returns one Mod per valid mod subdirectory.
// Missing roots are skipped to support optional local/workshop layouts.
func ScanDirs(dirPaths []string) ([]Mod, error) {
	return scanDirsWithWorkers(dirPaths, defaultScanWorkerCount())
}

func scanDirsWithWorkers(dirPaths []string, workers int) ([]Mod, error) {
	candidates, err := collectScanCandidates(dirPaths)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []Mod{}, nil
	}

	if workers <= 1 {
		return scanCandidatesSequential(candidates), nil
	}
	return scanCandidatesConcurrent(candidates, workers), nil
}

func collectScanCandidates(dirPaths []string) ([]scanCandidate, error) {
	candidates := make([]scanCandidate, 0)
	index := 0
	for _, root := range dirPaths {
		if strings.TrimSpace(root) == "" {
			continue
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("resolve absolute mod root %q: %w", root, err)
		}

		entries, err := os.ReadDir(absRoot)
		if err != nil {
			if os.IsNotExist(err) {
				logging.Debugf("mods: skipping missing root %q", absRoot)
				continue
			}
			return nil, fmt.Errorf("read mod root %q: %w", absRoot, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			modDir := filepath.Join(absRoot, entry.Name())
			descriptorPath := filepath.Join(modDir, "descriptor.mod")
			if _, err := os.Stat(descriptorPath); err != nil {
				jsonFallback := filepath.Join(modDir, ".metadata", "metadata.json")
				if _, fallbackErr := os.Stat(jsonFallback); fallbackErr != nil {
					continue
				}
				descriptorPath = jsonFallback
			}

			candidates = append(candidates, scanCandidate{
				index:          index,
				id:             entry.Name(),
				modDir:         modDir,
				descriptorPath: descriptorPath,
			})
			index++
		}
	}
	return candidates, nil
}

func scanCandidatesSequential(candidates []scanCandidate) []Mod {
	modsByID := make(map[string]Mod)
	errorCount := 0
	errorSamples := make([]string, 0, 3)

	for _, candidate := range candidates {
		if _, exists := modsByID[candidate.id]; exists {
			continue
		}
		name, version, description, tags, err := ParseDescriptor(candidate.descriptorPath)
		if err != nil {
			errorCount++
			if len(errorSamples) < 3 {
				errorSamples = append(errorSamples, fmt.Sprintf("%q (%v)", candidate.modDir, err))
			}
			continue
		}
		modsByID[candidate.id] = Mod{
			ID:          candidate.id,
			Name:        name,
			Version:     version,
			Tags:        tags,
			Description: description,
			DirPath:     candidate.modDir,
		}
	}

	logParseErrorSummary(errorCount, errorSamples)
	return modsFromMapSorted(modsByID)
}

func scanCandidatesConcurrent(candidates []scanCandidate, workers int) []Mod {
	jobs := make(chan scanCandidate)
	results := make(chan scanResult, len(candidates))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				name, version, description, tags, err := ParseDescriptor(candidate.descriptorPath)
				if err != nil {
					results <- scanResult{index: candidate.index, err: fmt.Errorf("%q (%v)", candidate.modDir, err)}
					continue
				}
				results <- scanResult{index: candidate.index, mod: Mod{
					ID:          candidate.id,
					Name:        name,
					Version:     version,
					Tags:        tags,
					Description: description,
					DirPath:     candidate.modDir,
				}}
			}
		}()
	}

	go func() {
		for _, candidate := range candidates {
			jobs <- candidate
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	byIndex := make(map[int]scanResult, len(candidates))
	errorCount := 0
	errorSamples := make([]string, 0, 3)
	for result := range results {
		byIndex[result.index] = result
		if result.err != nil {
			errorCount++
			if len(errorSamples) < 3 {
				errorSamples = append(errorSamples, result.err.Error())
			}
		}
	}

	modsByID := make(map[string]Mod)
	for _, candidate := range candidates {
		if _, exists := modsByID[candidate.id]; exists {
			continue
		}
		result, ok := byIndex[candidate.index]
		if !ok || result.err != nil {
			continue
		}
		modsByID[candidate.id] = result.mod
	}

	logParseErrorSummary(errorCount, errorSamples)
	return modsFromMapSorted(modsByID)
}

func modsFromMapSorted(modsByID map[string]Mod) []Mod {
	mods := make([]Mod, 0, len(modsByID))
	for _, mod := range modsByID {
		mods = append(mods, mod)
	}
	sort.Slice(mods, func(i, j int) bool {
		return mods[i].ID < mods[j].ID
	})
	return mods
}

func defaultScanWorkerCount() int {
	count := runtime.NumCPU()
	if count < 1 {
		return 1
	}
	if count > maxScanWorkers {
		return maxScanWorkers
	}
	return count
}

func logParseErrorSummary(total int, samples []string) {
	if total == 0 {
		return
	}
	if len(samples) == 0 {
		logging.Warnf("mods: skipped %d entries due to descriptor errors", total)
		return
	}
	logging.Warnf("mods: skipped %d entries due to descriptor errors (examples: %s)", total, strings.Join(samples, "; "))
}
