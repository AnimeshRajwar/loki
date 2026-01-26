package commands

import (
	"fmt"
	"loki/internal/core"
)

func Status() {
	repo := core.OpenRepository()
	files := repo.Status()
	if( len(files) == 0 ) {
		fmt.Println("No files staged to commit")
		return;
	}
	fmt.Println("Staged files:")
	for _, f := range files {
		fmt.Println(" ", f)
		}
}
