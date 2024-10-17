package yunikorn

import (
	"encoding/json"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/utils"
)

const (
	Yunikorn            = "yunikorn"
	AppID               = "yunikorn.apache.org/app-id"
	Queue               = "yunikorn.apache.org/queue"
	TaskGroupNameKey    = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey       = "yunikorn.apache.org/task-groups"
	TaskGroupParameters = "yunikorn.apache.org/schedulingPolicyParameters"
)

func MutateRayJob(app *rayv1.RayJob) error {
	appID := GenerateTaskGroupAppID()
	rayjobSpec := &app.Spec
	appSpec := rayjobSpec.RayClusterSpec
	TaskGroups := make([]TaskGroup, 1)
	for index := range appSpec.WorkerGroupSpecs {
		worker := &appSpec.WorkerGroupSpecs[index]
		worker.Template.Spec.SchedulerName = Yunikorn
		meta := worker.Template.ObjectMeta
		spec := worker.Template.Spec
		name := GenerateTaskGroupName(false, index)
		TaskGroups = append(TaskGroups, TaskGroup{
			Name:                      name,
			MinMember:                 *worker.Replicas,
			Labels:                    meta.Labels,
			Annotations:               meta.Annotations,
			MinResource:               Allocation(spec.Containers),
			NodeSelector:              spec.NodeSelector,
			Affinity:                  spec.Affinity,
			TopologySpreadConstraints: spec.TopologySpreadConstraints,
		})
		meta.Annotations[TaskGroupNameKey] = name
		meta.Annotations[AppID] = appID
	}
	headSpec := &appSpec.HeadGroupSpec
	headSpec.Template.Spec.SchedulerName = Yunikorn
	meta := headSpec.Template.ObjectMeta
	spec := headSpec.Template.Spec
	headName := GenerateTaskGroupName(true, 0)
	res := Allocation(spec.Containers)
	if ok := *appSpec.EnableInTreeAutoscaling; ok {
		res2 := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		}
		res = Add(res, res2)
	}
	TaskGroups[0] = TaskGroup{
		Name:                      headName,
		MinMember:                 1,
		Labels:                    meta.Labels,
		Annotations:               meta.Annotations,
		MinResource:               res,
		NodeSelector:              spec.NodeSelector,
		Affinity:                  spec.Affinity,
		TopologySpreadConstraints: spec.TopologySpreadConstraints,
	}
	meta.Annotations[TaskGroupNameKey] = headName
	info, err := json.Marshal(TaskGroups)
	if err != nil {
		return err
	}
	meta.Annotations[TaskGroupsKey] = string(info[:])
	meta.Annotations[AppID] = appID
	return nil
}

func UpdateGangSchedulingParameters(parameters string, objectMeta *metav1.ObjectMeta) {
	if len(parameters) == 0 {
		return
	}
	utils.UpdateAnnotations(
		map[string]string{TaskGroupParameters: parameters},
		objectMeta,
	)
}

func UpdateAnnotations(labels map[string]string, app *rayv1.RayJob) {
	appSpec := app.Spec.RayClusterSpec
	headSpec := appSpec.HeadGroupSpec
	utils.UpdatePodTemplateAnnotatations(labels, &headSpec.Template)
	for index := range appSpec.WorkerGroupSpecs {
		worker := appSpec.WorkerGroupSpecs[index]
		utils.UpdatePodTemplateAnnotatations(labels, &worker.Template)
	}
}

func Allocation(containers []v1.Container) v1.ResourceList {
	totalResources := v1.ResourceList{}
	for _, c := range containers {
		for name, q := range c.Resources.Limits {
			if _, exists := totalResources[name]; !exists {
				totalResources[name] = q.DeepCopy()
				continue
			}
			total := totalResources[name]
			total.Add(q)
			totalResources[name] = total
		}
	}
	return totalResources
}

func Add(left v1.ResourceList, right v1.ResourceList) v1.ResourceList {
	result := left
	for name, value := range left {
		sum := value
		if value2, ok := right[name]; ok {
			sum.Add(value2)
			result[name] = sum
		} else {
			result[name] = value
		}
	}
	for name, value := range right {
		if _, ok := left[name]; !ok {
			result[name] = value
		}
	}
	return result
}
