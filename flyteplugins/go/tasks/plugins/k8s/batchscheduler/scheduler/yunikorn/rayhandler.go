package yunikorn

import (
	"encoding/json"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func MutateRayJob(parameters string, app *rayv1.RayJob) error {
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
			Name:      name,
			MinMember: *worker.Replicas,
			//Labels:                    meta.Labels,
			//Annotations:               meta.Annotations,
			MinResource: Allocation(spec.Containers),
			//NodeSelector:              spec.NodeSelector,
			//Affinity:                  spec.Affinity,
			//TopologySpreadConstraints: spec.TopologySpreadConstraints,
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
		//tmp, _ := json.Marshal(res2)
		//meta.Annotations["tmp"] = string(tmp)
		//tmp, _ = json.Marshal(Add(res, res2))
		//meta.Annotations["Sum"] = string(tmp)
		res = Add(res, res2)
	}
	TaskGroups[0] = TaskGroup{
		Name:      headName,
		MinMember: 1,
		//Labels:                    meta.Labels,
		//Annotations:               meta.Annotations,
		MinResource: res,
		//NodeSelector:              spec.NodeSelector,
		//Affinity:                  spec.Affinity,
		//TopologySpreadConstraints: spec.TopologySpreadConstraints,
	}
	meta.Annotations[TaskGroupNameKey] = headName
	info, err := json.Marshal(TaskGroups)
	if err != nil {
		return err
	}
	meta.Annotations[TaskGroupsKey] = string(info[:])
	meta.Annotations[TaskGroupParameters] = parameters
	meta.Annotations[AppID] = appID
	return nil
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

func Add(a v1.ResourceList, b v1.ResourceList) v1.ResourceList {
	result := a
	for name, value := range a {
		sum := &value
		if value2, ok := b[name]; ok {
			sum.Add(value2)
			result[name] = *sum
		} else {
			result[name] = value
		}
	}
	for name, value := range b {
		if _, ok := a[name]; !ok {
			result[name] = value
		}
	}
	return result
}
