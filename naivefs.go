package naivefs

func Touch(file File) error {
	return file.fs.Touch(file.Path)
}

func MkDir(file File) error {
	return file.fs.MkDir(file.Path)
}

func Remove(file File) error {
	return file.fs.Remove(file.Path)
}

func Write(file File, buf []byte) error {
	return file.fs.Write(file.Path, buf)
}

// TODO: perf opt
func Copy(src File, dst File) error {
	buf, err := Read(src)
	if err != nil {
		return err
	}
	return Write(dst, buf)
}

// TODO: perf opt
func Move(src File, dst File) error {
	if err := Copy(src, dst); err != nil {
		return err
	}
	return Remove(src)
}

func Read(file File) ([]byte, error) {
	return file.fs.Read(file.Path)
}

func Exists(file File) bool {
	return file.fs.Exists(file.Path)
}

func IsDir(file File) bool {
	return file.fs.IsDir(file.Path)
}

func IsFile(file File) bool {
	return file.fs.IsFile(file.Path)
}
