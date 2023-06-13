package tasklog

import (
	"reflect"

	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestTemplateLog(t *testing.T) {
	p := NewTemplateLogPlugin([]string{"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.{{.podName}}_{{.podUID}}_{{.namespace}}_{{.containerName}}-{{.containerId}}.log"}, core.TaskLog_JSON)
	tl, err := p.GetTaskLog(
		"f-uuid-driver",
		"pod-uid",
		"flyteexamples-production",
		"spark-kubernetes-driver",
		"cri-o://abc",
		"main_logs",
		"2015-03-14T17:08:14+01:00",
		"2021-06-15T20:47:57+02:00",
		1426349294,
		1623782877,
	)
	assert.NoError(t, err)
	assert.Equal(t, tl.GetName(), "main_logs")
	assert.Equal(t, tl.GetMessageFormat(), core.TaskLog_JSON)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.f-uuid-driver_pod-uid_flyteexamples-production_spark-kubernetes-driver-abc.log", tl.Uri)
}

// Latest Run: Benchmark_mustInitTemplateRegexes-16    	   45960	     26914 ns/op
func Benchmark_mustInitTemplateRegexes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mustInitTemplateRegexes()
	}
}

func Test_templateLogPlugin_Regression(t *testing.T) {
	type fields struct {
		templateURI   string
		messageFormat core.TaskLog_MessageFormat
	}
	type args struct {
		podName              string
		podUID               string
		namespace            string
		containerName        string
		containerID          string
		logName              string
		podRFC3339StartTime  string
		podRFC3339FinishTime string
		podUnixStartTime     int64
		podUnixFinishTime    int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    core.TaskLog
		wantErr bool
	}{
		{
			"cloudwatch",
			fields{
				templateURI:   "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.{{.podName}}_{{.namespace}}_{{.containerName}}-{{.containerId}}.log",
				messageFormat: core.TaskLog_JSON,
			},
			args{
				podName:              "f-uuid-driver",
				podUID:               "pod-uid",
				namespace:            "flyteexamples-production",
				containerName:        "spark-kubernetes-driver",
				containerID:          "cri-o://abc",
				logName:              "main_logs",
				podRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
				podRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
				podUnixStartTime:     123,
				podUnixFinishTime:    12345,
			},
			core.TaskLog{
				Uri:           "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.f-uuid-driver_flyteexamples-production_spark-kubernetes-driver-abc.log",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			},
			false,
		},
		{
			"stackdriver",
			fields{
				templateURI:   "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3D{{.podName}}",
				messageFormat: core.TaskLog_JSON,
			},
			args{
				podName:              "podName",
				podUID:               "pod-uid",
				namespace:            "flyteexamples-production",
				containerName:        "spark-kubernetes-driver",
				containerID:          "cri-o://abc",
				logName:              "main_logs",
				podRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
				podRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
				podUnixStartTime:     123,
				podUnixFinishTime:    12345,
			},
			core.TaskLog{
				Uri:           "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3DpodName",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			},
			false,
		},
		{
			"kubernetes",
			fields{
				templateURI:   "https://dashboard.k8s.net/#!/log/{{.namespace}}/{{.podName}}/pod?namespace={{.namespace}}",
				messageFormat: core.TaskLog_JSON,
			},
			args{
				podName:              "flyteexamples-development-task-name",
				podUID:               "pod-uid",
				namespace:            "flyteexamples-development",
				containerName:        "ignore",
				containerID:          "ignore",
				logName:              "main_logs",
				podRFC3339StartTime:  "1970-01-01T01:02:03+01:00",
				podRFC3339FinishTime: "1970-01-01T04:25:45+01:00",
				podUnixStartTime:     123,
				podUnixFinishTime:    12345,
			},
			core.TaskLog{
				Uri:           "https://dashboard.k8s.net/#!/log/flyteexamples-development/flyteexamples-development-task-name/pod?namespace=flyteexamples-development",
				MessageFormat: core.TaskLog_JSON,
				Name:          "main_logs",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TemplateLogPlugin{
				templateUris:  []string{tt.fields.templateURI},
				messageFormat: tt.fields.messageFormat,
			}

			got, err := s.GetTaskLog(tt.args.podName, tt.args.podUID, tt.args.namespace, tt.args.containerName, tt.args.containerID, tt.args.logName, tt.args.podRFC3339FinishTime, tt.args.podRFC3339FinishTime, tt.args.podUnixStartTime, tt.args.podUnixFinishTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetTaskLog() got = %v, want %v, diff: %v", got, tt.want, diff)
			}
		})
	}
}

func TestTemplateLogPlugin_NewTaskLog(t *testing.T) {
	type fields struct {
		templateURI   string
		messageFormat core.TaskLog_MessageFormat
	}
	type args struct {
		input Input
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Output
		wantErr bool
	}{
		{
			"splunk",
			fields{
				templateURI:   "https://prd-p-ighar.splunkcloud.com/en-US/app/search/search?q=search%20container_name%3D%22{{ .containerName }}%22",
				messageFormat: core.TaskLog_JSON,
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
			false,
		},
		{
			"ddog",
			fields{
				templateURI:   "https://app.datadoghq.com/logs?event&from_ts={{ .podUnixStartTime }}&live=true&query=pod_name%3A{{ .podName }}&to_ts={{ .podUnixFinishTime }}",
				messageFormat: core.TaskLog_JSON,
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
			false,
		},
		{
			"stackdriver-with-rfc3339-timestamp",
			fields{
				templateURI:   "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3D{{.podName}}%20%22{{.podRFC3339StartTime}}%22",
				messageFormat: core.TaskLog_JSON,
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
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TemplateLogPlugin{
				templateUris:  []string{tt.fields.templateURI},
				messageFormat: tt.fields.messageFormat,
			}
			got, err := s.GetTaskLogs(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTaskLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTaskLog() got = %v, want %v", got, tt.want)
			}
		})
	}
}
