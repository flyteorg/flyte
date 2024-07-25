package batchscheduler

import (
	"encoding/json"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Pod lebel
	Yunikorn             = "yunikorn"
	TaskGroupNameKey     = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey        = "yunikorn.apache.org/task-groups"
	TaskGroupPrarameters = "yunikorn.apache.org/schedulingPolicyParameters"
	TaskGroupGenericName = "task-group"
)

type YunikornGangSchedulingConfig struct {
	Annotations map[string]map[string]string
	Parameters string
}

func NewYunikornPlugin() *YunikornGangSchedulingConfig {
	return &YunikornGangSchedulingConfig{
		Annotations: nil,
	}
}

func (s *YunikornGangSchedulingConfig) GetSchedulerName() string { return Yunikorn }

func (s *YunikornGangSchedulingConfig) ParseJob(config *Config, metadata *metav1.ObjectMeta, workerGroupsSpec []*plugins.WorkerGroupSpec, pod *v1.PodSpec, primaryContainerIdx int) error {
	s.Annotations = nil
	s.Parameters = config.GetParameters()
	return s.BuildGangInfo(metadata, workerGroupsSpec, pod, primaryContainerIdx)
}

func (s *YunikornGangSchedulingConfig) ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec) {
	s.SetSchedulerName(head)
	s.AddGangSchedulingAnnotations(GenerateTaskGroupName(true, 0), metadata)
}

func (s *YunikornGangSchedulingConfig) ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int) {
	s.SetSchedulerName(worker)
	s.AddGangSchedulingAnnotations(GenerateTaskGroupName(false, index), metadata)
}

func (s *YunikornGangSchedulingConfig) AfterProcess(metadata *metav1.ObjectMeta) {
	RemoveGangSchedulingAnnotations(metadata)
}

type TaskGroup struct {
	Name                      string
	MinMember                 int32
	Labels                    map[string]string
	Annotations               map[string]string
	MinResource               v1.ResourceList
	NodeSelector              map[string]string
	Tolerations               []v1.Toleration
	Affinity                  *v1.Affinity
	TopologySpreadConstraints []v1.TopologySpreadConstraint
}

func GenerateTaskGroupName(master bool, index int) string {
	if master {
		return fmt.Sprintf("%s-%s", TaskGroupGenericName, "head")
	}
	return fmt.Sprintf("%s-%s-%d", TaskGroupGenericName, "worker", index)
}

func (s *YunikornGangSchedulingConfig) SetSchedulerName(spec *v1.PodSpec) {
	spec.SchedulerName = s.GetSchedulerName()
}

func (s *YunikornGangSchedulingConfig) BuildGangInfo(
	metadata *metav1.ObjectMeta,
	workerGroupsSpec []*plugins.WorkerGroupSpec,
	pod *v1.PodSpec,
	primaryContainerIdx int,
) error {
	// Parsing placeholders from the pod resource among head and workers
	s.Annotations = make(map[string]map[string]string, 0)
	TaskGroups := make([]TaskGroup, 0)
	headName := GenerateTaskGroupName(true, 0)
	TaskGroups = append(TaskGroups, TaskGroup{
		Name:                      headName,
		MinMember:                 1,
		Labels:                    metadata.Labels,
		Annotations:               metadata.Annotations,
		MinResource:               pod.Containers[primaryContainerIdx].Resources.Requests,
		NodeSelector:              pod.NodeSelector,
		Affinity:                  pod.Affinity,
		TopologySpreadConstraints: pod.TopologySpreadConstraints,
	})
	for index, spec := range workerGroupsSpec {
		name := GenerateTaskGroupName(false, index)
		tg := TaskGroup{
			Name:                      name,
			MinMember:                 spec.Replicas,
			Labels:                    metadata.Labels,
			Annotations:               metadata.Annotations,
			MinResource:               pod.Containers[primaryContainerIdx].Resources.Requests,
			NodeSelector:              pod.NodeSelector,
			Affinity:                  pod.Affinity,
			TopologySpreadConstraints: pod.TopologySpreadConstraints,
		}
		s.Annotations[name] = map[string]string{
			TaskGroupNameKey: name,
		}
		TaskGroups = append(TaskGroups, tg)
	}
	// Yunikorn head gang scheduling annotations
	var info []byte
	var err error
	if info, err = json.Marshal(TaskGroups); err != nil {
		s.Annotations = nil
		return err
	}
	headAnnotations := make(map[string]string, 0)
	headAnnotations[TaskGroupNameKey] = headName
	headAnnotations[TaskGroupsKey] = string(info[:])
	if len(s.Parameters) > 0 {
		headAnnotations[TaskGroupPrarameters] = s.Parameters
	}
	s.Annotations[headName] = headAnnotations
	return nil
}

func (s *YunikornGangSchedulingConfig) AddGangSchedulingAnnotations(name string, metadata *metav1.ObjectMeta) {
	if s.Annotations == nil {
		return
	}

	if _, ok := s.Annotations[name]; !ok {
		return
	}

	// Updating Yunikorn gang scheduling annotations
	annotations := s.Annotations[name]
	if _, ok := metadata.Annotations[TaskGroupNameKey]; !ok {
		if _, ok = annotations[TaskGroupNameKey]; ok {
			metadata.Annotations[TaskGroupNameKey] = annotations[TaskGroupNameKey]
		}
	}
	if _, ok := metadata.Annotations[TaskGroupsKey]; !ok {
		if _, ok = annotations[TaskGroupsKey]; ok {
			metadata.Annotations[TaskGroupsKey] = annotations[TaskGroupsKey]
		}
	}
	if _, ok := metadata.Annotations[TaskGroupPrarameters]; !ok {
		if _, ok = annotations[TaskGroupPrarameters]; ok {
			metadata.Annotations[TaskGroupPrarameters] = annotations[TaskGroupPrarameters]
		}
	}
}

func RemoveGangSchedulingAnnotations(metadata *metav1.ObjectMeta) {
	if metadata == nil {
		return
	}
	delete(metadata.Annotations, TaskGroupNameKey)
	delete(metadata.Annotations, TaskGroupsKey)
	delete(metadata.Annotations, TaskGroupPrarameters)
}
