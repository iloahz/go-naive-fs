package naivefs

type FS interface {
	Type() FSType
	File(string) *File
	Touch(string) error
	MkDir(string) error
	Remove(string) error
	Write(string, []byte) error
	Read(string) ([]byte, error)
	Exists(string) bool
	IsDir(string) bool
}
