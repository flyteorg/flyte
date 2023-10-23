package register

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	ghMocks "github.com/flyteorg/flytectl/pkg/github/mocks"

	"github.com/flyteorg/flyte/flytestdlib/utils"

	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"

	"github.com/google/go-github/v42/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func registerFilesSetup() {
	httpClient = &MockHTTPClient{}
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

	rconfig.DefaultFilesConfig.AssumableIamRole = ""
	rconfig.DefaultFilesConfig.K8sServiceAccount = ""
	rconfig.DefaultFilesConfig.OutputLocationPrefix = ""
	rconfig.DefaultFilesConfig.EnableSchedule = true
}

func TestGetSortedArchivedFileWithParentFolderList(t *testing.T) {
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/valid-parent-folder-register.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/valid-register.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/valid-unordered-register.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/invalid.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, len(fileList), 0)
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedTgzList(t *testing.T) {
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/valid-register.tgz"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	s := setup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/invalid.tgz"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedInvalidArchiveFileList(t *testing.T) {
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"testdata/invalid-extension-register.zip"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("only .tar, .tar.gz and .tgz extension archives are supported"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughInvalidHttpList(t *testing.T) {
	s := setup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"http://invalidhost:invalidport/testdata/valid-register.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughValidHttpList(t *testing.T) {
	s := setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"http://dummyhost:80/testdata/valid-register.tar"}
	fileList, tmpDir, err := GetSerializeOutputFiles(s.Ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args := []string{"http://dummyhost:80/testdata/valid-register.tar"}
	var ctx context.Context = nil
	fileList, tmpDir, err := GetSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("net/http: nil Context"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func Test_getTotalSize(t *testing.T) {
	b := bytes.NewBufferString("hello world")
	size, err := getTotalSize(b)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), size)
}

func TestRegisterFile(t *testing.T) {
	t.Run("Successful run", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		args := []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Nil(t, err)
	})
	t.Run("Failed Scheduled launch plan registration", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		s.MockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		variableMap := map[string]*core.Variable{
			"var1": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var1",
			},
			"var2": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var2 long descriptions probably needs truncate",
			},
		}
		wf := &admin.Workflow{
			Closure: &admin.WorkflowClosure{
				CompiledWorkflow: &core.CompiledWorkflowClosure{
					Primary: &core.CompiledWorkflow{
						Template: &core.WorkflowTemplate{
							Interface: &core.TypedInterface{
								Inputs: &core.VariableMap{
									Variables: variableMap,
								},
							},
						},
					},
				},
			},
		}
		s.FetcherExt.OnFetchWorkflowVersionMatch(s.Ctx, "core.scheduled_workflows.lp_schedules.date_formatter_wf", mock.Anything, "dummyProject", "dummyDomain").Return(wf, nil)
		args := []string{"testdata/152_my_cron_scheduled_lp_3.pb"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.Contains(t, results[0].Info, "param values are missing on scheduled workflow for the following params")
		assert.NotNil(t, err)
	})
	t.Run("Non existent file", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		args := []string{"testdata/non-existent.pb"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.Equal(t, "Error reading file due to open testdata/non-existent.pb: no such file or directory", results[0].Info)
		assert.NotNil(t, err)
	})
	t.Run("unmarhal failure", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		args := []string{"testdata/valid-register.tar"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.True(t, strings.HasPrefix(results[0].Info, "Error unmarshalling file due to failed unmarshalling file testdata/valid-register.tar"))
		assert.NotNil(t, err)
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil,
			status.Error(codes.AlreadyExists, "AlreadyExists"))
		args := []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Success", results[0].Status)
		assert.Equal(t, "AlreadyExists", results[0].Info)
		assert.Nil(t, err)
	})
	t.Run("Registration Error", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		s.MockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil,
			status.Error(codes.InvalidArgument, "Invalid"))
		args := []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(s.Ctx, args[0], registerResults, s.CmdCtx, "", *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.Equal(t, "Error registering file due to rpc error: code = InvalidArgument desc = Invalid", results[0].Info)
		assert.NotNil(t, err)
	})
}

