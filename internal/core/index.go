package core

import (
	"os"

	"encoding/json"
	"loki/internal/models"
)

type FileStatus struct {
	Name   string
	Status string `json:"entries"`
}

type Index struct {
	Entries map[string]string
}

func LoadIndex() *Index {
	data, err := os.ReadFile(".loki/index")
	if err != nil {
		return &Index{Entries: make(map[string]string)}
	}
	var idx Index
	_ = json.Unmarshal(data, &idx)

	if idx.Entries == nil {
		idx.Entries = make(map[string]string)
	}
	return &idx
}

func (i *Index) Add(path string, hash string) {
	i.Entries[path] = hash
}

func (i *Index) Save() {
	data, _ := json.Marshal(i)
	_ = os.WriteFile(".loki/index", data, 0644)
}

func (i *Index) Files() []FileStatus {
	var files []FileStatus
	for name, hash := range i.Entries {
		files = append(files, FileStatus{Name: name, Status: hash})
	}

	return files
}

func (i *Index) Clear() {
	i.Entries = make(map[string]string)
	i.Save()
}

type ObjectStore interface {
	WriteObject(data []byte) string
}

func (i *Index) WriteTree(store ObjectStore) string {
	var entries []models.TreeEntry
	for name, hash := range i.Entries {
		entries = append(entries, models.TreeEntry{
			Mode: "100644",
			Name: name,
			Hash: decodeHash(hash),
		})
	}
	tree := &models.Tree{Entries: entries}
	return store.WriteObject(tree.Serialize())
}
