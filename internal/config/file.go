package config

import (
	"os"
	"path/filepath"
)

func FindRepoRoot(start string) string {
	dir := start
	for {
		headPath := filepath.Join(dir, ".loki", "HEAD")
		if _, err := os.Stat(headPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
