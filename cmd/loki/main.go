package main

import (
	"fmt"
	"os"
	"path/filepath"

	"loki/internal/commands"
	"loki/internal/core"
)

func main() {
	if len(os.Args) < 2 {
		commands.Help()
		return
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(cwd)
	if os.Args[0] == "./loki" && os.Args[1] != "help" && os.Args[1] != "init" {
		path, check := core.IsRepoInitialized(absPath)
		if !check {
			fmt.Println("fatal:" + path + " not a loki repository (or any of the parent directories)")
			return
		}
	}

	switch os.Args[1] {
	case "init":
		commands.Init()
	case "add":
		commands.Add(os.Args[2:])
	case "commit":
		commands.Commit(os.Args[2:])
	case "status":
		commands.Status()
	case "log":
		commands.Log()
	case "help":
		commands.Help()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		commands.Help()
	}
}
