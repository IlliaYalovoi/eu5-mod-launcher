package mods

import (
	"eu5-mod-launcher/internal/logging"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const maxScanWorkers = 8

const maxErrorSamples = 3

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

		fromRoot, nextIndex, err := collectCandidatesFromRoot(root, index)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, fromRoot...)
		index = nextIndex
	}
	return candidates, nil
}

func collectCandidatesFromRoot(root string, startIndex int) ([]scanCandidate, int, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, startIndex, fmt.Errorf("resolve absolute mod root %q: %w", root, err)
	}

	entries, err := os.ReadDir(absRoot)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debugf("mods: skipping missing root %q", absRoot)
			return nil, startIndex, nil
		}

		return nil, startIndex, fmt.Errorf("read mod root %q: %w", absRoot, err)
	}

	out := make([]scanCandidate, 0, len(entries))
	index := startIndex
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modDir := filepath.Join(absRoot, entry.Name())
		descriptorPath, ok := resolveDescriptorPath(modDir)
		if !ok {
			continue
		}

		out = append(out, scanCandidate{
			index:          index,
			id:             entry.Name(),
			modDir:         modDir,
			descriptorPath: descriptorPath,
		})
		index++
	}

	return out, index, nil
}

func resolveDescriptorPath(modDir string) (string, bool) {
	descriptorPath := filepath.Join(modDir, "descriptor.mod")
	if _, err := os.Stat(descriptorPath); err == nil {
		return descriptorPath, true
	}

	jsonFallback := filepath.Join(modDir, ".metadata", "metadata.json")
	if _, fallbackErr := os.Stat(jsonFallback); fallbackErr != nil {
		return "", false
	}

	return jsonFallback, true
}

func scanCandidatesSequential(candidates []scanCandidate) []Mod {
	modsByID := make(map[string]Mod)
	errorCount := 0
	errorSamples := make([]string, 0, 3)

	for i := range candidates {
		candidate := candidates[i]
		if _, exists := modsByID[candidate.id]; exists {
			continue
		}
		descriptor, err := ParseDescriptor(candidate.descriptorPath)
		if err != nil {
			errorCount++
			if len(errorSamples) < maxErrorSamples {
				errorSamples = append(errorSamples, fmt.Sprintf("%q (%v)", candidate.modDir, err))
			}
			continue
		}
		modsByID[candidate.id] = Mod{
			ID:               candidate.id,
			Name:             descriptor.Name,
			Version:          descriptor.Version,
			SupportedVersion: descriptor.SupportedVersion,
			Tags:             descriptor.Tags,
			Description:      descriptor.Description,
			DirPath:          candidate.modDir,
		}
	}

	logParseErrorSummary(errorCount, errorSamples)
	return modsFromMapSorted(modsByID)
}

func scanCandidatesConcurrent(candidates []scanCandidate, workers int) []Mod {
	jobs := make(chan scanCandidate)
	results := make(chan scanResult, len(candidates))

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			runScanWorker(jobs, results)
		})
	}

	go func() {
		enqueueCandidates(candidates, jobs)
		close(jobs)
		wg.Wait()
		close(results)
	}()

	byIndex, errorCount, errorSamples := collectScanResults(results, len(candidates))
	modsByID := collectModsByID(candidates, byIndex)

	logParseErrorSummary(errorCount, errorSamples)
	return modsFromMapSorted(modsByID)
}

func runScanWorker(jobs <-chan scanCandidate, results chan<- scanResult) {
	for candidate := range jobs {
		descriptor, err := ParseDescriptor(candidate.descriptorPath)
		if err != nil {
			results <- scanResult{
				index: candidate.index,
				err:   fmt.Errorf("%q: %w", candidate.modDir, err),
			}
			continue
		}

		results <- scanResult{index: candidate.index, mod: Mod{
			ID:               candidate.id,
			Name:             descriptor.Name,
			Version:          descriptor.Version,
			SupportedVersion: descriptor.SupportedVersion,
			Tags:             descriptor.Tags,
			Description:      descriptor.Description,
			DirPath:          candidate.modDir,
		}}
	}
}

func enqueueCandidates(candidates []scanCandidate, jobs chan<- scanCandidate) {
	for i := range candidates {
		candidate := candidates[i]
		jobs <- candidate
	}
}

func collectScanResults(results <-chan scanResult, size int) (map[int]scanResult, int, []string) {
	byIndex := make(map[int]scanResult, size)
	errorCount := 0
	errorSamples := make([]string, 0, maxErrorSamples)
	for result := range results {
		byIndex[result.index] = result
		if result.err == nil {
			continue
		}

		errorCount++
		if len(errorSamples) < maxErrorSamples {
			errorSamples = append(errorSamples, result.err.Error())
		}
	}

	return byIndex, errorCount, errorSamples
}

func collectModsByID(candidates []scanCandidate, byIndex map[int]scanResult) map[string]Mod {
	modsByID := make(map[string]Mod)
	for i := range candidates {
		candidate := candidates[i]
		if _, exists := modsByID[candidate.id]; exists {
			continue
		}

		result, ok := byIndex[candidate.index]
		if !ok || result.err != nil {
			continue
		}

		modsByID[candidate.id] = result.mod
	}

	return modsByID
}

func modsFromMapSorted(modsByID map[string]Mod) []Mod {
	mods := make([]Mod, 0, len(modsByID))
	for id := range modsByID {
		mod := modsByID[id]
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
	logging.Warnf(
		"mods: skipped %d entries due to descriptor errors (examples: %s)",
		total,
		strings.Join(samples, "; "),
	)
}
