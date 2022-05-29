package naivefs

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/iloahz/go-naive-fs/utils"
)

func existAndNoError(t *testing.T, fs FS, name string) {
	exists, err := fs.Exists(name)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(fmt.Sprintf("expect %s to exist", name))
	}
}

func notExistAndNoError(t *testing.T, fs FS, name string) {
	exists, err := fs.Exists(name)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal(fmt.Sprintf("expect %s to not exist", name))
	}
}

func testFSTouch(t *testing.T, fs FS) {
	name := "some_file.txt"
	err := fs.Touch(name)
	if err != nil {
		t.Fatal(err)
	}
	existAndNoError(t, fs, name)
	if isDir, _ := fs.IsDir(name); isDir {
		t.Fatal("expected to be file, but is dir")
	}
}

func testFSMkDir(t *testing.T, fs FS) {
	if fs.Type() == FSTypeMinio {
		// minio does not support empty dir
		return
	}
	name := "some_dir"
	err := fs.MkDir(name)
	if err != nil {
		t.Fatal(err)
	}
	existAndNoError(t, fs, name)
	if isDir, _ := fs.IsDir(name); !isDir {
		t.Fatal("expected to be dir, but not")
	}
}

func testFSRemove(t *testing.T, fs FS) {
	files := []string{
		"some_file.txt",
		"some_dir/some_file.txt",
	}
	for _, file := range files {
		err := fs.Touch(file)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range files {
		existAndNoError(t, fs, file)
		err := fs.Remove(file)
		if err != nil {
			t.Fatal(err)
		}
		notExistAndNoError(t, fs, file)
	}
}

func testFSWrite(t *testing.T, fs FS) {
	name := "some_file.txt"
	data := []byte{1, 2, 3, 5, 8}
	err := fs.Write(name, data)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := fs.Read(name)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, buf) {
		t.Fatal("write data mismatch")
	}
}

func testFSExists(t *testing.T, fs FS) {
	name := "some_dir/some_file.txt"
	data := []byte{1, 2, 3, 5, 8}
	utils.Must(fs.Write(name, data))
	existAndNoError(t, fs, "some_dir")
	existAndNoError(t, fs, "some_dir/some_file.txt")
}

func testFSGeneral(t *testing.T, fs FS) {
	testFSTouch(t, fs)
	testFSMkDir(t, fs)
	testFSRemove(t, fs)
	testFSWrite(t, fs)
	testFSExists(t, fs)
	fs.Remove(".")
}

func TestFSLocal(t *testing.T) {
	fs := NewFSLocal(path.Join(os.TempDir(), "naivefs_test"))
	testFSGeneral(t, fs)
}

func TestFSMinio(t *testing.T) {
	config := &MinioConfig{
		Endpoint:        os.Getenv("NAIVEFS_TEST_MINIO_ENDPOINT"),
		AccessKeyID:     os.Getenv("NAIVEFS_TEST_MINIO_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("NAIVEFS_TEST_MINIO_SECRET_ACCESS_KEY"),
		UseSSL:          true,
		BucketName:      os.Getenv("NAIVEFS_TEST_MINIO_BUCKET_NAME"),
		BaseDir:         os.Getenv("NAIVEFS_TEST_MINIO_BASE_DIR"),
	}
	fs := NewFSMinio(config)
	testFSGeneral(t, fs)
}
