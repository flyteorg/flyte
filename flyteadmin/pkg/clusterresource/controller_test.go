package clusterresource

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var testScope = mockScope.NewTestScope()

type mockFileInfo struct {
	name    string
	modTime time.Time
}

func (i *mockFileInfo) Name() string {
	return i.name
}

func (i *mockFileInfo) Size() int64 {
	return 0
}

func (i *mockFileInfo) Mode() os.FileMode {
	return os.ModeExclusive
}

func (i *mockFileInfo) ModTime() time.Time {
	return i.modTime
}

func (i *mockFileInfo) IsDir() bool {
	return false
}

func (i *mockFileInfo) Sys() interface{} {
	return nil
}

func TestTemplateAlreadyApplied(t *testing.T) {
	const namespace = "namespace"
	const fileName = "fileName"
	var lastModifiedTime = time.Now()
	testController := controller{
		metrics: newMetrics(testScope),
	}
	mockFile := mockFileInfo{
		name:    fileName,
		modTime: lastModifiedTime,
	}
	assert.False(t, testController.templateAlreadyApplied(namespace, &mockFile))

	testController.appliedTemplates = make(map[string]map[string]time.Time)
	testController.appliedTemplates[namespace] = make(map[string]time.Time)
	assert.False(t, testController.templateAlreadyApplied(namespace, &mockFile))

	testController.appliedTemplates[namespace][fileName] = lastModifiedTime.Add(-10 * time.Minute)
	assert.False(t, testController.templateAlreadyApplied(namespace, &mockFile))

	testController.appliedTemplates[namespace][fileName] = lastModifiedTime
	assert.True(t, testController.templateAlreadyApplied(namespace, &mockFile))
}

func TestPopulateTemplateValues(t *testing.T) {
	const testEnvVarName = "TEST_FOO"
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString("3")
	if err != nil {
		t.Fatalf("Cannot write to temporary file: %v", err)
	}

	data := map[string]runtimeInterfaces.DataSource{
		"directValue": {
			Value: "1",
		},
		"envValue": {
			ValueFrom: runtimeInterfaces.DataSourceValueFrom{
				EnvVar: testEnvVarName,
			},
		},
		"filePath": {
			ValueFrom: runtimeInterfaces.DataSourceValueFrom{
				FilePath: tmpFile.Name(),
			},
		},
	}
	origEnvVar := os.Getenv(testEnvVarName)
	defer os.Setenv(testEnvVarName, origEnvVar)
	os.Setenv(testEnvVarName, "2")

	templateValues, err := populateTemplateValues("namespacey", data)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"{{ namespace }}":   "namespacey",
		"{{ directValue }}": "1",
		"{{ envValue }}":    "2",
		"{{ filePath }}":    "3",
	}, templateValues)
}
