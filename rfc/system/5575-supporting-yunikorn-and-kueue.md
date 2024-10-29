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

```yaml
queueconfig:
  scheduler: yunikorn
  general:
  - org: org1
    users: "*" 
    priorityclass: priority-default
    acl:
      - jobs: "ray,dask"
        users: "user1,user2"
        priorityclass: priority-ray
        gangscheduling: true
      - jobs: "spark"
        users: "user1,user3"
        priorityclass: priority-spark
        gangscheduling: true
    namespace:
      default: "ns1"
      namespaces: "ns1|ns2"
  - org: org2
    users: "user4, user5"
    acl:
      - jobs: "*"
        users: "*"
    namespace:
      namespaces: "*"
```

Mentioned configuration indicates what queues exist for an org.
Hierarchucak queues will be structured as follows.
root.org1.ray、root.org1.spark and root.org1.default allowing submissions from ns1 and ns2 namespace.
root.org2."CRDs" and root.org2.default allowing submissions from any namespaces.

ResourceFlavor allocates resource based on labels which indicates that category-based resource allocation by organization label is available.
Thus, a clusterQueue including multiple resources represents the total acessaible resource for an organization.  
| clusterQueue | localQueue |
| --- | --- |
| Org | ray、spark、default |
A tenant can submit organization-specific tasks to queues such as org.ray, org.spark and org.default to track which job types are submittable. 


A SchedulerConfigManager maintains config from mentioned yaml.
It patches labels or annotations on k8s resources after they pass rules specified in the configuration.

```go
type SchedulerHelper interface {
  Patch(obj client.object) (error, string)
  GangScheduling(string) error
}

type PodLabels struct {
  PodName string
  labels map[string]string
  annotations map[string]string
}
```
Creat a scheduler helper according to the queueconfig.scheduler.
Its basic responsibility validate whether submitted application is accepted. 
When a Yunikorn scheduler helper created, it will create applicationID、queue name and preemption labels .
in the other hand, a Kueue scheduler helper constructs labels including localQueueName, preemption.

```go
func (e *PluginManager) launchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
  o, err := e.plugin.BuildResource(ctx, k8sTaskCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}
  if 
  if err, jobtype := e.SchedulerHelper.Patch(o); err == nil {
    if e.SchedulerHelper.GangScheduling(jobtype, o); err != nil {
       return pluginsCore.UnknownTransition, err
    }
  } else {
     return pluginsCore.UnknownTransition, err
  }
}
```
When batchscheduler in flyte is yunikorn, some examples are like following.
For example, this appoarch submits a Ray job owned by user1 in org1 to "root.org1.ray".
A spark application in ns1 submitted by user4 in org1 is in "root.org1.ns1".
In the other hand, results of these examples are "org1-ray" and "org1-ns1" when adopting Kueue.

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
Yunikorn provides preemption calculating the priority of applications based on thier priority class and priority score of the queue where they are submitted, in order to trigger high-prioirty or emergency application immediately. 
Yunikorn's hierarchical queue includes grarateed resources settings and ACLs.

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [ ]  Copy template
- [ ]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
