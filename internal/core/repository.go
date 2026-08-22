package core

import (
	"bytes"
	"crypto/sha1"
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
	var commitHash string
	if len(ref) >= 5 && ref[:4] == "ref:" {
		refPath := ".loki/" + ref[5:]
		refHashData, err := os.ReadFile(refPath)
		if err != nil {
			return nil
		}
		commitHash = string(bytes.TrimSpace(refHashData))
	} else {
		commitHash = ref
	}
	// Read commit object
	objData, err := r.store.ReadObject(commitHash)
	if err != nil {
		return nil
	}
	// Parse commit to get tree hash
	idx := bytes.IndexByte(objData, 0)
	if idx >= 0 {
		objData = objData[idx+1:]
	}
	var treeHash string
	for _, line := range bytes.Split(content, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("tree ")) {
			treeHash = string(line[5:])
			break
		}
	}
	if treeHash == "" {
		return nil
	}
	// Read tree object
	treeData, err := r.store.ReadObject(treeHash)
	if err != nil {
		return nil
	}
	// Parse tree entries
	entries := []models.TreeEntry{}
	// Skip header ("tree <len>\0")
	idx = bytes.IndexByte(treeData, 0)
	if idx < 0 {
		return nil
	}
	// Parse tree entries
	entries := []models.TreeEntry{}
	for len(treeContent) > 0 {
		// Format: mode name\0hash(20 bytes)
		sp := bytes.IndexByte(treeContent, ' ')
		if sp < 0 {
			break
		}
		mode := string(treeContent[:sp])
		treeContent = treeContent[sp+1:]
		nul := bytes.IndexByte(treeContent, 0)
		if nul < 0 {
			break
		}
		name := string(treeContent[:nul])
		if len(treeContent) < nul+21 {
			break
		}
		hash := treeContent[nul+1 : nul+21]
		entries = append(entries, models.TreeEntry{Mode: mode, Name: name, Hash: hash})
		treeContent = treeContent[nul+21:]
	}
	return &models.Tree{Entries: entries}
}

func parseObject(data []byte) (string, []byte, error) {
	idx := bytes.IndexByte(data, 0)
	if idx < 0 {
		return "", nil, fmt.Errorf("invalid object: missing header separator")
	}
	header := string(data[:idx])
	content := data[idx+1:]
	parts := strings.Split(header, " ")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid object header format")
	}
	return parts[0], content, nil
}

// ReadBlob retrieves raw blob content by its hash.
func (r *Repository) ReadBlob(hash string) (string, error) {
	data, err := r.store.ReadObject(hash)
	if err != nil {
		return "", err
	}
	objType, content, err := parseObject(data)
	if err != nil {
		return "", err
	}
	if objType != "blob" {
		return "", fmt.Errorf("object is not a blob: %s", objType)
	}
	return string(content), nil
}

// GetHeadTreeEntries retrieves the file path to hex hash mappings for the HEAD commit.
func (r *Repository) GetHeadTreeEntries() (map[string]string, error) {
	tree := r.getLastCommitTree()
	entries := make(map[string]string)
	if tree == nil {
		return entries, nil
	}
	for _, entry := range tree.Entries {
		entries[entry.Name] = hex.EncodeToString(entry.Hash)
	}
	return entries, nil
}

// GetIndexEntries retrieves the current index entries.
func (r *Repository) GetIndexEntries() map[string]string {
	entries := make(map[string]string)
	for k, v := range r.index.Entries {
		entries[k] = v
	}
	return entries
}

