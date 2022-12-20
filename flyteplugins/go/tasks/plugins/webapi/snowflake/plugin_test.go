package snowflake

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
	MockDo func(req *http.Request) (*http.Response, error)
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
		taskInfo := createTaskInfo("d5493e36", "test-account")

		assert.Equal(t, 1, len(taskInfo.Logs))
		assert.Equal(t, taskInfo.Logs[0].Uri, "https://test-account.snowflakecomputing.com/console#/monitoring/queries/detail?queryId=d5493e36")
		assert.Equal(t, taskInfo.Logs[0].Name, "Snowflake Console")
	})
}

func TestBuildRequest(t *testing.T) {
	account := "test-account"
	token := "test-token"
	queryID := "019e70eb-0000-278b-0000-40f100012b1a"
	snowflakeEndpoint := ""
	snowflakeURL := "https://" + account + ".snowflakecomputing.com/api/v2/statements"
	t.Run("build http request for submitting a snowflake query", func(t *testing.T) {
		queryInfo := QueryInfo{
			Account:   account,
			Warehouse: "test-warehouse",
			Schema:    "test-schema",
			Database:  "test-database",
			Statement: "SELECT 1",
		}

		req, err := buildRequest(post, queryInfo, snowflakeEndpoint, account, token, queryID, false)
		header := http.Header{}
		header.Add("Authorization", "Bearer "+token)
		header.Add("X-Snowflake-Authorization-Token-Type", "KEYPAIR_JWT")
		header.Add("Content-Type", "application/json")
		header.Add("Accept", "application/json")

		assert.NoError(t, err)
		assert.Equal(t, header, req.Header)
		assert.Equal(t, snowflakeURL+"?async=true", req.URL.String())
		assert.Equal(t, post, req.Method)
	})
	t.Run("build http request for getting a snowflake query status", func(t *testing.T) {
		req, err := buildRequest(get, QueryInfo{}, snowflakeEndpoint, account, token, queryID, false)

		assert.NoError(t, err)
		assert.Equal(t, snowflakeURL+"/"+queryID, req.URL.String())
		assert.Equal(t, get, req.Method)
	})
	t.Run("build http request for deleting a snowflake query", func(t *testing.T) {
		req, err := buildRequest(post, QueryInfo{}, snowflakeEndpoint, account, token, queryID, true)

		assert.NoError(t, err)
		assert.Equal(t, snowflakeURL+"/"+queryID+"/cancel", req.URL.String())
		assert.Equal(t, post, req.Method)
	})
}

func TestBuildResponse(t *testing.T) {
	t.Run("build http response", func(t *testing.T) {
		bodyStr := `{"statementHandle":"019c06a4-0000", "message":"Statement executed successfully."}`
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
