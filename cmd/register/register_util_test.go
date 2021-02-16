package register

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

var (
	ctx       context.Context
	args      []string
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func setup() {
	ctx = context.Background()
	Client = &MockClient{}
	validTar, err := os.Open("testdata/valid-register.tar")
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
		os.Exit(-1)
	}
	response := &http.Response{
		Body: validTar,
	}
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return response, nil
	}
}

func TestGetSortedFileList(t *testing.T) {
	setup()
	filesConfig.Archive = false
	args = []string{"file2", "file1"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, "file1", fileList[0])
	assert.Equal(t, "file2", fileList[1])
	assert.Equal(t, tmpDir, "")
	assert.Nil(t, err)
}

func TestGetSortedArchivedFileWithParentFolderList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/valid-parent-folder-register.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, filepath.Join(tmpDir, "parentfolder", "014_recipes.core.basic.basic_workflow.t1_1.pb"), fileList[0])
	assert.Equal(t, filepath.Join(tmpDir, "parentfolder", "015_recipes.core.basic.basic_workflow.t2_1.pb"), fileList[1])
	assert.Equal(t, filepath.Join(tmpDir, "parentfolder", "016_recipes.core.basic.basic_workflow.my_wf_2.pb"), fileList[2])
	assert.Equal(t, filepath.Join(tmpDir, "parentfolder", "017_recipes.core.basic.basic_workflow.my_wf_3.pb"), fileList[3])
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.Nil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/valid-register.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, filepath.Join(tmpDir, "014_recipes.core.basic.basic_workflow.t1_1.pb"), fileList[0])
	assert.Equal(t, filepath.Join(tmpDir, "015_recipes.core.basic.basic_workflow.t2_1.pb"), fileList[1])
	assert.Equal(t, filepath.Join(tmpDir, "016_recipes.core.basic.basic_workflow.my_wf_2.pb"), fileList[2])
	assert.Equal(t, filepath.Join(tmpDir, "017_recipes.core.basic.basic_workflow.my_wf_3.pb"), fileList[3])
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.Nil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileUnorderedList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/valid-unordered-register.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, filepath.Join(tmpDir, "014_recipes.core.basic.basic_workflow.t1_1.pb"), fileList[0])
	assert.Equal(t, filepath.Join(tmpDir, "015_recipes.core.basic.basic_workflow.t2_1.pb"), fileList[1])
	assert.Equal(t, filepath.Join(tmpDir, "016_recipes.core.basic.basic_workflow.my_wf_2.pb"), fileList[2])
	assert.Equal(t, filepath.Join(tmpDir, "017_recipes.core.basic.basic_workflow.my_wf_3.pb"), fileList[3])
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.Nil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedCorruptedFileList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/invalid.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 0)
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedTgzList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/valid-register.tgz"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, filepath.Join(tmpDir, "014_recipes.core.basic.basic_workflow.t1_1.pb"), fileList[0])
	assert.Equal(t, filepath.Join(tmpDir, "015_recipes.core.basic.basic_workflow.t2_1.pb"), fileList[1])
	assert.Equal(t, filepath.Join(tmpDir, "016_recipes.core.basic.basic_workflow.my_wf_2.pb"), fileList[2])
	assert.Equal(t, filepath.Join(tmpDir, "017_recipes.core.basic.basic_workflow.my_wf_3.pb"), fileList[3])
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.Nil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedCorruptedTgzFileList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/invalid.tgz"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedInvalidArchiveFileList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"testdata/invalid-extension-register.zip"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("only .tar and .tgz extension archives are supported"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughInvalidHttpList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"http://invalidhost:invalidport/testdata/valid-register.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughValidHttpList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"http://dummyhost:80/testdata/valid-register.tar"}
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, filepath.Join(tmpDir, "014_recipes.core.basic.basic_workflow.t1_1.pb"), fileList[0])
	assert.Equal(t, filepath.Join(tmpDir, "015_recipes.core.basic.basic_workflow.t2_1.pb"), fileList[1])
	assert.Equal(t, filepath.Join(tmpDir, "016_recipes.core.basic.basic_workflow.my_wf_2.pb"), fileList[2])
	assert.Equal(t, filepath.Join(tmpDir, "017_recipes.core.basic.basic_workflow.my_wf_3.pb"), fileList[3])
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.Nil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughValidHttpWithNullContextList(t *testing.T) {
	setup()
	filesConfig.Archive = true
	args = []string{"http://dummyhost:80/testdata/valid-register.tar"}
	ctx = nil
	fileList, tmpDir, err := getSortedFileList(ctx, args)
	assert.Equal(t, len(fileList), 0)
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("net/http: nil Context"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}
