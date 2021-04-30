package filesystemutils

import (
	"os"
	"path/filepath"
)

var osUserHomDirFunc = os.UserHomeDir
var filePathJoinFunc = filepath.Join

// UserHomeDir Returns the users home directory or on error returns the current dir
func UserHomeDir() string {
	if homeDir, err := osUserHomDirFunc(); err == nil {
		return homeDir
	}
	return "."
}

// FilePathJoin Returns the file path obtained by joining various path elements.
func FilePathJoin(elems ...string) string {
	return filePathJoinFunc(elems...)
}
