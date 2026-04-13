package mods

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type parseDescriptorCase struct {
	name        string
	fileName    string
	content     string
	wantName    string
	wantVersion string
	wantDesc    string
	wantTags    []string
	wantErr     bool
}

func TestParseDescriptor(t *testing.T) {
	testCases := []parseDescriptorCase{
		{
			name:        "parses descriptor.mod fields",
			fileName:    "descriptor.mod",
			content:     "name=\"Mod A\"\nversion=\"1.2.3\"\ndescription=\"Desc\"\ntags={\"Gameplay\",\"Balance\"}\n",
			wantName:    "Mod A",
			wantVersion: "1.2.3",
			wantDesc:    "Desc",
			wantTags:    []string{"Gameplay", "Balance"},
		},
		{
			name:        "unknown keys are ignored and defaults are empty",
			fileName:    "descriptor.mod",
			content:     "foo=\"bar\"\nname=\"Only Name\"\n",
			wantName:    "Only Name",
			wantVersion: "",
			wantDesc:    "",
			wantTags:    nil,
		},
		{
			name:        "json descriptor supported",
			fileName:    "metadata.json",
			content:     "{\"name\":\"Mod J\",\"version\":\"9\",\"short_description\":\"JSON\",\"tags\":[\"UI\"]}",
			wantName:    "Mod J",
			wantVersion: "9",
			wantDesc:    "JSON",
			wantTags:    []string{"UI"},
		},
		{
			name:        "json descriptor with utf8 bom supported",
			fileName:    "metadata.json",
			content:     "\ufeff{\"name\":\"BOM Mod\",\"version\":\"1\",\"description\":\"BOM\",\"tags\":[\"Map\"]}",
			wantName:    "BOM Mod",
			wantVersion: "1",
			wantDesc:    "BOM",
			wantTags:    []string{"Map"},
		},
		{
			name:        "json descriptor with raw newline in string is tolerated",
			fileName:    "metadata.json",
			content:     "{\"name\":\"Broken JSON\",\"version\":\"1\",\"short_description\":\"line one\nline two\",\"tags\":[\"Fixes\"]}",
			wantName:    "Broken JSON",
			wantVersion: "1",
			wantDesc:    "line one\nline two",
			wantTags:    []string{"Fixes"},
		},
		{
			name:     "invalid json returns error",
			fileName: "metadata.json",
			content:  "{",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runParseDescriptorCase(t, tc)
		})
	}
}

func runParseDescriptorCase(t *testing.T, tc parseDescriptorCase) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, tc.fileName)
	if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	parsed, err := ParseDescriptor(path)
	if tc.wantErr {
		if err == nil {
			t.Fatalf("ParseDescriptor() error = nil, want error")
		}
		return
	}
	if err != nil {
		t.Fatalf("ParseDescriptor() unexpected error = %v", err)
	}

	if parsed.Name != tc.wantName || parsed.Version != tc.wantVersion || parsed.Description != tc.wantDesc {
		t.Fatalf(
			"ParseDescriptor() values = (%q, %q, %q), want (%q, %q, %q)",
			parsed.Name,
			parsed.Version,
			parsed.Description,
			tc.wantName,
			tc.wantVersion,
			tc.wantDesc,
		)
	}

	if !reflect.DeepEqual(parsed.Tags, tc.wantTags) {
		t.Fatalf("ParseDescriptor() tags = %v, want %v", parsed.Tags, tc.wantTags)
	}
}
