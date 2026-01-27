package core

import (
	"os"

	"encoding/json"
	"loki/internal/models"
)

type FileStatus struct {
	Name   string
	Status string // "modified", "added", "deleted"
}

type Index struct {
	FilesList []FileStatus
}

func LoadIndex() *Index {
	data, err := os.ReadFile(".loki/index")
	if err != nil {
		return &Index{}
	}
	var idx Index
	_ = json.Unmarshal(data, &idx)
	return &idx
}

func (i *Index) Add(file string, status string) {
	i.FilesList = append(i.FilesList, FileStatus{Name: file, Status: status})
}

func (i *Index) Save() {
	data, _ := json.Marshal(i)
	_ = os.WriteFile(".loki/index", data, 0644)
}

func (i *Index) Files() []FileStatus {
	return i.FilesList
}

func (i *Index) Clear() {
	i.FilesList = []FileStatus{}
	i.Save()
}

type ObjectStore interface {
	WriteObject(data []byte) string
}

func (i *Index) WriteTree(store ObjectStore) string {
	var entries []models.TreeEntry
	for _, fs := range i.FilesList {
		data, _ := os.ReadFile(fs.Name)
		blob := &models.Blob{Content: data}
		blobHash := store.WriteObject(blob.Serialize())
		entries = append(entries, models.TreeEntry{
			Mode: "100644",
			Name: fs.Name,
			Hash: decodeHash(blobHash),
		})
	}
	tree := &models.Tree{Entries: entries}
	return store.WriteObject(tree.Serialize())
}
