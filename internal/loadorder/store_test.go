package loadorder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "state", "loadorder.json")

	store, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	state := State{OrderedIDs: []string{"mod.a", "mod.b", "mod.c"}}
	if saveErr := store.Save(state); saveErr != nil {
		t.Fatalf("Save() error = %v", saveErr)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(loaded, state) {
		t.Fatalf("Load() = %#v, want %#v", loaded, state)
	}
}

func TestNewCreatesParentDir(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "nested", "deeper", "loadorder.json")

	if _, err := os.Stat(filepath.Dir(configPath)); !os.IsNotExist(err) {
		t.Fatalf("expected parent directory to not exist before New(), err = %v", err)
	}

	store, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := os.Stat(filepath.Dir(store.ConfigPath())); err != nil {
		t.Fatalf("parent directory should exist after New(): %v", err)
	}
}

func TestSaveAtomicWriteNoTmpLeftBehind(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "loadorder.json")

	store, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if firstSaveErr := store.Save(State{OrderedIDs: []string{"old.mod"}}); firstSaveErr != nil {
		t.Fatalf("first Save() error = %v", firstSaveErr)
	}
	if secondSaveErr := store.Save(State{OrderedIDs: []string{"new.mod", "next.mod"}}); secondSaveErr != nil {
		t.Fatalf("second Save() error = %v", secondSaveErr)
	}

	if _, statErr := os.Stat(configPath + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatalf("temporary file should not remain after Save(), stat err = %v", statErr)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := State{OrderedIDs: []string{"new.mod", "next.mod"}}
	if !reflect.DeepEqual(loaded, want) {
		t.Fatalf("Load() = %#v, want %#v", loaded, want)
	}
}
