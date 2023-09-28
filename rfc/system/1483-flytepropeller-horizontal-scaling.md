# [RFC] FlytePropeller Horizontal Scaling

**Authors:**

- @kumare3
- @hamersaw

This Document - https://hackmd.io/uVT5Q9zNS8SNMkEhs5pfqQ

## 1 Executive Summary

We propose an approach to facilitate automatic, horizontal scaling of FlytePropeller. This is important to combat increases in individual workflow execution latency as the number of concurrent workflow executions increases. The solution includes a new FlytePropeller Manager component which handles automatic scaling of additional FlytePropeller instances. Additionally, we rely on a configurable sharding mechanism to ensure deterministic, decentralized coordination of Flyte workflows.

## 2 Motivation

FlytePropeller is the engine that drives Flyte workflows. It is highly optimized and a single instance can run thousands of concurrent workflows. Internally, workflows are processed in a control loop which operates over a work queue. Consequently, as the number of concurrent workflows increases beyond a reasonable threshold, the latency of workflow completion will experience marginal increases as well.

The default FlytePropeller deployment strategy is a single instance and while it is possible to launch multiple FlytePropeller instances (ie. one per namespace) this has to be done manually. This approach has obvious limits to scalability. Therefore, to reduce workflow processing latency as the number of workflows increase, we need a new solution to increase the refresh rate per workflow.

In addition to the obvious scaling of workflow processing workloads, the system should be automated and should be self healing. It should have the following Safety and Liveness properties
- Safety properties
    - Every workflow should be executed: Once a workflow is being executed, it should continue to make progress
    - A workflow should not be lost, unless the k8s cluster is lost
    - The K8s API server should be safe
- Liveness properties
    - Once a Workflow starts executing, it should eventually either fail or succeed
    - As the number of workflows increase, existing workflows should reach completion
    - System should scale automatically and in case of failures should recover in a correct way - workflows should not get corrupted.
    - Correlated failures should be unlikely


## 3 Proposed Implementation

![](https://i.imgur.com/ru21IDU.png)

The proposed system intends to extend FlytePropeller to add a new mode called propeller-manager. When FlytePropeller is run in the manager mode, it can consume an additional configuration called the ShardingStrategy.

```yaml=
ShardingStrategy:
     Type: <name of strategy>
     Strategy: <strategy specific configuration>
```

Depending on the Sharding strategy the propeller-manager, launches one or more FlytePropeller instances. Ideally the FlytePropeller instances are identical to the propeller-manager (ie. k8s pod configuration) with minor changes that affect how the sharding works. The configuration will include FlytePropeller image definition, serviceAccounts, etc.

```yaml=
ConstantShardingStrategy:
    Replicas: 4
```

To begin, we want to by implement a ConstantShardingStrategy, where the distribution mechanism is similar to consistent hashing. FlytePropeller maintains a constant ring-size and FlyteAdmin submits CRD's with shard-bin tokens within that range. Specifically, FlyteAdmin assigns new workflow CRD's the following extra labels:

1. A shard-bin field with a number between \[0,256)
2. A project label and a domain label

The ConstantShardingStrategy configuration defines the number of FlytePropeller instances (ie. Replicas) that FlytePropeller Manager runs. Each replia computes its shard bin range using an assigned rank and the constant population size using the following algorithm. 

    [ 256*n/p, 256*(n+1)/p), where n = rank and p = population size

For example, with population = 4 and rank = 0 the range is \[0, 64), rank = 1 range is \[64, 128). Using this range as a template, FlytePropeller replicas are able to use k8s server side [label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) to filter workflows, reducing constraints on bandwidth.

FlytePropeller Manager will deploy replicas using k8s pods with deterministic naming. Because k8s resources are idempotent, attempts to create multiple instances the same ranked FlytePropeller replica (ie. based on name) will be rejected. Additionally, if a replica fails FlytePropeller Manager will automatically restart it, this functionality can me modeled as a control loop.

```go=
initialize p (population size)
while True:
    replicas = fetch_current_state
    if len(replicas) < p:
        missing_replicas = find_missing_replica()
        for r in missing_replicas:
            start(r)
```

**Shard System Extensibility:** Given the above setup we can extend the Sharding system to allow multiple strategies like - Dynamic load based sharding, namespace sharding, domain sharding etc. But we will start with the simplest strategy as this is the easiest to manage


## 4 Metrics & Dashboards

- As the number of workflows increase, the latency of workflow state transitions should remain constant or within a pre-determined bound.

## 5 Drawbacks

Support for additional sharding pardigms may require coordinated changes in FlytePropeller and FlyteAdmin. For example:
1. Sharding strategies that do not comply with the proposed bin-based scheme (ex. namespace / domain sharding) require FlyteAdmin to include additional CRD values
2. Support for dynamic load-balanced sharding requires resource utilization metrics to be reported to FlyteAdmin

## 6 Alternatives

**Assigning Workflow Replica in FlytePropeller Manager:** We could have simply assigned the workflows to a replica in the manager, but this has a performance overhead, where the manager process has to consume each workflow CRD and then apply the label for the bin size. Instead of this, moving the labeling to be a constant ring-size in FlyteAdmin reduces overhead with the following advantages:
- Reduces Acceptance latency (startup latency). This is a very critical metric for Flyte users. This is because we do not need to assign this value as an extra step in the flytepropeller-manager.
- Since the bins are statically assigned as labels, Flytepropeller can use server side label queries to filter the workflows that are retrieved by every instance. This reduces the total workflows that need to be processed and reduces constraints on bandwidth. Read more about it in Labels and Selectors.
- Improves debuggability of the system.

**Using Leases to Assign Replica Rank:** Each FlytePropeller instance gets the population size as an input or queries k8s for its replica count. With a known p value, each FlytePropeller now tries to acquire a lease - using the k8s lease object. Each FlytePropeller instance does not know which leases are available, so it can either test all leases, or try to acquire a lease at random or acquire them in serial order. Lease acquisition in the absence of a true-time or known clock uses a bounded-drift style algorithm, which delays a FlytePropeller instance’s lease acquisition time and this slows down the startup. The slowdown is directly proportional to the number of failed lease acquisition attempts and is not deterministic.
A much simpler model is to use a central controller, which is a leader elected single process and hence can take decisions like a control loop. This closely resembles K8s mode and is simpler to implement - relatively bug-free.

**Using Deployment Directly:** The biggest challenge with using deployment is that each propeller needs to know its rank. Deployments do not currently have a way to pass this in, even as an environment variable. To achieve this, we will have to do 2 additional things which complicate the system especially during re-balancing and recovery from failure.
1. Listen to the deployment and look at the replicas to determine p
2. Come up with a dynamic algorithm that tries to acquire different ranks or ranks in order.


## 7 Potential Impact and Dependencies

Existing deployment of FlytePropeller should continue to run without interruptions. If users wish to upgrade their FlytePropeller deployments - then it should be as simple as changing the execution command or a configuration.

## 8 Unresolved questions

_NA_

## 9 Conclusion

The solution presented in this document facilitates efficient horizontal scaling of FlytePropeller. This will effectively reduce the latency of individual workflow execution as the number of concurrent workflow executions increase. The approach presents an extensible, decentralized solution in adherence with existing k8s principles.