// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func inTempDir(fun func(dir string)) error {

	curdir, err := os.Getwd()
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "awgo-util-")
	if err != nil {
		return err
	}
	// TempDir() returns a symlink on my macOS :(
	if dir, err = filepath.EvalSymlinks(dir); err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()

	// Change to temporary directory
	if err := os.Chdir(dir); err != nil {
		return err
	}

	// Change back after we're done
	defer func() {
		if err := os.Chdir(curdir); err != nil {
			panic(err)
		}
	}()

	fun(dir)

	return nil
}

func TestMustExist(t *testing.T) {

	err := inTempDir(func(dir string) {

		name := "testdir"

		// Create directory
		s := MustExist(name)
		if s != name {
			t.Errorf("Bad Dirname. Expected=%s, Got=%s", name, s)
		}

		if _, err := os.Stat(s); err != nil {
			t.Errorf("Couldn't stat dir %#v: %v", s, err)
		}

		// Check path is as expected
		p := filepath.Join(dir, name)
		p2, err := filepath.Abs(s)
		if err != nil {
			t.Fatal(err)
		}

		if p != p2 {
			t.Errorf("Bad Path. Expected=%v, Got=%v", p2, p)
		}

	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestPathExists(t *testing.T) {

	err := inTempDir(func(dir string) {

		name := "existingdir"
		path := filepath.Join(dir, name)
		badName := "nodir"
		badPath := filepath.Join(dir, badName)

		if err := os.MkdirAll(name, 0700); err != nil {
			t.Fatal(err)
		}

		data := []struct {
			p string
			x bool
		}{
			{dir, true},
			{name, true},
			{path, true},
			{badName, false},
			{badPath, false},
		}

		for _, td := range data {
			v := PathExists(td.p)
			if v != td.x {
				t.Errorf("Bad PathExists for %#v. Expected=%v, Got=%v", td.p, td.x, v)
			}

		}

	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestClearDirectory(t *testing.T) {
	err := inTempDir(func(dir string) {
		names := []string{"./root/one", "./root/two", "./root/three"}
		for _, s := range names {
			if err := os.MkdirAll(s, 0700); err != nil {
				t.Fatal(err)
			}
		}

		for _, s := range names {
			_, err := os.Stat(s)
			if err != nil {
				t.Error(err)
			}
		}

		if err := ClearDirectory("./root"); err != nil {
			t.Error(err)
		}

		for _, s := range names {
			_, err := os.Stat(s)
			if !os.IsNotExist(err) {
				t.Errorf("file %q exists", s)
			}
		}

	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteFile(t *testing.T) {
	err := inTempDir(func(dir string) {
		var (
			name    = "test.txt"
			content = []byte(`test`)
		)

		if PathExists(name) {
			t.Fatal("Path already exists.")
		}

		if err := WriteFile(name, content, 0600); err != nil {
			t.Fatal(err)
		}

		if !PathExists(name) {
			t.Errorf("Path doesn't exist: %s", name)
		}

		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(data, content) {
			t.Errorf("Bad Content. Expected=%q, Got=%q.", string(content), string(data))
		}

		fi, err := os.Stat(name)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode() != 0600 {
			t.Errorf("Bad FileMode. Expected=0600, Got=%v", fi.Mode())
		}

		infos, err := ioutil.ReadDir(".")
		if err != nil {
			t.Fatal(err)
		}
		if len(infos) != 1 {
			t.Errorf("Bad no. of files. Expected=1, Got=%d", len(infos))
		}

	})
	if err != nil {
		t.Fatal(err)
	}
}
