package brwsr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var root string

// Entry corresponds to an item on the path - can be either a file or a folder
type Entry struct {
	Name   string
	Folder bool
}

// Mount checks if the path is "mountable", meaning, it's accessible
// and writeable.
func Mount(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	if path == ".." {
		return fmt.Errorf("tried mounting parent directory")
	}

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}

		if os.IsPermission(err) {
			return fmt.Errorf("permission issue: %s", err.Error())
		}

		return fmt.Errorf("unknown issue: %s", err.Error())
	}

	root = path

	return nil
}

// List returns the entries in a given path relative to the root path
func List(relPath string) ([]Entry, error) {
	if strings.HasPrefix(relPath, "..") {
		return nil, fmt.Errorf("relpath starting with '..'")
	}

	list, err := ioutil.ReadDir(fullPath(relPath))
	if err != nil {
		return nil, fmt.Errorf("failed reading dir: %s", err.Error())
	}

	var res []Entry

	for _, item := range list {
		entry := Entry{
			Name:   item.Name(),
			Folder: item.IsDir(),
		}

		res = append(res, entry)
	}

	return res, nil
}

func fullPath(relPath string) string {
	return filepath.Join(root, relPath)
}
