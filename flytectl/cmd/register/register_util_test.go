package register

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flyteorg/flytestdlib/utils"

	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	ctx = u.Ctx
	mockAdminClient = u.MockClient
	cmdCtx = cmdCore.NewCommandContext(mockAdminClient, u.MockOutStream)

	rconfig.DefaultFilesConfig.AssumableIamRole = ""
	rconfig.DefaultFilesConfig.K8sServiceAccount = ""
	rconfig.DefaultFilesConfig.OutputLocationPrefix = ""
}

func TestGetSortedArchivedFileWithParentFolderList(t *testing.T) {
	setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/valid-parent-folder-register.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/valid-register.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/valid-unordered-register.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/invalid.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, len(fileList), 0)
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedTgzList(t *testing.T) {
	setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/valid-register.tgz"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/invalid.tgz"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedInvalidArchiveFileList(t *testing.T) {
	setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"testdata/invalid-extension-register.zip"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("only .tar, .tar.gz and .tgz extension archives are supported"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughInvalidHttpList(t *testing.T) {
	setup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"http://invalidhost:invalidport/testdata/valid-register.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, 0, len(fileList))
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestGetSortedArchivedFileThroughValidHttpList(t *testing.T) {
	setup()
	registerFilesSetup()
	rconfig.DefaultFilesConfig.Archive = true
	args = []string{"http://dummyhost:80/testdata/valid-register.tar"}
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
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
	args = []string{"http://dummyhost:80/testdata/valid-register.tar"}
	ctx = nil
	fileList, tmpDir, err := getSerializeOutputFiles(ctx, args, rconfig.DefaultFilesConfig.Archive)
	assert.Equal(t, len(fileList), 0)
	assert.True(t, strings.HasPrefix(tmpDir, "/tmp/register"))
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("net/http: nil Context"), err)
	// Clean up the temp directory.
	assert.Nil(t, os.RemoveAll(tmpDir), "unable to delete temp dir %v", tmpDir)
}

func TestRegisterFile(t *testing.T) {
	t.Run("Successful run", func(t *testing.T) {
		setup()
		registerFilesSetup()
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		args = []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(ctx, args[0], "", registerResults, cmdCtx, *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Nil(t, err)
	})
	t.Run("Non existent file", func(t *testing.T) {
		setup()
		registerFilesSetup()
		args = []string{"testdata/non-existent.pb"}
		var registerResults []Result
		results, err := registerFile(ctx, args[0], "", registerResults, cmdCtx, *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.Equal(t, "Error reading file due to open testdata/non-existent.pb: no such file or directory", results[0].Info)
		assert.NotNil(t, err)
	})
	t.Run("unmarhal failure", func(t *testing.T) {
		setup()
		registerFilesSetup()
		args = []string{"testdata/valid-register.tar"}
		var registerResults []Result
		results, err := registerFile(ctx, args[0], "", registerResults, cmdCtx, *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Failed", results[0].Status)
		assert.Equal(t, "Error unmarshalling file due to failed unmarshalling file testdata/valid-register.tar", results[0].Info)
		assert.NotNil(t, err)
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		setup()
		registerFilesSetup()
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil,
			status.Error(codes.AlreadyExists, "AlreadyExists"))
		args = []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(ctx, args[0], "", registerResults, cmdCtx, *rconfig.DefaultFilesConfig)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Success", results[0].Status)
		assert.Equal(t, "AlreadyExists", results[0].Info)
		assert.Nil(t, err)
	})
	t.Run("Registration Error", func(t *testing.T) {
		setup()
		registerFilesSetup()
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil,
			status.Error(codes.InvalidArgument, "Invalid"))
		args = []string{"testdata/69_core.flyte_basics.lp.greet_1.pb"}
		var registerResults []Result
		results, err := registerFile(ctx, args[0], "", registerResults, cmdCtx, *rconfig.DefaultFilesConfig)
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
	})
	t.Run("k8sService account override", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.K8sServiceAccount = "k8Account"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.AuthRole{KubernetesServiceAccount: "k8Account"}, lpSpec.AuthRole)
	})
	t.Run("Both k8sService and IamRole", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.AssumableIamRole = "iamRole"
		rconfig.DefaultFilesConfig.K8sServiceAccount = "k8Account"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.AuthRole{AssumableIamRole: "iamRole",
			KubernetesServiceAccount: "k8Account"}, lpSpec.AuthRole)
	})
	t.Run("Output prefix", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.OutputLocationPrefix = "prefix"
		lpSpec := &admin.LaunchPlanSpec{}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
		assert.Equal(t, &admin.RawOutputDataConfig{OutputLocationPrefix: "prefix"}, lpSpec.RawOutputDataConfig)
	})
	t.Run("Validation successful", func(t *testing.T) {
		lpSpec := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "foo",
					},
					KickoffTimeInputArg: "kickoff_time_arg",
				},
			},
			FixedInputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{},
			},
		}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
	})
	t.Run("Validation failure", func(t *testing.T) {
		lpSpec := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "expr",
					},
					KickoffTimeInputArg: "kickoff_time_arg",
				},
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"bar": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
					},
				},
			},
			FixedInputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{},
			},
		}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.NotNil(t, err)
	})
	t.Run("Validation failed with fixed inputs empty", func(t *testing.T) {
		lpSpec := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "expr",
					},
					KickoffTimeInputArg: "kickoff_time_arg",
				},
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"bar": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
					},
				},
			},
		}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.NotNil(t, err)
	})
	t.Run("Validation success with fixed", func(t *testing.T) {
		lpSpec := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "expr",
					},
					KickoffTimeInputArg: "kickoff_time_arg",
				},
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"bar": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
					},
				},
			},
			FixedInputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"bar": coreutils.MustMakeLiteral("bar-value"),
				},
			},
		}
		err := hydrateLaunchPlanSpec(rconfig.DefaultFilesConfig.AssumableIamRole, rconfig.DefaultFilesConfig.K8sServiceAccount, rconfig.DefaultFilesConfig.OutputLocationPrefix, lpSpec)
		assert.Nil(t, err)
	})
}

