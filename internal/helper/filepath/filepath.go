// Package filepath provides extended file path utilities beyond the
// standard library's filepath package, including relative path resolution.
package filepath

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// GetAbsPath converts a relative path to an absolute path relative to
// the calling source file.
// If the input path is already absolute, it is returned unchanged.
// This is useful for resolving paths in development environments
// where relative paths might be used.
func GetAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("get caller info")
	}
	baseDir := filepath.Dir(filename)
	absolutePath := filepath.Join(baseDir, path)
	return absolutePath, nil
}
