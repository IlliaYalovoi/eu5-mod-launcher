package mods

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type scanDirCase struct {
	name         string
	setup        func(t *testing.T, root string)
	wantIDs      []string
	wantNames    []string
	wantVersions []string
}

func TestScanDir(t *testing.T) {
	testCases := []scanDirCase{
		{
			name: "scans valid mods and skips missing descriptor",
			setup: func(t *testing.T, root string) {
				writeDescriptor(t, filepath.Join(root, "modA"), "name=\"Alpha\"\nversion=\"1.0\"\n")
				writeDescriptor(t, filepath.Join(root, "modB"), "name=\"Beta\"\nversion=\"2.0\"\n")
				writeDescriptor(t, filepath.Join(root, "modC"), "name=\"Gamma\"\nversion=\"3.0\"\n")

				if err := os.Mkdir(filepath.Join(root, "empty"), 0o750); err != nil {
					t.Fatalf("Mkdir() error = %v", err)
				}
			},
			wantIDs:      []string{"modA", "modB", "modC"},
			wantNames:    []string{"Alpha", "Beta", "Gamma"},
			wantVersions: []string{"1.0", "2.0", "3.0"},
		},
		{
			name: "supports .metadata metadata.json fallback",
			setup: func(t *testing.T, root string) {
				jsonDir := filepath.Join(root, "jsonOnly", ".metadata")
				if err := os.MkdirAll(jsonDir, 0o750); err != nil {
					t.Fatalf("MkdirAll() error = %v", err)
				}
				content := "{\"name\":\"From JSON\",\"version\":\"9.9\",\"short_description\":\"desc\",\"tags\":[\"UI\"]}"
				if err := os.WriteFile(filepath.Join(jsonDir, "metadata.json"), []byte(content), 0o600); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			wantIDs:      []string{"jsonOnly"},
			wantNames:    []string{"From JSON"},
			wantVersions: []string{"9.9"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runScanDirCase(t, tc)
		})
	}
}

func runScanDirCase(t *testing.T, tc scanDirCase) {
	t.Helper()

	root := t.TempDir()
	tc.setup(t, root)

	mods, err := ScanDir(root)
	if err != nil {
		t.Fatalf("ScanDir() error = %v", err)
	}

	gotIDs := make([]string, 0, len(mods))
	gotNames := make([]string, 0, len(mods))
	gotVersions := make([]string, 0, len(mods))
	for i := range mods {
		mod := mods[i]
		gotIDs = append(gotIDs, mod.ID)
		gotNames = append(gotNames, mod.Name)
		gotVersions = append(gotVersions, mod.Version)
	}

	if !reflect.DeepEqual(gotIDs, tc.wantIDs) {
		t.Fatalf("ScanDir() IDs = %v, want %v", gotIDs, tc.wantIDs)
	}
	if !reflect.DeepEqual(gotNames, tc.wantNames) {
		t.Fatalf("ScanDir() names = %v, want %v", gotNames, tc.wantNames)
	}
	if !reflect.DeepEqual(gotVersions, tc.wantVersions) {
		t.Fatalf("ScanDir() versions = %v, want %v", gotVersions, tc.wantVersions)
	}
}

func TestScanDirs_MultiRootAndDeduplicate(t *testing.T) {
	localRoot := t.TempDir()
	workshopRoot := t.TempDir()

	writeDescriptor(t, filepath.Join(localRoot, "shared"), "name=\"Local\"\nversion=\"1.0\"\n")
	writeDescriptor(t, filepath.Join(localRoot, "localOnly"), "name=\"Local Only\"\nversion=\"1.1\"\n")
	writeDescriptor(t, filepath.Join(workshopRoot, "shared"), "name=\"Workshop\"\nversion=\"2.0\"\n")
	writeDescriptor(t, filepath.Join(workshopRoot, "workshopOnly"), "name=\"Workshop Only\"\nversion=\"2.1\"\n")

	mods, err := ScanDirs([]string{localRoot, workshopRoot, filepath.Join(workshopRoot, "missing")})
	if err != nil {
		t.Fatalf("ScanDirs() error = %v", err)
	}

	if len(mods) != 3 {
		t.Fatalf("ScanDirs() returned %d mods, want 3", len(mods))
	}

	byID := make(map[string]Mod, len(mods))
	for i := range mods {
		mod := mods[i]
		byID[mod.ID] = mod
	}

	if byID["shared"].Name != "Local" {
		t.Fatalf("duplicate ID should keep first root mod, got %q", byID["shared"].Name)
	}
	if _, ok := byID["localOnly"]; !ok {
		t.Fatalf("expected localOnly mod to be present")
	}
	if _, ok := byID["workshopOnly"]; !ok {
		t.Fatalf("expected workshopOnly mod to be present")
	}
}

func TestScanDirs_DeterministicWithConcurrency(t *testing.T) {
	root := t.TempDir()
	writeDescriptor(t, filepath.Join(root, "a_mod"), "name=\"A\"\nversion=\"1\"\n")
	writeDescriptor(t, filepath.Join(root, "b_mod"), "name=\"B\"\nversion=\"1\"\n")
	writeDescriptor(t, filepath.Join(root, "c_mod"), "name=\"C\"\nversion=\"1\"\n")

	var baseline []string
	for i := 0; i < 20; i++ {
		mods, err := scanDirsWithWorkers([]string{root}, 4)
		if err != nil {
			t.Fatalf("scanDirsWithWorkers() error = %v", err)
		}
		ids := make([]string, 0, len(mods))
		for j := range mods {
			mod := mods[j]
			ids = append(ids, mod.ID)
		}
		if i == 0 {
			baseline = ids
			continue
		}
		if !reflect.DeepEqual(ids, baseline) {
			t.Fatalf("run %d IDs = %v, want %v", i, ids, baseline)
		}
	}
}

func TestScanDirs_SkipsBrokenDescriptor(t *testing.T) {
	root := t.TempDir()
	writeDescriptor(t, filepath.Join(root, "good"), "name=\"Good\"\nversion=\"1\"\n")
	if err := os.MkdirAll(filepath.Join(root, "broken"), 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "broken", "descriptor.mod"), []byte("tags={\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	mods, err := scanDirsWithWorkers([]string{root}, 4)
	if err != nil {
		t.Fatalf("scanDirsWithWorkers() error = %v", err)
	}
	if len(mods) != 1 || mods[0].ID != "good" {
		t.Fatalf("scanDirsWithWorkers() = %v, want only good mod", mods)
	}
}

func writeDescriptor(t *testing.T, modDir, descriptorContent string) {
	t.Helper()

	if err := os.MkdirAll(modDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "descriptor.mod"), []byte(descriptorContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
