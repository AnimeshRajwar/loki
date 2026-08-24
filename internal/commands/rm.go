package commands

import (
	"fmt"
	"os"

	"loki/internal/core"
	"loki/internal/utils"
)

func Rm(files []string) {
	cached := false
	var targetFiles []string
	for _, f := range files {
		if f == "--cached" {
			cached = true
		} else {
			targetFiles = append(targetFiles, f)
		}
	}

	if len(targetFiles) == 0 {
		fmt.Println(utils.ColorText("error: no files specified", "error"))
		return
	}

	repo := core.OpenRepository()
	removedAny := false

	for _, f := range targetFiles {
		indexEntries := repo.GetIndexEntries()
		if _, ok := indexEntries[f]; !ok {
			fmt.Println(utils.ColorText("error: file not tracked -> "+f, "error"))
			continue
		}

		if !cached {
			err := os.Remove(f)
			if err != nil && !os.IsNotExist(err) {
				fmt.Println(utils.ColorText("error: failed to remove file -> "+f, "error"))
				continue
			}
		}

		repo.RemoveFile(f)
		fmt.Println(utils.ColorText("removed: "+f, "success"))
		removedAny = true
	}

	if removedAny {
		fmt.Println(utils.ColorText("Files removed", "success"))
	}
}
