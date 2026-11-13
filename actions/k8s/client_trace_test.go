package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func TestEnqueue_TraceAnnotations(t *testing.T) {
	traceData := map[string]string{
		"traceparent": "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01",
		"tracestate":  "vendor=value",
		"baggage":     "tenant=example",
	}
	for _, kind := range []string{"task", "condition"} {
		for _, root := range []bool{true, false} {
			for _, contextKind := range []string{"full", "trace only", "baggage only", "empty"} {
				name := kind + "/child/" + contextKind
				if root {
					name = kind + "/root/" + contextKind
				}
				t.Run(name, func(t *testing.T) {
					c := newTaskClient(t, nil)
					var action *actions.Action
					if kind == "task" {
						action = newChildTaskAction()
					} else {
						action = newConditionAction(core.SimpleType_BOOLEAN)
					}
					if root {
						action.ParentActionName = nil
					}
					annotations := map[string]string{"owner": "sdk", "traceparent": "stale", "tracestate": "stale=value", "baggage": "stale=value"}
					parent := &executorv1.TaskAction{}
					require.NoError(t, c.k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "flyte", Name: "run1-a0"}, parent))
					parent.Annotations = annotations
					require.NoError(t, c.k8sClient.Update(context.Background(), parent))
					runSpec := &task.RunSpec{Annotations: &task.Annotations{Values: annotations}}
					expected := map[string]string{}
					switch contextKind {
					case "full":
						for k, v := range traceData {
							expected[k] = v
						}
					case "trace only":
						expected["traceparent"] = traceData["traceparent"]
					case "baggage only":
						expected["baggage"] = traceData["baggage"]
					}
					ctx := otelutils.RPCPropagator().Extract(context.Background(), propagation.MapCarrier(expected))
					require.NoError(t, c.Enqueue(ctx, action, runSpec))
					created, err := c.GetTaskAction(ctx, action.ActionId)
					require.NoError(t, err)
					if kind == "task" {
						expected["owner"] = "sdk"
					}
					if len(expected) == 0 {
						assert.Empty(t, created.Annotations)
					} else {
						assert.Equal(t, expected, created.Annotations)
					}
					unchanged := &executorv1.TaskAction{}
					require.NoError(t, c.k8sClient.Get(ctx, client.ObjectKeyFromObject(parent), unchanged))
					assert.Equal(t, annotations, unchanged.Annotations)
					assert.Equal(t, "stale", runSpec.Annotations.Values["traceparent"])
				})
			}
		}
	}
}
