package demo

import (
	"context"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var fakePod = corev1.Pod{
	Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{},
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:   "dummyflytepod",
		Labels: map[string]string{"app": "flyte"},
	},
}

func TestDemoReload(t *testing.T) {
	ctx := context.Background()
	commandCtx := cmdCore.CommandContext{}

	t.Run("No errors", func(t *testing.T) {
		client := testclient.NewSimpleClientset()
		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		k8s.Client = client
		err = reloadDemoCluster(ctx, []string{}, commandCtx)
		assert.NoError(t, err)
	})

	t.Run("Multiple pods will error", func(t *testing.T) {
		client := testclient.NewSimpleClientset()
		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		fakePod.SetName("othername")
		_, err = client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		k8s.Client = client
		err = reloadDemoCluster(ctx, []string{}, commandCtx)
		assert.Errorf(t, err, "should only have one pod")
	})
}
