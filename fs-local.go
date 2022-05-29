package naivefs

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type FSLocal struct {
	baseDir string
}

func NewFSLocal(baseDir string) *FSLocal {
	return &FSLocal{
		baseDir: baseDir,
	}
}

func (local *FSLocal) Type() FSType {
	return FSTypeLocal
}

func (local *FSLocal) toAbs(name string) string {
	if path.IsAbs(name) {
		return name
	}
	return path.Join(local.baseDir, name)
}

func (local *FSLocal) File(name string) *File {
	name = local.toAbs(name)
	return &File{
		fs:   local,
		name: name,
	}
}

func (local *FSLocal) Touch(name string) error {
	name = local.toAbs(name)
	exists, err := local.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	local.MkDir(path.Dir(name))
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	return file.Close()
}

func (local *FSLocal) MkDir(name string) error {
	name = local.toAbs(name)
	exists, err := local.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return os.MkdirAll(name, os.ModePerm)
}

func (local *FSLocal) Remove(name string) error {
	name = local.toAbs(name)
	exists, err := local.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.RemoveAll(name)
}

func (local *FSLocal) Write(name string, data []byte) error {
	name = local.toAbs(name)
	return os.WriteFile(name, data, os.ModePerm)
}

func (local *FSLocal) Read(name string) ([]byte, error) {
	name = local.toAbs(name)
	return os.ReadFile(name)
}

func (local *FSLocal) Exists(name string) (bool, error) {
	name = local.toAbs(name)
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (local *FSLocal) IsDir(name string) (bool, error) {
	name = local.toAbs(name)
	stat, err := os.Stat(name)
	if err == nil {
		return stat.IsDir(), nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (local *FSLocal) ReadDir(name string) ([]FileInfo, error) {
	name = local.toAbs(name)
	osFileInfos, err := ioutil.ReadDir(name)
	if err != nil {
		return nil, err
	}
	var res []FileInfo
	for _, osFileInfo := range osFileInfos {
		fileInfo := FileInfo{
			Name:  osFileInfo.Name(),
			Size:  osFileInfo.Size(),
			IsDir: osFileInfo.IsDir(),
		}
		res = append(res, fileInfo)
	}
	return res, nil
}
