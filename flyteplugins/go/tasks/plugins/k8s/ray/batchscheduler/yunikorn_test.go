package batchscheduler

import (
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	podSpec = &v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						"cpu":    resource.MustParse("500m"),
						"memory": resource.MustParse("1Gi"),
					},
				},
			},
		},
		NodeSelector:              nil,
		Affinity:                  nil,
		TopologySpreadConstraints: nil,
	}
)

func TestSetSchedulerName(t *testing.T) {
	t.Run("Set Scheduler Name", func(t *testing.T) {
		p := NewYunikornPlugin()
		p.SetSchedulerName(podSpec)
		assert.Equal(t, p.GetSchedulerName(), podSpec.SchedulerName)
		podSpec.SchedulerName = ""
	})
}

func TestBuildGangInfo(t *testing.T) {
	names := []string{GenerateTaskGroupName(true, 0)}
	res := v1.ResourceList{
		"cpu":    resource.MustParse("500m"),
		"memory": resource.MustParse("1Gi"),
	}
	for index := 0; index < 2; index++ {
		names = append(names, GenerateTaskGroupName(false, index))
	}
	var tests = []struct {
		workerGroupNum int
		taskGroups     []TaskGroup
	}{
		{
			workerGroupNum: 2,
			taskGroups: []TaskGroup{
				{
					Name:                      names[0],
					MinMember:                 int32(1),
					Labels:                    nil,
					Annotations:               map[string]string{"others": "extra"},
					MinResource:               res,
					NodeSelector:              nil,
					Tolerations:               nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				{
					Name:                      names[1],
					MinMember:                 int32(1),
					Labels:                    nil,
					Annotations:               map[string]string{"others": "extra"},
					MinResource:               res,
					NodeSelector:              nil,
					Tolerations:               nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				{
					Name:                      names[2],
					MinMember:                 int32(2),
					Labels:                    nil,
					Annotations:               map[string]string{"others": "extra"},
					MinResource:               res,
					NodeSelector:              nil,
					Tolerations:               nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("Create Yunikorn gang scheduling annotations", func(t *testing.T) {
			workersSpec := make([]*plugins.WorkerGroupSpec, 0)
			for index := 0; index < tt.workerGroupNum; index++ {
				count := 1 * (1 + index)
				max := 2 * (1 + index)
				workersSpec = append(workersSpec, &plugins.WorkerGroupSpec{
					Replicas:    int32(count),
					MinReplicas: int32(count),
					MaxReplicas: int32(max),
				})
			}
			metadata := &metav1.ObjectMeta{
				Annotations: map[string]string{"others": "extra"},
			}
			p := NewYunikornPlugin()
			err := p.BuildGangInfo(metadata, workersSpec, podSpec, 0)
			assert.Nil(t, err)
			// test worker name
			for index := 0; index < tt.workerGroupNum; index++ {
				workerIndex := index + 1
				name := names[workerIndex]
				if annotations, ok := p.Annotations[name]; ok {
					assert.Equal(t, 1, len(annotations))
					assert.Equal(t, name, annotations[TaskGroupNameKey])
				} else {
					t.Errorf("Worker group %d annotatiosn miss", index)
				}
			}
			// Test head name and groups
			headName := names[0]
			if annotations, ok := p.Annotations[headName]; ok {
				info, err := json.Marshal(tt.taskGroups)
				assert.Nil(t, err)
				assert.Equal(t, 2, len(annotations))
				assert.Equal(t, headName, annotations[TaskGroupNameKey])
				assert.Equal(t, string(info[:]), annotations[TaskGroupsKey])
			} else {
				t.Error("Head annotations miss")
			}
		})
	}
}

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
			assert.Equal(t, tt.expect, GenerateTaskGroupName(tt.master, tt.index))
		})
	}
}

func TestRemoveGangSchedulingAnnotations(t *testing.T) {
	var tests = []struct {
		input  *metav1.ObjectMeta
		expect int
	}{
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"others":             "extra",
					TaskGroupNameKey:     "TGName",
					TaskGroupsKey:        "TGs",
					TaskGroupPrarameters: "parameters",
				},
			},
			expect: 1,
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey:     "TGName",
					TaskGroupsKey:        "TGs",
					TaskGroupPrarameters: "parameters",
				},
			},
			expect: 0,
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey: "TGName",
					TaskGroupsKey:    "TGs",
				},
			},
			expect: 0,
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey: "TGName",
				},
			},
			expect: 0,
		},
		{
			input:  &metav1.ObjectMeta{},
			expect: 0,
		},
	}
	for _, tt := range tests {
		t.Run("Remove Gang scheduling labels", func(t *testing.T) {
			RemoveGangSchedulingAnnotations(tt.input)
			assert.Equal(t, tt.expect, len(tt.input.Annotations))
		})
	}
}
