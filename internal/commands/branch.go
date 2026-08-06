package commands

import (
	"bytes"
	"fmt"
	"loki/internal/utils"
	"os"
	"path/filepath"
)

func Branch(args []string) {
	if len(args) == 0 {
		// Determine current branch
		headData, err := os.ReadFile(".loki/HEAD")
		var currentBranch string
		if err == nil {
			ref := string(bytes.TrimSpace(headData))
			if len(ref) >= 5 && ref[:4] == "ref:" {
				currentBranch = filepath.Base(ref[5:])
			}
		}

		// List branches
		entries, err := os.ReadDir(".loki/refs/heads")
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.Name() == currentBranch {
				fmt.Printf("* %s\n", utils.ColorText(entry.Name(), "success"))
			} else {
				fmt.Printf("  %s\n", entry.Name())
			}
		}
		return
	}

	if len(args) == 1 {
		// Create branch
		branchName := args[0]
		
		headData, err := os.ReadFile(".loki/HEAD")
		if err != nil {
			fmt.Println("Error reading HEAD")
			return
		}
		ref := string(bytes.TrimSpace(headData))
		
		var commitHash string
		if len(ref) >= 5 && ref[:4] == "ref:" {
			refPath := ".loki/" + ref[5:]
			refHashData, err := os.ReadFile(refPath)
			if err != nil {
				fmt.Println("Error reading ref")
				return
			}
			commitHash = string(bytes.TrimSpace(refHashData))
		} else {
			commitHash = ref
		}

		if commitHash == "" {
			fmt.Println("No commits yet")
			return
		}

		branchPath := filepath.Join(".loki/refs/heads", branchName)
		err = os.WriteFile(branchPath, []byte(commitHash+"\n"), 0644)
		if err != nil {
			fmt.Println("Error creating branch")
			return
		}
		return
	}

	if len(args) == 2 && args[0] == "-d" {
		branchName := args[1]
		branchPath := filepath.Join(".loki/refs/heads", branchName)
		err := os.Remove(branchPath)
		if err != nil {
			fmt.Printf("Error deleting branch %s\n", branchName)
		} else {
			fmt.Printf("Deleted branch %s\n", branchName)
		}
		return
	}
	
	fmt.Println("Invalid branch command")
}
