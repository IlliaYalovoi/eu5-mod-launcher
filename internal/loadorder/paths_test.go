package loadorder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigPathForOS(t *testing.T) {
	testCases := []struct {
		name    string
		goos    string
		env     map[string]string
		want    string
		wantErr bool
	}{
		{
			name: "windows path from APPDATA",
			goos: "windows",
			env: map[string]string{
				"APPDATA": `C:\Users\Alice\AppData\Roaming`,
			},
			want: `C:\Users\Alice\AppData\Roaming\EU5ModLauncher\loadorder.json`,
		},
		{
			name:    "windows requires APPDATA",
			goos:    "windows",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name: "linux path from XDG_CONFIG_HOME",
			goos: "linux",
			env: map[string]string{
				"XDG_CONFIG_HOME": "/home/alice/.cfg",
			},
			want: "/home/alice/.cfg/eu5-mod-launcher/loadorder.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tc.env[key]
			}

			got, err := defaultConfigPathForOS(tc.goos, getenv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("defaultConfigPathForOS() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("defaultConfigPathForOS() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("defaultConfigPathForOS() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDefaultConfigPathLinuxFallbackUsesHomeConfigDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, err := defaultConfigPathForOS("linux", func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return ""
		}
		return ""
	})
	if err != nil {
		t.Fatalf("defaultConfigPathForOS() error = %v", err)
	}

	if !strings.Contains(got, filepath.Join(".config", "eu5-mod-launcher", "loadorder.json")) {
		t.Fatalf("linux fallback path %q does not include expected suffix", got)
	}
}

func TestParseLibraryFoldersVDF(t *testing.T) {
	root := t.TempDir()
	vdfPath := filepath.Join(root, "libraryfolders.vdf")
	content := `"libraryfolders"
{
	"0"
	{
		"path"		"D:\\\\SteamLibrary"
	}
	"1"
	{
		"path"		"E:\\\\Games\\\\Steam"
	}
}`

	if err := os.WriteFile(vdfPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	paths := parseLibraryFoldersVDF(vdfPath)
	if len(paths) != 2 {
		t.Fatalf("parseLibraryFoldersVDF() returned %d paths, want 2", len(paths))
	}

	if paths[0] != filepath.Clean(`D:\SteamLibrary`) {
		t.Fatalf("first parsed path = %q", paths[0])
	}
	if paths[1] != filepath.Clean(`E:\Games\Steam`) {
		t.Fatalf("second parsed path = %q", paths[1])
	}
}

func TestDiscoverGameExePathFromSteamLibraries(t *testing.T) {
	steamRoot := t.TempDir()
	libraryB := filepath.Join(t.TempDir(), "LibraryB")
	if err := os.MkdirAll(filepath.Join(steamRoot, "steamapps"), 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	vdf := `"libraryfolders"
{
	"0" { "path" "` + strings.ReplaceAll(filepath.ToSlash(libraryB), "/", "\\") + `" }
}`
	if err := os.WriteFile(filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf"), []byte(vdf), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gameExe := filepath.Join(libraryB, "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
	if err := os.MkdirAll(filepath.Dir(gameExe), 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(gameExe, []byte(""), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	originalFinder := steamInstallPathFinder
	steamInstallPathFinder = func() string { return steamRoot }
	t.Cleanup(func() { steamInstallPathFinder = originalFinder })

	got := discoverGameExePath()
	if got != filepath.Clean(gameExe) {
		t.Fatalf("discoverGameExePath() = %q, want %q", got, filepath.Clean(gameExe))
	}
}
