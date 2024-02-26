# [RFC] Select cluster execution when triggering an execution

**Authors:**

- @RRap0so
- @andresgomezfrr

## 1 Executive Summary

When triggering a Flyte LaunchPlan, Workflow or Task, Flyte permits the user to define a ClusterPool from a predefined list of pools that are set in the FlyteAdmin configuration. For example with `flytectl create execution
--clusterPool string` 

This provides greater flexibility compared to the existing feature where an [Execution Cluster Label](https://docs.flyte.org/en/latest/deployment/deployment/multicluster.html#configure-execution-cluster-labels) can be set at the Project/Domain level.

## 2 Motivation

The primary motivation for this proposal is to select the clusters where an execution is to be performed dynamically at execution time, without the constant need to overwrite the Execution Cluster Label.

Additionally, we would like to have the possibility of different executions targeting different clusters, even if they are published under the same Project/Domain. This is not entirely feasible using the current Execution Cluster Label.


## 3 Proposed Implementation

We plan to use the `cluster_assignement` field (refer to [proto](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/admin/execution.proto#L312)) in each execution, this is a ClusterAssignment object with a `cluster_pool_name` which would be the field that contains a string with the corresponding ClusterPool ID.

In the FlyteAdmin configuration we have a list of `clusterpool_config` with all the different available pools and corresponding clusters.

**Note**: Today there are configurations to assign domains to specific pools, see [PR](https://github.com/flyteorg/flyte/pull/4208) and `cluster_pool` [interface](https://github.com/flyteorg/flyte/blob/962d45827646f106265630a0fb1d9ae74968bc32/flyteadmin/pkg/runtime/interfaces/cluster_pools.go). We propose another configuration to assign Clusters to a clusterPoolName by having another map under cluster object with weights and cluster assignments.

Proposed configuration:
```
    clusters:
       poolNameClusterMap:
        pool_1:
        - id: flyte-1-k8s
          weight: 1
        - id: flyte-2-k8s
	      weight: 2
        pool_2:
        - id: flyte-3-k8s
	      weight: 6
        - id : flyte-4-k8s
	      weight: 99
      labelClusterMap:
        production:
        - id: flyte-1-k8s
          weight: 1
      clusterConfigs:
      - name: flyte-1-k8s
        endpoint: 127.0.0.1
        enabled: true
        auth:
          type: file_path
          tokenPath: /etc/…
          certPath: /etc/…
    - name: flyte-2-k8s
        endpoint: 127.0.0.2
        enabled: true
...
```

At the start of FlyteAdmin it will create a new RandomClusterSelector with a `poolWeightedRandomMap`. 


### When an execution is launched

Flyteadmin runs the [Execute](https://github.com/flyteorg/flyte/blob/master/flyteadmin/pkg/workflowengine/impl/k8s_executor.go#L36) method and builds the workflow. Now we need to know which Kubernetes Cluster to hit. We will have the pool information available in the ExecutionData that is already passed into the Execute method. 

**If no clusterPool is defined**: No change in current behavior.

**If clusterPool is defined**: Then we will add the poolName into the [ExecutionTargetSpec](https://github.com/flyteorg/flyte/blob/master/flyteadmin/pkg/executioncluster/execution_target.go#L13-L20) (we will need to modify the struct). This is passed to the getTarget where we check if it has a poolName defined and get the value from the poolWeightedRandomMap. There are still open questions regarding the priority of labels / pool names.


## 4 Metrics & Dashboards

If implemented correctly, we should be able to capture where executions are being triggered. This would be extra useful for our use case to make sure we’re routing executions to the correct pool of clusters

## 5 Drawbacks

Other than the usual drawbacks of implementing a new feature (the introduction of new bugs/problems) we don’t see any potential drawbacks with this implementation. The fields are already exposed but they are only missing the implementation. 

We don’t expect any additional overhead for execution or impact in terms of API for flyteadmin.

## 6 Alternatives

As mentioned above, the only alternative we have implemented today is using the [Execution Cluster Labels](https://docs.flyte.org/en/latest/deployment/deployment/multicluster.html#configure-execution-cluster-labels). This can bring further problems when we want to route executions to different clusters from the same Project/Domain.

## 7 Potential Impact and Dependencies

We do not see any.


## 8 Unresolved questions

What takes priority Labels or Pools? 
Since we have two different ways to route traffic we need to decide what takes priority.

## 9 Conclusion

With this we want to bring to the finish line an idea that we believe was going to be done but was not yet implemented. We will reuse as many already existing objects as possible and will be able to smoothly select clusters at a moment's notice which for our use case is extremely important to ensure the timely and reliable execution of LaunchPlans/Workflows/Tasks.
