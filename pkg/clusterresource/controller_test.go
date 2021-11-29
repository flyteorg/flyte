package clusterresource

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

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

func Test_controller_createResourceFromTemplate(t *testing.T) {
	type args struct {
		ctx                  context.Context
		templateDir          string
		templateFileName     string
		project              models.Project
		domain               runtimeInterfaces.Domain
		namespace            NamespaceName
		templateValues       templateValuesType
		customTemplateValues templateValuesType
	}
	tests := []struct {
		name            string
		args            args
		wantK8sManifest string
		wantErr         bool
	}{
		{
			name: "test create resource from namespace.yaml",
			args: args{
				ctx:              context.Background(),
				templateDir:      "testdata",
				templateFileName: "namespace.yaml",
				project: models.Project{
					Name:       "my-project",
					Identifier: "my-project",
				},
				domain: runtimeInterfaces.Domain{
					ID:   "dev",
					Name: "dev",
				},
				namespace:            "my-project-dev",
				templateValues:       templateValuesType{},
				customTemplateValues: templateValuesType{},
			},
			wantK8sManifest: `apiVersion: v1
kind: Namespace
metadata:
  name: my-project-dev
spec:
  finalizers:
  - kubernetes
`,
			wantErr: false,
		},
		{
			name: "test create resource from docker.yaml",
			args: args{
				ctx:              context.Background(),
				templateDir:      "testdata",
				templateFileName: "docker.yaml",
				project: models.Project{
					Name:       "my-project",
					Identifier: "my-project",
				},
				domain: runtimeInterfaces.Domain{
					ID:   "dev",
					Name: "dev",
				},
				namespace: "my-project-dev",
				templateValues: templateValuesType{
					"{{ dockerSecret }}": "myDockerSecret",
					"{{ dockerAuth }}":   "myDockerAuth",
				},
				customTemplateValues: templateValuesType{},
			},
			wantK8sManifest: `apiVersion: v1
kind: Secret
metadata:
  name: dockerhub
  namespace: my-project-dev
stringData:
  .dockerconfigjson: '{"auths":{"docker.io":{"username":"mydockerusername","password":"myDockerSecret","email":"none","auth":" myDockerAuth"}}}'
type: kubernetes.io/dockerconfigjson
`,
			wantErr: false,
		},
		{
			name: "test create resource from imagepullsecrets.yaml",
			args: args{
				ctx:              context.Background(),
				templateDir:      "testdata",
				templateFileName: "imagepullsecrets.yaml",
				project: models.Project{
					Name:       "my-project",
					Identifier: "my-project",
				},
				domain: runtimeInterfaces.Domain{
					ID:   "dev",
					Name: "dev",
				},
				namespace:            "my-project-dev",
				templateValues:       templateValuesType{},
				customTemplateValues: templateValuesType{},
			},
			wantK8sManifest: `apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: my-project-dev
imagePullSecrets:
- name: dockerhub
`,
			wantErr: false,
		},
		{
			name: "test create resource from gsa.yaml",
			args: args{
				ctx:              context.Background(),
				templateDir:      "testdata",
				templateFileName: "gsa.yaml",
				project: models.Project{
					Name:       "my-project",
					Identifier: "my-project",
				},
				domain: runtimeInterfaces.Domain{
					ID:   "dev",
					Name: "dev",
				},
				namespace:            "my-project-dev",
				templateValues:       templateValuesType{},
				customTemplateValues: templateValuesType{},
			},
			wantK8sManifest: `apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMServiceAccount
metadata:
  name: my-project-dev-gsa
  namespace: my-project-dev
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository := repositoryMocks.NewMockRepository()
			mockCluster := &mocks.MockCluster{}
			mockPromScope := mockScope.NewTestScope()
			c := NewClusterResourceController(mockRepository, mockCluster, mockPromScope)
			testController := c.(*controller)

			gotK8sManifest, err := testController.createResourceFromTemplate(tt.args.ctx, tt.args.templateDir, tt.args.templateFileName, tt.args.project, tt.args.domain, tt.args.namespace, tt.args.templateValues, tt.args.customTemplateValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("createResourceFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantK8sManifest, gotK8sManifest)
		})
	}
}

func Test_controller_createPatch(t *testing.T) {
	type args struct {
		gvk       schema.GroupVersionKind
		current   []byte
		modified  []byte
		namespace string
	}

	tests := []struct {
		name          string
		args          args
		want          []byte
		wantPatchType types.PatchType
		wantErr       bool
	}{
		{
			name: "no modification for native resource",
			args: args{
				gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
				current: []byte(`{
									"apiVersion": "v1",
									"kind": "Namespace",
									"metadata": {
										"creationTimestamp": "2021-10-04T09:36:08Z",
										"name": "flyteexamples",
										"resourceVersion": "15127693",
										"selfLink": "/api/v1/namespaces/flyteexamples",
										"uid": "6341347f-af77-498b-ad52-5e70713d91b2"
									},
									"spec": {
										"finalizers": [
											"kubernetes"
										]
									},
									"status": {
										"phase": "Active"
									}
								}
								`),
				modified: []byte(`{
								"apiVersion": "v1",
								"kind": "Namespace",
								"metadata": {
									"name": "flyteexamples"
								}
							}`),
				namespace: "flyteexamples",
			},
			want:          []byte("{}"),
			wantPatchType: types.StrategicMergePatchType,
			wantErr:       false,
		},
		{
			name: "no modification for custom resource",
			args: args{
				gvk: schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
				current: []byte(`{
						"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
						"kind": "IAMServiceAccount",
						"metadata": {
							"annotations": {
								"cnrm.cloud.google.com/management-conflict-prevention-policy": "none",
								"cnrm.cloud.google.com/project-id": "my-project",
								"cnrm.cloud.google.com/state-into-spec": "merge"
							},
							"creationTimestamp": "2021-11-23T14:48:03Z",
							"generation": 1,
							"name": "my-project-gsa",
							"namespace": "my-project",
							"resourceVersion": "54948113",
							"selfLink": "/apis/iam.cnrm.cloud.google.com/v1beta1/namespaces/my-project/iamserviceaccounts/my-project-gsa",
							"uid": "ec76cccc-6b23-4638-b2ee-7cd886cbc031"
						},
						"status": {
							"conditions": [
								{
									"lastTransitionTime": "2021-11-23T14:48:03Z",
									"status": "True",
									"type": "Ready"
								}
							],
							"observedGeneration": 1
						}
					}`),
				modified: []byte(`{
								"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
								"kind": "IAMServiceAccount",
								"metadata": {
									"name": "my-project-gsa",
									"namespace": "my-project",
									"annotations": {
										"cnrm.cloud.google.com/project-id": "my-project"
									}
								}
							}`),
				namespace: "my-project",
			},
			want:          []byte("{}"),
			wantPatchType: types.MergePatchType,
			wantErr:       false,
		},
		{
			name: "patch namespace",
			args: args{
				gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
				current: []byte(`{
									"apiVersion": "v1",
									"kind": "Namespace",
									"metadata": {
										"creationTimestamp": "2021-10-04T09:36:08Z",
										"name": "flyteexamples",
										"resourceVersion": "15127693",
										"selfLink": "/api/v1/namespaces/flyteexamples",
										"uid": "6341347f-af77-498b-ad52-5e70713d91b2"
									},
									"spec": {
										"finalizers": [
											"kubernetes"
										]
									},
									"status": {
										"phase": "Active"
									}
								}
								`),
				modified: []byte(`{
								"apiVersion": "v1",
								"kind": "Namespace",
								"metadata": {
									"name": "flyteexamples",
									"annotations": {
										"cnrm.cloud.google.com/project-id": "my-project"
									}
								}
							}`),
				namespace: "flyteexamples",
			},
			want:          []byte(`{"metadata":{"annotations":{"cnrm.cloud.google.com/project-id":"my-project"},"resourceVersion":"15127693"}}`),
			wantPatchType: types.StrategicMergePatchType,
			wantErr:       false,
		},
		{
			name: "patch resource quota",
			args: args{
				gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ResourceQuota"},
				current: []byte(`{
								"apiVersion": "v1",
								"kind": "ResourceQuota",
								"metadata": {
									"creationTimestamp": "2021-10-04T09:36:08Z",
									"name": "project-quota",
									"namespace": "flyteexamples",
									"resourceVersion": "54942626",
									"selfLink": "/api/v1/namespaces/flyteexamples/resourcequotas/project-quota",
									"uid": "a2776675-8b00-423f-aa13-20fd80b87e7c"
								},
								"spec": {
									"hard": {
										"limits.cpu": "32",
										"limits.memory": "64Gi"
									}
								},
								"status": {
									"hard": {
										"limits.cpu": "32",
										"limits.memory": "64Gi"
									},
									"used": {
										"limits.cpu": "0",
										"limits.memory": "0"
									}
								}
							}
							`),
				modified: []byte(`{
								"apiVersion": "v1",
								"kind": "ResourceQuota",
								"spec": {
									"hard": {
										"limits.cpu": "64",
										"limits.memory": "128Gi"
									}
								}
							}`),
				namespace: "flyteexamples",
			},
			want:          []byte(`{"metadata":{"resourceVersion":"54942626"},"spec":{"hard":{"limits.cpu":"64","limits.memory":"128Gi"}}}`),
			wantPatchType: types.StrategicMergePatchType,
			wantErr:       false,
		},
		{
			name: "patch service account",
			args: args{
				gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
				current: []byte(`{
							"apiVersion": "v1",
							"kind": "ServiceAccount",
							"metadata": {
								"creationTimestamp": "2021-09-26T11:05:03Z",
								"name": "default",
								"namespace": "flyte",
								"resourceVersion": "8902329",
								"selfLink": "/api/v1/namespaces/flyte/serviceaccounts/default",
								"uid": "87c18ff7-f16c-4af3-b08a-8419134f2674"
							},
							"secrets": [
								{
									"name": "default-token-lgqjs"
								}
							]
						}
					`),
				modified: []byte(`{
								"apiVersion": "v1",
								"kind": "ServiceAccount",
								"metadata": {
									"name": "default",
									"namespace": "flyte",
									"annotations": {
										"cnrm.cloud.google.com/project-id": "my-project"
									}
								}
							}`),
				namespace: "flyteexamples",
			},
			want:          []byte(`{"metadata":{"annotations":{"cnrm.cloud.google.com/project-id":"my-project"},"resourceVersion":"8902329"}}`),
			wantPatchType: types.StrategicMergePatchType,
			wantErr:       false,
		},
		{
			name: "patch custom resource",
			args: args{
				gvk: schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
				current: []byte(`{
						"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
						"kind": "IAMServiceAccount",
						"metadata": {
							"annotations": {
								"cnrm.cloud.google.com/management-conflict-prevention-policy": "none",
								"cnrm.cloud.google.com/project-id": "my-project",
								"cnrm.cloud.google.com/state-into-spec": "merge"
							},
							"creationTimestamp": "2021-11-23T14:48:03Z",
							"generation": 1,
							"name": "my-project-gsa",
							"namespace": "my-project",
							"resourceVersion": "54948113",
							"selfLink": "/apis/iam.cnrm.cloud.google.com/v1beta1/namespaces/my-project/iamserviceaccounts/my-project-gsa",
							"uid": "ec76cccc-6b23-4638-b2ee-7cd886cbc031"
						},
						"status": {
							"conditions": [
								{
									"lastTransitionTime": "2021-11-23T14:48:03Z",
									"status": "True",
									"type": "Ready"
								}
							],
							"observedGeneration": 1
						}
					}`),
				modified: []byte(`{
								"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
								"kind": "IAMServiceAccount",
								"metadata": {
									"name": "my-project-gsa",
									"namespace": "my-project",
									"annotations": {
										"cnrm.cloud.google.com/project-id": "my-new-project"
									}
								}
							}`),
				namespace: "my-project",
			},
			want:          []byte(`{"metadata":{"annotations":{"cnrm.cloud.google.com/project-id":"my-new-project"},"resourceVersion":"54948113"}}`),
			wantPatchType: types.MergePatchType,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentObj, _, err := unstructured.UnstructuredJSONScheme.Decode(tt.args.current, nil, nil)
			assert.NoError(t, err)

			c := &controller{}
			got, gotPatchType, err := c.createPatch(tt.args.gvk, currentObj.(*unstructured.Unstructured), tt.args.modified, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, string(tt.want), string(got))
			assert.Equal(t, string(tt.wantPatchType), string(gotPatchType))
		})
	}
}