func TestUploadFastRegisterArtifact(t *testing.T) {
	t.Run("Successful upload", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
		s, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, testScope.NewSubScope("flytectl"))
		assert.Nil(t, err)
		Client = s
		err = uploadFastRegisterArtifact(ctx, "testdata/flytesnacks-core.tgz", "flytesnacks-core.tgz", "", &rconfig.DefaultFilesConfig.SourceUploadPath)
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
		err = uploadFastRegisterArtifact(ctx, "testdata/flytesnacks-core.tgz", "", "", &rconfig.DefaultFilesConfig.SourceUploadPath)
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
		err = uploadFastRegisterArtifact(ctx, "testdata/flytesnacksre.tgz", "", "", &rconfig.DefaultFilesConfig.SourceUploadPath)
		assert.NotNil(t, err)
	})
}

func TestGetStorageClient(t *testing.T) {
	t.Run("Failed to create storage client", func(t *testing.T) {
		Client = nil
		s, err := getStorageClient(ctx)
		assert.NotNil(t, err)
		assert.Nil(t, s)
	})
}

func TestGetAllFlytesnacksExample(t *testing.T) {
	t.Run("Failed to get manifest with wrong name", func(t *testing.T) {
		_, tag, err := getAllFlytesnacksExample("no////ne", "no////ne", "")
		assert.NotNil(t, err)
		assert.Equal(t, len(tag), 0)
	})
	t.Run("Failed to get release", func(t *testing.T) {
		_, tag, err := getAllFlytesnacksExample("flyteorg", "homebrew-tap", "")
		assert.NotNil(t, err)
		assert.Equal(t, len(tag), 0)
	})
	t.Run("Successfully get examples", func(t *testing.T) {
		assets, tag, err := getAllFlytesnacksExample("flyteorg", "flytesnacks", "v0.2.175")
		assert.Nil(t, err)
		assert.Greater(t, len(tag), 0)
		assert.Greater(t, len(assets), 0)
	})
}

func TestRegister(t *testing.T) {
	t.Run("Failed to register", func(t *testing.T) {
		setup()
		registerFilesSetup()
		node := &admin.NodeExecution{}
		err := register(ctx, node, cmdCtx, rconfig.DefaultFilesConfig.DryRun, rconfig.DefaultFilesConfig.Version)
		assert.NotNil(t, err)
	})
}

func TestHydrateNode(t *testing.T) {
	t.Run("Failed hydrate node", func(t *testing.T) {
		setup()
		registerFilesSetup()
		node := &core.Node{}
		err := hydrateNode(node, rconfig.DefaultFilesConfig.Version)
		assert.NotNil(t, err)
	})

	t.Run("hydrateSpec with wrong type", func(t *testing.T) {
		setup()
		registerFilesSetup()
		task := &admin.Task{}
		err := hydrateSpec(task, "", *rconfig.DefaultFilesConfig)
		assert.NotNil(t, err)
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
	err = hydrateTaskSpec(task, "sourcey", rconfig.DefaultFilesConfig.SourceUploadPath, rconfig.DefaultFilesConfig.Version)
	assert.NoError(t, err)
	var hydratedPodSpec = v1.PodSpec{}
	err = utils.UnmarshalStructToObj(task.Template.GetK8SPod().PodSpec, &hydratedPodSpec)
	assert.NoError(t, err)
	assert.Len(t, hydratedPodSpec.Containers[1].Args, 2)
	assert.True(t, strings.HasSuffix(hydratedPodSpec.Containers[1].Args[1], "sourcey"))
}
