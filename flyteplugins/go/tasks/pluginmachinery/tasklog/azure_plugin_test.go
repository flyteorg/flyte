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

func TestAzureTemplateLogPlugin(t *testing.T) {
	const baseURI = "https://portal.azure.com#@test-tenantID/blade/Microsoft_OperationsManagementSuite_Workspace/Logs.ReactView/resourceId/%%2Fsubscriptions%%2Ftest-subscriptionID%%2FresourceGroups%%2Ftest-resourceGroupName/source/LogsBlade.AnalyticsShareLinkToQuery/q/"

	plugin := AzureLogsTemplatePlugin{
		TemplateLogPlugin: TemplateLogPlugin{
			Name:         "Azure Logs",
			DisplayName:  "Azure Logs",
			TemplateURIs: []TemplateURI{baseURI},
		},
	}
	input := Input{
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
	}

	got, err := plugin.GetTaskLogs(input)
	require.NoError(t, err)
	require.Len(t, got.TaskLogs, 1)
	assert.Equal(t, "Azure Logsmain_logs", got.TaskLogs[0].Name)
	assert.Equal(t, core.TaskLog_JSON, got.TaskLogs[0].MessageFormat)
	require.True(t, strings.HasPrefix(got.TaskLogs[0].Uri, baseURI))

	encodedQuery := strings.TrimPrefix(got.TaskLogs[0].Uri, baseURI)
	base64Query, err := url.QueryUnescape(encodedQuery)
	require.NoError(t, err)
	compressedQuery, err := base64.StdEncoding.DecodeString(base64Query)
	require.NoError(t, err)
	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedQuery))
	require.NoError(t, err)
	query, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	require.NoError(t, gzipReader.Close())

	assert.Equal(t, replaceAll(defaulQueryFormat, input.templateVars()), string(query))
}
