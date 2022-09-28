package snowflake

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/tests"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEndToEnd(t *testing.T) {
	server := newFakeSnowflakeServer()
	defer server.Close()

	iter := func(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
		return nil
	}

	cfg := defaultConfig
	cfg.snowflakeEndpoint = server.URL
	cfg.DefaultWarehouse = "test-warehouse"
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	err := SetConfig(&cfg)
	assert.NoError(t, err)

	pluginEntry := pluginmachinery.CreateRemotePlugin(newSnowflakeJobTaskPlugin())
	plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext())
	assert.NoError(t, err)

	t.Run("SELECT 1", func(t *testing.T) {
		config := make(map[string]string)
		config["database"] = "my-database"
		config["account"] = "snowflake"
		config["schema"] = "my-schema"
		config["warehouse"] = "my-warehouse"

		inputs, _ := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		template := flyteIdlCore.TaskTemplate{
			Type:   "snowflake",
			Config: config,
			Target: &coreIdl.TaskTemplate_Sql{Sql: &coreIdl.Sql{Statement: "SELECT 1", Dialect: coreIdl.Sql_ANSI}},
		}

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)

		assert.Equal(t, true, phase.Phase().IsSuccess())
	})
}

func newFakeSnowflakeServer() *httptest.Server {
	statementHandle := "019e7546-0000-278c-0000-40f10001a082"
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v2/statements" && request.Method == "POST" {
			writer.WriteHeader(202)
			bytes := []byte(fmt.Sprintf(`{
			  "statementHandle": "%v",
			  "message": "Asynchronous execution in progress."
			}`, statementHandle))
			_, _ = writer.Write(bytes)
			return
		}

		if request.URL.Path == "/api/v2/statements/"+statementHandle && request.Method == "GET" {
			writer.WriteHeader(200)
			bytes := []byte(fmt.Sprintf(`{
			  "statementHandle": "%v",
			  "message": "Statement executed successfully."
			}`, statementHandle))
			_, _ = writer.Write(bytes)
			return
		}

		if request.URL.Path == "/api/v2/statements/"+statementHandle+"/cancel" && request.Method == "POST" {
			writer.WriteHeader(200)
			return
		}

		writer.WriteHeader(500)
	}))
}

func newFakeSetupContext() *pluginCoreMocks.SetupContext {
	fakeResourceRegistrar := pluginCoreMocks.ResourceRegistrar{}
	fakeResourceRegistrar.On("RegisterResourceQuota", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	labeled.SetMetricKeys(contextutils.NamespaceKey)

	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))
	fakeSetupContext.OnResourceRegistrar().Return(&fakeResourceRegistrar)

	return &fakeSetupContext
}
