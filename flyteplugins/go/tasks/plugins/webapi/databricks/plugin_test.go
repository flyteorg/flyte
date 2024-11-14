package databricks

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flytestdlib/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type MockClient struct {
	MockDo func(req *http.Request) (*http.Response, error)
}

func (m MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

var (
	testInstance = "test-account.cloud.databricks.com"
)

func TestPlugin(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))

	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
		client: &MockClient{func(req *http.Request) (*http.Response, error) {
			return nil, nil
		}},
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

func TestSendRequest(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test1"))
	databricksJob := map[string]interface{}{"sparkConfig": map[string]interface{}{"sparkVersion": "7.3.x-scala2.12"}}
	token := "token"

	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
		client: &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodPost)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutils.NewBytesReadCloser([]byte(`{"id":"someID","data":"someData"}`)),
			}, nil
		}},
	}

	t.Run("create a Databricks job", func(t *testing.T) {
		data, err := plugin.sendRequest(create, databricksJob, token, "")
		assert.NotNil(t, data)
		assert.Equal(t, "someID", data["id"])
		assert.Equal(t, "someData", data["data"])
		assert.Nil(t, err)
	})

	t.Run("failed to create a Databricks job", func(t *testing.T) {
		plugin.client = &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodPost)
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutils.NewBytesReadCloser([]byte(`{"message":"failed"}`)),
			}, nil
		}}
		data, err := plugin.sendRequest(create, databricksJob, token, "")
		assert.Nil(t, data)
		assert.Equal(t, err.Error(), "failed to create Databricks job with error [failed]")
	})

	t.Run("failed to send request to Databricks", func(t *testing.T) {
		plugin.client = &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodPost)
			return nil, errors.New("failed to send request")
		}}
		data, err := plugin.sendRequest(create, databricksJob, token, "")
		assert.Nil(t, data)
		assert.Equal(t, err.Error(), "failed to send request to Databricks platform with err: [failed to send request]")
	})

	t.Run("failed to send request to Databricks", func(t *testing.T) {
		plugin.client = &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodPost)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutils.NewBytesReadCloser([]byte(`123`)),
			}, nil
		}}
		data, err := plugin.sendRequest(create, databricksJob, token, "")
		assert.Nil(t, data)
		assert.Equal(t, err.Error(), "failed to parse response with err: [json: cannot unmarshal number into Go value of type map[string]interface {}]")
	})

	t.Run("get a Databricks job", func(t *testing.T) {
		plugin.client = &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodGet)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutils.NewBytesReadCloser([]byte(`{"message":"ok"}`)),
			}, nil
		}}
		data, err := plugin.sendRequest(get, databricksJob, token, "")
		assert.NotNil(t, data)
		assert.Nil(t, err)
	})

	t.Run("cancel a Databricks job", func(t *testing.T) {
		plugin.client = &MockClient{MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, req.Method, http.MethodPost)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutils.NewBytesReadCloser([]byte(`{"message":"ok"}`)),
			}, nil
		}}
		data, err := plugin.sendRequest(cancel, databricksJob, token, "")
		assert.NotNil(t, data)
		assert.Nil(t, err)
	})
}

func TestCreateTaskInfo(t *testing.T) {
	t.Run("create task info", func(t *testing.T) {
		taskInfo := createTaskInfo("run-id", "job-id", testInstance)

		assert.Equal(t, 1, len(taskInfo.Logs))
		assert.Equal(t, taskInfo.Logs[0].GetUri(), "https://test-account.cloud.databricks.com/#job/job-id/run/run-id")
		assert.Equal(t, taskInfo.Logs[0].GetName(), "Databricks Console")
	})
}
