package create

import (
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytectl/cmd/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	executionCreateResponse *admin.ExecutionCreateResponse
	relaunchRequest         *admin.ExecutionRelaunchRequest
	recoverRequest          *admin.ExecutionRecoverRequest
)

// This function needs to be called after testutils.Steup()
func createExecutionUtilSetup() {
	executionCreateResponse = &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}
	relaunchRequest = &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    "execName",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	recoverRequest = &admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    "execName",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	executionConfig = &ExecutionConfig{}
}

func TestCreateExecutionForRelaunch(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(executionCreateResponse, nil)
	err := relaunchExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
	assert.Nil(t, err)
}

func TestCreateExecutionForRelaunchNotFound(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(nil, errors.New("unknown execution"))
	err := relaunchExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")

	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}

func TestCreateExecutionForRecovery(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRecoverExecutionMatch(s.Ctx, recoverRequest).Return(executionCreateResponse, nil)
	err := recoverExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
	assert.Nil(t, err)
}

func TestCreateExecutionForRecoveryNotFound(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRecoverExecutionMatch(s.Ctx, recoverRequest).Return(nil, errors.New("unknown execution"))
	err := recoverExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}

func TestCreateExecutionRequestForWorkflow(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		launchPlan := &admin.LaunchPlan{}
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan, nil)
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("successful with envs", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		launchPlan := &admin.LaunchPlan{}
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan, nil)
		var executionConfigWithEnvs = &ExecutionConfig{
			Envs: map[string]string{"foo": "bar"},
		}
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfigWithEnvs, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("successful with empty envs", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		launchPlan := &admin.LaunchPlan{}
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan, nil)
		var executionConfigWithEnvs = &ExecutionConfig{
			Envs: map[string]string{},
		}
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfigWithEnvs, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("failed literal conversion", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				DefaultInputs: &core.ParameterMap{
					Parameters: map[string]*core.Parameter{"nilparam": nil},
				},
			},
		}
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan, nil)
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.NotNil(t, err)
		assert.Nil(t, execCreateRequest)
		assert.Equal(t, fmt.Errorf("parameter [nilparam] has nil Variable"), err)
	})
	t.Run("failed fetch", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.NotNil(t, err)
		assert.Nil(t, execCreateRequest)
		assert.Equal(t, err, errors.New("failed"))
	})
	t.Run("with security context", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		executionConfig.KubeServiceAcct = "default"
		launchPlan := &admin.LaunchPlan{}
		s.FetcherExt.OnFetchLPVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan, nil)
		s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, mock.Anything).Return(launchPlan, nil)
		execCreateRequest, err := createExecutionRequestForWorkflow(s.Ctx, "wfName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
		executionConfig.KubeServiceAcct = ""
	})
}

func TestCreateExecutionRequestForTask(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		task := &admin.Task{
			Id: &core.Identifier{
				Name: "taskName",
			},
		}
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(task, nil)
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("successful with envs", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		task := &admin.Task{
			Id: &core.Identifier{
				Name: "taskName",
			},
		}
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(task, nil)
		var executionConfigWithEnvs = &ExecutionConfig{
			Envs: map[string]string{"foo": "bar"},
		}
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfigWithEnvs, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("successful with empty envs", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		task := &admin.Task{
			Id: &core.Identifier{
				Name: "taskName",
			},
		}
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(task, nil)
		var executionConfigWithEnvs = &ExecutionConfig{
			Envs: map[string]string{},
		}
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfigWithEnvs, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
	})
	t.Run("failed literal conversion", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		task := &admin.Task{
			Closure: &admin.TaskClosure{
				CompiledTask: &core.CompiledTask{
					Template: &core.TaskTemplate{
						Interface: &core.TypedInterface{
							Inputs: &core.VariableMap{
								Variables: map[string]*core.Variable{
									"nilvar": nil,
								},
							},
						},
					},
				},
			},
		}
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(task, nil)
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.NotNil(t, err)
		assert.Nil(t, execCreateRequest)
		assert.Equal(t, fmt.Errorf("variable [nilvar] has nil type"), err)
	})
	t.Run("failed fetch", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.NotNil(t, err)
		assert.Nil(t, execCreateRequest)
		assert.Equal(t, err, errors.New("failed"))
	})
	t.Run("with security context", func(t *testing.T) {
		s := setup()
		createExecutionUtilSetup()
		executionConfig.KubeServiceAcct = "default"
		task := &admin.Task{
			Id: &core.Identifier{
				Name: "taskName",
			},
		}
		s.FetcherExt.OnFetchTaskVersionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(task, nil)
		execCreateRequest, err := createExecutionRequestForTask(s.Ctx, "taskName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
		assert.Nil(t, err)
		assert.NotNil(t, execCreateRequest)
		executionConfig.KubeServiceAcct = ""
	})
}

func Test_resolveOverrides(t *testing.T) {
	executionConfig.KubeServiceAcct = "k8s-acct"
	executionConfig.IamRoleARN = "iam-role"
	executionConfig.TargetProject = "t-proj"
	executionConfig.TargetDomain = "t-domain"
	executionConfig.Version = "v1"
	executionConfig.ClusterPool = "gpu"
	cfg := &ExecutionConfig{}

	resolveOverrides(cfg, "p1", "d1")

	assert.Equal(t, "k8s-acct", cfg.KubeServiceAcct)
	assert.Equal(t, "iam-role", cfg.IamRoleARN)
	assert.Equal(t, "t-proj", cfg.TargetProject)
	assert.Equal(t, "t-domain", cfg.TargetDomain)
	assert.Equal(t, "v1", cfg.Version)
	assert.Equal(t, "gpu", cfg.ClusterPool)
}

func TestCreateExecutionForRelaunchOverwritingCache(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	executionConfig.OverwriteCache = true
	relaunchRequest.OverwriteCache = true // ensure request has overwriteCache param set
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(executionCreateResponse, nil)
	err := relaunchExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx, executionConfig, "")
	assert.Nil(t, err)
}
