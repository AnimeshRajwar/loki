package commands

import (
	"fmt"
	"loki/internal/core"
)

func Checkout(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: loki checkout [-b] <branch-name-or-commit-hash>")
		return
	}

	target := args[0]
	isNewBranch := false

	if len(args) == 2 && args[0] == "-b" {
		isNewBranch = true
		target = args[1]
	} else if len(args) != 1 {
		fmt.Println("Usage: loki checkout [-b] <branch-name-or-commit-hash>")
		return
	}

	repo := core.OpenRepository()
	
	if isNewBranch {
		Branch([]string{target})
	}

	err := repo.Checkout(target)
	if err != nil {
		fmt.Printf("Error during checkout: %v\n", err)
	} else {
		if isNewBranch {
			fmt.Printf("Switched to a new branch '%s'\n", target)
		} else {
			fmt.Printf("Successfully checked out %s\n", target)
		}
	}
}
