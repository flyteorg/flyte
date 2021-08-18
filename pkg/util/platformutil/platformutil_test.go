package platformutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArch(t *testing.T) {
	var amd64 = ArchAmd64
	assert.NotNil(t, amd64)
	assert.Equal(t, "amd64", amd64.String())

	var arch386 = Arch386
	assert.NotNil(t, arch386)
	assert.Equal(t, "386", arch386.String())

	var i386 = Archi386
	assert.NotNil(t, i386)
	assert.Equal(t, "i386", i386.String())

	var x8664 = ArchX86
	assert.NotNil(t, x8664)
	assert.Equal(t, "x86_64", x8664.String())
}

func TestGoosEnum(t *testing.T) {
	var linux = Linux
	assert.NotNil(t, linux)
	assert.Equal(t, "linux", linux.String())

	var windows = Windows
	assert.NotNil(t, windows)
	assert.Equal(t, "windows", windows.String())

	var darwin = Darwin
	assert.NotNil(t, darwin)
	assert.Equal(t, "darwin", darwin.String())
}
