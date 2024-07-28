package yunikorn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/ray/batchscheduler/config"
)

var (
	res = v1.ResourceList{
		"cpu":    resource.MustParse("500m"),
		"memory": resource.MustParse("1Gi"),
	}
)

func TestParseJob(t *testing.T) {
	type inputFormat struct {
		config         *schedulerConfig.Config
		metadata       *metav1.ObjectMeta
		workerGroupNum int
		podSpec        *v1.PodSpec
		index          int
	}
	type expectFormat struct {
		raiseErr   bool
		parameters string
		taskGroups []TaskGroup
	}
	var tests = []struct {
		input  inputFormat
		expect expectFormat
	}{
		{
			input: inputFormat{
				config: &schedulerConfig.Config{
					Scheduler:  "yunikorn",
					Parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				},
				workerGroupNum: 1,
				podSpec:        nil,
				metadata:       &metav1.ObjectMeta{},
				index:          0,
			},
			expect: expectFormat{
				raiseErr:   true,
				parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				taskGroups: []TaskGroup{
					{
						Name:                      GenerateTaskGroupName(true, 0),
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
						Name:                      GenerateTaskGroupName(false, 0),
						MinMember:                 int32(1),
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
		},
		{
			input: inputFormat{
				config: &schedulerConfig.Config{
					Scheduler:  "yunikorn",
					Parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				},
				workerGroupNum: 1,
				podSpec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: res,
							},
						},
					},
					NodeSelector:              nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
				index: 0,
			},
			expect: expectFormat{
				raiseErr:   false,
				parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				taskGroups: []TaskGroup{
					{
						Name:                      GenerateTaskGroupName(true, 0),
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
						Name:                      GenerateTaskGroupName(false, 0),
						MinMember:                 int32(1),
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
		},
	}
	for _, tt := range tests {
		t.Run("Yunikorn parse job", func(t *testing.T) {
			workersSpec := make([]*plugins.WorkerGroupSpec, 0)
			for index := 0; index < tt.input.workerGroupNum; index++ {
				count := 1 * (1 + index)
				max := 2 * (1 + index)
				workersSpec = append(workersSpec, &plugins.WorkerGroupSpec{
					Replicas:    int32(count),
					MinReplicas: int32(count),
					MaxReplicas: int32(max),
				})
			}
			p := NewYunikornPlugin()
			err := p.ParseJob(tt.input.config, tt.input.metadata, workersSpec, tt.input.podSpec, tt.input.index)
			if tt.expect.raiseErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, Yunikorn, p.GetSchedulerName())
				names := []string{GenerateTaskGroupName(true, 0)}
				for index := 0; index < tt.input.workerGroupNum; index++ {
					names = append(names, GenerateTaskGroupName(false, index))
				}
				// task-groups among head and workers
				assert.Equal(t, len(names), len(p.Annotations))
				// check head annotations
				head := p.Annotations[names[0]]
				assert.Equal(t, names[0], head[TaskGroupNameKey])
				assert.Equal(t, tt.expect.parameters, head[TaskGroupPrarameters])
				// task-groups in head
				var taskgroups []TaskGroup
				err = json.Unmarshal([]byte(head[TaskGroupsKey]), &taskgroups)
				assert.Nil(t, err)
				assert.Equal(t, len(names), len(taskgroups))
				for index, tg := range taskgroups {
					assert.Equal(t, names[index], tg.Name)
				}
			}
		})
	}
}

func TestProcessHead(t *testing.T) {
	type inputFormat struct {
		config         *schedulerConfig.Config
		metadata       *metav1.ObjectMeta
		workerGroupNum int
		podSpec        *v1.PodSpec
		index          int
	}
	type expectFormat struct {
		name          string
		taskgroupsNum int
		parameters    string
	}
	var tests = []struct {
		input  inputFormat
		expect expectFormat
	}{
		{
			input: inputFormat{
				config: &schedulerConfig.Config{
					Scheduler:  "yunikorn",
					Parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				},
				workerGroupNum: 1,
				podSpec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: res,
							},
						},
					},
					NodeSelector:              nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
				index: 0,
			},
			expect: expectFormat{
				name:          GenerateTaskGroupName(true, 0),
				taskgroupsNum: 2,
				parameters:    "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
			},
		},
	}
	for _, tt := range tests {
		t.Run("Yunikorn process head", func(t *testing.T) {
			workersSpec := make([]*plugins.WorkerGroupSpec, 0)
			for index := 0; index < tt.input.workerGroupNum; index++ {
				workersSpec = append(workersSpec, &plugins.WorkerGroupSpec{
					Replicas:    int32(1),
					MinReplicas: int32(1),
					MaxReplicas: int32(2),
				})
			}
			p := NewYunikornPlugin()
			err := p.ParseJob(tt.input.config, tt.input.metadata, workersSpec, tt.input.podSpec, tt.input.index)
			assert.Nil(t, err)
			p.ProcessHead(tt.input.metadata, tt.input.podSpec, tt.input.index)
			assert.Equal(t, Yunikorn, tt.input.podSpec.SchedulerName)
			assert.Equal(t, tt.expect.name, tt.input.metadata.Annotations[TaskGroupNameKey])
			assert.Equal(t, tt.expect.parameters, tt.input.metadata.Annotations[TaskGroupPrarameters])
			var taskgroups []TaskGroup
			err = json.Unmarshal([]byte(tt.input.metadata.Annotations[TaskGroupsKey]), &taskgroups)
			assert.Nil(t, err)
			assert.Equal(t, tt.expect.taskgroupsNum, len(taskgroups))
		})
	}
}

