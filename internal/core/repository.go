package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"loki/internal/models"
	"loki/internal/storage"
	"os"
	"path/filepath"
)

type Repository struct {
	store *storage.FileStorage
	index *Index
}

func OpenRepository() *Repository {
	cwd, err := os.Getwd()
	if err != nil {
		panic("Could not get current working directory")
	}
	repoRoot, ok := IsRepoInitialized(cwd + string(os.PathSeparator))
	if !ok {
		fmt.Fprintln(os.Stderr, "fatal: not a loki repository (or any of the parent directories)")
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
func (r *Repository) AddFile(path string) {
	lastTree := r.getLastCommitTree()
	var status string
	fileInLast := false
	var lastHash []byte
	if lastTree != nil {
		for _, entry := range lastTree.Entries {
			if entry.Name == path {
				fileInLast = true
				lastHash = entry.Hash
				break
			}
		}
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if fileInLast {
			status = "deleted"
		} else {
			// File neither in last commit nor in working dir, skip
			return
		}
	} else {
		if !fileInLast {
			status = "added"
		} else {
			// Compare content
			data, _ := ioutil.ReadFile(path)
			blob := &models.Blob{Content: data}
			hash := r.store.WriteObject(blob.Serialize())
			if !bytes.Equal(decodeHash(hash), lastHash) {
				status = "modified"
			} else {
				// No change, skip
				return
			}
		}
	}
	r.index.Add(path, status)
	r.index.Save()
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
	treeHash := r.index.WriteTree(r.store)
	commit := &models.Commit{
		Tree:    treeHash,
		Message: message,
	}
	commitHash := r.store.WriteObject(commit.Serialize())
	// Optionally update HEAD and log (not implemented here)
	r.index.Clear()
	return commitHash
}

func (r *Repository) Status() []FileStatus {
	return r.index.Files()
}

func (r *Repository) PrintLog() {
	logs, _ := os.ReadFile(".loki/commits.log")
	fmt.Println(string(logs))
}
