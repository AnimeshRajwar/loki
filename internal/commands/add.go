package commands

import (
	"fmt"

	"loki/internal/core"
	"loki/internal/utils"
)

func Add(files []string) {
	repo := core.OpenRepository()
	for _, f := range files {
		repo.AddFile(f)
	}
	fmt.Println(utils.ColorText("Files added to staging area", "success"))
}
