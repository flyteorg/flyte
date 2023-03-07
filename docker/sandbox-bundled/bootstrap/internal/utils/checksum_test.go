package utils

import (
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileChecksum(t *testing.T) {
	c, err := FileChecksum(filepath.Join("testdata/foo"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t,
		hex.EncodeToString(c),
		"b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		"checksums should match",
	)
}

func TestFileCollectionChecksum(t *testing.T) {
	c, err := FileCollectionChecksum(
		[]string{filepath.Join("testdata/foo"), filepath.Join("testdata/bar")},
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t,
		hex.EncodeToString(c),
		"2db7bade901cc7260437c7cdd91b9f0cdb34fe3e5c72c5d0e756f3ce7d0c3f6a",
		"checksums should match",
	)
}
