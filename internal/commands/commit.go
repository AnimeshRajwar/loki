package commands

import (
	"fmt"
	"os"

	"loki/internal/config"
	"loki/internal/core"
)

func Commit(args []string) {
	msg := "default commit"
	if len(args) >= 2 && args[0] == "-m" {
		msg = args[1]
	}

	cwd, _ := os.Getwd()
	repoRoot := config.FindRepoRoot(cwd)
	cfg := config.NewConfig()
	cfg.Load(repoRoot)

	author := cfg.Get("user.name")
	email := cfg.Get("user.email")

	if author == "" {
		author = "loki"
	}
	if email == "" {
		email = "loki@local"
	}

	repo := core.OpenRepository()
	hash := repo.Commit(msg, author, email)
	fmt.Println("Committed:", hash)
}
