package testutils

import "github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

var ExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"foo", "bar", "baz",
			},
		},
	},
}
