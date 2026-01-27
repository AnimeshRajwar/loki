package commands

import (
	"fmt"
	"loki/internal/core"
)

func Status() {
	repo := core.OpenRepository()
	files := repo.Status()
	if len(files) == 0 {
		fmt.Println("No files staged to commit")
		return
	}
	fmt.Println("Changes to be committed:")
	for _, fs := range files {
		fmt.Printf("        %s:   %s\n", fs.Status, fs.Name)
	}
}
