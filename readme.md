# Groot

[![GoDoc](https://godoc.org/github.com/ovila98/groot?status.svg)](https://godoc.org/github.com/ovila98/groot)
[![Email](https://img.shields.io/badge/email-ovila.acolatse.dev%40gmail.com-blue)](mailto:ovila.acolatse.dev@gmail.com)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-Ovila%20Acolatse-blue)](https://www.linkedin.com/in/ovila-acolatse/)

Groot is a Go package that solves the challenge of consistent project root management across different execution contexts. It handles the complexities of path resolution whether your code runs as an interpreted file (`go run`) or a compiled executable, ensuring reliable resource access in both development and production environments.

## Why Groot?

Go's file path handling can be tricky when dealing with:

- Different execution contexts (temp directory for `go run` vs actual location for binaries)
- Working directory variations
- Resource path resolution in development vs production
- Project portability across environments

Groot provides a unified solution to these challenges by establishing and maintaining a consistent project root, making your Go applications truly portable.

## Features

- Multiple ways to set project root:
  - Using entry files
  - Using Git repository detection
  - Using environment files
- Cross-platform path handling
- Flexible environment file loading
- Rich utility functions for path operations
- Clean error handling

## Installation

```bash
go get github.com/ovila98/groot
```

## Quick Start

```go
package main

import "github.com/ovila98/groot"

func main() {
    // Set root using app.id file and load env1.env, env2.env
    err := groot.SetRoot("app.id", "env1.env", "env2.env")
    if err != nil {
        panic(err)
    }

    // Get root directory
    root := groot.GetRoot()

    // Get path relative to root
    configPath := groot.FromRoot("config", "app.yaml")
}
```

## Usage Examples

### Setting Root Directory

```go
// Using entry file with env files
err := groot.SetRoot("app.id", "dev.env", "local.env")

// Using entry file without env files
err := groot.SetRootNoEnv("app.id")

// Using Git repository
err := groot.SetRootFromGit()

// Using environment file
err := groot.SetRootFromEnv(".env")
```

### Path Operations

```go
// Get path relative to root
configPath := groot.FromRoot("config", "settings.json")

// Check if path is in root
isInRoot := groot.IsInRoot("/path/to/file")

// Get relative path from root
relPath, err := groot.GetRelativeToRoot("/absolute/path")
```

### Root Information

```go
// Get root directory
root := groot.GetRoot()

// Get root directory name
name := groot.GetRootName()

// Get root parent directory
parent := groot.GetRootParent()

// Get root directory info
info, err := groot.GetRootInfo()
```

### File Operations

```go
// List files from root
files, err := groot.ListFilesFromRoot("*.go")

// Walk directory tree from root
err := groot.WalkFromRoot(func(path string, d fs.DirEntry, err error) error {
    // Process files
    return nil
})
```

## Complete API Reference

### Root Management

- `SetGrootKey(key string) error` - Change environment variable key for root
- `SetRoot(entryFile string, envFiles ...string) error` - Set root using entry file
- `SetRootNoEnv(entryFile string) error` - Set root without env files
- `SetRootFromEnv(entryFile string) error` - Set root using env file
- `SetRootFromGit() error` - Set root using Git repository
- `SetRootFromPath(path string) error` - Set root from absolute or relative path
- `GetRoot() string` - Get current root directory
- `MustGetRoot() string` - Get root directory or panic
- `ClearRoot()` - Clear root setting
- `IsTemporary() bool` - Check if current execution context is temporary

### Path Operations

- `FromRoot(path ...string) string` - Get path relative to root
- `IsRoot(path string) bool` - Check if path is root directory
- `IsInRoot(path string) bool` - Check if path is within root
- `GetRelativeToRoot(path string) (string, error)` - Get relative path from root
- `GetRootParent() string` - Get parent of root directory
- `GetRootName() string` - Get name of root directory

### File Operations

- `ListFilesFromRoot(pattern string) ([]string, error)` - List files matching pattern
- `WalkFromRoot(fn fs.WalkDirFunc) error` - Walk directory tree from root
- `GetRootInfo() (os.FileInfo, error)` - Get root directory information

### Validation

- `ValidateRoot() error` - Verify root is properly set and exists

## License

Apache License 2.0

## Author

Ovila Acolatse

- GitHub: [github.com/ovila98](https://github.com/ovila98)
- LinkedIn: [Ovila Acolatse](https://www.linkedin.com/in/ovila-acolatse/)
- Email: ovila.acolatse.dev@gmail.com
