package commands

import (
	"fmt"

	"loki/internal/core"
	"loki/internal/utils"
)

func Add(files []string) {

	if len(files) == 0 {
		fmt.Println(
			utils.ColorText("error: no files specified", "error"),
		)
		return
	}

	repo := core.OpenRepository()
	stagedAny := false

	for _, f := range files {

		info, err := repo.Stat(f)
		if err != nil {
			fmt.Println(
				utils.ColorText("error: file not found -> "+f, "error"),
			)
			continue
		}

		if info.IsDir() {
			fmt.Println(
				utils.ColorText("error: cannot add directory -> "+f, "error"),
			)
			continue
		}

		if repo.IsIgnored(f) {
			fmt.Println(utils.ColorText("warning: ignoring file -> "+f, "warning"))
			continue
		}

		if err := repo.AddFile(f); err != nil {
			fmt.Println(
				utils.ColorText("error: failed to stage -> "+f, "error"),
			)
			continue
		}

		fmt.Println(
			utils.ColorText("staged: "+f, "success"),
		)
		stagedAny = true
	}

	if stagedAny {
		fmt.Println(
			utils.ColorText("index updated", "info"),
		)
	} else {
		fmt.Println(
			utils.ColorText("nothing to stage", "warning"),
		)
	}
	fmt.Println(utils.ColorText("Files added to staging area", "success"))
}
