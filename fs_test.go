package naivefs

import (
	"errors"
	"os"
	"path"
	"reflect"
	"testing"
)

func testFSTouch(t *testing.T, fs FS) {
	name := "some_file.txt"
	err := fs.Touch(name)
	if err != nil {
		t.Fatal(err)
	}
	if !fs.Exists(name) {
		t.Fatal("expected file to exist")
	}
	if fs.IsDir(name) {
		t.Fatal("expected to be file, but is dir")
	}
}

func testFSMkDir(t *testing.T, fs FS) {
	name := "some_folder"
	err := fs.MkDir(name)
	if err != nil {
		t.Fatal(err)
	}
	if !fs.Exists(name) {
		t.Fatal("expected file to exist")
	}
	if !fs.IsDir(name) {
		t.Fatal("expected to be dir, but not")
	}
}

func testFSRemove(t *testing.T, fs FS) {
	files := []string{
		"some_file.txt",
		"some_folder/some_file.txt",
	}
	for _, file := range files {
		err := fs.Touch(file)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range files {
		if !fs.Exists(file) {
			t.Fatal(errors.New("touch is not working"))
		}
		err := fs.Remove(file)
		if err != nil {
			t.Fatal(err)
		}
		if fs.Exists(file) {
			t.Fatal("expect file to be removed, but still exists")
		}
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

func testFSGeneral(t *testing.T, fs FS) {
	testFSTouch(t, fs)
	testFSMkDir(t, fs)
	testFSRemove(t, fs)
	testFSWrite(t, fs)
	fs.Remove(".")
}

func TestAllFS(t *testing.T) {
	testFSGeneral(t, NewFSLocal(path.Join(os.TempDir(), "naivefs_test")))
}
