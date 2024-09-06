package plugin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/config"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

func TestIsValidEnvironmentSpec(t *testing.T) {
	podTemplateSpec := &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Command: []string{"bar"},
				},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.Nil(t, err)

	tests := []struct {
		name                    string
		executionEnvironmentID  core.ExecutionEnvID
		fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec
		expectedError           string
	}{
		{
			name:                    "EmptyExecutionEnvironmentName",
			executionEnvironmentID:  core.ExecutionEnvID{},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{},
			expectedError:           "execution environment name is required",
		},
		{
			name: "EmptyExecutionEnvironmentVersion",
			executionEnvironmentID: core.ExecutionEnvID{
				Name: "foo",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{},
			expectedError:           "execution environment version is required",
		},
		{
			name: "NegativeBacklogLength",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength: -1,
			},
			expectedError: "backlog length must be greater than or equal to 0",
		},
		{
			name: "ZeroParallelism",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength: 0,
				Parallelism:   0,
			},
			expectedError: "parallelism must be greater than 0",
		},
		{
			name: "ZeroTTLSeconds",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength: 0,
				Parallelism:   1,
				TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
					TtlSeconds: 0,
				},
			},
			expectedError: "ttlSeconds must be greater than 0 if terminationCriteria is set",
		},
		{
			name: "InvalidPodTemplateSpec",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength: 0,
				Parallelism:   1,
				TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
					TtlSeconds: 30,
				},
			},
			expectedError: "unable to unmarshal PodTemplateSpec [[]], Err: [unexpected end of JSON input]",
		},
		{
			name: "ZeroReplicaCount",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength:   0,
				Parallelism:     1,
				PodTemplateSpec: podTemplateSpecBytes,
				ReplicaCount:    0,
				TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
					TtlSeconds: 30,
				},
			},
			expectedError: "replica count must be greater than 0",
		},
		{
			name: "Success",
			executionEnvironmentID: core.ExecutionEnvID{
				Name:    "foo",
				Version: "bar",
			},
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				BacklogLength:   0,
				Parallelism:     1,
				PodTemplateSpec: podTemplateSpecBytes,
				ReplicaCount:    1,
				TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
					TtlSeconds: 30,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidEnvironmentSpec(tt.executionEnvironmentID, tt.fastTaskEnvironmentSpec)
			if len(tt.expectedError) > 0 {
				assert.NotNil(t, actual)
				assert.Contains(t, actual.Error(), tt.expectedError)
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}

func TestGetTTLSeconds(t *testing.T) {
	tests := []struct {
		name                    string
		fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec
		defaultTTL              time.Duration
		expected                time.Duration
	}{
		{
			name:                    "Default",
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{},
			defaultTTL:              time.Second * time.Duration(90),
			expected:                time.Second * time.Duration(90),
		},
		{
			name: "TerminationCriteria",
			fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
				TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
					TtlSeconds: 120,
				},
			},
			defaultTTL: time.Second * time.Duration(90),
			expected:   time.Second * time.Duration(120),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetConfig().DefaultTTL = config.Duration{Duration: tt.defaultTTL}

			actual := getTTLOrDefault(tt.fastTaskEnvironmentSpec)
			assert.Equal(t, time.Duration(tt.expected), actual)
		})
	}
}

func TestSanitizePodName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "Base",
			input:  "test",
			output: "test",
		},
		{
			name:   "ReplaceUnderScoreWithDash",
			input:  "t_e_s_t",
			output: "t-e-s-t",
		},
		{
			name:   "ToLower",
			input:  "TEST",
			output: "test",
		},
		{
			name:   "RemoveSpecialCharacters",
			input:  "t!e@s#t$",
			output: "test",
		},
		{
			name:   "StripLeadingDashOrDot",
			input:  "-.test",
			output: "test",
		},
		{
			name:   "CutOffLongName",
			input:  "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
			output: "testtesttesttesttesttesttesttesttesttesttesttestte",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := sanitizeEnvName(test.input)
			assert.Equal(t, test.output, output)
		})
	}
}
