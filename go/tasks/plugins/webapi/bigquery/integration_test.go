package bigquery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginUtils "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/tests"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/bigquery/v2"
)

func TestEndToEnd(t *testing.T) {
	server := newFakeBigQueryServer()
	defer server.Close()

	iter := func(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
		return nil
	}

	cfg := defaultConfig
	cfg.bigQueryEndpoint = server.URL
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	err := SetConfig(&cfg)
	assert.NoError(t, err)

	pluginEntry := pluginmachinery.CreateRemotePlugin(newBigQueryJobTaskPlugin())
	plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext())
	assert.NoError(t, err)

	t.Run("SELECT 1", func(t *testing.T) {
		queryJobConfig := QueryJobConfig{
			ProjectID: "flyte",
		}

		inputs, _ := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		custom, _ := pluginUtils.MarshalObjToStruct(queryJobConfig)
		template := flyteIdlCore.TaskTemplate{
			Type:   bigqueryQueryJobTask,
			Custom: custom,
			Target: &flyteIdlCore.TaskTemplate_Sql{Sql: &flyteIdlCore.Sql{Statement: "SELECT 1", Dialect: flyteIdlCore.Sql_ANSI}},
		}

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)

		assert.Equal(t, true, phase.Phase().IsSuccess())
	})
}

func newFakeBigQueryServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/projects/flyte/jobs" && request.Method == "POST" {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: "RUNNING"}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if strings.HasPrefix(request.URL.Path, "/projects/flyte/jobs/") && request.Method == "GET" {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: "DONE"}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
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
