# [Newbie] Supporting Yunikorn and Kueue

**Authors:**

- @yuteng

## 1 Executive Summary

Providing k8s resource management, gang scheduling and preemption for flyte apllications by 3rd software including Apache Yunikorn and Kueue.

## 2 Motivation

Flyte support multi-tenant and many k8s plugins.

Kueue and Yunikorn support gang scheduling and preemption.
Gang scheduling gurantees some avialable k8s crd services including spark, ray service with sufficient resource and preemption make sure high priority task execute immediately.

Flyte doesn't maintain a resource management for multi-tenant which Yunikorn can solve it with hierarchy resource queues.

## 3 Proposed Implementation

```yaml
queueconfig:
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
Hierachy queues will be like following.
root.org1.ray、root.org1.spark and root.org1.default allowing summission from ns1 and ns1 namespace.
root.org2."CRDs" and root.org2.default allowing summission from any namespaces.

ResourceFlavor allocate resource based on labels which indicates thant category resource allocatiom by org label is available.
So a clusterQueue including multiple resources is a total acessaible resource for a org.  
| clusterQueue | localQueue |
| --- | --- |
| Org | ray、spark、default |
A tenant can submit org task to queue such org.ray, org.spark and org.default to trace what job types are submitable. 


A SchedulerConfigManager maintains config from mentioned yaml.
It patch labels or annotations of k8s resources after they pass rules in config.

```go
func (e *PluginManager) launchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
  o, err := e.plugin.BuildResource(ctx, k8sTaskCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}
  if err := e.SchedulerConfigManager.Label(0); err != nil {
    return pluginsCore.UnknownTransition, err
  }
}
```
When batchscheduler in flyte is yunikorn, some examples are like following.
For exmaple, this appoarch submit a Rayjob owned by user1 in org1 to "root.org1.ray".
A spark application in ns1 submitted by user4 in org1 is in "root.org1.ns1".
In the other hand, results of these examples are "org1-ray" and "org1-ns1" when adopting Kueue.

## 4 Metrics & Dashboards

1. Yunikorn scheduler add applications to a specific queue based on thier user info, queue name for any application type.
2. Yunikorn and Kueue provides gang scheduling based annotations For Ray and spark.
3. Preemption behavior meets user-defined configuration in yunikorn.

## 5 Drawbacks

This appoarch doens't provide a way to keep consistent between the accuate resource quato of groups and configuration in scheduler.

## 6 Alternatives

## 7 Potential Impact and Dependencies

Flyte support spark, ray and kubeflow CRD including pytorch and tfjobs.
Spark and Ray operator have been support Yunikorn gang scheduling after taskgroup calculation is implemented in these operators.
Taskgroup caclucation implementation in pods aspect in flyte or kubeflow is required for supporting kubeflow CRDs.
In the other hand, Kueue currently doesn't support Spark CRD.
| Operator | Yunikorn | Kueue |
| --- | --- | --- |
| Spark | v | x |
| Ray | v | v |
| Kubeflow | x | v |

## 8 Unresolved questions


## 9 Conclusion

Yunikorn and Kueue support gang scheduling to run all neccesary pods at same time when required resource are available.
Yunikorn provide preemption which calculate the priority of application based on its priority class an priority score of the queue where it submitted in order to trigger high-prioirty or emergency application immediately. 
Yunikorn hierachy queue includes grarateed resources setting and acls.

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
