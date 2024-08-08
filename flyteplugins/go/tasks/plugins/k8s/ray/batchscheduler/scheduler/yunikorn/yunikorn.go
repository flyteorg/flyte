package yunikorn

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/ray/batchscheduler/config"
)

const (
	// Pod lebel
	Yunikorn             = "yunikorn"
	TaskGroupNameKey     = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey        = "yunikorn.apache.org/task-groups"
	TaskGroupPrarameters = "yunikorn.apache.org/schedulingPolicyParameters"
)

type Plugin struct {
	Annotations map[string]map[string]string
	Parameters  string
}

func NewYunikornPlugin() *Plugin {
	return &Plugin{
		Annotations: nil,
		Parameters:  "",
	}
}

func (s *Plugin) GetSchedulerName() string { return Yunikorn }

func (s *Plugin) ParseJob(config *schedulerConfig.Config, metadata *metav1.ObjectMeta, workerGroupsSpec []*plugins.WorkerGroupSpec, pod *v1.PodSpec, primaryContainerIdx int) error {
	s.Annotations = nil
	if parameters := config.GetParameters(); len(parameters) > 0 {
		s.Parameters = parameters
	}
	return s.BuildGangInfo(metadata, workerGroupsSpec, pod, primaryContainerIdx)
}

func (s *Plugin) ProcessHead(metadata *metav1.ObjectMeta, head *v1.PodSpec, index int) {
	s.SetSchedulerName(head)
	s.AddGangSchedulingAnnotations(GenerateTaskGroupName(true, index), metadata)
}

func (s *Plugin) ProcessWorker(metadata *metav1.ObjectMeta, worker *v1.PodSpec, index int) {
	s.SetSchedulerName(worker)
	s.AddGangSchedulingAnnotations(GenerateTaskGroupName(false, index), metadata)
}

func (s *Plugin) AfterProcess(metadata *metav1.ObjectMeta) {
	if metadata == nil {
		return
	}
	delete(metadata.Annotations, TaskGroupNameKey)
	delete(metadata.Annotations, TaskGroupsKey)
	delete(metadata.Annotations, TaskGroupPrarameters)
}

func (s *Plugin) SetSchedulerName(spec *v1.PodSpec) {
	spec.SchedulerName = s.GetSchedulerName()
}

func (s *Plugin) BuildGangInfo(
	metadata *metav1.ObjectMeta,
	workerGroupsSpec []*plugins.WorkerGroupSpec,
	pod *v1.PodSpec,
	primaryContainerIdx int,
) error {
	if pod == nil {
		return errors.New("Ray gang scheduling: pod is nil")
	}
	// Parsing placeholders from the pod resource among head and workers
	var labels, annotations map[string]string = nil, nil
	if metadata != nil {
		labels = metadata.Labels
		annotations = metadata.Annotations
	}
	TaskGroups := make([]TaskGroup, 0)
	headName := GenerateTaskGroupName(true, 0)
	TaskGroups = append(TaskGroups, TaskGroup{
		Name:                      headName,
		MinMember:                 1,
		Labels:                    labels,
		Annotations:               annotations,
		MinResource:               pod.Containers[primaryContainerIdx].Resources.Requests,
		NodeSelector:              pod.NodeSelector,
		Affinity:                  pod.Affinity,
		TopologySpreadConstraints: pod.TopologySpreadConstraints,
	})

	s.Annotations = make(map[string]map[string]string, 0)
	for index, spec := range workerGroupsSpec {
		name := GenerateTaskGroupName(false, index)
		tg := TaskGroup{
			Name:                      name,
			MinMember:                 spec.Replicas,
			Labels:                    labels,
			Annotations:               annotations,
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
	info, _ = Marshal(TaskGroups)
	headAnnotations := make(map[string]string, 0)
	headAnnotations[TaskGroupNameKey] = headName
	headAnnotations[TaskGroupsKey] = string(info[:])
	if len(s.Parameters) > 0 {
		headAnnotations[TaskGroupPrarameters] = s.Parameters
	}
	s.Annotations[headName] = headAnnotations
	return nil
}

func (s *Plugin) AddGangSchedulingAnnotations(name string, metadata *metav1.ObjectMeta) {
	if s.Annotations == nil || metadata == nil {
		return
	}

	if _, ok := s.Annotations[name]; !ok {
		return
	}

	if metadata.Annotations == nil {
		metadata.Annotations = make(map[string]string, 0)
	}

	// Updating Yunikorn gang scheduling annotations
	annotations := s.Annotations[name]
	if _, ok := annotations[TaskGroupNameKey]; ok {
		metadata.Annotations[TaskGroupNameKey] = annotations[TaskGroupNameKey]
	}
	if _, ok := annotations[TaskGroupsKey]; ok {
		metadata.Annotations[TaskGroupsKey] = annotations[TaskGroupsKey]
	}
	if _, ok := metadata.Annotations[TaskGroupPrarameters]; !ok {
		if parameters, ok := annotations[TaskGroupPrarameters]; ok && len(parameters) > 0 {
			metadata.Annotations[TaskGroupPrarameters] = parameters
		}
	}
}