func TestHydrateLaunchPlanSpec(t *testing.T) {
	t.Run("IamRole override", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.AssumableIamRole = "iamRole"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.AuthRole{AssumableIamRole: "iamRole"}, lpSpec.AuthRole)
		assert.Equal(t, &core.SecurityContext{RunAs: &core.Identity{IamRole: "iamRole"}}, lpSpec.SecurityContext)
	})
	t.Run("k8sService account override", func(t *testing.T) {
		registerFilesSetup()
		rconfig.DefaultFilesConfig.K8sServiceAccount = "k8Account"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.AuthRole{KubernetesServiceAccount: "k8Account"}, lpSpec.AuthRole)
		assert.Equal(t, &core.SecurityContext{RunAs: &core.Identity{K8SServiceAccount: "k8Account"}}, lpSpec.SecurityContext)
	})
	t.Run("Both k8sService and IamRole", func(t *testing.T) {
		registerFilesSetup()
		rconfig.DefaultFilesConfig.AssumableIamRole = "iamRole"
		rconfig.DefaultFilesConfig.K8sServiceAccount = "k8Account"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.AuthRole{AssumableIamRole: "iamRole",
			KubernetesServiceAccount: "k8Account"}, lpSpec.AuthRole)
		assert.Equal(t, &core.SecurityContext{RunAs: &core.Identity{IamRole: "iamRole", K8SServiceAccount: "k8Account"}}, lpSpec.SecurityContext)
	})
	t.Run("Output prefix", func(t *testing.T) {
		registerFilesSetup()
		rconfig.DefaultFilesConfig.OutputLocationPrefix = "prefix"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.RawOutputDataConfig{OutputLocationPrefix: "prefix"}, lpSpec.RawOutputDataConfig)
	})
}

func TestUploadFastRegisterArtifact(t *testing.T) {
	t.Run("Successful upload", func(t *testing.T) {
		s := setup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = store
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(s.Ctx, &service.CreateUploadLocationRequest{
			Project:    "flytesnacks",
			Domain:     "development",
			Filename:   "flytesnacks-core.tgz",
			ContentMd5: []uint8{0x19, 0x72, 0x39, 0xcd, 0x85, 0x2d, 0xf1, 0x79, 0x8f, 0x6b, 0x3, 0xb3, 0xa9, 0x6c, 0xec, 0xa0},
		}).Return(&service.CreateUploadLocationResponse{}, nil)
		_, err = uploadFastRegisterArtifact(s.Ctx, "flytesnacks", "development", "testdata/flytesnacks-core.tgz", "", s.MockClient.DataProxyClient(), rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath)
		assert.Nil(t, err)
	})
	t.Run("Failed upload", func(t *testing.T) {
		s := setup()
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		store, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = store
		s.MockClient.DataProxyClient().(*mocks.DataProxyServiceClient).OnCreateUploadLocationMatch(s.Ctx, &service.CreateUploadLocationRequest{
			Project:    "flytesnacks",
			Domain:     "development",
			Filename:   "flytesnacks-core.tgz",
			ContentMd5: []uint8{0x19, 0x72, 0x39, 0xcd, 0x85, 0x2d, 0xf1, 0x79, 0x8f, 0x6b, 0x3, 0xb3, 0xa9, 0x6c, 0xec, 0xa0},
		}).Return(&service.CreateUploadLocationResponse{}, nil)
		_, err = uploadFastRegisterArtifact(context.Background(), "flytesnacks", "development", "testdata/flytesnacks-core.tgz", "", s.MockClient.DataProxyClient(), rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath)
		assert.Nil(t, err)
	})
	t.Run("Failed upload", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = s
		_, err = uploadFastRegisterArtifact(context.Background(), "flytesnacks", "development", "testdata/flytesnacksre.tgz", "", nil, rconfig.DefaultFilesConfig.DeprecatedSourceUploadPath)
		assert.NotNil(t, err)
	})
}

