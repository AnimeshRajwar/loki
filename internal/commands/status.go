package commands

import (
	"fmt"
	"loki/internal/core"
	"loki/internal/utils"
)

func Status() {
	repo := core.OpenRepository()
	files := repo.Status()
	if len(files) == 0 {
		fmt.Println(utils.ColorText("No files staged to commit", "warning"))
		return
	}
	fmt.Println(utils.ColorText("Changes to be committed:", "info"))
	for _, fs := range files {
		fmt.Printf("        %s:   %s\n", fs.Status, fs.Name)
	}
}
