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
	BatchSchedulerLabel  = "batch-scheduler"
	SchedulerLabel       = "scheduler"
	SchedulerName        = "yunikorn"
	TaskGroupNameKey     = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey        = "yunikorn.apache.org/task-groups"
	TaskGroupPrarameters = "yunikorn.apache.org/schedulingPolicyParameters"
	TaskGroupGenericName = "task-group"
)

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

func SetSchedulerNameAndBuildGangInfo(config BatchSchedulerConfig, metadata *metav1.ObjectMeta, workerGroupsSpec []*plugins.WorkerGroupSpec, head, worker *v1.PodSpec) (map[string]map[string]string, error) {
	if config.Scheduler != SchedulerName {
		return nil, nil
	}
	head.SchedulerName = SchedulerName
	worker.SchedulerName = SchedulerName

	TaskGroupsAnnotations := make(map[string]map[string]string, 0)
	// Parsing placeholders from the pod resource among head and workers
	TaskGroups := make([]TaskGroup, 0)
	headName := GenerateTaskGroupName(true, 0)
	TaskGroups = append(TaskGroups, TaskGroup{
		Name:                      headName,
		MinMember:                 1,
		Labels:                    metadata.Labels,
		Annotations:               metadata.Annotations,
		MinResource:               head.Containers[0].Resources.Requests,
		NodeSelector:              head.NodeSelector,
		Affinity:                  head.Affinity,
		TopologySpreadConstraints: head.TopologySpreadConstraints,
	})

	for index, spec := range workerGroupsSpec {
		name := GenerateTaskGroupName(false, index)
		tg := TaskGroup{
			Name:                      name,
			MinMember:                 spec.Replicas,
			Labels:                    metadata.Labels,
			Annotations:               metadata.Annotations,
			MinResource:               worker.Containers[0].Resources.Requests,
			NodeSelector:              worker.NodeSelector,
			Affinity:                  worker.Affinity,
			TopologySpreadConstraints: worker.TopologySpreadConstraints,
		}
		TaskGroupsAnnotations[name] = map[string]string{
			TaskGroupNameKey: name,
		}
		TaskGroups = append(TaskGroups, tg)
	}

	// Yunikorn head gang scheduling annotations
	info, err := json.Marshal(TaskGroups)
	if err != nil {
		return nil, err
	}
	headAnnotations := make(map[string]string, 0)
	headAnnotations[TaskGroupNameKey] = headName
	headAnnotations[TaskGroupsKey] = string(info[:])
	headAnnotations[TaskGroupPrarameters] = config.Parameters
	TaskGroupsAnnotations[headName] = headAnnotations
	return TaskGroupsAnnotations, nil
}

func AddGangSchedulingAnnotations(name string, metadata *metav1.ObjectMeta, TGAnnotations map[string]map[string]string) {
	if TGAnnotations == nil {
		return
	}

	if _, ok := TGAnnotations[name]; !ok {
		return
	}

	annotations := TGAnnotations[name]
	if _, ok := metadata.Annotations[TaskGroupNameKey]; !ok {
		metadata.Annotations[TaskGroupNameKey] = annotations[TaskGroupNameKey]
	}
	if _, ok := metadata.Annotations[TaskGroupsKey]; !ok {
		metadata.Annotations[TaskGroupsKey] = annotations[TaskGroupsKey]
	}
	if _, ok := metadata.Annotations[TaskGroupPrarameters]; !ok {
		if _, ok = annotations[TaskGroupPrarameters]; !ok {
			return
		}
		metadata.Annotations[TaskGroupPrarameters] = annotations[TaskGroupPrarameters]
	}
	return
}

func RemoveGangSchedulingAnnotations(metadata *metav1.ObjectMeta) {
	if _, ok := metadata.Annotations[TaskGroupNameKey]; ok {
		delete(metadata.Annotations, TaskGroupNameKey)
	}
	if _, ok := metadata.Annotations[TaskGroupsKey]; ok {
		delete(metadata.Annotations, TaskGroupsKey)
	}
	if _, ok := metadata.Annotations[TaskGroupPrarameters]; ok {
		delete(metadata.Annotations, TaskGroupPrarameters)
	}
	return
}
