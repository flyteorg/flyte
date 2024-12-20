# [Newbie] Supporting Yunikorn and Kueue

**Authors:**

- @yuteng

## 1 Executive Summary

Providing kubernetes (k8s) resource management, gang scheduling and preemption for flyte applications by third-party software, including Apache Yunikorn and Kueue.

## 2 Motivation

Flyte supports multi-tenancy and various Kubernetes plugins.
Some Kubernetes plugins may encounter into resource wastage when jobs partially start without performing any meaningful work. 
A solution to this issue is gang scheduling, which guarantees that all worker pods derived from a CRD are scheduled simultaneously.
Kueue or Apache Yunikorn support this mechanism. 
Additionally, Yunikorn can map tenants and organizations to hierarchical queues to define resource quotas.
Based on this setting, access control lists can be configured to grant access to users and groups.

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

`root.organization1.ray` is the queue of the ray job submitted by user1 belonging organization1. 

ResourceFlavor allocates resource based on labels which indicates that category-based resource allocation by organization label is available.
Thus, a clusterQueue including multiple resources represents the total acessaible resource for an organization.  
| clusterQueue | localQueue |
| --- | --- |
| <organization name> | ray、spark、default |
A tenant can submit organization-specific tasks to queues such as organization.ray, organization.spark and organization.default to track which job types are submittable.

A scheduling plugin implements functions `SetSchedulerName`, `CreateLabels` and `CreateGroupLabels` to create labels and `schedulerName`.
`CreateLabels` patches necassary labels, such as `queuename`, `user-info` and `applcationID`, to jobs.
`CreateGroupLabels` supports creating `group-pod` and `task-group` labels based on incoming CRD if need. 
`SetSchedulerName` set `schedulerName` field in `podTemplate`.

```go
type SchedulePlugin interface {
  CreateLabels(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig)
  CreateGroupLabels(ctx context.Context, object client.Object, taskTmpl *core.TaskTemplate)
  GetGroupLabels() (labels, annotations map[string]string)
  SetSchedulerName(object client.Object)
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

When a job comes, following things happens.
1. `SetSchedulerName` sets the `schedulerName` with the specific scheduler name
2. `CreateLabels` new basic labels based on the scheduler.
3. `CreateGroupLabels` creates `kueue.x-k8s.io/pod-group-name` or `yunikorn.apache.org/task-groups` according to the calculatied results from CRD.
4. Merging labels and annotations from `CreateLabels` and `CreateGroupLabels` to the CRD.

```go
type PluginManager struct {
	id              string
	plugin          k8s.Plugin
	resourceToWatch runtime.Object
	kubeClient      pluginsCore.KubeClient
	metrics         PluginMetrics
	// Per namespace-resource
	backOffController    *backoff.Controller
	resourceLevelMonitor *ResourceLevelMonitor
	eventWatcher         EventWatcher
}

func (e *PluginManager) launchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
  o, err := e.plugin.BuildResource(ctx, k8sTaskCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}
  if p, ok := e.plugin.(k8s.ScheduablePlugin); ok {
		o, err = p.SetSchedulerName(o)
		if err != nil {
			return pluginsCore.UnknownTransition, err
		}
	}
}

func (e *PluginManager) addObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
  var schedulerLabels, schedulerAnnotations map[string]string
  if p, ok := e.plugin.(k8s.ScheduablePlugin); ok {
		o, err = p.SetSchedulerName(o)
		if err != nil {
			return pluginsCore.UnknownTransition, err
		}
     p.CreateLabels(taskCtx, o)
     p.CreateGroupLabels(taskCtx, o)
     schedulerLabels, schedulerAnnotations = e.plugin.GetLabels()
	}
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations(), schedulerAnnotations)))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels(), schedulerLabels)))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())
}
```

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
