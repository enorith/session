package handlers

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/enorith/supports/file"
)

var (
	FileMode = os.FileMode(0666)
)

type FileSessionHandler struct {
	dir string
}

func (f *FileSessionHandler) Init(id string) error {
	if ok, _ := file.PathExists(f.dir); !ok {
		return os.MkdirAll(f.dir, FileMode)
	}

	if ok, _ := file.PathExists(f.resolvePath(id)); !ok {
		f, e := os.Create(f.resolvePath(id))
		if e == nil {
			defer f.Close()
		}
		return e
	}

	return nil
}

func (f *FileSessionHandler) Read(id string) ([]byte, error) {
	row, e := os.ReadFile(f.resolvePath(id))
	if e == io.EOF {
		return nil, nil
	}

	return row, e
}

func (f *FileSessionHandler) Write(id string, data []byte) error {
	return os.WriteFile(f.resolvePath(id), data, FileMode)
}

func (f *FileSessionHandler) Destroy(id string) error {
	return os.Remove(f.resolvePath(id))
}

func (f *FileSessionHandler) GC(maxLifeTime time.Duration) error {
	return filepath.WalkDir(f.dir, func(path string, d fs.DirEntry, err error) error {
		fileInfo, e := d.Info()
		if e != nil {
			return e
		}

		if fileInfo.IsDir() {
			return nil
		}

		if fileInfo.ModTime().Before(time.Now().Add(-maxLifeTime)) {
			return os.Remove(path)
		}

		return err
	})
}

func (f *FileSessionHandler) resolvePath(id string) string {
	return filepath.Join(f.dir, id)
}

func NewFileSessionHandler(dir string) *FileSessionHandler {
	return &FileSessionHandler{dir: dir}
}
