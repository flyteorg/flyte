package sagemaker

import (
	"context"
	"reflect"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"

	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
)

func Test_awsSagemakerPlugin_BuildIdentityResource(t *testing.T) {
	ctx := context.TODO()
	type fields struct {
		TaskType pluginsCore.TaskType
	}
	type args struct {
		in0 context.Context
		in1 pluginsCore.TaskExecutionMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    client.Object
		wantErr bool
	}{
		{name: "Training Job Identity Resource", fields: fields{TaskType: trainingJobTaskType},
			args: args{in0: ctx, in1: genMockTaskExecutionMetadata()}, want: &trainingjobv1.TrainingJob{}, wantErr: false},
		{name: "HPO Job Identity Resource", fields: fields{TaskType: hyperparameterTuningJobTaskType},
			args: args{in0: ctx, in1: genMockTaskExecutionMetadata()}, want: &hpojobv1.HyperparameterTuningJob{}, wantErr: false},
		{name: "Unsupported Job Identity Resource", fields: fields{TaskType: "bad type"},
			args: args{in0: ctx, in1: genMockTaskExecutionMetadata()}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := awsSagemakerPlugin{
				TaskType: tt.fields.TaskType,
			}
			got, err := m.BuildIdentityResource(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildIdentityResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildIdentityResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProperties(t *testing.T) {
	plugin := awsSagemakerPlugin{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, plugin.GetProperties())
}

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}
