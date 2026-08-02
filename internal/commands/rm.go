package commands

import (
	"fmt"

	"loki/internal/core"
	"loki/internal/utils"
)

func Rm(files []string) {

	if len(files) == 0 {
		fmt.Println(utils.ColorText("error: no files specified", "error"))
		return
	}

	repo := core.OpenRepository()

	for _, file := range files {

		err := repo.RemoveFile(file)

		if err != nil {
			fmt.Println(utils.ColorText("error: failed to remove -> "+file, "error"))
			continue
		}

		fmt.Println(utils.ColorText("removed: "+file, "success"))
	}

	fmt.Println(utils.ColorText("index updated", "info"))
}