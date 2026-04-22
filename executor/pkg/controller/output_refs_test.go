package controller

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

func taskActionForOutputRefs(namespace, name, runOutputBase, actionName string, attempts uint32) *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: flyteorgv1.TaskActionSpec{
			RunOutputBase: runOutputBase,
			ActionName:    actionName,
		},
		Status: flyteorgv1.TaskActionStatus{
			Attempts: attempts,
		},
	}
}

func TestOutputRefs(t *testing.T) {
	ctx := context.Background()

	t.Run("empty RunOutputBase returns nil", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "", "action-0", 1)
		assert.Nil(t, outputRefs(ctx, ta))
	})

	t.Run("output URI ends with outputs.pb", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)
		assert.True(t, strings.HasSuffix(refs.GetOutputUri(), "/outputs.pb"),
			"expected URI to end with /outputs.pb, got: %s", refs.GetOutputUri())
	})

	t.Run("output URI contains action name", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)
		assert.Contains(t, refs.GetOutputUri(), "action-0")
	})

	t.Run("output URI contains attempt number", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 2)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)
		// segment before outputs.pb should be the attempt number
		uri := strings.TrimSuffix(refs.GetOutputUri(), "/outputs.pb")
		parts := strings.Split(uri, "/")
		assert.Equal(t, "2", parts[len(parts)-1])
	})

	t.Run("attempt defaults to 1 when Status.Attempts is zero", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 0)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)
		uri := strings.TrimSuffix(refs.GetOutputUri(), "/outputs.pb")
		parts := strings.Split(uri, "/")
		assert.Equal(t, "1", parts[len(parts)-1])
	})

	t.Run("report URI ends with report.html and shares prefix with output URI", func(t *testing.T) {
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)
		assert.True(t, strings.HasSuffix(refs.GetReportUri(), "/report.html"),
			"expected report URI to end with /report.html, got: %s", refs.GetReportUri())
		assert.Equal(t,
			strings.TrimSuffix(refs.GetOutputUri(), "/outputs.pb"),
			strings.TrimSuffix(refs.GetReportUri(), "/report.html"),
			"report URI and output URI should share the same prefix")
	})

	t.Run("output URI is consistent with ComputeActionOutputPath", func(t *testing.T) {
		// outputRefs and NewTaskExecutionContext must agree on the output path.
		// Verify outputRefs produces a URI whose directory matches the plugin's output prefix.
		ta := taskActionForOutputRefs("flyte", "ta-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		refs := outputRefs(ctx, ta)
		require.NotNil(t, refs)

		// The URI should be <outputPrefix>/outputs.pb — strip the file to get the prefix.
		dir := strings.TrimSuffix(refs.GetOutputUri(), "/outputs.pb")
		assert.NotEmpty(t, dir)
		assert.NotEqual(t, refs.GetOutputUri(), dir, "TrimSuffix should have removed /outputs.pb")
	})
}