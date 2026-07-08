package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"loki/internal/models"
	"loki/internal/storage"
	"loki/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	store *storage.FileStorage
	index *Index
}

func (r *Repository) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func OpenRepository() *Repository {
	cwd, err := os.Getwd()
	if err != nil {
		panic(utils.ColorText("Could not get current working directory", "error"))
	}
	repoRoot, ok := IsRepoInitialized(cwd + string(os.PathSeparator))
	if !ok {
		fmt.Fprintln(os.Stderr, utils.ColorText("fatal: not a loki repository (or any of the parent directories)", "error"))
		os.Exit(1)
	}
	return &Repository{
		store: storage.NewFileStorage(filepath.Join(repoRoot, ".loki")),
		index: LoadIndex(),
	}
}

// Check for loki repo
func IsRepoInitialized(path string) (string, bool) {
	cur_path := path
	for {
		loki_check := filepath.Join(cur_path, ".loki")

		if info, err := os.Stat(loki_check); err == nil && info.IsDir() {
			return cur_path, true
		}

		parent := filepath.Dir(cur_path)

		if parent == cur_path {
			break
		}

		cur_path = parent
	}

	return "", false
}

// Detects and sets status: "new file", "modified", or "deleted"
func (r *Repository) AddFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	blob := &models.Blob{Content: data}
	hash := r.store.WriteObject(blob.Serialize())

	r.index.Add(path, hash)
	r.index.Save()

	return nil
}

// Helper: get last commit's tree (if any)
func (r *Repository) getLastCommitTree() *models.Tree {
	// Try to read HEAD ref
	headData, err := os.ReadFile(".loki/HEAD")
	if err != nil {
		return nil
	}
	ref := string(bytes.TrimSpace(headData))
	if len(ref) < 5 || ref[:4] != "ref:" {
		return nil
	}
	refPath := ".loki/" + ref[5:]
	refHashData, err := os.ReadFile(refPath)
	if err != nil {
		return nil
	}
	commitHash := string(bytes.TrimSpace(refHashData))
	// Read commit object
	objPath := ".loki/objects/" + commitHash[:2] + "/" + commitHash[2:]
	objData, err := os.ReadFile(objPath)
	if err != nil {
		return nil
	}
	// Parse commit to get tree hash
	var treeHash string
	for _, line := range bytes.Split(objData, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("tree ")) {
			treeHash = string(line[5:])
			break
		}
	}
	if treeHash == "" {
		return nil
	}
	// Read tree object
	treeObjPath := ".loki/objects/" + treeHash[:2] + "/" + treeHash[2:]
	treeData, err := os.ReadFile(treeObjPath)
	if err != nil {
		return nil
	}
	// Parse tree entries
	entries := []models.TreeEntry{}
	// Skip header ("tree <len>\0")
	idx := bytes.IndexByte(treeData, 0)
	if idx < 0 {
		return nil
	}
	treeData = treeData[idx+1:]
	for len(treeData) > 0 {
		// Format: mode name\0hash(20 bytes)
		sp := bytes.IndexByte(treeData, ' ')
		if sp < 0 {
			break
		}
		mode := string(treeData[:sp])
		treeData = treeData[sp+1:]
		nul := bytes.IndexByte(treeData, 0)
		if nul < 0 {
			break
		}
		name := string(treeData[:nul])
		hash := treeData[nul+1 : nul+21]
		entries = append(entries, models.TreeEntry{Mode: mode, Name: name, Hash: hash})
		treeData = treeData[nul+21:]
	}
	return &models.Tree{Entries: entries}
}

func (r *Repository) Commit(message string) string {
	// 1. Write the tree and get its hash
	treeHash := r.index.WriteTree(r.store)

	// 2. Create a proper Commit object using the model
	commitModel := &models.Commit{
		Tree:    treeHash,
		Message: message,
	}

	// 3. Serialize and write using the standard WriteObject (Git-style)
	commitHash := r.store.WriteObject(commitModel.Serialize())

	// 4. Update the log
	f, _ := os.OpenFile(filepath.Join(r.store.GiveRoot(), "commits.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(commitHash + " " + message + "\n")

	// 5. Fix: update branch ref so getLastCommitTree works
	headData, err := os.ReadFile(".loki/HEAD")
	if err == nil {
		ref := string(bytes.TrimSpace(headData))
		if len(ref) >= 5 && ref[:4] == "ref:" {
			refPath := ".loki/" + ref[5:]
			os.MkdirAll(filepath.Dir(refPath), 0755)
			os.WriteFile(refPath, []byte(commitHash+"\n"), 0644)
		}
	}

	r.index.Clear()
	return commitHash
}

func (r *Repository) Status() []FileStatus {
	lastTree := r.getLastCommitTree()

	var results []FileStatus

	for path, indexHash := range r.index.Entries {
		status := "added"

		if lastTree != nil {
			for _, entry := range lastTree.Entries {
				if entry.Name == path {
					if hex.EncodeToString(entry.Hash) == indexHash {
						status = "staged (unchanged)"
					} else {
						status = "modified"
					}

					break
				}
			}
		}
		results = append(results, FileStatus{Name: path, Status: status})
	}
	return results
}

func (r *Repository) PrintLog() {
	logs, err := os.ReadFile(filepath.Join(r.store.GiveRoot(), "commits.log"))
	if err != nil {
		fmt.Println(utils.ColorText("No commit found.", "error"))
		return
	}
	lines := strings.Split(string(logs), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		fmt.Printf(utils.ColorText("%s %s\n", "info"), parts[0], parts[1])
	}
}
