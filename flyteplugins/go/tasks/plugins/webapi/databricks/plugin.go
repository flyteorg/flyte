package databricks

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	pluginErrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	create          string = "create"
	get             string = "get"
	cancel          string = "cancel"
	databricksAPI   string = "/api/2.1/jobs/runs"
	newCluster      string = "new_cluster"
	dockerImage     string = "docker_image"
	sparkConfig     string = "spark_conf"
	sparkPythonTask string = "spark_python_task"
	pythonFile      string = "python_file"
	parameters      string = "parameters"
	url             string = "url"
)

// HTTPClient for mocking/testing purposes, and we'll override this method
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
	client      HTTPClient
}

type ResourceWrapper struct {
	StatusCode     int
	LifeCycleState string
	ResultState    string
	JobID          string
	Message        string
}

type ResourceMetaWrapper struct {
	RunID              string
	DatabricksInstance string
	Token              string
}

func (p Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	token, err := taskCtx.SecretManager().Get(ctx, p.cfg.TokenKey)
	if err != nil {
		return nil, nil, err
	}

	container := taskTemplate.GetContainer()
	sparkJob := plugins.SparkJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sparkJob)
	if err != nil {
		return nil, nil, errors.Wrapf(pluginErrors.BadTaskSpecification, err, "invalid TaskSpecification [%v], failed to unmarshal", taskTemplate.GetCustom())
	}

	// override the default token in propeller
	if len(sparkJob.DatabricksToken) != 0 {
		token = sparkJob.DatabricksToken
	}
	modifiedArgs, err := template.Render(ctx, container.GetArgs(), template.Parameters{
		TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
		Inputs:           taskCtx.InputReader(),
		OutputPath:       taskCtx.OutputWriter(),
		Task:             taskCtx.TaskReader(),
	})
	if err != nil {
		return nil, nil, err
	}

	databricksJob := make(map[string]interface{})
	err = utils.UnmarshalStructToObj(sparkJob.DatabricksConf, &databricksJob)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal databricksJob: %v: %v", sparkJob.DatabricksConf, err)
	}

	// If "existing_cluster_id" is in databricks_job, then we don't need to set "new_cluster"
	// Refer the docs here: https://docs.databricks.com/en/workflows/jobs/jobs-2.0-api.html#request-structure
	if clusterConfig, ok := databricksJob[newCluster].(map[string]interface{}); ok {
		if dockerConfig, ok := clusterConfig[dockerImage].(map[string]interface{}); !ok || dockerConfig[url] == nil {
			clusterConfig[dockerImage] = map[string]string{url: container.Image}
		}

		if clusterConfig[sparkConfig] == nil && len(sparkJob.SparkConf) != 0 {
			clusterConfig[sparkConfig] = sparkJob.SparkConf
		}
	}
	databricksJob[sparkPythonTask] = map[string]interface{}{pythonFile: p.cfg.EntrypointFile, parameters: modifiedArgs}

	data, err := p.sendRequest(create, databricksJob, token, "")
	if err != nil {
		return nil, nil, err
	}

	if _, ok := data["run_id"]; !ok {
		return nil, nil, errors.Errorf("CorruptedPluginState", "can't get the run_id")
	}
	runID := fmt.Sprintf("%.0f", data["run_id"])

	return ResourceMetaWrapper{runID, p.cfg.DatabricksInstance, token}, nil, nil
}

func (p Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	exec := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	res, err := p.sendRequest(get, nil, exec.Token, exec.RunID)
	if err != nil {
		return nil, err
	}
	if _, ok := res["state"]; !ok {
		return nil, errors.Errorf("CorruptedPluginState", "can't get the job state")
	}
	jobState := res["state"].(map[string]interface{})
	jobID := fmt.Sprintf("%.0f", res["job_id"])
	message := fmt.Sprintf("%s", jobState["state_message"])
	lifeCycleState := fmt.Sprintf("%s", jobState["life_cycle_state"])
	var resultState string
	if _, ok := jobState["result_state"]; !ok {
		// The result_state is not available until the job is finished.
		// https://docs.databricks.com/en/workflows/jobs/jobs-2.0-api.html#runresultstate
		resultState = ""
	} else {
		resultState = fmt.Sprintf("%s", jobState["result_state"])
	}
	return ResourceWrapper{
		JobID:          jobID,
		LifeCycleState: lifeCycleState,
		ResultState:    resultState,
		Message:        message,
	}, nil
}

func (p Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	exec := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	_, err := p.sendRequest(cancel, nil, exec.Token, exec.RunID)
	if err != nil {
		return err
	}
	logger.Info(ctx, "Deleted Databricks job execution.")

	return nil
}

