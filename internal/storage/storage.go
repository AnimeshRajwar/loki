package storage

import (
	"crypto/sha1"
	"fmt"
	"os"
)

type ObjectStore interface {
	WriteObject(data []byte) string
}

type FileStorage struct {
	root string
}

func NewFileStorage(root string) *FileStorage {
	return &FileStorage{root: root}
}

func (fs *FileStorage) WriteCommit(tree, msg string) string {
	data := tree + "\n" + msg
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(data)))

	_ = os.WriteFile(fs.root+"/objects/"+hash, []byte(data), 0644)
	f, _ := os.OpenFile(fs.root+"/commits.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(hash + " " + msg + "\n")

	return hash
}
func(fs *FileStorage) GiveRoot() string {
	return fs.root
}