func TestProcessWorker(t *testing.T) {
	type inputFormat struct {
		config         *schedulerConfig.Config
		metadata       *metav1.ObjectMeta
		workerGroupNum int
		podSpec        *v1.PodSpec
		index          int
	}
	type expectFormat struct {
		name          string
		taskgroupsNum int
	}
	var tests = []struct {
		input  inputFormat
		expect expectFormat
	}{
		{
			input: inputFormat{
				config: &schedulerConfig.Config{
					Scheduler:  "yunikorn",
					Parameters: "placeholderTimeoutInSeconds=15 gangSchedulingStyle=Soft",
				},
				workerGroupNum: 1,
				podSpec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: res,
							},
						},
					},
					NodeSelector:              nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
				index: 0,
			},
			expect: expectFormat{
				name:          GenerateTaskGroupName(false, 0),
				taskgroupsNum: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run("Yunikorn process worker", func(t *testing.T) {
			workersSpec := make([]*plugins.WorkerGroupSpec, 0)
			for index := 0; index < tt.input.workerGroupNum; index++ {
				workersSpec = append(workersSpec, &plugins.WorkerGroupSpec{
					Replicas:    int32(1),
					MinReplicas: int32(1),
					MaxReplicas: int32(2),
				})
			}
			p := NewYunikornPlugin()
			err := p.ParseJob(tt.input.config, tt.input.metadata, workersSpec, tt.input.podSpec, tt.input.index)
			assert.Nil(t, err)
			p.ProcessWorker(tt.input.metadata, tt.input.podSpec, tt.input.index)
			assert.Equal(t, Yunikorn, tt.input.podSpec.SchedulerName)
			assert.Equal(t, tt.expect.name, tt.input.metadata.Annotations[TaskGroupNameKey])
		})
	}
}

func TestAfterProcess(t *testing.T) {
	type expectFormat struct {
		isNil  bool
		length int
	}
	var tests = []struct {
		input  *metav1.ObjectMeta
		expect expectFormat
	}{
		{
			input:  nil,
			expect: expectFormat{isNil: true, length: -1},
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"others":             "extra",
					TaskGroupNameKey:     "TGName",
					TaskGroupsKey:        "TGs",
					TaskGroupPrarameters: "parameters",
				},
			},
			expect: expectFormat{isNil: false, length: 1},
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey:     "TGName",
					TaskGroupsKey:        "TGs",
					TaskGroupPrarameters: "parameters",
				},
			},
			expect: expectFormat{isNil: false, length: 0},
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey: "TGName",
					TaskGroupsKey:    "TGs",
				},
			},
			expect: expectFormat{isNil: false, length: 0},
		},
		{
			input: &metav1.ObjectMeta{
				Annotations: map[string]string{
					TaskGroupNameKey: "TGName",
				},
			},
			expect: expectFormat{isNil: false, length: 0},
		},
		{
			input:  &metav1.ObjectMeta{},
			expect: expectFormat{isNil: false, length: 0},
		},
	}
	for _, tt := range tests {
		t.Run("Remove Gang scheduling labels", func(t *testing.T) {
			p := NewYunikornPlugin()
			p.AfterProcess(tt.input)
			if tt.expect.isNil {
				assert.Nil(t, tt.input)
			} else {
				assert.NotNil(t, tt.input)
				assert.Equal(t, tt.expect.length, len(tt.input.Annotations))
			}
		})
	}
}

func TestSetSchedulerName(t *testing.T) {
	t.Run("Set Scheduler Name", func(t *testing.T) {
		p := NewYunikornPlugin()
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: res,
					},
				},
			},
			NodeSelector:              nil,
			Affinity:                  nil,
			TopologySpreadConstraints: nil,
		}
		p.SetSchedulerName(podSpec)
		assert.Equal(t, p.GetSchedulerName(), podSpec.SchedulerName)
		podSpec.SchedulerName = ""
	})
}

