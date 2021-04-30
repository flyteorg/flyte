package filesystemutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	homeDirVal = "/home/user"
	homeDirErr error
)

func FakeUserHomeDir() (string, error) {
	return homeDirVal, homeDirErr
}

func TestUserHomeDir(t *testing.T) {
	t.Run("User home dir", func(t *testing.T) {
		osUserHomDirFunc = FakeUserHomeDir
		homeDir := UserHomeDir()
		assert.Equal(t, homeDirVal, homeDir)
	})
	t.Run("User home dir fail", func(t *testing.T) {
		homeDirErr = fmt.Errorf("failed to get users home directory")
		homeDirVal = "."
		osUserHomDirFunc = FakeUserHomeDir
		homeDir := UserHomeDir()
		assert.Equal(t, ".", homeDir)
		// Reset
		homeDirErr = nil
		homeDirVal = "/home/user"
	})
}

func TestFilePathJoin(t *testing.T) {
	t.Run("File path join", func(t *testing.T) {
		homeDir := FilePathJoin("/", "home", "user")
		assert.Equal(t, "/home/user", homeDir)
	})
}