func (p Plugin) sendRequest(method string, databricksJob map[string]interface{}, token string, runID string) (map[string]interface{}, error) {
	var databricksURL string
	// for mocking/testing purposes
	if p.cfg.databricksEndpoint == "" {
		databricksURL = fmt.Sprintf("https://%v%v", p.cfg.DatabricksInstance, databricksAPI)
	} else {
		databricksURL = fmt.Sprintf("%v%v", p.cfg.databricksEndpoint, databricksAPI)
	}

	// build the request spec
	var body io.Reader
	var httpMethod string
	switch method {
	case create:
		databricksURL += "/submit"
		mJSON, err := json.Marshal(databricksJob)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the job request: %v", err)
		}
		body = bytes.NewBuffer(mJSON)
		httpMethod = http.MethodPost
	case get:
		databricksURL += "/get?run_id=" + runID
		httpMethod = http.MethodGet
	case cancel:
		databricksURL += "/cancel"
		body = bytes.NewBuffer([]byte(fmt.Sprintf("{ \"run_id\": %v }", runID)))
		httpMethod = http.MethodPost
	}

	req, err := http.NewRequest(httpMethod, databricksURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Databricks platform with err: [%v]", err)
	}
	defer resp.Body.Close()

	// Parse the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}

	if len(responseBody) != 0 {
		err = json.Unmarshal(responseBody, &data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response with err: [%v]", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		message := ""
		if v, ok := data["message"]; ok {
			message = v.(string)
		}
		return nil, fmt.Errorf("failed to %v Databricks job with error [%v]", method, message)
	}
	return data, nil
}

func (p Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	exec := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	resource := taskCtx.Resource().(ResourceWrapper)
	message := resource.Message
	jobID := resource.JobID
	lifeCycleState := resource.LifeCycleState
	resultState := resource.ResultState

	taskInfo := createTaskInfo(exec.RunID, jobID, exec.DatabricksInstance)
	switch lifeCycleState {
	// Job response format. https://docs.databricks.com/en/workflows/jobs/jobs-2.0-api.html#runlifecyclestate
	case "QUEUED":
		return core.PhaseInfoQueued(time.Now(), core.DefaultPhaseVersion, message), nil
	case "PENDING":
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, message, taskInfo), nil
	case "RUNNING":
		fallthrough
	case "BLOCKED":
		fallthrough
	case "WAITING_FOR_RETRY":
		fallthrough
	case "TERMINATING":
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case "TERMINATED":
		if resultState == "SUCCESS" {
			// Result state details. https://docs.databricks.com/en/workflows/jobs/jobs-2.0-api.html#runresultstate
			if err := writeOutput(ctx, taskCtx); err != nil {
				return core.PhaseInfoFailure(string(rune(http.StatusInternalServerError)), "failed to write output", taskInfo), nil
			}
			return core.PhaseInfoSuccess(taskInfo), nil
		} else if resultState == "FAILED" {
			return core.PhaseInfoRetryableFailure("job failed", message, taskInfo), nil
		}
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, message, taskInfo), nil
	case "SKIPPED":
		return core.PhaseInfoFailure(string(rune(http.StatusConflict)), message, taskInfo), nil
	case "INTERNAL_ERROR":
		return core.PhaseInfoRetryableFailure(string(rune(http.StatusInternalServerError)), message, taskInfo), nil
	}
	return core.PhaseInfoUndefined, pluginErrors.Errorf(pluginsCore.SystemErrorCode, "unknown execution phase [%v].", lifeCycleState)
}

func writeOutput(ctx context.Context, taskCtx webapi.StatusContext) error {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}
	if taskTemplate.Interface == nil || taskTemplate.Interface.Outputs == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Infof(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	outputReader := ioutils.NewRemoteFileOutputReader(ctx, taskCtx.DataStore(), taskCtx.OutputWriter(), 0)
	return taskCtx.OutputWriter().Put(ctx, outputReader)
}

func createTaskInfo(runID, jobID, databricksInstance string) *core.TaskInfo {
	timeNow := time.Now()

	return &core.TaskInfo{
		OccurredAt: &timeNow,
		Logs: []*flyteIdlCore.TaskLog{
			{
				Uri: fmt.Sprintf("https://%s/#job/%s/run/%s",
					databricksInstance,
					jobID,
					runID),
				Name: "Databricks Console",
			},
		},
	}
}

func newDatabricksJobTaskPlugin() webapi.PluginEntry {
	return webapi.PluginEntry{
		ID:                 "databricks",
		SupportedTaskTypes: []core.TaskType{"spark"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         GetConfig(),
				client:      &http.Client{},
			}, nil
		},
	}
}

func init() {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newDatabricksJobTaskPlugin())
}
