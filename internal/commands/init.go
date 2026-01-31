package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init() {
	dirs := []string{
		".loki",
		".loki/objects",
		".loki/refs",
	}

	for _, d := range dirs {
		_ = os.MkdirAll(d, 0755)
	}

	_ = os.WriteFile(".loki/HEAD", []byte("ref: refs/main"), 0644)

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(cwd)
	fmt.Printf("Initialized empty Loki repository at %s\n", absPath)
}
