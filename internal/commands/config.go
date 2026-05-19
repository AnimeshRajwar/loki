package commands

import (
	"fmt"
	"loki/internal/config"
	"loki/internal/utils"
	"os"
	"strings"
)

func Config(args []string) {
	if len(args) < 1 {
		fmt.Println(utils.ColorText("Usage: loki config [--local|--global|--system] key [value]", "warning"))
		return
	}

	level := "local" // default
	key := ""
	value := ""
	for _, arg := range args {
		if arg == "--local" || arg == "--global" || arg == "--system" {
			level = strings.TrimPrefix(arg, "--")
		} else if key == "" {
			key = arg
		} else {
			value = arg
		}
	}

	repoRoot := config.FindRepoRoot(os.Getenv("PWD"))
	cfg := config.NewConfig()
	cfg.Load(repoRoot)

	if value == "" {
		// Get
		v := cfg.Get(key)
		if v == "" {
			fmt.Println(utils.ColorText(fmt.Sprintf("%s not set", key), "error"))
		} else {
			fmt.Println(utils.ColorText(fmt.Sprintf("%s=%s", key, v), "info"))
		}
	} else {
		// Set
		err := cfg.Set(level, repoRoot, key, value)
		if err != nil {
			fmt.Println(utils.ColorText(fmt.Sprintf("Error setting %s: %v", key, err), "error"))
		} else {
			fmt.Println(utils.ColorText(fmt.Sprintf("Set %s=%s (%s)", key, value, level), "success"))
		}
	}
}
