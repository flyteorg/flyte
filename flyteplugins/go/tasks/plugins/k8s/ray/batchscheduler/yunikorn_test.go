package batchscheduler

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
)

var (
	podSpec = &v1.PodSpec{
		Containers: []v1.Container{
			v1.Container{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList {
						"cpu": resource.MustParse("500m"),
						"memory": resource.MustParse("1Gi"),
					},
				},
			},
		},
		NodeSelector: nil,
		Affinity: nil,
		TopologySpreadConstraints: nil,
	}
	rayWorkersSpec = []*plugins.WorkerGroupSpec{
		&plugins.WorkerGroupSpec{
			GroupName: "group1",
			Replicas: int32(1),
			MinReplicas: int32(1),
			MaxReplicas: int32(2),
			RayStartParams: nil,
		},
		&plugins.WorkerGroupSpec{
			GroupName: "group2",
			Replicas: int32(1),
			MinReplicas: int32(1),
			MaxReplicas: int32(2),
			RayStartParams: nil,
		},
	}
)

func TestGenerateTaskGroupName(t *testing.T) {
	var tests = []struct {
		master bool
		index  int
		expect string
	}{
		{true, 0, GenerateTaskGroupName(true, 0)},
		{false, 0, GenerateTaskGroupName(false, 0)},
		{false, 1, GenerateTaskGroupName(false, 1)},
	}
	for _, tt := range tests {
		t.Run("Generating Task group name", func(t *testing.T) {
			if got := GenerateTaskGroupName(tt.master, tt.index); got != tt.expect {
				t.Errorf("got %s, expect %s", got, tt.expect)
			}
		})
	}
}

func TestSetSchedulerName(t *testing.T) {
	head := podSpec.DeepCopy()
	worker := podSpec.DeepCopy()
	var tests = []struct {
		schedulerConfig BatchSchedulerConfig
		expect string
	}{
		{schedulerConfig: BatchSchedulerConfig{Scheduler:"", Parameters:""}, expect: ""},
		{schedulerConfig: BatchSchedulerConfig{Scheduler:SchedulerName, Parameters:"gangSchedulingStyle=Hard"}, expect: SchedulerName},
		{schedulerConfig: BatchSchedulerConfig{Scheduler:"other", Parameters:""}, expect: ""},
	}
	for _, tt := range tests {
		t.Run("Scheduler Name", func(t *testing.T) {
			SetSchedulerNameAndBuildGangInfo(
				tt.schedulerConfig,
				&metav1.ObjectMeta{
					Labels: map[string]string{},
					Annotations: map[string]string{},
				},
				rayWorkersSpec,
				head,
				worker,
			)
			if got := head.SchedulerName; got != tt.expect {
				t.Errorf("head pod scheduler name: expect %s, got %s", tt.expect, got)
			}
			if got := worker.SchedulerName; got != tt.expect {
				t.Errorf("worker pod scheduler name: expect %s, got %s", tt.expect, got)
			}
			head.SchedulerName = ""
			worker.SchedulerName = ""
		})
	}
}

func TestRemoveGangSchedulingAnnotations(t *testing.T) {
	var tests = []struct {
		input  *metav1.ObjectMeta
		expect int
	}{
		{input: &metav1.ObjectMeta{Annotations: map[string]string{"others": "extra", TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs", TaskGroupPrarameters: "parameters"}}, expect: 1},
		{input: &metav1.ObjectMeta{Annotations: map[string]string{TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs", TaskGroupPrarameters: "parameters"}}, expect: 0},
		{input: &metav1.ObjectMeta{Annotations: map[string]string{TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs"}}, expect: 0},
		{input: &metav1.ObjectMeta{Annotations: map[string]string{TaskGroupNameKey: "TGName"}}, expect: 0},
		{input: &metav1.ObjectMeta{}, expect: 0},
	}
	for _, tt := range tests {
		t.Run("Remove Gang scheduling labels", func(t *testing.T) {
			RemoveGangSchedulingAnnotations(tt.input)
			if got := len(tt.input.Annotations); got != tt.expect {
				t.Errorf("got %d, expect %d", got, tt.expect)
			}
		})
	}
}
