package clusterresource

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
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

	templateValues, err := populateTemplateValues(data)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"{{ directValue }}": "1",
		"{{ envValue }}":    "2",
		"{{ filePath }}":    "3",
	}, templateValues)
}

func TestPopulateDefaultTemplateValues(t *testing.T) {
	testDefaultData := map[runtimeInterfaces.DomainName]runtimeInterfaces.TemplateData{
		"production": {
			"var1": {
				Value: "prod1",
			},
			"var2": {
				Value: "prod2",
			},
		},
		"development": {
			"var1": {
				Value: "dev1",
			},
			"var2": {
				Value: "dev2",
			},
		},
	}
	templateValues, err := populateDefaultTemplateValues(testDefaultData)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]templateValuesType{
		"production": {
			"{{ var1 }}": "prod1",
			"{{ var2 }}": "prod2",
		},
		"development": {
			"{{ var1 }}": "dev1",
			"{{ var2 }}": "dev2",
		},
	}, templateValues)
}

func TestGetCustomTemplateValues(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	projectDomainAttributes := admin.ProjectDomainAttributes{
		Project: "project-foo",
		Domain:  "domain-bar",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ClusterResourceAttributes{ClusterResourceAttributes: &admin.ClusterResourceAttributes{
				Attributes: map[string]string{
					"var1": "val1",
					"var2": "val2",
				},
			},
			},
		},
	}
	resourceModel, err := transformers.ProjectDomainAttributesToResourceModel(projectDomainAttributes, admin.MatchableResource_CLUSTER_RESOURCE)
	assert.Nil(t, err)
	mockRepository.ResourceRepo().(*repositoryMocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (resource models.Resource, e error) {
		assert.Equal(t, "project-foo", ID.Project)
		assert.Equal(t, "domain-bar", ID.Domain)
		return resourceModel, nil
	}
	testController := controller{
		db:              mockRepository,
		resourceManager: resources.NewResourceManager(mockRepository, testutils.GetApplicationConfigWithDefaultDomains()),
	}
	domainTemplateValues := templateValuesType{
		"{{ var1 }}": "i'm getting overwritten",
		"{{ var3 }}": "persist",
	}
	customTemplateValues, err := testController.getCustomTemplateValues(context.Background(), "project-foo",
		"domain-bar", domainTemplateValues)
	assert.Nil(t, err)
	assert.EqualValues(t, templateValuesType{
		"{{ var1 }}": "val1",
		"{{ var2 }}": "val2",
		"{{ var3 }}": "persist",
	}, customTemplateValues)

	assert.NotEqual(t, domainTemplateValues, customTemplateValues)
}

func TestGetCustomTemplateValues_NothingToOverride(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	testController := controller{
		db:              mockRepository,
		resourceManager: resources.NewResourceManager(mockRepository, testutils.GetApplicationConfigWithDefaultDomains()),
	}
	customTemplateValues, err := testController.getCustomTemplateValues(context.Background(), "project-foo", "domain-bar", templateValuesType{
		"{{ var1 }}": "val1",
		"{{ var2 }}": "val2",
	})
	assert.Nil(t, err)
	assert.EqualValues(t, templateValuesType{
		"{{ var1 }}": "val1",
		"{{ var2 }}": "val2",
	}, customTemplateValues,
		"missing project-domain combinations in the db should result in the config defaults being applied")
}

func TestGetCustomTemplateValues_InvalidDBModel(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	mockRepository.ResourceRepo().(*repositoryMocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (resource models.Resource, e error) {
		return models.Resource{
			Attributes: []byte("i'm invalid"),
		}, nil
	}
	testController := controller{
		db:              mockRepository,
		resourceManager: resources.NewResourceManager(mockRepository, testutils.GetApplicationConfigWithDefaultDomains()),
	}
	_, err := testController.getCustomTemplateValues(context.Background(), "project-foo", "domain-bar", templateValuesType{
		"{{ var1 }}": "val1",
		"{{ var2 }}": "val2",
	})
	assert.NotNil(t, err,
		"invalid project-domain combinations in the db should result in the config defaults being applied")
}
