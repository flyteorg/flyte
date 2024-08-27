package yunikorn

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
)

func ProcessRay(paras string, app *rayv1.RayJob) error {
	jobname :=  GenerateTaskGroupName(true, 0)
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
			//Labels:                    meta.Labels,
			//Annotations:               meta.Annotations,
			MinResource:               Allocation(spec.Containers),
			//NodeSelector:              spec.NodeSelector,
			//Affinity:                  spec.Affinity,
			//TopologySpreadConstraints: spec.TopologySpreadConstraints,
		})
		meta.Annotations[TaskGroupNameKey] =  name
		meta.Annotations[AppID] = jobname
	}
	headSpec := &appSpec.HeadGroupSpec
	headSpec.Template.Spec.SchedulerName = Yunikorn
	meta := headSpec.Template.ObjectMeta
	spec := headSpec.Template.Spec
	headName := GenerateTaskGroupName(true, 0)
	TaskGroups[0] = TaskGroup{
		Name:                      headName,
		MinMember:                 1,
		//Labels:                    meta.Labels,
		//Annotations:               meta.Annotations,
		MinResource:               Allocation(spec.Containers),
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
	meta.Annotations[TaskGroupPrarameters] = paras
	meta.Annotations[AppID] = jobname
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