package commands

import (
	"fmt"
	"strings"

	"loki/internal/core"
	"loki/internal/utils"
)

func Reset(args []string) {
	mode := "mixed" // default
	target := ""

	for _, arg := range args {
		if arg == "--soft" {
			mode = "soft"
		} else if arg == "--mixed" {
			mode = "mixed"
		} else if arg == "--hard" {
			mode = "hard"
		} else if !strings.HasPrefix(arg, "-") {
			target = arg
		}
	}

	repo := core.OpenRepository()
	if err := repo.Reset(target, mode); err != nil {
		fmt.Println(utils.ColorText("error: "+err.Error(), "error"))
		return
	}

	fmt.Printf(utils.ColorText("Reset complete (mode: %s)\n", "success"), mode)
}
