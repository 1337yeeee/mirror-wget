package storage

import (
	"os"
	"path/filepath"
)

func Save(path string, data []byte) (int, error) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return 0, err
	}

	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	n, err := file.Write(data)
	if err != nil {
		return n, err
	}

	return n, nil
}
