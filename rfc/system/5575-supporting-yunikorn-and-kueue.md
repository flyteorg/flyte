# [Newbie] Supporting Yunikorn and Kueue

**Authors:**

- @yuteng

## 1 Executive Summary

Providing kubernetes (k8s) resource management, gang scheduling and preemption for flyte applications by third-party software, including Apache Yunikorn and Kueue.

## 2 Motivation

Flyte support multi-tenancy and various k8s plugins.

Kueue and Yunikorn support gang scheduling and preemption.
Gang scheduling guarantees the availability of certain K8s crd services, such as Spark, Ray, with sufficient resource and preemption make sure high priority task execute immediately.

Flyte doesn't provide resource management for multi-tenancy, which hierarchical resource queues of Yunikorn can solve.

## 3 Proposed Implementation

Kueue

```yaml
queueconfig:
  scheduler: yunikorn
  jobs:
    - type: "ray"
      gangscheduling: "placeholderTimeoutInSeconds=60 gangSchedulingStyle=hard"
      allow-preemption: false
    - type: "spark"
      gangscheduling: "placeholderTimeoutInSeconds=30 gangSchedulingStyle=hard"
      allow-preemption: true
```

Mentioned configuration indicates what queues exist for an organization.
Hierarchical queues will be structured as follows.
root.organization1.ray、root.organization1.spark and root.organization1.default".

ResourceFlavor allocates resource based on labels which indicates that category-based resource allocation by organization label is available.
Thus, a clusterQueue including multiple resources represents the total acessaible resource for an organization.  
| clusterQueue | localQueue |
| --- | --- |
| <organization name> | ray、spark、default |
A tenant can submit organization-specific tasks to queues such as organization.ray, organization.spark and organization.default to track which job types are submittable. 


It patches labels or annotations on k8s resources after they pass rules specified in the configuration.

```go
type SchedulePlugin interface {
  CreateLabels(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig)
  CreateGroupLabels(ctx context.Context, object client.Object, taskTmpl *core.TaskTemplate)
  GetGroupLabels() (labels, annotations map[string]string)
}

type YunikornScheduablePlugin struct {
	jobs map[string]string
  Labels map[string]string
  Annotations map[string]string
}

func (yk *YunikornSchedulPlugin) GetGroupLabels() (labels, annotations map[string]string) {
  return yk.Labels, yk.Annotations
}

func (yk *YunikornSchedulePlugin) CreateLabels(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) (labels, annotations map[string]string) {
  // Set queue name based on the job type and flyteidl.Identifier fields including "ResourceType", "Org" and "Name".
  // 1.Clean yk.Labels and yk.Annotations
  // 2.Add yunikorn.apache.org/user.info = <organization>.<Name>
  // 3.Add yunikorn.apache.org/app-id = <ResourceType>-<uuid>
  // 4.Add yunikorn.apache.org/queue = <organization>.<jobType>
}

func (yk *YunikornSchedulePlugin) CreateGroupLabels(ctx context.Context, object client.Object, taskTmpl *core.TaskTemplate) {
  // 1.Add yunikorn.apache.org/task-group-name = yk.CreateTaskgroupName(ResourceType)
  // 2.Add yunikorn.apache.org/task-groups = yk.CreateTaskgroup(object)
  // 3.Add yunikorn.apache.org/schedulingPolicyParameters = yk.jobs[ResourceType]
  // 4.Add yunikorn.apache.org/allow-preemption = true/false
}

type KueueScheduablePlugin struct {
	jobs map[string]string
  Labels map[string]string
  Annotations map[string]string
}

func (k *KueueScheduablePlugin) GetGroupLabels() (labels, annotations map[string]string) {
  return k.Labels, k.Annotations
}

func (k *KueueScheduablePlugin) CreateLabels(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) (labels, annotations map[string]string) {
  // Set queue name based on the job type and flyteidl.Identifier field "Org".
  // Clean k.Labels and k.Annotations
  // 1.Add kueue.x-k8s.io/queue-name = <organization>.<jobtype>
  // Update k.Labels and k.Annotations
}

func (k *KueueScheduablePlugin) CreateGroupLabels(ctx context.Context, object client.Object, taskTmpl *core.TaskTemplate) {
  // Add Label "kueue.x-k8s.io/pod-group-name" and "kueue.x-k8s.io/pod-group-total-count" for spark、dask.
  // If object type is ray CRD and kubeflow CRD which are supported by Kueue then skips.
  // Update k.Labels and k.Annotations
}
```

Creat a scheduler plugin according to the queueconfig.scheduler.
Its basic responsibility validate whether submitted application is accepted. 
When a Yunikorn scheduler plugin created, it will create applicationID and queue name.
in the other hand, a Kueue scheduler plugin constructs labels including localQueueName, preemption.

```go
func (e *PluginManager) addObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
  e.SchedulerPlugin.CreateLabels(taskCtx, o)
  e.SchedulerPlugin.CreateGroupLabels(taskCtx, o)
  schedulerLabels, schedulerAnnotations := e.SchedulerPlugin.GetLabels()
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations(), schedulerAnnotations)))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels(), schedulerLabels)))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())
}
```
When batchscheduler in flyte is yunikorn, some examples are like following.
For example, this appoarch submits a Ray job owned by user1 in organization1 to "root.organization1.ray".
A spark application in ns1 submitted by user4 in organization1 is in "root.organization1.ns1".
In the other hand, results of these examples are "organization1-ray" and "organization1-ns1" when adopting Kueue.

## 4 Metrics & Dashboards

1. The Yunikorn scheduler add applications to a specific queue based on their user info, queue name for any application type.
2. Yunikorn and Kueue provide gang scheduling through annotations For Ray and spark.
3. Preemption behavior aligns with user-defined configuration in yunikorn.

## 5 Drawbacks

This appoarch doesn't offer a way to maintain consistency between the accuate resource quotas of groups and the configuration in scheduler.

## 6 Alternatives

## 7 Potential Impact and Dependencies

Flyte support Spark, Ray and Kubeflow CRDs including Pytorch and TFjobs.
The Spark and Ray operators have supported Yunikorn gang scheduling since task group calculation were implemented in these operators.
Taskgroup calculation implementation in pods aspect in flyte or kubeflow is required for supporting kubeflow CRDs.
In the other hand, Kueue currently doesn't support Spark CRD.
| Operator | Yunikorn | Kueue |
| --- | --- | --- |
| Spark | v | x |
| Ray | v | v |
| Kubeflow | x | v |

## 8 Unresolved questions


## 9 Conclusion

Yunikorn and Kueue support gang scheduling allowing all necassary pods to run sumultaneously when required resource are available.
Yunikorn provides preemption calculating the priority of applications based on their priority class and priority score of the queue where they are submitted, in order to trigger high-prioirty or emergency application immediately. 
Yunikorn's hierarchical queue includes grarateed resources settings and ACLs.
