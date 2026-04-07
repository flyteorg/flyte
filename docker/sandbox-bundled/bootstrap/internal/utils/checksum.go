package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func FileChecksum(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func FileCollectionChecksum(paths []string) ([]byte, error) {
	// Compute individual checksums and sort
	var checksumPaths []string
	for _, p := range paths {
		c, err := FileChecksum(p)
		if err != nil {
			return nil, err
		}
		checksumPaths = append(checksumPaths, fmt.Sprintf("%s\t%s", hex.EncodeToString(c), p))
	}
	sort.Strings(checksumPaths)

	// Compute checksum of sorted checksum path pairs
	h := sha256.New()
	h.Write([]byte(strings.Join(checksumPaths, "\n")))
	return h.Sum(nil), nil
}
