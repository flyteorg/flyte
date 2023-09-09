package databricks

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
}

var (
	MockDo       func(req *http.Request) (*http.Response, error)
	testInstance = "test-account.cloud.databricks.com"
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return MockDo(req)
}

func TestPlugin(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))

	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
		client:      &MockClient{},
	}
	t.Run("get config", func(t *testing.T) {
		cfg := defaultConfig
		cfg.WebAPI.Caching.Workers = 1
		cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
		err := SetConfig(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg.WebAPI, plugin.GetConfig())
	})
	t.Run("get ResourceRequirements", func(t *testing.T) {
		namespace, constraints, err := plugin.ResourceRequirements(context.TODO(), nil)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.ResourceNamespace("default"), namespace)
		assert.Equal(t, plugin.cfg.ResourceConstraints, constraints)
	})
}

func TestCreateTaskInfo(t *testing.T) {
	t.Run("create task info", func(t *testing.T) {
		taskInfo := createTaskInfo("run-id", "job-id", testInstance)

		assert.Equal(t, 1, len(taskInfo.Logs))
		assert.Equal(t, taskInfo.Logs[0].Uri, "https://test-account.cloud.databricks.com/#job/job-id/run/run-id")
		assert.Equal(t, taskInfo.Logs[0].Name, "Databricks Console")
	})
}

func TestBuildRequest(t *testing.T) {
	token := "test-token"
	runID := "019e70eb"
	databricksEndpoint := ""
	databricksURL := "https://" + testInstance + "/api/2.0/jobs/runs"
	t.Run("build http request for submitting a databricks job", func(t *testing.T) {
		req, err := buildRequest(post, nil, databricksEndpoint, testInstance, token, runID, false)
		header := http.Header{}
		header.Add("Authorization", "Bearer "+token)
		header.Add("Content-Type", "application/json")

		assert.NoError(t, err)
		assert.Equal(t, header, req.Header)
		assert.Equal(t, databricksURL+"/submit", req.URL.String())
		assert.Equal(t, post, req.Method)
	})
	t.Run("Get a databricks spark job status", func(t *testing.T) {
		req, err := buildRequest(get, nil, databricksEndpoint, testInstance, token, runID, false)

		assert.NoError(t, err)
		assert.Equal(t, databricksURL+"/get?run_id="+runID, req.URL.String())
		assert.Equal(t, get, req.Method)
	})
	t.Run("Cancel a spark job", func(t *testing.T) {
		req, err := buildRequest(post, nil, databricksEndpoint, testInstance, token, runID, true)

		assert.NoError(t, err)
		assert.Equal(t, databricksURL+"/cancel", req.URL.String())
		assert.Equal(t, post, req.Method)
	})
}

func TestBuildResponse(t *testing.T) {
	t.Run("build http response", func(t *testing.T) {
		bodyStr := `{"job_id":"019c06a4-0000", "message":"Statement executed successfully."}`
		responseBody := ioutil.NopCloser(strings.NewReader(bodyStr))
		response := &http.Response{Body: responseBody}
		actualData, err := buildResponse(response)
		assert.NoError(t, err)

		bodyByte, err := ioutil.ReadAll(strings.NewReader(bodyStr))
		assert.NoError(t, err)
		var expectedData map[string]interface{}
		err = json.Unmarshal(bodyByte, &expectedData)
		assert.NoError(t, err)
		assert.Equal(t, expectedData, actualData)
	})
}
