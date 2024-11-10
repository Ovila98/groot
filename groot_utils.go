package groot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ovila98/ers"
)

// function copied from filepath.ToSlash()
func replaceStringByte(s string, old, new byte) string {
	if strings.IndexByte(s, old) == -1 {
		return s
	}
	n := []byte(s)
	for i := range n {
		if n[i] == old {
			n[i] = new
		}
	}
	return string(n)
}

// ensureCleanPath trims spaces, normalizes separators and removes duplicate separators
func ensureCleanPath(path string) string {
	path = replaceStringByte(strings.TrimSpace(path), os.PathSeparator, '/')
	path = replaceStringByte(strings.TrimSpace(path), '/', os.PathSeparator)
	return strings.ReplaceAll(
		path,
		fmt.Sprintf("%c%c", os.PathSeparator, os.PathSeparator),
		string(os.PathSeparator),
	)
}

// cleanFilenames removes duplicate filenames and returns a slice of unique filenames
func cleanFilenames(filenames ...string) []string {
	uniqueFilenames := make(map[string]struct{})
	for _, filename := range filenames {
		filename = replaceStringByte(strings.TrimSpace(filename), os.PathSeparator, '/')
		if filename == "" || strings.Contains(filename, "/") {
			// skip empty filenames and filenames with slashes (paths)
			continue
		}
		uniqueFilenames[filename] = struct{}{}
	}
	uniqueFilenamesSlice := make([]string, 0, len(uniqueFilenames))
	for filename := range uniqueFilenames {
		uniqueFilenamesSlice = append(uniqueFilenamesSlice, filename)
	}
	return uniqueFilenamesSlice
}

// findFiles returns a slice of found files in a directory
func findFiles(dirPath string, fileNames []string) ([]string, error) {
	var foundFiles []string
	for _, fileName := range fileNames {
		files, err := filepath.Glob(filepath.Join(dirPath, fileName))
		if err != nil {
			return nil, ers.Wrap(err)
		}
		foundFiles = append(foundFiles, files...)
	}
	return foundFiles, nil
}
