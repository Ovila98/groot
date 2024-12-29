package groot

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ovila98/ers"
)

// Root env key.
var grootEnv = "GROOT"

// SetGrootKey changes the environment variable key used to store the root path.
// Returns error if key is empty.
func SetGrootKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ers.New("key cannot be empty")
	}
	grootEnv = key
	return nil
}

// ErrNoEnvDefined indicates no environment files were defined or found
var ErrNoEnvDefined = errors.New("no env defined")

// ErrMissingEnvs indicates one or more specified environment files were not found
var ErrMissingEnvs = errors.New("missing env files")

// ErrBadEnvsDefined indicates invalid environment filenames were provided
var ErrBadEnvsDefined = errors.New("bad env files defined")

// IterateThroughPath returns a slice of paths starting from the given path
// up to the filesystem root. The returned paths are valid but may not exist.
// Absolute paths are recommended.
func IterateThroughPath(path string) []string {
	path = ensureCleanPath(path)

	var paths []string
	for path != filepath.Dir(path) {
		paths = append(paths, path)
		path = filepath.Dir(path)
	}
	paths = append(paths, path)
	return paths
}

// SetRoot establishes the project root directory and loads environment files.
//
// The root is set to the directory containing the first occurrence of entryFile,
// searching upward from the current directory.
//
// Environment files are loaded from all directories up to root:
//
// - Only filenames should be provided (no paths)
//
// - Duplicate filenames are treated as one
//
// - All occurrences of each env file are loaded
//
// - The entry file is loaded if it ends in .env
//
// Returns:
//
// - nil on success
//
// - ErrNoEnvDefined if no env files found/specified
//
// - ErrMissingEnvs if any specified env file not found
//
// - ErrBadEnvsDefined if invalid env filenames provided
func SetRoot(entryFile string, envFiles ...string) error {
	entryFile = strings.TrimSpace(entryFile)
	if entryFile == "" {
		return ers.New("entry file not defined")
	}
	definedEnvsFlag := strings.TrimSpace(strings.Join(envFiles, "")) != ""
	cleanEnvFilenames := cleanFilenames(envFiles...)
	if definedEnvsFlag && len(cleanEnvFilenames) == 0 {
		return ers.Wrap(ErrBadEnvsDefined)
	}

	projectDir, err := GetProjectDir()
	if err != nil {
		return ers.Wrap(err)
	}

	foundEnvPaths := make([]string, 0)
	for _, path := range IterateThroughPath(projectDir) {
		found, err := findFiles(path, cleanEnvFilenames)
		if err != nil {
			return ers.Wrap(err)
		}
		foundEnvPaths = append(foundEnvPaths, found...)
		if f, err := os.Stat(filepath.Join(path, entryFile)); err == nil && !f.IsDir() {
			os.Setenv(grootEnv, path)
			break
		}
	}
	root := os.Getenv(grootEnv)

	if root == "" {
		return ers.New("no root found")
	}

	if strings.HasSuffix(entryFile, ".env") {
		foundEnvPaths = append(foundEnvPaths, filepath.Join(root, entryFile))
	}

	if len(envFiles) == 0 || !definedEnvsFlag {
		if len(foundEnvPaths) != 0 {
			// If no env files are provided and entryFile is *.env then use it
			err := godotenv.Load(foundEnvPaths...)
			return ers.Wrap(err)
		}
		return ers.Wrap(ErrNoEnvDefined)
	}

	if definedEnvsFlag {
		// Convert foundEnvPaths to just filenames for comparison
		foundFilenames := make(map[string]struct{})
		for _, path := range foundEnvPaths {
			foundFilenames[filepath.Base(path)] = struct{}{}
		}

		// Check if each required env file was found
		for _, requiredFile := range cleanEnvFilenames {
			if _, exists := foundFilenames[requiredFile]; !exists {
				return ers.Wrap(ErrMissingEnvs)
			}
		}
	}
	if len(foundEnvPaths) > 0 {
		err := godotenv.Load(foundEnvPaths...)
		if err != nil {
			return ers.Wrap(err)
		}
	}

	return nil
}

// SetRootNoEnv sets the project root without requiring environment files.
// Ignores ErrNoEnvDefined and returns other errors.
func SetRootNoEnv(entryFile string) error {
	err := SetRoot(entryFile)
	if errors.Is(err, ErrNoEnvDefined) {
		return nil
	}
	return ers.Wrap(err)
}

// SetRootFromEnv sets the root directory and loads the entry file as an env file.
func SetRootFromEnv(entryFile string) error {
	err := SetRoot(entryFile, entryFile)
	if err != nil {
		return ers.Wrap(err)
	}
	return nil
}

// SetRootFromGit sets the root to the nearest parent git repository.
func SetRootFromGit() error {
	projectDir, err := GetProjectDir()
	if err != nil {
		return ers.Wrap(err)
	}
	root := FindGitRootFrom(projectDir)
	if root == "" {
		return ers.New("no git root found")
	}
	os.Setenv(grootEnv, root)
	return nil
}

// SetRootFromPath sets the root directory from the given path.
// If path is absolute, sets root to that path.
// If path is relative, resolves it from the project directory.
// Returns error if path is empty, invalid, or does not exist.
func SetRootFromPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ers.New("path cannot be empty")
	}

	if !filepath.IsAbs(path) {
		projectDir, err := GetProjectDir()
		if err != nil {
			return ers.Wrap(err)
		}
		path = filepath.Join(projectDir, path)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return ers.Wrap(err)
	}

	if !fi.IsDir() {
		return ers.New("path is not a directory")
	}

	os.Setenv(grootEnv, path)
	return nil
}

// GetRoot returns the current project root directory.
// Returns empty string if not set.
func GetRoot() string {
	return os.Getenv(grootEnv)
}

