package brwsr

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	testRoot           = "dumps"
	testFolder         = "folder"
	testFolderInFolder = "another"
	testFileName       = "file.txt"
)

func setup() {
	os.Mkdir(testRoot, os.ModePerm)
	os.Mkdir(filepath.Join(testRoot, testFolder), os.ModePerm)
	os.Mkdir(filepath.Join(testRoot, testFolder, testFolderInFolder), os.ModePerm)
	file, _ := os.OpenFile(filepath.Join(testRoot, testFileName), os.O_RDWR|os.O_CREATE, 0755)

	file.Close()
}

func teardown() {
	os.RemoveAll(testRoot)
}

func TestMount(t *testing.T) {
	setup()
	defer teardown()

	err := Mount("..")
	if err == nil {
		t.Errorf("Mount('..') should have failed")
	}

	err = Mount("")
	if err == nil {
		t.Errorf("Mount('') should have failed")
	}

	err = Mount("NonExistent")
	if err == nil {
		t.Errorf("Mount('NonExistent') should have failed")
	}

	err = Mount(testRoot)
	if err != nil {
		t.Errorf("Mount(%q) failed to mount: %s", testRoot, err.Error())
	}
}

func TestList(t *testing.T) {
	setup()
	defer teardown()

	Mount(testRoot)

	list, err := List("")
	if err != nil {
		t.Errorf("List('') returned error: %s", err.Error())
	}

	if len(list) != 2 {
		t.Errorf("Lenght of List('') should be 2, is %d", len(list))
	}

	for _, item := range list {
		if item.Folder {
			if item.Name != testFolder {
				t.Errorf("Folder should be %q, is %q instead", testFolder, item.Name)
			}
		} else {
			if item.Name != testFileName {
				t.Errorf("File name should be %q, is %q instead", testFileName, item.Name)
			}
		}
	}

	list, err = List(testFolder)
	if err != nil {
		t.Errorf("List(%q) returned error: %s", testFolder, err.Error())
	}

	if len(list) != 1 {
		t.Errorf("Length of List('') should be 1, is %d", len(list))
	}

	for _, item := range list {
		if !item.Folder {
			t.Errorf("%q should've been a folder, is a file instead", item.Name)
		}

		if item.Name != testFolderInFolder {
			t.Errorf("Folder name should be %q, is %q instead", testFolderInFolder, item.Name)
		}
	}

}

func TestFullPath(t *testing.T) {
	root = "/some/path"

	var tests = []struct {
		input string
		want  string
	}{
		{"hello/relative", root + "/hello/relative"},
		{"/hello/relative", root + "/hello/relative"},
		{"", root},
		{".", root},
		{"/", root},
	}

	for _, test := range tests {
		if out := fullPath(test.input); out != test.want {
			t.Errorf("[root = %q] => fullPath(%q) = %q, want %q", root, test.input, out, test.want)
		}
	}
}