func TestGetStorageClient(t *testing.T) {
	t.Run("Failed to create storage client", func(t *testing.T) {
		Client = nil
		s, err := getStorageClient(context.Background())
		assert.NotNil(t, err)
		assert.Nil(t, s)
	})
}

func TestGetAllFlytesnacksExample(t *testing.T) {
	t.Run("Failed to get manifest with wrong name", func(t *testing.T) {
		mockGh := &ghMocks.GHRepoService{}
		mockGh.OnGetLatestReleaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failed"))
		_, _, err := getAllExample("no////ne", "", mockGh)
		assert.NotNil(t, err)
	})
	t.Run("Failed to get release", func(t *testing.T) {
		mockGh := &ghMocks.GHRepoService{}
		tag := "v0.15.0"
		sandboxManifest := "flyte_sandbox_manifest.tgz"
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
			TagName: &tag,
			Assets: []*github.ReleaseAsset{{
				Name: &sandboxManifest,
			}},
		}, nil, fmt.Errorf("failed"))
		_, _, err := getAllExample("homebrew-tap", "1.0", mockGh)
		assert.NotNil(t, err)
	})
	t.Run("Successfully get examples", func(t *testing.T) {
		mockGh := &ghMocks.GHRepoService{}
		tag := "v0.15.0"
		sandboxManifest := "flyte_sandbox_manifest.tgz"
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, tag).Return(&github.RepositoryRelease{
			TagName: &tag,
			Assets: []*github.ReleaseAsset{{
				Name: &sandboxManifest,
			}},
		}, nil, nil)
		assets, r, err := getAllExample("flytesnacks", tag, mockGh)
		assert.Nil(t, err)
		assert.Greater(t, len(*r.TagName), 0)
		assert.Greater(t, len(assets), 0)
	})
}

func TestRegister(t *testing.T) {
	t.Run("Failed to register", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		node := &admin.NodeExecution{}
		err := register(s.Ctx, node, s.CmdCtx, rconfig.DefaultFilesConfig.DryRun, rconfig.DefaultFilesConfig.EnableSchedule)
		assert.NotNil(t, err)
	})
}

func TestHydrateNode(t *testing.T) {
	t.Run("Failed hydrate node", func(t *testing.T) {
		registerFilesSetup()
		node := &core.Node{}
		err := hydrateNode(node, rconfig.DefaultFilesConfig.Version, true)
		assert.NotNil(t, err)
	})

	t.Run("hydrateSpec with wrong type", func(t *testing.T) {
		registerFilesSetup()
		task := &admin.Task{}
		err := hydrateSpec(task, "", *rconfig.DefaultFilesConfig)
		assert.NotNil(t, err)
	})
}

func TestHydrateArrayNode(t *testing.T) {
	registerFilesSetup()
	node := &core.Node{
		Target: &core.Node_ArrayNode{
			ArrayNode: &core.ArrayNode{
				Node: &core.Node{
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{
									ResourceType: core.ResourceType_TASK,
									Project:      "flytesnacks",
									Domain:       "development",
									Name:         "n1",
									Version:      "v1",
								},
							},
						},
					},
				},
			},
		},
	}
	err := hydrateNode(node, rconfig.DefaultFilesConfig.Version, true)
	assert.Nil(t, err)
}

