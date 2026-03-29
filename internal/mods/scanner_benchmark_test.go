package mods

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScanDirs_SequentialVsConcurrent(b *testing.B) {
	root := b.TempDir()
	for i := 0; i < 300; i++ {
		id := fmt.Sprintf("mod_%03d", i)
		modDir := filepath.Join(root, id)
		if err := os.MkdirAll(modDir, 0o755); err != nil {
			b.Fatalf("MkdirAll() error = %v", err)
		}
		content := fmt.Sprintf("name=\"Mod %d\"\nversion=\"1\"\n", i)
		if err := os.WriteFile(filepath.Join(modDir, "descriptor.mod"), []byte(content), 0o644); err != nil {
			b.Fatalf("WriteFile() error = %v", err)
		}
	}

	b.Run("sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mods, err := scanDirsWithWorkers([]string{root}, 1)
			if err != nil {
				b.Fatalf("scanDirsWithWorkers() error = %v", err)
			}
			if len(mods) != 300 {
				b.Fatalf("mods len = %d, want 300", len(mods))
			}
		}
	})

	b.Run("concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mods, err := scanDirsWithWorkers([]string{root}, defaultScanWorkerCount())
			if err != nil {
				b.Fatalf("scanDirsWithWorkers() error = %v", err)
			}
			if len(mods) != 300 {
				b.Fatalf("mods len = %d, want 300", len(mods))
			}
		}
	})
}