// FromRoot joins the given path elements with the root directory.
// If root is not set or first path is absolute, joins paths without root.
func FromRoot(path ...string) string {
	root := GetRoot()
	if root == "" || filepath.IsAbs(path[0]) {
		return filepath.Join(path...)
	}
	return filepath.Join(root, filepath.Join(path...))
}

// FindGitRootFrom locates the nearest parent git repository from startPath.
// Returns empty string if none found.
func FindGitRootFrom(startPath string) string {
	paths := IterateThroughPath(startPath)

	for _, path := range paths {
		if f, err := os.Stat(filepath.Join(path, ".git")); err == nil && f.IsDir() {
			return path
		}
	}
	return ""
}

func GetMainFile() (string, error) {
	callFrame := 0
	for {
		_, _, _, ok := runtime.Caller(callFrame)
		if !ok {
			break
		}
		callFrame++
	}
	_, goFile, _, _ := runtime.Caller(callFrame - 3)
	if !strings.HasSuffix(goFile, ".go") {
		return "", ers.New("main *.go file not found")
	}
	if len(goFile) > 1 {
		goFile = strings.ToUpper(goFile[0:1]) + goFile[1:]
	}
	return goFile, nil
}

// GetProjectDir returns either the directory containing the executable
// or the directory containing the file containing main() depending on
// calling context ('go run' or standalone executable).
func GetProjectDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", ers.Wrap(err)
	}
	execDir := filepath.Dir(execPath)

	goFile, err := GetMainFile()
	if err != nil {
		return "", ers.Wrap(err)
	}
	goDir := filepath.Dir(goFile)

	tempDir := os.TempDir()

	// If execPath contains temp dir,
	// that means that 'go run' has been called
	// So get the file containing main()
	// And return its directory
	if strings.Contains(execDir, tempDir) {
		return goDir, nil
	}

	// Ensure execDir is a directory
	fi, err := os.Stat(execDir)
	if err != nil {
		return "", ers.Wrap(err)
	}

	if fi.IsDir() {
		return execDir, nil
	}

	return "", ers.New("unable to get project dir")
}

// IsRoot checks if the provided path is the project root directory
func IsRoot(path string) bool {
	root := GetRoot()
	if root == "" {
		return false
	}
	cleanPath := ensureCleanPath(path)
	cleanRoot := ensureCleanPath(root)
	return cleanPath == cleanRoot
}

// IsTemporary checks wether the current execution context is temporary.
// (i.e. if 'go run' has been called).
func IsTemporary() bool {
	executable, _ := os.Executable()
	return strings.Contains(executable, "go-build")
}

// GetRootParent returns the parent directory of the project root.
// Returns an empty string if root is not set or if root is the filesystem root.
func GetRootParent() string {
	root := GetRoot()
	if root == "" {
		return ""
	}
	parent := filepath.Dir(root)
	if parent == root {
		return ""
	}
	return parent
}

// GetRelativeToRoot returns the relative path from root to the given path.
// Returns an error if root is not set or if path is not under root.
func GetRelativeToRoot(path string) (string, error) {
	root := GetRoot()
	if root == "" {
		return "", ers.New("root not set")
	}

	cleanPath := ensureCleanPath(path)
	cleanRoot := ensureCleanPath(root)

	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return "", ers.Wrap(err)
	}

	return rel, nil
}

// ListFilesFromRoot returns a slice of file paths matching the given pattern relative to root.
// Pattern follows filepath.Glob syntax.
func ListFilesFromRoot(pattern string) ([]string, error) {
	root := GetRoot()
	if root == "" {
		return nil, ers.New("root not set")
	}

	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return nil, ers.Wrap(err)
	}

	return matches, nil
}

// WalkFromRoot walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
func WalkFromRoot(fn fs.WalkDirFunc) error {
	root := GetRoot()
	if root == "" {
		return ers.New("root not set")
	}

	err := filepath.WalkDir(root, fn)
	if err != nil {
		return ers.Wrap(err)
	}

	return nil
}

// GetRootInfo returns FileInfo for the root directory.
// Returns error if root is not set or cannot be accessed.
func GetRootInfo() (os.FileInfo, error) {
	root := GetRoot()
	if root == "" {
		return nil, ers.New("root not set")
	}

	fi, err := os.Stat(root)
	if err != nil {
		return nil, ers.Wrap(err)
	}

	return fi, nil
}

// GetRootName returns the name of the root directory.
// Returns empty string if root is not set.
func GetRootName() string {
	fi, err := GetRootInfo()
	if err != nil {
		return ""
	}
	return fi.Name()
}

// ValidateRoot verifies that root is properly set and exists on the filesystem
func ValidateRoot() error {
	root := GetRoot()
	if root == "" {
		return ers.New("root not set")
	}

	fi, err := os.Stat(root)
	if err != nil {
		return ers.Wrap(err)
	}

	if !fi.IsDir() {
		return ers.New("root is not a directory")
	}

	return nil
}

// ClearRoot unsets the GROOT environment variable
func ClearRoot() {
	os.Unsetenv(grootEnv)
}

// IsInRoot checks if the given path is within the project root directory
func IsInRoot(path string) bool {
	root := GetRoot()
	if root == "" {
		return false
	}

	cleanPath := ensureCleanPath(path)
	cleanRoot := ensureCleanPath(root)

	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return false
	}

	// Check if path attempts to traverse outside root with ../
	return !strings.HasPrefix(rel, "..")
}

// MustGetRoot returns the root directory of the project.
// Panics if root is not set.
func MustGetRoot() string {
	root := GetRoot()
	if root == "" {
		panic("root not set")
	}
	return root
}
