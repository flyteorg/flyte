package tasklog

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// decodeAzureQuery reverses the URL-encode, base64 and gzip steps applied to the query.
// The gzip byte stream is not stable across Go releases, so tests compare the decoded query.
func decodeAzureQuery(t *testing.T, encoded string) string {
	t.Helper()
	unescaped, err := url.QueryUnescape(encoded)
	require.NoError(t, err)
	compressed, err := base64.StdEncoding.DecodeString(unescaped)
	require.NoError(t, err)
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)
	raw, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(raw)
}

func TestAzureTemplateLogPlugin(t *testing.T) {
	const baseURI = "https://portal.azure.com#@test-tenantID/blade/Microsoft_OperationsManagementSuite_Workspace/Logs.ReactView/resourceId/%%2Fsubscriptions%%2Ftest-subscriptionID%%2FresourceGroups%%2Ftest-resourceGroupName/source/LogsBlade.AnalyticsShareLinkToQuery/q/"
	type args struct {
		input Input
	}
	tests := []struct {
		name      string
		plugin    AzureLogsTemplatePlugin
		args      args
		wantName  string
		wantQuery string
	}{
		{
			"test azure template log plugin",
			AzureLogsTemplatePlugin{
				TemplateLogPlugin: TemplateLogPlugin{
					Name:         "Azure Logs",
					DisplayName:  "Azure Logs",
					TemplateURIs: []TemplateURI{baseURI},
				},
			},
			args{
				input: Input{
					HostName:             "test-host",
					PodName:              "test-pod",
					Namespace:            "test-namespace",
					ContainerName:        "test-container",
					ContainerID:          "test-containerID",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
					TaskExecutionID:      dummyTaskExecID(),
				},
			},
			"Azure Logsmain_logs",
			`let StartTime = datetime_add('hour', -1, datetime("1970-01-01T01:02:03+01:00"));
let FinishTime = datetime_add('hour', 1, datetime("1970-01-01T04:25:45+01:00"));
ContainerLogV2
| where TimeGenerated between (StartTime .. FinishTime)
 and ContainerName == "test-container"
 and PodName == "test-pod"
 and PodNamespace == "test-namespace"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.plugin.GetTaskLogs(tt.args.input)
			assert.NoError(t, err)
			require.Len(t, got.TaskLogs, 1)
			taskLog := got.TaskLogs[0]
			assert.Equal(t, tt.wantName, taskLog.GetName())
			assert.Equal(t, core.TaskLog_JSON, taskLog.GetMessageFormat())
			require.True(t, strings.HasPrefix(taskLog.GetUri(), baseURI), "unexpected URI prefix: %s", taskLog.GetUri())
			assert.Equal(t, tt.wantQuery, decodeAzureQuery(t, strings.TrimPrefix(taskLog.GetUri(), baseURI)))
		})
	}
}