func TestHydrateGateNode(t *testing.T) {
	t.Run("Hydrate Sleep", func(t *testing.T) {
		registerFilesSetup()
		// Write a node that contains a GateNode
		node := &core.Node{
			Target: &core.Node_GateNode{
				GateNode: &core.GateNode{
					Condition: &core.GateNode_Sleep{
						Sleep: &core.SleepCondition{
							Duration: &durationpb.Duration{
								Seconds: 10,
							},
						},
					},
				},
			},
		}
		err := hydrateNode(node, rconfig.DefaultFilesConfig.Version, true)
		assert.Nil(t, err)
	})

	t.Run("Hydrate Signal", func(t *testing.T) {
		registerFilesSetup()
		// Write a node that contains a GateNode
		node := &core.Node{
			Target: &core.Node_GateNode{
				GateNode: &core.GateNode{
					Condition: &core.GateNode_Signal{
						Signal: &core.SignalCondition{
							SignalId: "abc",
						},
					},
				},
			},
		}
		err := hydrateNode(node, rconfig.DefaultFilesConfig.Version, true)
		assert.Nil(t, err)
	})

	t.Run("Hydrate Approve", func(t *testing.T) {
		registerFilesSetup()
		// Write a node that contains a GateNode
		node := &core.Node{
			Target: &core.Node_GateNode{
				GateNode: &core.GateNode{
					Condition: &core.GateNode_Approve{
						Approve: &core.ApproveCondition{
							SignalId: "abc",
						},
					},
				},
			},
		}
		err := hydrateNode(node, rconfig.DefaultFilesConfig.Version, true)
		assert.Nil(t, err)
	})
}

func TestHydrateTaskSpec(t *testing.T) {
	testScope := promutils.NewTestScope()
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
	s, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, testScope.NewSubScope("flytectl"))
	assert.Nil(t, err)
	Client = s

	metadata := &core.K8SObjectMetadata{
		Labels: map[string]string{
			"l": "a",
		},
		Annotations: map[string]string{
			"a": "b",
		},
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Args: []string{"foo", "bar"},
			},
			{
				Args: []string{"baz", registrationRemotePackagePattern},
			},
		},
	}
	podSpecStruct, err := utils.MarshalObjToStruct(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	task := &admin.TaskSpec{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_K8SPod{
				K8SPod: &core.K8SPod{
					Metadata: metadata,
					PodSpec:  podSpecStruct,
				},
			},
		},
	}
	err = hydrateTaskSpec(task, storage.DataReference("file://somewhere"), "sourcey")
	assert.NoError(t, err)
	var hydratedPodSpec = v1.PodSpec{}
	err = utils.UnmarshalStructToObj(task.Template.GetK8SPod().PodSpec, &hydratedPodSpec)
	assert.NoError(t, err)
	assert.Len(t, hydratedPodSpec.Containers[1].Args, 2)
	assert.Contains(t, hydratedPodSpec.Containers[1].Args[1], "somewhere")
}

func TestLeftDiff(t *testing.T) {
	t.Run("empty slices", func(t *testing.T) {
		c := leftDiff(nil, nil)
		assert.Empty(t, c)
	})
	t.Run("right empty slice", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		c := leftDiff(a, nil)
		sort.Strings(a)
		sort.Strings(c)
		assert.Equal(t, a, c)
	})
	t.Run("non empty slices without intersection", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		b := []string{"5", "6", "7"}
		c := leftDiff(a, b)
		sort.Strings(a)
		sort.Strings(c)
		assert.Equal(t, a, c)
	})
	t.Run("non empty slices with some intersection", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		b := []string{"2", "5", "7"}
		c := leftDiff(a, b)
		expected := []string{"1", "3"}
		sort.Strings(expected)
		sort.Strings(c)
		assert.Equal(t, expected, c)
	})

	t.Run("non empty slices with full intersection same order", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		b := []string{"1", "2", "3"}
		c := leftDiff(a, b)
		var expected []string
		sort.Strings(c)
		assert.Equal(t, expected, c)
	})

	t.Run("non empty slices with full intersection diff order", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		b := []string{"2", "3", "1"}
		c := leftDiff(a, b)
		var expected []string
		sort.Strings(c)
		assert.Equal(t, expected, c)
	})
}

