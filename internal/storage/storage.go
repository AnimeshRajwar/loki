package storage

type ObjectStore interface {
	WriteObject(data []byte) string
}

type FileStorage struct {
	root string
}

func NewFileStorage(root string) *FileStorage {
	return &FileStorage{root: root}
}

func (fs *FileStorage) GiveRoot() string {
	return fs.root
}
