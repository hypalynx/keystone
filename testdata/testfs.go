package testdata

import (
	"io/fs"
	"os"
	"path/filepath"
)

type TestFS struct {
	rootDir string
}

func NewTestFS(rootDir string) *TestFS {
	return &TestFS{rootDir: rootDir}
}

func (tfs *TestFS) Open(name string) (fs.File, error) {
	fullPath := filepath.Join(tfs.rootDir, name)
	return os.Open(fullPath)
}

func (tfs *TestFS) ReadDir(name string) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(tfs.rootDir, name)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			return nil, err
		}
	}
	return os.ReadDir(fullPath)
}

func (tfs *TestFS) ReadFile(name string) ([]byte, error) {
	fullPath := filepath.Join(tfs.rootDir, name)
	return os.ReadFile(fullPath)
}
