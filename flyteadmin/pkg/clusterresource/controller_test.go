package clusterresource

import (
	"context"
	"crypto/md5" // #nosec
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	pluginMocks "github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	execClusterMocks "github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/mocks"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	configurationMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

const org = "org-baz"
const proj = "project-foo"
const domain = "domain-bar"

var testScope = mockScope.NewTestScope()

func TestTemplateAlreadyApplied(t *testing.T) {
	ctx := context.TODO()
	const namespace = "namespace"
	const fileName = "fileName"
	testController := controller{
		metrics: newMetrics(testScope),
	}
	checksum1 := md5.Sum([]byte("template1")) // #nosec
	checksum2 := md5.Sum([]byte("template2")) // #nosec
	assert.False(t, testController.templateAlreadyApplied(ctx, namespace, fileName, checksum1))

	testController.appliedNamespaceTemplateChecksums = sync.Map{}
	testController.setTemplateChecksum(ctx, namespace, fileName, checksum1)
	assert.True(t, testController.templateAlreadyApplied(ctx, namespace, fileName, checksum1))
	assert.False(t, testController.templateAlreadyApplied(ctx, namespace, fileName, checksum2))

	testController.setTemplateChecksum(ctx, namespace, fileName, checksum2)
	assert.True(t, testController.templateAlreadyApplied(ctx, namespace, fileName, checksum2))
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
	adminDataProvider := mocks.FlyteAdminDataProvider{}
	adminDataProvider.OnGetClusterResourceAttributesMatch(mock.Anything, org, proj, domain).Return(&admin.ClusterResourceAttributes{
		Attributes: map[string]string{
			"var1": "val1",
			"var2": "val2",
		},
	}, nil)

	testController := controller{
		adminDataProvider: &adminDataProvider,
	}
	domainTemplateValues := templateValuesType{
		"{{ var1 }}": "i'm getting overwritten",
		"{{ var3 }}": "persist",
	}
	customTemplateValues, err := testController.getCustomTemplateValues(context.Background(), org, proj, domain, domainTemplateValues)
	assert.Nil(t, err)
	assert.EqualValues(t, templateValuesType{
		"{{ var1 }}": "val1",
		"{{ var2 }}": "val2",
		"{{ var3 }}": "persist",
	}, customTemplateValues)

	assert.NotEqual(t, domainTemplateValues, customTemplateValues)
}

func TestGetCustomTemplateValues_NothingToOverride(t *testing.T) {
	adminDataProvider := mocks.FlyteAdminDataProvider{}
	adminDataProvider.OnGetClusterResourceAttributesMatch(mock.Anything, org, proj, domain).Return(
		nil, errors.NewFlyteAdminError(codes.NotFound, "foo"))
	testController := controller{
		adminDataProvider: &adminDataProvider,
	}
	customTemplateValues, err := testController.getCustomTemplateValues(context.Background(), org, proj, domain, templateValuesType{
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

func Test_controller_createResourceFromTemplate(t *testing.T) {
	type args struct {
		ctx                  context.Context
		templateDir          string
		templateFileName     string
		project              *admin.Project
		domain               *admin.Domain
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
				project: &admin.Project{
					Name: "my-project",
					Id:   "my-project",
				},
				domain: &admin.Domain{
					Id:   "dev",
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
				project: &admin.Project{
					Name: "my-project",
					Id:   "my-project",
				},
				domain: &admin.Domain{
					Id:   "dev",
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
				project: &admin.Project{
					Name: "my-project",
					Id:   "my-project",
				},
				domain: &admin.Domain{
					Id:   "dev",
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
				project: &admin.Project{
					Name: "my-project",
					Id:   "my-project",
				},
				domain: &admin.Domain{
					Id:   "dev",
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
		{
			name: "test create resource from templatized imagepullsecrets.yaml",
			args: args{
				ctx:              context.Background(),
				templateDir:      "testdata",
				templateFileName: "imagepullsecrets_templatized.yaml",
				project: &admin.Project{
					Name: "my-project",
					Id:   "my-project",
				},
				domain: &admin.Domain{
					Id:   "dev",
					Name: "dev",
				},
				namespace: "my-project-dev",
				templateValues: templateValuesType{
					"{{ imagePullSecretsName }}": "default,",
				},
				customTemplateValues: templateValuesType{
					"{{ imagePullSecretsName }}": "custom",
				},
			},
			wantK8sManifest: `apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: my-project-dev
imagePullSecrets:
  - name: custom
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminDataProvider := mocks.FlyteAdminDataProvider{}
			adminDataProvider.OnGetClusterResourceAttributesMatch(mock.Anything, mock.Anything, mock.Anything).Return(&admin.ClusterResourceAttributes{}, nil)
			mockPromScope := mockScope.NewTestScope()

			c := NewClusterResourceController(&adminDataProvider, &execClusterMocks.ListTargetsInterface{}, mockPromScope, plugins.NewRegistry())
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

// This test ensures the order of processing projects in batch
// If you modify the order, remember to also update other parts of the code that rely on this order,
// specifically the orgs collection right after the projects are processed
func Test_Controller_ProcessProjectsInBatch_Order(t *testing.T) {
	ctx := context.TODO()
	mockConfig := configurationMocks.Configuration{}
	mockConfig.OnClusterResourceConfiguration().Return(configurationMocks.MockClusterResourceConfiguration{
		UnionProjectSyncConfig: runtimeInterfaces.UnionProjectSyncConfig{
			BatchSize: 7,
		},
	})
	mockNamespaceMappingConfiguration := &configurationMocks.NamespaceMappingConfiguration{}
	mockNamespaceMappingConfiguration.OnGetNamespaceTemplate().Return("{{ project }}-{{ domain }}")
	mockConfig.OnNamespaceMappingConfiguration().Return(mockNamespaceMappingConfiguration)
	mockController := &controller{
		config:  &mockConfig,
		metrics: newMetrics(mockScope.NewTestScope()),
	}

	var projects []*admin.Project
	for i := 0; i < 100; i++ {
		projects = append(projects, &admin.Project{
			Id:      strconv.Itoa(i),
			Domains: []*admin.Domain{{Id: "0"}, {Id: "1"}},
		})
	}

	processor := func(ctx context.Context, project *admin.Project, domain *admin.Domain, namespace NamespaceName) (ResourceSyncStats, error) {
		projectId, err := strconv.Atoi(project.Id)
		assert.Nil(t, err)
		domainId, err := strconv.Atoi(domain.Id)
		assert.Nil(t, err)
		return ResourceSyncStats{
			Created: projectId*10 + domainId,
		}, nil
	}

	stats := mockController.processProjectsInBatch(ctx, projects, processor)
	for i, stat := range stats {
		projectId := i / 2
		domainId := i % 2
		assert.Equal(t, projectId*10+domainId, stat.Created)
	}
}

func Test_Controller_ReportClusterResourceStateUpdated(t *testing.T) {
	ctx := context.TODO()
	mockConfig := configurationMocks.Configuration{}
	mockConfig.OnClusterResourceConfiguration().Return(configurationMocks.MockClusterResourceConfiguration{
		ClusterName: "test-cluster",
	})
	mockPlugin := pluginMocks.ClusterResourcePlugin{}
	pluginRegistry := plugins.NewRegistry()
	pluginRegistry.RegisterDefault(plugins.PluginIDClusterResource, &mockPlugin)
	mockController := &controller{
		config:                          &mockConfig,
		metrics:                         newMetrics(mockScope.NewTestScope()),
		pluginRegistry:                  pluginRegistry,
		clusterResourceStateReportCache: sync.Map{},
	}

	testOrgs := sets.Set[string]{}
	for i := 0; i < 6; i++ {
		testOrgs.Insert(strconv.Itoa(i))
	}

	// Mock cache
	mockController.clusterResourceStateReportCache.Store("0", plugin.ClusterResourceStateProvisioned)
	mockController.clusterResourceStateReportCache.Store("1", plugin.ClusterResourceStateProvisioned)
	mockController.clusterResourceStateReportCache.Store("2", plugin.ClusterResourceStateDeprovisioned)
	mockController.clusterResourceStateReportCache.Store("3", plugin.ClusterResourceStateDeprovisioned)

	// Mock plugin
	mockPlugin.On("BatchUpdateClusterResourceState", ctx, mock.Anything).Return(plugin.BatchUpdateClusterResourceStateOutput{}, []plugin.BatchUpdateClusterResourceStateError{
		{
			OrgName: "3",
		},
		{
			OrgName: "5",
		},
	}, nil)

	mockController.reportClusterResourceStateUpdated(ctx, testOrgs, plugin.ClusterResourceStateProvisioned)
	expectedCache := map[string]plugin.ClusterResourceState{
		"0": plugin.ClusterResourceStateProvisioned,
		"1": plugin.ClusterResourceStateProvisioned,
		"2": plugin.ClusterResourceStateProvisioned,
		"4": plugin.ClusterResourceStateProvisioned,
	}
	for i := 0; i < 6; i++ {
		testOrg := strconv.Itoa(i)
		if value, ok := expectedCache[testOrg]; ok {
			value2, ok2 := mockController.clusterResourceStateReportCache.Load(testOrg)
			assert.True(t, ok2)
			assert.Equal(t, value, value2)
		} else {
			_, ok2 := mockController.clusterResourceStateReportCache.Load(testOrg)
			assert.False(t, ok2)
		}
	}
}
