package commands

import (
	"fmt"
	"loki/internal/core"
	"loki/internal/utils"
	"os"
	"sort"
	"strings"
)

// Diff compares working directory against Index, or Index against HEAD tree.
func Diff(args []string) {
	staged := false
	for _, arg := range args {
		if arg == "--staged" {
			staged = true
		} else {
			fmt.Printf("Unknown argument: %s\n", arg)
			return
		}
	}

	repo := core.OpenRepository()

	if staged {
		// Compare Index vs HEAD tree
		indexEntries := repo.GetIndexEntries()
		headEntries, err := repo.GetHeadTreeEntries()
		if err != nil {
			fmt.Println(utils.ColorText("error: failed to read HEAD tree: "+err.Error(), "error"))
			return
		}

		// Collect all unique file paths that are staged
		pathsMap := make(map[string]bool)
		for p := range indexEntries {
			pathsMap[p] = true
		}

		var paths []string
		for p := range pathsMap {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		for _, p := range paths {
			indexHash, inIndex := indexEntries[p]
			headHash, inHead := headEntries[p]

			var headContent, indexContent string
			if inHead {
				hc, err := repo.ReadBlob(headHash)
				if err != nil {
					fmt.Printf("error: failed to read HEAD blob for %s: %s\n", p, err)
					continue
				}
				headContent = hc
			}
			if inIndex {
				ic, err := repo.ReadBlob(indexHash)
				if err != nil {
					fmt.Printf("error: failed to read Index blob for %s: %s\n", p, err)
					continue
				}
				indexContent = ic
			}

			if headContent != indexContent {
				linesHead := splitLines(headContent)
				linesIndex := splitLines(indexContent)
				diffOut := utils.FormatDiff(linesHead, linesIndex, p)
				fmt.Print(diffOut)
			}
		}
	} else {
		// Compare Working Directory vs Index
		indexEntries := repo.GetIndexEntries()
		var paths []string
		for p := range indexEntries {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		for _, p := range paths {
			indexHash := indexEntries[p]
			indexContent, err := repo.ReadBlob(indexHash)
			if err != nil {
				fmt.Printf("error: failed to read Index blob for %s: %s\n", p, err)
				continue
			}

			// Read file from working directory
			var workdirContent string
			workdirData, err := os.ReadFile(p)
			if err != nil {
				if os.IsNotExist(err) {
					// File deleted in working directory
					workdirContent = ""
				} else {
					fmt.Printf("error: failed to read working directory file %s: %s\n", p, err)
					continue
				}
			} else {
				workdirContent = string(workdirData)
			}

			if indexContent != workdirContent {
				linesIndex := splitLines(indexContent)
				linesWorkdir := splitLines(workdirContent)
				diffOut := utils.FormatDiff(linesIndex, linesWorkdir, p)
				fmt.Print(diffOut)
			}
		}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\r\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}
