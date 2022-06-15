package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testVersion = "v0.1.20"

func TestWriteIntoFile(t *testing.T) {
	t.Run("Successfully write into a file", func(t *testing.T) {
		err := WriteIntoFile([]byte(""), "version.yaml")
		assert.Nil(t, err)
	})
	t.Run("Error in writing file", func(t *testing.T) {
		err := WriteIntoFile([]byte(""), "version.yaml")
		assert.Nil(t, err)
	})
}

func TestSetupFlyteDir(t *testing.T) {
	assert.Nil(t, SetupFlyteDir())
}

func TestPrintSandboxMessage(t *testing.T) {
	t.Run("Print Sandbox Message", func(t *testing.T) {
		PrintSandboxMessage(SandBoxConsolePort)
	})
}

func TestSendRequest(t *testing.T) {
	t.Run("Successful get request", func(t *testing.T) {
		response, err := SendRequest("GET", "https://github.com", nil)
		assert.Nil(t, err)
		assert.NotNil(t, response)
	})
	t.Run("Successful get request failed", func(t *testing.T) {
		response, err := SendRequest("GET", "htp://github.com", nil)
		assert.NotNil(t, err)
		assert.Nil(t, response)
	})
	t.Run("Successful get request failed", func(t *testing.T) {
		response, err := SendRequest("GET", "https://github.com/evalsocket/flyte/archive/refs/tags/source-code.zip", nil)
		assert.NotNil(t, err)
		assert.Nil(t, response)
	})
}

func TestIsVersionGreaterThan(t *testing.T) {
	t.Run("Compare FlyteCTL version when upgrade available", func(t *testing.T) {
		_, err := IsVersionGreaterThan("v1.1.21", testVersion)
		assert.Nil(t, err)
	})
	t.Run("Compare FlyteCTL version greater then", func(t *testing.T) {
		ok, err := IsVersionGreaterThan("v1.1.21", testVersion)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
	})
	t.Run("Compare FlyteCTL version greater then for equal value", func(t *testing.T) {
		ok, err := IsVersionGreaterThan(testVersion, testVersion)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
	})
	t.Run("Compare FlyteCTL version smaller then", func(t *testing.T) {
		ok, err := IsVersionGreaterThan("v0.1.19", testVersion)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
	})
	t.Run("Compare FlyteCTL version", func(t *testing.T) {
		_, err := IsVersionGreaterThan(testVersion, testVersion)
		assert.Nil(t, err)
	})
	t.Run("Error in compare FlyteCTL version", func(t *testing.T) {
		_, err := IsVersionGreaterThan("vvvvvvvv", testVersion)
		assert.NotNil(t, err)
	})
	t.Run("Error in compare FlyteCTL version", func(t *testing.T) {
		_, err := IsVersionGreaterThan(testVersion, "vvvvvvvv")
		assert.NotNil(t, err)
	})
}
