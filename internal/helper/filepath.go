package helper

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func GetAbsFilePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller info")
	}
	baseDir := filepath.Dir(filename)
	absolutePath := filepath.Join(baseDir, path)
	return absolutePath, nil
}
