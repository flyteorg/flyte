package tasklog

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestAzureTemplateLogPlugin(t *testing.T) {
	type args struct {
		input Input
	}
	tests := []struct {
		name   string
		plugin AzureLogsTemplatePlugin
		args   args
		want   Output
	}{
		{
			"test azure template log plugin",
			AzureLogsTemplatePlugin{
				TemplateLogPlugin: TemplateLogPlugin{
					Name:         "Azure Logs",
					DisplayName:  "Azure Logs",
					TemplateURIs: []TemplateURI{"https://portal.azure.com#@test-tenantID/blade/Microsoft_OperationsManagementSuite_Workspace/Logs.ReactView/resourceId/%%2Fsubscriptions%%2Ftest-subscriptionID%%2FresourceGroups%%2Ftest-resourceGroupName/source/LogsBlade.AnalyticsShareLinkToQuery/q/"},
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
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Name:          "Azure Logsmain_logs",
						Uri:           "https://portal.azure.com#@test-tenantID/blade/Microsoft_OperationsManagementSuite_Workspace/Logs.ReactView/resourceId/%%2Fsubscriptions%%2Ftest-subscriptionID%%2FresourceGroups%%2Ftest-resourceGroupName/source/LogsBlade.AnalyticsShareLinkToQuery/q/H4sIAAAAAAAA%2F3yPwUrFMBBF9%2F2KIZvX4ktJaosY6UrQjYhgcStjM9iATUo60o0fL0Fa24XuhnsuhzsfxPDMGLlzI0ELFpnYjfSK1uanIXzG0xmkPm8gF%2Fr6SkmlpdKd0kZVRl1epEOJorjJkvDOeTcP%2Fxn%2FFNamakzd7IS3wTM6T%2FEhvL9U2RcsA0WCZL8nTxGZLLwRL0Qe8t9fynK3o8gAvYXN9YhpWwuCaWbZr7H4qT0FeyxMwR7RPGG%2F436NxHcAAAD%2F%2F4NTt6FQAQAA",
						MessageFormat: core.TaskLog_JSON,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.plugin.GetTaskLogs(tt.args.input)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskLogs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
