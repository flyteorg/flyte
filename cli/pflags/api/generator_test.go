package api

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Make sure existing config file(s) parse correctly before overriding them with this flag!
var update = flag.Bool("update", false, "Updates testdata")

func TestNewGenerator(t *testing.T) {
	g, err := NewGenerator(".", "TestType")
	assert.NoError(t, err)

	ctx := context.Background()
	p, err := g.Generate(ctx)
	assert.NoError(t, err)

	codeOutput, err := ioutil.TempFile("", "output-*.go")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(codeOutput.Name())) }()

	testOutput, err := ioutil.TempFile("", "output-*_test.go")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(testOutput.Name())) }()

	assert.NoError(t, p.WriteCodeFile(codeOutput.Name()))
	assert.NoError(t, p.WriteTestFile(testOutput.Name()))

	codeBytes, err := ioutil.ReadFile(codeOutput.Name())
	assert.NoError(t, err)

	testBytes, err := ioutil.ReadFile(testOutput.Name())
	assert.NoError(t, err)

	goldenFilePath := filepath.Join("testdata", "testtype.go")
	goldenTestFilePath := filepath.Join("testdata", "testtype_test.go")
	if *update {
		assert.NoError(t, ioutil.WriteFile(goldenFilePath, codeBytes, os.ModePerm))
		assert.NoError(t, ioutil.WriteFile(goldenTestFilePath, testBytes, os.ModePerm))
	}

	goldenOutput, err := ioutil.ReadFile(filepath.Clean(goldenFilePath))
	assert.NoError(t, err)
	assert.Equal(t, goldenOutput, codeBytes)

	goldenTestOutput, err := ioutil.ReadFile(filepath.Clean(goldenTestFilePath))
	assert.NoError(t, err)
	assert.Equal(t, string(goldenTestOutput), string(testBytes))
}
