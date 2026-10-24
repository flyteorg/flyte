package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func recoverRunSpec(sourceRun string, forceRerun ...string) *task.RunSpec {
	spec := &task.RunSpec{
		Relation: &common.Relation{
			RelatedTo: &common.RunIdentifier{
				Org: "org1", Project: "proj", Domain: "dev", Name: sourceRun,
			},
			RelationType: common.RelationType_RELATION_TYPE_RECOVER,
		},
	}
	if len(forceRerun) > 0 {
		spec.Recover = &task.Recover{ForceRerunActions: forceRerun}
	}
	return spec
}

func TestApplyRunSpecToTaskAction_StampsRecoveryContext(t *testing.T) {
	taskAction := &executorv1.TaskAction{Spec: executorv1.TaskActionSpec{}}

	require.NoError(t, applyRunSpecToTaskAction(taskAction, recoverRunSpec("r1", "a3", "a7")))

	recoveryContext := taskAction.Spec.RecoveryContext
	require.NotNil(t, recoveryContext)
	assert.Equal(t, []string{"a3", "a7"}, recoveryContext.ForceRerunActions)

	relation := &common.Relation{}
	require.NoError(t, proto.Unmarshal(recoveryContext.Relation, relation))
	assert.Equal(t, common.RelationType_RELATION_TYPE_RECOVER, relation.GetRelationType())
	assert.Equal(t, "r1", relation.GetRelatedTo().GetName())
}

// rerun and spawn share RunSpec.relation with recover; only the type makes it a recovery.
func TestApplyRunSpecToTaskAction_NonRecoveryRelationStampsNothing(t *testing.T) {
	for _, relationType := range []common.RelationType{
		common.RelationType_RELATION_TYPE_RERUN,
		common.RelationType_RELATION_TYPE_SPAWN,
		common.RelationType_RELATION_TYPE_UNSPECIFIED,
	} {
		t.Run(relationType.String(), func(t *testing.T) {
			spec := recoverRunSpec("r1")
			spec.Relation.RelationType = relationType

			taskAction := &executorv1.TaskAction{Spec: executorv1.TaskActionSpec{}}
			require.NoError(t, applyRunSpecToTaskAction(taskAction, spec))
			assert.Nil(t, taskAction.Spec.RecoveryContext)
		})
	}
}

func TestApplyRunSpecToTaskAction_NilRunSpecClearsRecoveryContext(t *testing.T) {
	taskAction := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			RecoveryContext: &executorv1.RecoveryContext{Relation: []byte("stale")},
		},
	}

	require.NoError(t, applyRunSpecToTaskAction(taskAction, nil))
	assert.Nil(t, taskAction.Spec.RecoveryContext)
}

func TestInheritRunContextFromParentTaskAction_CopiesRecoveryContext(t *testing.T) {
	parent := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			RecoveryContext: recoveryContextFor("source-run", "a3"),
		},
	}
	child := &executorv1.TaskAction{Spec: executorv1.TaskActionSpec{}}

	inheritRunContextFromParentTaskAction(child, parent)

	require.NotNil(t, child.Spec.RecoveryContext)
	assert.Equal(t, []string{"a3"}, child.Spec.RecoveryContext.ForceRerunActions)

	child.Spec.RecoveryContext.ForceRerunActions[0] = "mutated"
	child.Spec.RecoveryContext.Relation[0] = 'X'
	assert.Equal(t, []string{"a3"}, parent.Spec.RecoveryContext.ForceRerunActions)
	assert.Equal(t, recoveryContextFor("source-run").Relation, parent.Spec.RecoveryContext.Relation)
}

func TestInheritRunContextFromParentTaskAction_NoRecoveryContextOnParent(t *testing.T) {
	child := &executorv1.TaskAction{Spec: executorv1.TaskActionSpec{}}
	inheritRunContextFromParentTaskAction(child, &executorv1.TaskAction{})
	assert.Nil(t, child.Spec.RecoveryContext)
}

// A condition action used to inherit nothing from its parent, so a subtree beneath one lost
// the run context entirely.
func TestEnqueueCondition_InheritsRunContextFromParent(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, executorv1.AddToScheme(scheme))

	interruptible := true
	parent := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "run1-a0",
			Namespace:   "flyte",
			Annotations: map[string]string{"owner": "sdk"},
			Labels:      map[string]string{"team": "platform"},
		},
		Spec: executorv1.TaskActionSpec{
			EnvVars:         map[string]string{"TRACE_ID": "abc123"},
			Interruptible:   &interruptible,
			RecoveryContext: recoveryContextFor("source-run", "a3"),
		},
	}
	c := &ActionsClient{
		recordedFilter: testFilter(),
		namespace:      "flyte",
		k8sClient: fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(parent).
			WithStatusSubresource(&executorv1.TaskAction{}).
			Build(),
	}

	action := newConditionAction(core.SimpleType_BOOLEAN)
	require.NoError(t, c.Enqueue(context.Background(), action, nil))

	created, err := c.GetTaskAction(context.Background(), action.ActionId)
	require.NoError(t, err)
	require.NotNil(t, created.Spec.RecoveryContext)
	assert.Equal(t, recoveryContextFor("source-run", "a3").Relation, created.Spec.RecoveryContext.Relation)
	assert.Equal(t, []string{"a3"}, created.Spec.RecoveryContext.ForceRerunActions)
	assert.Equal(t, "abc123", created.Spec.EnvVars["TRACE_ID"])
	require.NotNil(t, created.Spec.Interruptible)
	assert.True(t, *created.Spec.Interruptible)
	assert.Equal(t, "sdk", created.Annotations["owner"])
	assert.Equal(t, "platform", created.Labels["team"])
}
