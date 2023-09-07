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

const (
	httpPost string = "POST"
	httpGet  string = "GET"
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

	inputs, _ := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
	template := flyteIdlCore.TaskTemplate{
		Type:   bigqueryQueryJobTask,
		Target: &flyteIdlCore.TaskTemplate_Sql{Sql: &flyteIdlCore.Sql{Statement: "SELECT 1", Dialect: flyteIdlCore.Sql_ANSI}},
	}

	t.Run("SELECT 1", func(t *testing.T) {
		queryJobConfig := QueryJobConfig{
			ProjectID: "flyte",
		}

		custom, _ := pluginUtils.MarshalObjToStruct(queryJobConfig)
		template.Custom = custom

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)

		assert.Equal(t, true, phase.Phase().IsSuccess())
	})

	t.Run("cache job result", func(t *testing.T) {
		queryJobConfig := QueryJobConfig{
			ProjectID: "cache",
		}

		custom, _ := pluginUtils.MarshalObjToStruct(queryJobConfig)
		template.Custom = custom

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)

		assert.Equal(t, true, phase.Phase().IsSuccess())
	})

	t.Run("pending job", func(t *testing.T) {
		queryJobConfig := QueryJobConfig{
			ProjectID: "pending",
		}

		custom, _ := pluginUtils.MarshalObjToStruct(queryJobConfig)
		template.Custom = custom

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)

		assert.Equal(t, true, phase.Phase().IsSuccess())
	})
}

func newFakeBigQueryServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/projects/flyte/jobs" && request.Method == httpPost {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusRunning}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if strings.HasPrefix(request.URL.Path, "/projects/flyte/jobs/") && request.Method == httpGet {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusDone},
				Configuration: &bigquery.JobConfiguration{
					Query: &bigquery.JobConfigurationQuery{
						DestinationTable: &bigquery.TableReference{
							ProjectId: "project", DatasetId: "dataset", TableId: "table"}}}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if request.URL.Path == "/projects/cache/jobs" && request.Method == httpPost {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusDone}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if strings.HasPrefix(request.URL.Path, "/projects/cache/jobs/") && request.Method == httpGet {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusDone},
				Configuration: &bigquery.JobConfiguration{
					Query: &bigquery.JobConfigurationQuery{
						DestinationTable: &bigquery.TableReference{
							ProjectId: "project", DatasetId: "dataset", TableId: "table"}}}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if request.URL.Path == "/projects/pending/jobs" && request.Method == httpPost {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusPending}}
			bytes, _ := json.Marshal(job)
			_, _ = writer.Write(bytes)
			return
		}

		if strings.HasPrefix(request.URL.Path, "/projects/pending/jobs/") && request.Method == httpGet {
			writer.WriteHeader(200)
			job := bigquery.Job{Status: &bigquery.JobStatus{State: bigqueryStatusDone}}
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
