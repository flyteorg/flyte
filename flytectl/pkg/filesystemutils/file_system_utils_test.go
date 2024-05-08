package filesystemutils

import (
	"archive/tar"
	"fmt"
	"os"
	"testing"
	"time"

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

func TestTaring(t *testing.T) {
	// Create a fake tar file in tmp.
	text := "a: b"
	fo, err := os.CreateTemp("", "sampledata")
	assert.NoError(t, err)
	tarWriter := tar.NewWriter(fo)
	err = tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "flyte.yaml",
		Size:     4,
		Mode:     0640,
		ModTime:  time.Unix(1245206587, 0),
	})
	assert.NoError(t, err)
	cnt, err := tarWriter.Write([]byte(text))
	assert.NoError(t, err)
	assert.Equal(t, 4, cnt)
	tarWriter.Close()
	fo.Close()

	t.Run("Basic testing", func(t *testing.T) {
		destFile, err := os.CreateTemp("", "sampledata")
		assert.NoError(t, err)
		reader, err := os.Open(fo.Name())
		assert.NoError(t, err)
		err = ExtractTar(reader, destFile.Name())
		assert.NoError(t, err)
		fileBytes, err := os.ReadFile(destFile.Name())
		assert.NoError(t, err)
		readString := string(fileBytes)
		assert.Equal(t, text, readString)

		// Try to extract the file we just extracted again. It's not a tar file obviously so it should error
		reader, err = os.Open(destFile.Name())
		assert.NoError(t, err)
		err = ExtractTar(reader, destFile.Name())
		assert.Errorf(t, err, "unexpected EOF")
	})
}

func TestTarBadHeader(t *testing.T) {
	// Create a fake tar file in tmp.
	fo, err := os.CreateTemp("", "sampledata")
	assert.NoError(t, err)
	tarWriter := tar.NewWriter(fo)
	// Write a symlink, we should not know how to parse.
	err = tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeLink,
		Name:     "flyte.yaml",
		Size:     4,
		Mode:     0640,
		ModTime:  time.Unix(1245206587, 0),
	})
	assert.NoError(t, err)
	tarWriter.Close()
	fo.Close()

	t.Run("Basic testing", func(t *testing.T) {
		destFile, err := os.CreateTemp("", "sampledata")
		assert.NoError(t, err)
		reader, err := os.Open(fo.Name())
		assert.NoError(t, err)
		err = ExtractTar(reader, destFile.Name())
		assert.Errorf(t, err, "ExtractTarGz: unknown type")
	})
}