func TestValidateLaunchSpec(t *testing.T) {
	ctx := context.Background()
	t.Run("nil launchplan spec", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		err := validateLaunchSpec(ctx, nil, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("launchplan spec with nil workflow id", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		lpSpec := &admin.LaunchPlanSpec{}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("launchplan spec with empty metadata", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("launchplan spec with metadata and empty schedule", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("validate spec failed to fetch workflow", func(t *testing.T) {
		s := setup()
		registerFilesSetup()

		s.FetcherExt.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "kick_off_time_arg",
				},
			},
		}
		lp := &admin.LaunchPlan{
			Spec: lpSpec,
		}
		err := validateSpec(ctx, lp, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, "failed", err.Error())
	})
	t.Run("failed to fetch workflow", func(t *testing.T) {
		s := setup()
		registerFilesSetup()

		s.FetcherExt.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "kick_off_time_arg",
				},
			},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, "failed", err.Error())
	})
	t.Run("launchplan spec missing required param schedule", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		variableMap := map[string]*core.Variable{
			"var1": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var1",
			},
			"var2": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var2 long descriptions probably needs truncate",
			},
		}
		wf := &admin.Workflow{
			Closure: &admin.WorkflowClosure{
				CompiledWorkflow: &core.CompiledWorkflowClosure{
					Primary: &core.CompiledWorkflow{
						Template: &core.WorkflowTemplate{
							Interface: &core.TypedInterface{
								Inputs: &core.VariableMap{
									Variables: variableMap,
								},
							},
						},
					},
				},
			},
		}
		s.FetcherExt.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "kick_off_time_arg",
				},
			},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "param values are missing on scheduled workflow for the following params")
	})
	t.Run("launchplan spec non empty schedule default param success", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		variableMap := map[string]*core.Variable{
			"var1": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var1",
			},
			"var2": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var2 long descriptions probably needs truncate",
			},
		}
		wf := &admin.Workflow{
			Closure: &admin.WorkflowClosure{
				CompiledWorkflow: &core.CompiledWorkflowClosure{
					Primary: &core.CompiledWorkflow{
						Template: &core.WorkflowTemplate{
							Interface: &core.TypedInterface{
								Inputs: &core.VariableMap{
									Variables: variableMap,
								},
							},
						},
					},
				},
			},
		}
		s.FetcherExt.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "kick_off_time_arg",
				},
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"var1": {
						Var: &core.Variable{
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
						Behavior: &core.Parameter_Default{
							Default: &core.Literal{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_Primitive{
											Primitive: &core.Primitive{
												Value: &core.Primitive_Integer{
													Integer: 10,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			FixedInputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"var2": {
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{
										Value: &core.Primitive_Integer{
											Integer: 10,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.Nil(t, err)
	})

	t.Run("launchplan spec non empty schedule required param without value fail", func(t *testing.T) {
		s := setup()
		registerFilesSetup()
		variableMap := map[string]*core.Variable{
			"var1": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "var1",
			},
		}
		wf := &admin.Workflow{
			Closure: &admin.WorkflowClosure{
				CompiledWorkflow: &core.CompiledWorkflowClosure{
					Primary: &core.CompiledWorkflow{
						Template: &core.WorkflowTemplate{
							Interface: &core.TypedInterface{
								Inputs: &core.VariableMap{
									Variables: variableMap,
								},
							},
						},
					},
				},
			},
		}
		s.FetcherExt.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		lpSpec := &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "projectValue",
				Domain:  "domainValue",
				Name:    "workflowNameValue",
				Version: "workflowVersionValue",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "kick_off_time_arg",
				},
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"var1": {
						Var: &core.Variable{
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		}
		err := validateLaunchSpec(ctx, lpSpec, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("param values are missing on scheduled workflow for the following params [var1]. Either specify them having a default or fixed value"), err)
	})
}
