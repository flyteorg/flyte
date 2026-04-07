package config

import (
	"os"
	"path/filepath"
	"strings"
)

func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	return filepath.Clean(path)
}
