package logs

import (
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func Test_stackdriverLogPlugin_GetTaskLog(t *testing.T) {
	type fields struct {
		gcpProject  string
		logResource string
	}
	type args struct {
		podName       string
		namespace     string
		containerName string
		containerID   string
		logName       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    core.TaskLog
		wantErr bool
	}{
		{
			"podName-proj1",
			fields{gcpProject: "test-gcp-project", logResource: "aws_ec2_instance"},
			args{podName: "podName"},
			core.TaskLog{Uri: "https://console.cloud.google.com/logs/viewer?project=test-gcp-project&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dtest-gcp-project&resource=aws_ec2_instance&advancedFilter=resource.labels.pod_name%3DpodName", MessageFormat: core.TaskLog_JSON}, false,
		},
		{
			"podName2-proj2",
			fields{gcpProject: "proj2", logResource: "res1"},
			args{podName: "long-pod-name-xyyyx"},
			core.TaskLog{Uri: "https://console.cloud.google.com/logs/viewer?project=proj2&angularJsUrl=%2Flogs%2Fviewer%3Fproject%3Dproj2&resource=res1&advancedFilter=resource.labels.pod_name%3Dlong-pod-name-xyyyx", MessageFormat: core.TaskLog_JSON}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stackdriverLogPlugin{
				gcpProject:  tt.fields.gcpProject,
				logResource: tt.fields.logResource,
			}
			got, err := s.GetTaskLog(tt.args.podName, tt.args.namespace, tt.args.containerName, tt.args.containerID, tt.args.logName)
			if (err != nil) != tt.wantErr {
				t.Errorf("stackdriverLogPlugin.GetTaskLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stackdriverLogPlugin.GetTaskLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
