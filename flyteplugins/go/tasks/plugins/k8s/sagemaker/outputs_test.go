package sagemaker

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
)

func Test_createModelOutputPath(t *testing.T) {
	type args struct {
		job            client.Object
		prefix         string
		bestExperiment string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "training job: simple output path", args: args{job: &trainingjobv1.TrainingJob{}, prefix: "s3://my-bucket", bestExperiment: "job-ABC"},
			want: "s3://my-bucket/training_outputs/job-ABC/output/model.tar.gz"},
		{name: "hpo job: simple output path", args: args{job: &hpojobv1.HyperparameterTuningJob{}, prefix: "s3://my-bucket", bestExperiment: "job-ABC"},
			want: "s3://my-bucket/hyperparameter_tuning_outputs/job-ABC/output/model.tar.gz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createModelOutputPath(tt.args.job, tt.args.prefix, tt.args.bestExperiment); got != tt.want {
				t.Errorf("createModelOutputPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
