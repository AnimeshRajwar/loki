package commands

import (
	"fmt"
	"loki/internal/utils"
	"os"
	"path/filepath"
)

func Init() {
	dirs := []string{
		".loki",
		".loki/objects",
		".loki/refs",
		".loki/refs/heads",
	}

	for _, d := range dirs {
		_ = os.MkdirAll(d, 0755)
	}

	_ = os.WriteFile(".loki/HEAD", []byte("ref: refs/heads/main"), 0644)

	configPath := filepath.Join(".loki", "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		_ = os.WriteFile(configPath, []byte(""), 0644)
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(cwd)
	fmt.Printf(utils.ColorText("Initialized empty Loki repository at %s\n", "success"), utils.ColorText(absPath, "notice"))
}
