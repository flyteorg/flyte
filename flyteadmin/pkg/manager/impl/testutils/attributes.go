package testutils

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

var ExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"foo", "bar", "baz",
			},
		},
	},
}

var WorkflowExecutionConfigSample = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
		WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
			MaxParallelism: 5,
			RawOutputDataConfig: &admin.RawOutputDataConfig{
				OutputLocationPrefix: "s3://test-bucket",
			},
			Labels: &admin.Labels{
				Values: map[string]string{"lab1": "val1"},
			},
			Annotations: nil,
		},
	},
}
