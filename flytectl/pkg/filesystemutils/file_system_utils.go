package filesystemutils

import (
	"archive/tar"
	"fmt"
	"io"
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

func ExtractTar(ss io.Reader, destination string) error {
	tarReader := tar.NewReader(ss)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			fmt.Printf("Creating Flyte configuration file at: %s\n", destination)
			outFile, err := os.Create(destination)
			if err != nil {
				return err
			}
			for {
				// Read one 1MB at a time.
				if _, err := io.CopyN(outFile, tarReader, 1024*1024); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
			}
			outFile.Close()

		default:
			return fmt.Errorf("ExtractTarGz: unknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return nil
}