func (r *Repository) Commit(message, author, email string) string {
	// Write the tree and get its hash
	treeHash := r.index.WriteTree(r.store)

	// Create a proper Commit object using the model
	commitModel := &models.Commit{
		Tree:    treeHash,
		Message: message,
		Author:  author,
		Email:   email,
	}

	// Serialize and write using the standard WriteObject (Git-style)
	commitHash := r.store.WriteObject(commitModel.Serialize())

	// Update the log
	f, _ := os.OpenFile(filepath.Join(r.store.GiveRoot(), "commits.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(commitHash + " " + message + " " + author + " <" + email + ">\n")

	// update branch ref so getLastCommitTree works
	headData, err := os.ReadFile(".loki/HEAD")
	if err == nil {
		ref := string(bytes.TrimSpace(headData))
		if len(ref) >= 5 && ref[:4] == "ref:" {
			refPath := ".loki/" + ref[5:]
			os.MkdirAll(filepath.Dir(refPath), 0755)
			os.WriteFile(refPath, []byte(commitHash+"\n"), 0644)
		} else {
			os.WriteFile(".loki/HEAD", []byte(commitHash+"\n"), 0644)
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
		parts := strings.Split(line, " ")
		if len(parts) >= 4 {
			hash := parts[0]
			email := parts[len(parts)-1]
			author := parts[len(parts)-2]
			msg := strings.Join(parts[1:len(parts)-2], " ")
			fmt.Printf(utils.ColorText("Commit: %s\nMessage: %s\nAuthor: %s\n%s\n\n", "info"), hash, msg, author, email)
		} else {
			p := strings.SplitN(line, " ", 2)
			if len(p) >= 2 {
				fmt.Printf(utils.ColorText("Commit: %s\n%s\n\n", "info"), p[0], p[1])
			}
		}
	}
}

func (r *Repository) Checkout(target string) error {
	commitHash, headContent, err := r.resolveCheckoutTarget(target)
	if err != nil {
		return err
	}

	treeHash, err := r.commitTreeHash(commitHash)
	if err != nil {
		return err
	}

	if err := r.ensureCleanForCheckout(); err != nil {
		return err
	}

	currentTracked := map[string]string{}
	if headCommit, err := r.resolveHeadCommitHash(); err == nil && headCommit != "" {
		if headTreeHash, err := r.commitTreeHash(headCommit); err == nil && headTreeHash != "" {
			if err := r.collectTreeFiles(headTreeHash, "", currentTracked); err != nil {
				return err
			}
		}
	}

	for path := range currentTracked {
		if strings.HasPrefix(path, ".loki/") || path == ".loki" {
			continue
		}
		_ = os.Remove(filepath.FromSlash(path))
	}

	targetFiles := map[string]string{}
	if err := r.collectTreeFiles(treeHash, "", targetFiles); err != nil {
		return err
	}

	r.index.Entries = make(map[string]string)
	for path, blobHash := range targetFiles {
		blobData, err := r.store.ReadObject(blobHash)
		if err != nil {
			return fmt.Errorf("failed to read blob %s: %v", blobHash, err)
		}
		content := objectBody(blobData)

		osPath := filepath.FromSlash(path)
		if err := os.MkdirAll(filepath.Dir(osPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(osPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", path, err)
		}
		r.index.Add(path, blobHash)
	}
	r.index.Save()

	if err := os.WriteFile(".loki/HEAD", []byte(headContent+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to update HEAD: %v", err)
	}

	return nil
}

func objectBody(data []byte) []byte {
	idx := bytes.IndexByte(data, 0)
	if idx >= 0 {
		return data[idx+1:]
	}
	return data
}

func normalizeRepoPath(path string) string {
	cleaned := filepath.Clean(path)
	if cleaned == "." {
		return ""
	}
	return filepath.ToSlash(cleaned)
}

func blobHashForContent(data []byte) string {
	blob := (&models.Blob{Content: data}).Serialize()
	sum := sha1.Sum(blob)
	return hex.EncodeToString(sum[:])
}

func sameFileMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func (r *Repository) resolveHeadCommitHash() (string, error) {
	headData, err := os.ReadFile(".loki/HEAD")
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(string(headData))
	if strings.HasPrefix(ref, "ref: ") {
		refPath := filepath.Join(".loki", strings.TrimPrefix(ref, "ref: "))
		refHashData, err := os.ReadFile(refPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
		return strings.TrimSpace(string(refHashData)), nil
	}
	return ref, nil
}

func (r *Repository) commitTreeHash(commitHash string) (string, error) {
	objData, err := r.store.ReadObject(commitHash)
	if err != nil {
		return "", fmt.Errorf("target commit %s not found", commitHash)
	}
	body := objectBody(objData)
	for _, line := range bytes.Split(body, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("tree ")) {
			return strings.TrimSpace(string(line[5:])), nil
		}
	}
	return "", fmt.Errorf("commit %s has no tree", commitHash)
}

func (r *Repository) resolveCheckoutTarget(target string) (string, string, error) {
	refPath := filepath.Join(".loki", "refs", "heads", target)
	if _, err := os.Stat(refPath); err == nil {
		hashData, err := os.ReadFile(refPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to read branch file: %v", err)
		}
		commitHash := strings.TrimSpace(string(hashData))
		if commitHash == "" {
			return "", "", fmt.Errorf("branch %s has no commit", target)
		}
		return commitHash, "ref: refs/heads/" + target, nil
	}

	return target, target, nil
}

func (r *Repository) collectTreeFiles(treeHash, prefix string, out map[string]string) error {
	treeData, err := r.store.ReadObject(treeHash)
	if err != nil {
		return fmt.Errorf("failed to read tree object %s: %v", treeHash, err)
	}
	body := objectBody(treeData)
	for len(body) > 0 {
		sp := bytes.IndexByte(body, ' ')
		if sp < 0 {
			break
		}
		mode := string(body[:sp])
		body = body[sp+1:]

		nul := bytes.IndexByte(body, 0)
		if nul < 0 {
			break
		}
		name := string(body[:nul])
		if len(body) < nul+21 {
			break
		}
		hashBytes := body[nul+1 : nul+21]
		hashHex := hex.EncodeToString(hashBytes)
		body = body[nul+21:]

		joined := name
		if prefix != "" {
			joined = prefix + "/" + name
		}
		joined = normalizeRepoPath(joined)

		if mode == "40000" {
			if err := r.collectTreeFiles(hashHex, joined, out); err != nil {
				return err
			}
			continue
		}

		if joined != "" {
			out[joined] = hashHex
		}
	}
	return nil
}

func (r *Repository) ensureCleanForCheckout() error {
	headFiles := map[string]string{}
	headCommit, err := r.resolveHeadCommitHash()
	if err != nil {
		return fmt.Errorf("failed to resolve HEAD: %v", err)
	}
	if headCommit != "" {
		headTreeHash, err := r.commitTreeHash(headCommit)
		if err != nil {
			return err
		}
		if err := r.collectTreeFiles(headTreeHash, "", headFiles); err != nil {
			return err
		}
	}

	for path, expectedHash := range headFiles {
		content, err := os.ReadFile(filepath.FromSlash(path))
		if err != nil {
			return fmt.Errorf("working directory is not clean, missing tracked file: %s", path)
		}
		if blobHashForContent(content) != expectedHash {
			return fmt.Errorf("working directory is not clean, unstaged changes in: %s", path)
		}
	}

	return nil
}
