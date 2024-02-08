package tasklog

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

// Latest Run: Benchmark_mustInitTemplateRegexes-16    	   45960	     26914 ns/op
func Benchmark_initDefaultRegexes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		initDefaultRegexes()
	}
}

func dummyTaskExecID() pluginCore.TaskExecutionID {
	tID := &coreMocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("generated-name")
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "my-task-name",
			Project:      "my-task-project",
			Domain:       "my-task-domain",
			Version:      "1",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my-execution-name",
				Project: "my-execution-project",
				Domain:  "my-execution-domain",
			},
		},
		RetryAttempt: 1,
	})
	tID.OnGetUniqueNodeID().Return("n0-0-n0")
	return tID
}

func Test_Input_templateVars(t *testing.T) {
	testRegexes := struct {
		Foo      *regexp.Regexp
		Bar      *regexp.Regexp
		Baz      *regexp.Regexp
		Ham      *regexp.Regexp
		Spam     *regexp.Regexp
		LinkType *regexp.Regexp
		Port     *regexp.Regexp
	}{
		MustCreateRegex("foo"),
		MustCreateRegex("bar"),
		MustCreateRegex("baz"),
		MustCreateRegex("ham"),
		MustCreateRegex("spam"),
		MustCreateDynamicLogRegex("link_type"),
		MustCreateDynamicLogRegex("port"),
	}
	podBase := Input{
		HostName:             "my-host",
		PodName:              "my-pod",
		PodUID:               "my-pod-uid",
		Namespace:            "my-namespace",
		ContainerName:        "my-container",
		ContainerID:          "docker://containerID",
		LogName:              "main_logs",
		PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
		PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
		PodUnixStartTime:     123,
		PodUnixFinishTime:    12345,
	}
	taskExecutionBase := Input{
		LogName:              "main_logs",
		TaskExecutionID:      dummyTaskExecID(),
		PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
		PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
		PodUnixStartTime:     123,
		PodUnixFinishTime:    12345,
	}

	tests := []struct {
		name        string
		baseVars    Input
		extraVars   []TemplateVar
		exact       []TemplateVar
		contains    []TemplateVar
		notContains []TemplateVar
	}{
		{
			"pod happy path",
			podBase,
			nil,
			[]TemplateVar{
				{defaultRegexes.LogName, "main_logs"},
				{defaultRegexes.PodName, "my-pod"},
				{defaultRegexes.PodUID, "my-pod-uid"},
				{defaultRegexes.Namespace, "my-namespace"},
				{defaultRegexes.ContainerName, "my-container"},
				{defaultRegexes.ContainerID, "containerID"},
				{defaultRegexes.Hostname, "my-host"},
				{defaultRegexes.PodRFC3339StartTime, "1970-01-01T01:02:03+01:00"},
				{defaultRegexes.PodRFC3339FinishTime, "1970-01-01T04:25:45+01:00"},
				{defaultRegexes.PodUnixStartTime, "123"},
				{defaultRegexes.PodUnixFinishTime, "12345"},
			},
			nil,
			nil,
		},
		{
			"pod with extra vars",
			podBase,
			[]TemplateVar{
				{testRegexes.Foo, "foo"},
				{testRegexes.Bar, "bar"},
				{testRegexes.Baz, "baz"},
			},
			nil,
			[]TemplateVar{
				{testRegexes.Foo, "foo"},
				{testRegexes.Bar, "bar"},
				{testRegexes.Baz, "baz"},
			},
			nil,
		},
		{
			"task execution happy path",
			taskExecutionBase,
			nil,
			[]TemplateVar{
				{defaultRegexes.LogName, "main_logs"},
				{defaultRegexes.PodName, ""},
				{defaultRegexes.PodUID, ""},
				{defaultRegexes.Namespace, ""},
				{defaultRegexes.ContainerName, ""},
				{defaultRegexes.ContainerID, ""},
				{defaultRegexes.Hostname, ""},
				{defaultRegexes.NodeID, "n0-0-n0"},
				{defaultRegexes.GeneratedName, "generated-name"},
				{defaultRegexes.TaskRetryAttempt, "1"},
				{defaultRegexes.TaskID, "my-task-name"},
				{defaultRegexes.TaskVersion, "1"},
				{defaultRegexes.TaskProject, "my-task-project"},
				{defaultRegexes.TaskDomain, "my-task-domain"},
				{defaultRegexes.ExecutionName, "my-execution-name"},
				{defaultRegexes.ExecutionProject, "my-execution-project"},
				{defaultRegexes.ExecutionDomain, "my-execution-domain"},
				{defaultRegexes.PodRFC3339StartTime, "1970-01-01T01:02:03+01:00"},
				{defaultRegexes.PodRFC3339FinishTime, "1970-01-01T04:25:45+01:00"},
				{defaultRegexes.PodUnixStartTime, "123"},
				{defaultRegexes.PodUnixFinishTime, "12345"},
			},
			nil,
			nil,
		},
		{
			"task execution with extra vars",
			taskExecutionBase,
			[]TemplateVar{
				{testRegexes.Foo, "foo"},
				{testRegexes.Bar, "bar"},
				{testRegexes.Baz, "baz"},
			},
			nil,
			[]TemplateVar{
				{testRegexes.Foo, "foo"},
				{testRegexes.Bar, "bar"},
				{testRegexes.Baz, "baz"},
			},
			nil,
		},
		{
			"pod with port not affected",
			podBase,
			nil,
			nil,
			nil,
			[]TemplateVar{
				{testRegexes.Port, "1234"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.baseVars
			base.ExtraTemplateVars = tt.extraVars
			got := base.templateVars()
			if tt.exact != nil {
				assert.Equal(t, got, tt.exact)
			}
			if tt.contains != nil {
				for _, c := range tt.contains {
					assert.Contains(t, got, c)
				}
			}
			if tt.notContains != nil {
				for _, c := range tt.notContains {
					assert.NotContains(t, got, c)
				}
			}
		})
	}
}

func TestTemplateLogPlugin(t *testing.T) {
	type args struct {
		input Input
	}
	tests := []struct {
		name   string
		plugin TemplateLogPlugin
		args   args
		want   Output
	}{
		{
			"cloudwatch",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.{{.podName}}_{{.namespace}}_{{.containerName}}-{{.containerId}}.log"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					PodName:              "f-uuid-driver",
					PodUID:               "pod-uid",
					Namespace:            "flyteexamples-production",
					ContainerName:        "spark-kubernetes-driver",
					ContainerID:          "cri-o://abc",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{TaskLogs: []*core.TaskLog{{
				Uri:           "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.f-uuid-driver_flyteexamples-production_spark-kubernetes-driver-abc.log",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			}}},
		},
		{
			"stackdriver",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3D{{.podName}}"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					PodName:              "podName",
					PodUID:               "pod-uid",
					Namespace:            "flyteexamples-production",
					ContainerName:        "spark-kubernetes-driver",
					ContainerID:          "cri-o://abc",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{TaskLogs: []*core.TaskLog{{
				Uri:           "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3DpodName",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			}}},
		},
		{
			"kubernetes",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://dashboard.k8s.net/#!/log/{{.namespace}}/{{.podName}}/pod?namespace={{.namespace}}"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					PodName:              "flyteexamples-development-task-name",
					PodUID:               "pod-uid",
					Namespace:            "flyteexamples-development",
					ContainerName:        "ignore",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{TaskLogs: []*core.TaskLog{{
				Uri:           "https://dashboard.k8s.net/#!/log/flyteexamples-development/flyteexamples-development-task-name/pod?namespace=flyteexamples-development",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			}}},
		},
		{
			"splunk",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://prd-p-ighar.splunkcloud.com/en-US/app/search/search?q=search%20container_name%3D%22{{ .containerName }}%22"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					HostName:             "my-host",
					PodName:              "my-pod",
					Namespace:            "my-namespace",
					ContainerName:        "my-container",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Uri:           "https://prd-p-ighar.splunkcloud.com/en-US/app/search/search?q=search%20container_name%3D%22my-container%22",
						MessageFormat: core.TaskLog_JSON,
						Name:          "main_logs",
					},
				},
			},
		},
		{
			"ddog",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://app.datadoghq.com/logs?event&from_ts={{ .podUnixStartTime }}&live=true&query=pod_name%3A{{ .podName }}&to_ts={{ .podUnixFinishTime }}"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					HostName:             "my-host",
					PodName:              "my-pod",
					Namespace:            "my-namespace",
					ContainerName:        "my-container",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Uri:           "https://app.datadoghq.com/logs?event&from_ts=123&live=true&query=pod_name%3Amy-pod&to_ts=12345",
						MessageFormat: core.TaskLog_JSON,
						Name:          "main_logs",
					},
				},
			},
		},
		{
			"stackdriver-with-rfc3339-timestamp",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3D{{.podName}}%20%22{{.podRFC3339StartTime}}%22"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					HostName:             "my-host",
					PodName:              "my-pod",
					Namespace:            "my-namespace",
					ContainerName:        "my-container",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Uri:           "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3Dmy-pod%20%221970-01-01T01:02:03+01:00%22",
						MessageFormat: core.TaskLog_JSON,
						Name:          "main_logs",
					},
				},
			},
		},
		{
			"task-with-task-execution-identifier",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://flyte.corp.net/console/projects/{{ .executionProject }}/domains/{{ .executionDomain }}/executions/{{ .executionName }}/nodeId/{{ .nodeID }}/taskId/{{ .taskID }}/attempt/{{ .taskRetryAttempt }}/view/logs"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					HostName:             "my-host",
					PodName:              "my-pod",
					Namespace:            "my-namespace",
					ContainerName:        "my-container",
					ContainerID:          "ignore",
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
						Uri:           "https://flyte.corp.net/console/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/n0-0-n0/taskId/my-task-name/attempt/1/view/logs",
						MessageFormat: core.TaskLog_JSON,
						Name:          "main_logs",
					},
				},
			},
		},
		{
			"mapped-task-with-task-execution-identifier",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://flyte.corp.net/console/projects/{{ .executionProject }}/domains/{{ .executionDomain }}/executions/{{ .executionName }}/nodeId/{{ .nodeID }}/taskId/{{ .taskID }}/attempt/{{ .subtaskParentRetryAttempt }}/mappedIndex/{{ .subtaskExecutionIndex }}/mappedAttempt/{{ .subtaskRetryAttempt }}/view/logs"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					HostName:             "my-host",
					PodName:              "my-pod",
					Namespace:            "my-namespace",
					ContainerName:        "my-container",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
					TaskExecutionID:      dummyTaskExecID(),
					ExtraTemplateVars: []TemplateVar{
						{MustCreateRegex("subtaskExecutionIndex"), "1"},
						{MustCreateRegex("subtaskRetryAttempt"), "1"},
						{MustCreateRegex("subtaskParentRetryAttempt"), "0"},
					},
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Uri:           "https://flyte.corp.net/console/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/n0-0-n0/taskId/my-task-name/attempt/0/mappedIndex/1/mappedAttempt/1/view/logs",
						MessageFormat: core.TaskLog_JSON,
						Name:          "main_logs",
					},
				},
			},
		},
		{
			"flyteinteractive",
			TemplateLogPlugin{
				Name:                "vscode",
				DynamicTemplateURIs: []TemplateURI{"vscode://flyteinteractive:{{ .taskConfig.port }}/{{ .podName }}"},
				MessageFormat:       core.TaskLog_JSON,
			},
			args{
				input: Input{
					PodName: "my-pod-name",
					TaskTemplate: &core.TaskTemplate{
						Config: map[string]string{
							"link_type": "vscode",
							"port":      "1234",
						},
					},
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{
					{
						Uri:           "vscode://flyteinteractive:1234/my-pod-name",
						MessageFormat: core.TaskLog_JSON,
					},
				},
			},
		},
		{
			"flyteinteractive - no link_type in task template",
			TemplateLogPlugin{
				Name:                "vscode",
				DynamicTemplateURIs: []TemplateURI{"vscode://flyteinteractive:{{ .taskConfig.port }}/{{ .podName }}"},
				MessageFormat:       core.TaskLog_JSON,
				DisplayName:         "Flyteinteractive Logs",
			},
			args{
				input: Input{
					PodName: "my-pod-name",
				},
			},
			Output{
				TaskLogs: []*core.TaskLog{},
			},
		},
		{
			"kubernetes",
			TemplateLogPlugin{
				TemplateURIs:  []TemplateURI{"https://dashboard.k8s.net/#!/log/{{.namespace}}/{{.podName}}/pod?namespace={{.namespace}}"},
				MessageFormat: core.TaskLog_JSON,
			},
			args{
				input: Input{
					PodName:              "flyteexamples-development-task-name",
					PodUID:               "pod-uid",
					Namespace:            "flyteexamples-development",
					ContainerName:        "ignore",
					ContainerID:          "ignore",
					LogName:              "main_logs",
					PodRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
					PodRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
					PodUnixStartTime:     123,
					PodUnixFinishTime:    12345,
				},
			},
			Output{TaskLogs: []*core.TaskLog{{
				Uri:           "https://dashboard.k8s.net/#!/log/flyteexamples-development/flyteexamples-development-task-name/pod?namespace=flyteexamples-development",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			}}},
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