func TestBuildGangInfo(t *testing.T) {
	names := []string{GenerateTaskGroupName(true, 0)}
	for index := 0; index < 2; index++ {
		names = append(names, GenerateTaskGroupName(false, index))
	}
	type inputFormat struct {
		workerGroupNum int
		podSpec        *v1.PodSpec
		metadata       *metav1.ObjectMeta
	}
	var tests = []struct {
		input      inputFormat
		taskGroups []TaskGroup
	}{
		{
			input: inputFormat{
				workerGroupNum: 1,
				podSpec:        nil,
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
			},
			taskGroups: nil,
		},
		{
			input: inputFormat{
				workerGroupNum: 1,
				podSpec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: res,
							},
						},
					},
					NodeSelector:              nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
			},
			taskGroups: []TaskGroup{
				{
					Name:                      GenerateTaskGroupName(true, 0),
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
					Name:                      GenerateTaskGroupName(false, 0),
					MinMember:                 int32(1),
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
		{
			input: inputFormat{
				workerGroupNum: 2,
				podSpec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: res,
							},
						},
					},
					NodeSelector:              nil,
					Affinity:                  nil,
					TopologySpreadConstraints: nil,
				},
				metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{"others": "extra"},
				},
			},
			taskGroups: []TaskGroup{
				{
					Name:                      GenerateTaskGroupName(true, 0),
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
					Name:                      GenerateTaskGroupName(false, 0),
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
					Name:                      GenerateTaskGroupName(false, 1),
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
			for index := 0; index < tt.input.workerGroupNum; index++ {
				count := 1 * (1 + index)
				max := 2 * (1 + index)
				workersSpec = append(workersSpec, &plugins.WorkerGroupSpec{
					Replicas:    int32(count),
					MinReplicas: int32(count),
					MaxReplicas: int32(max),
				})
			}
			p := NewYunikornPlugin()
			if err := p.BuildGangInfo(tt.input.metadata, workersSpec, tt.input.podSpec, 0); tt.input.podSpec == nil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				// test worker name
				for index := 0; index < tt.input.workerGroupNum; index++ {
					name := GenerateTaskGroupName(false, index)
					if annotations, ok := p.Annotations[name]; ok {
						assert.Equal(t, 1, len(annotations))
						assert.Equal(t, name, annotations[TaskGroupNameKey])
					} else {
						t.Errorf("Worker group %d annotatiosn miss", index)
					}
				}
				// Test head name and groups
				headName := GenerateTaskGroupName(true, 0)
				if annotations, ok := p.Annotations[headName]; ok {
					info, err := json.Marshal(tt.taskGroups)
					assert.Nil(t, err)
					assert.Equal(t, 2, len(annotations))
					assert.Equal(t, headName, annotations[TaskGroupNameKey])
					assert.Equal(t, string(info[:]), annotations[TaskGroupsKey])
				} else {
					t.Error("Head annotations miss")
				}
			}
		})
	}
}

func TestAddGangSchedulingAnnotations(t *testing.T) {
	taskGroupsAnnotations := map[string]map[string]string{
		GenerateTaskGroupName(true, 0): {
			TaskGroupNameKey:     GenerateTaskGroupName(true, 0),
			TaskGroupsKey:        "TGs",
			TaskGroupPrarameters: "parameters",
		},
		GenerateTaskGroupName(false, 0): {
			TaskGroupNameKey: GenerateTaskGroupName(false, 0),
		},
	}
	type inputFormat struct {
		annotations map[string]map[string]string
		metadata    *metav1.ObjectMeta
		name        string
	}
	var tests = []struct {
		input  inputFormat
		expect *metav1.ObjectMeta
	}{
		{
			input: inputFormat{
				annotations: nil,
				metadata:    nil,
				name:        "",
			},
			expect: nil,
		},
		{
			input: inputFormat{
				annotations: taskGroupsAnnotations,
				metadata:    nil,
				name:        "",
			},
			expect: nil,
		},
		{
			input: inputFormat{
				annotations: taskGroupsAnnotations,
				metadata:    &metav1.ObjectMeta{},
				name:        "Unknown",
			},
			expect: &metav1.ObjectMeta{},
		},
		{
			input: inputFormat{
				annotations: taskGroupsAnnotations,
				metadata:    &metav1.ObjectMeta{},
				name:        GenerateTaskGroupName(true, 0),
			},
			expect: &metav1.ObjectMeta{
				Annotations: taskGroupsAnnotations[GenerateTaskGroupName(true, 0)],
			},
		},
		{
			input: inputFormat{
				annotations: taskGroupsAnnotations,
				metadata:    &metav1.ObjectMeta{},
				name:        GenerateTaskGroupName(false, 0),
			},
			expect: &metav1.ObjectMeta{
				Annotations: taskGroupsAnnotations[GenerateTaskGroupName(false, 0)],
			},
		},
	}
	for _, tt := range tests {
		t.Run("Check gang scheduling annotatiosn after labeling", func(t *testing.T) {
			p := NewYunikornPlugin()
			p.Annotations = tt.input.annotations
			p.AddGangSchedulingAnnotations(tt.input.name, tt.input.metadata)
			if tt.expect == nil {
				assert.Nil(t, tt.expect, tt.input.metadata)
			} else {
				assert.Equal(t, tt.expect.Annotations, tt.input.metadata.Annotations)
			}
		})
	}
}
