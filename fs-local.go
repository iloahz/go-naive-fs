package naivefs

import (
	"errors"
	"os"
)

type FSLocal struct {
}

func NewFSLocal() *FSLocal {
	return &FSLocal{}
}

func (local *FSLocal) File(path string) *File {
	return &File{
		fs:   local,
		Path: path,
	}
}

func (local *FSLocal) Touch(path string) error {
	if local.Exists(path) {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

func (local *FSLocal) MkDir(path string) error {
	if local.Exists(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func (local *FSLocal) Remove(path string) error {
	if !local.Exists(path) {
		return nil
	}
	return os.RemoveAll(path)
}

func (local *FSLocal) Write(path string, data []byte) error {
	return os.WriteFile(path, data, os.ModePerm)
}

func (local *FSLocal) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (local *FSLocal) Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func (local *FSLocal) IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
