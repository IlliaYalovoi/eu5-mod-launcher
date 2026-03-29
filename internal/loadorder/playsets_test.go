package loadorder

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadStateFromPlaysets(t *testing.T) {
	root := t.TempDir()
	playsetsPath := filepath.Join(root, "playsets.json")
	content := `{
  "file_version": "1.0.0",
  "playsets": [
    {
      "name": "Default",
      "orderedListMods": [
        {"path": "C:/Steam/workshop/content/3450310/111/", "isEnabled": true},
        {"path": "C:/Steam/workshop/content/3450310/222/", "isEnabled": false},
        {"path": "C:/Users/Alice/Documents/Paradox Interactive/Europa Universalis V/mod/local_mod/", "isEnabled": true}
      ]
	},
	{
	  "name": "Secondary",
	  "orderedListMods": [
		{"path": "C:/Steam/workshop/content/3450310/333/", "isEnabled": true}
	  ]
    }
  ]
}`
	if err := os.WriteFile(playsetsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	state, idToPath, err := LoadStateFromPlaysets(playsetsPath, 1)
	if err != nil {
		t.Fatalf("LoadStateFromPlaysets() error = %v", err)
	}

	want := State{OrderedIDs: []string{"333"}}
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("LoadStateFromPlaysets() state = %#v, want %#v", state, want)
	}
	if idToPath["333"] == "" {
		t.Fatalf("LoadStateFromPlaysets() should provide path mapping for loaded IDs")
	}
}

func TestListPlaysetsDetectsGameActive(t *testing.T) {
	root := t.TempDir()
	playsetsPath := filepath.Join(root, "playsets.json")
	content := `{
  "file_version": "1.0.0",
  "playsets": [
    {"name": "A", "orderedListMods": []},
    {"name": "B", "isActive": true, "orderedListMods": []}
  ]
}`
	if err := os.WriteFile(playsetsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	names, gameActive, err := ListPlaysets(playsetsPath)
	if err != nil {
		t.Fatalf("ListPlaysets() error = %v", err)
	}
	if !reflect.DeepEqual(names, []string{"A", "B"}) {
		t.Fatalf("ListPlaysets() names = %v", names)
	}
	if gameActive != 1 {
		t.Fatalf("ListPlaysets() gameActive = %d, want 1", gameActive)
	}
}

func TestSaveStateToPlaysetsUpdatesSelectedPlaysetOnly(t *testing.T) {
	root := t.TempDir()
	playsetsPath := filepath.Join(root, "playsets.json")
	original := `{
  "file_version": "1.0.0",
  "playsets": [
    {
      "name": "Primary",
      "isAutomaticallySorted": true,
      "orderedListMods": []
    },
    {
      "name": "Secondary",
      "orderedListMods": [{"path":"C:/old/entry/","isEnabled":true}]
    }
  ]
}`
	if err := os.WriteFile(playsetsPath, []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	state := State{OrderedIDs: []string{"modA", "modB"}}
	idToPath := map[string]string{
		"modA": `C:\Mods\modA`,
		"modB": `D:\Workshop\3450310\modB`,
	}

	if err := SaveStateToPlaysets(playsetsPath, 1, state, idToPath); err != nil {
		t.Fatalf("SaveStateToPlaysets() error = %v", err)
	}

	updated, err := os.ReadFile(playsetsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(updated)
	if !strings.Contains(text, `"name": "Primary"`) {
		t.Fatalf("updated playsets should preserve first playset metadata")
	}
	if !strings.Contains(text, `"name": "Primary"`) || !strings.Contains(text, `"orderedListMods": []`) {
		t.Fatalf("primary playset should remain unchanged, got: %s", text)
	}
	if !strings.Contains(text, `"path": "C:/Mods/modA/"`) {
		t.Fatalf("expected modA path entry in selected playset, got: %s", text)
	}
	if !strings.Contains(text, `"path": "D:/Workshop/3450310/modB/"`) {
		t.Fatalf("expected modB path entry in selected playset, got: %s", text)
	}
	if !strings.Contains(text, `"name": "Secondary"`) {
		t.Fatalf("other playsets should remain present")
	}
}
