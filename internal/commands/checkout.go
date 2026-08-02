package commands

import (
	"fmt"

	"loki/internal/core"
)

func Checkout(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: loki checkout <branch|commit>")
		return
	}

	repo := core.OpenRepository()

	err := repo.Checkout(args[0])
	if err != nil {
		fmt.Println(err)
	}
}