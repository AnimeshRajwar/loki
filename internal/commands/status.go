package commands

import (
	"fmt"
	"strings"

	"loki/internal/core"
	"loki/internal/utils"
)

func Status() {
	repo := core.OpenRepository()
	files := repo.Status()

	var staged, unstaged, untracked []core.FileStatus
	for _, fs := range files {
		if strings.HasPrefix(fs.Status, "staged") {
			staged = append(staged, fs)
		} else if fs.Status == "modified" || fs.Status == "deleted" {
			unstaged = append(unstaged, fs)
		} else if fs.Status == "untracked" {
			untracked = append(untracked, fs)
		}
	}

	if len(staged) > 0 {
		fmt.Println(utils.ColorText("Changes to be committed:", "info"))
		for _, fs := range staged {
			fmt.Printf("        %s:   %s\n", fs.Status, fs.Name)
		}
		fmt.Println()
	} else {
		fmt.Println(utils.ColorText("No files staged to commit", "warning"))
		fmt.Println()
	}

	if len(unstaged) > 0 {
		fmt.Println(utils.ColorText("Changes not staged for commit:", "error"))
		for _, fs := range unstaged {
			fmt.Printf("        %s:   %s\n", fs.Status, fs.Name)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println(utils.ColorText("Untracked files:", "error"))
		for _, fs := range untracked {
			fmt.Printf("        untracked:   %s\n", fs.Name)
		}
		fmt.Println()
	}
}
