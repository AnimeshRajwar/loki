package commands

import (
	"fmt"
	"loki/internal/core"
	"loki/internal/utils"
)

func Status() {
	repo := core.OpenRepository()
	stagedFiles := repo.Status()
	untrackedFiles := repo.GetUntrackedFiles()

	if len(stagedFiles) == 0 && len(untrackedFiles) == 0 {
		fmt.Println(utils.ColorText("No files staged to commit", "warning"))
		return
	}

	if len(stagedFiles) > 0 {
		fmt.Println(utils.ColorText("Changes to be committed:", "info"))
		for _, fs := range stagedFiles {
			fmt.Printf("        %s:   %s\n", fs.Status, fs.Name)
		}
	} else {
		fmt.Println(utils.ColorText("No files staged to commit", "warning"))
	}

	if len(untrackedFiles) > 0 {
		fmt.Println(utils.ColorText("Untracked files:", "info"))
		for _, u := range untrackedFiles {
			fmt.Printf("        untracked:   %s\n", u)
		}
	}
}
