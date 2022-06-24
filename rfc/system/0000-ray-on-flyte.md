# [RR] Ray on Flyte

**Authors:**

- [@hamersaw](https://github.com/hamersaw)

## 1 Executive Summary

This RFC proposes an integration of Ray with the Flyte ecosystem. Specifically we propose a solution to (1) dynamically manage cluster resources within Flyte workflows, enabling cluster reuse between sub workflows and tasks and (2) seamlessly integrate Ray jobs into the Flyte task / workflow definitions ensuring ease of management and performance for AI/ML operations.

## 2 Motivation

Ray is solidifying it's place in the AI/ML framework stack. It offers performant model training at scale, proving to be both stable and extensible. We think the integration of Flyte with Ray serves both communities, offering advantages in the ease of management and performance in AI/ML operations.

Flyte can further benefit from support for dynamically managing cluster resources. Currently, Flyte manages clusters on a per-task basis, rather than reusing clusters between tasks / workflows. This does not fit well with the Ray model, mitigating many of the inherit advantages. To take full advantage of Ray benefits, multiple tasks are required to execute on the same cluster to facilitates efficient data placement / retrieval and job scheduling. Beyond Ray, this functionally may be easily extended to support a variety of cluster types becoming a very powerful paradigm in Flyte.

## 3 Proposed Implementation
We believe that the majority of user interest is in the FlyteKit API rather than backend design, therefore we present it first. However, perhaps the most important details for Ray integration is seamless dynamic management of Ray clusters. From an implementation perspective, this introduces the majority of challenges, and therefore requires the majority of concentration. If this is designed correctly, it results in significant improvements in the utility of Flyte, offering future support for dynamically managing diverse clusters

TODO - architecture diagram

### FlyteKit API
Our proposal for support of dynamically managed clusters involves the addition of named cluster resources defined at both the workflow and task levels. This API focuses on simplicity, with deeper integrated support for the most complex cases. An example highlighting this API is provided below:


    @task(config=Spark(...), resource=SparkCluster(...))
    def my_task():
        …

    @workflow(resources={“x1”:RayCluster(...), “x2”: SparkCluster(...)}
    def my_sub_wf(): 
        …

    @workflow(resources={“y1”:RayCluster(...), “y2”: RayCluster(...)}
    def my_sub_wf2(): 
        …

    @workflow(resources={
            “z1”: RayCluster(image={{.image.default}}),
            “z2”: SparkCluster(spark_config_vars: “abc”)
        })
    def wf():
        # execute task by implicitely using the SparkCluster "z2"
        o1 = my_task(a=1)

        # execute my_sub_wf by implicitely reusing clusters "z1" and "z2"
        my_sub_wf(....)

        # do not reuse clusters z1 and z2 when executing my_sub_wf by explicitly ignoring them
        my_sub_wf.with_override(ignore_parent_resources=True)(....)

        # must explicitely set clusters because there is no exact type matching
        my_sub_wf2.with_resources("y1":"z1", "y2":"z2")(...)

        lp(....)

    # use the workflows default cluster configuration
    lp = launchplan.create(wf)

    # implicitely override the cluster configuration for the r1 cluster
    lp2 = launchplan.create(wf, inputs=..., resources={“r1”: RayCluster(image=””)})

Explanation of this API proposal may be best approached by dissecting individual components:

**Resource Definitions:** We introduce a new configuration option on workflows and tasks, namely `resources`. These are responsible for setting the specific configuration options for each cluster that will be managed by the defined entity. The specific options are defined by each individual cluster type.

By default clusters defined in a workflow are started at the beginning of workflow execution and will be deleted on workflow termination (whether successful or failure).

**Resource Reuse and Overrides:** Flyte will implicitly reuse clusters for child tasks / subworkflows if there is a 1 to 1 type match from child to parent definitions. This means the child will not start and manage the clusters defined in their configuration, rather the clusters of the parent are used for all jobs. Sometimes, this will not or can not be the desired behavior. Therefore, we offer multiple options to override this:

- _Ignore Overrides_: This allows child entities to manage their own clusters even if there is a 1 to 1 type match from child to parent types. Essentially, the child ignores the clusters of the parent, instead choosing to manage it's own clusters.
- _Manually Mapped Clusters_: Users may manually define mapping of cluster variable names between parents and children. This will be required for scenarios where the child does not have a 1 to 1 mapping of cluster types with the parent. For example, a child with ray cluster ("b1") and spark cluster ("b2") may implicitly reuse clusters from a parent with a ray cluster ("a1") and spark cluster ("a2"), but a child with a single ray cluster ("y1") and a parent with two ray clusters ("x1" ,"x2") will require explicit mapping.

For now the design for the FlyteKit API for Ray specific tasks is an open question. Further exploration of Ray is required. For example, can we just write the Ray python in a task and execute that directly?

### FlytePropeller
In this proposal, FlytePropeller will be responsible for dynamically managing clusters. This involves starting clusters when a workflow execution starts and ensuring clusters are deleted when the workflow completes. There are multiple extensions necessary in within FlytePropeller to achieve this functionality:

**Start / End Node Plugins:** Workflows definitions in Flyte are automatically anchored by these nodes to facilitate input / output management. We can introduce specific plugins so that these nodes can handle cluster management as part of the pre / post execution process. For Ray specifically, we will leverage the [KubeRay](https://github.com/ray-project/kuberay) project to start a Ray cluster on k8s.

One significant challenge in this process is to ensure successful cluster shutdown in all of the workflow termination scenarios (i.e., success, failure, etc). A few options include:

- _Leveraging k8s Ownership:_ k8s provides a resource ownership model. We can set the owner of the cluster to the FlyteWorkflow CRD for the managing workflow. Therefore, when the FlyteWorkflow CRD is garbage collected the cluster will be automatically cleaned up as well.

- _Try / Catch / Finally on FlyteWorkflow:_ Discussions on introducing default failure handling on FlyteWorkflows has been on the radar. However, there has not yet been a resounding need. Cleanup of managed clusters, seems to be an ideal fit for this model.

**Decorating ExecutionContext with Cluster Information:** To support reuse of clusters we will include additional metadata within the ExecutionContext so that child entities have configuration details for cluster submission.

_Note:_ Specific implementation changes are far too verbose to summarize within this document.

### FlytePlugins

FlytePlugins will need to introduce a specific Ray plugin. This will be responsible for reading cluster configuration from the TaskExecutionContext and submitting / monitoring Ray jobs on the configured cluster. Without an intimate understanding of Rays API it is difficult to expose specific details here. This will need to be explored further.

### FlyteIDL

FlyteIDL will need updates to reflect the addition of resources at the workflow, task, and launchplan levels. These must be designed to support integration of additional cluster types. We propose the follow IDL message definitions:

    /*
     * define resource message
     */

    message Resource {
        enum ResourceKind {
            RAY = 0,
            SPARK = 1,
        }

        ResourceKind kind = 1;
        RayCluster ray = 2;
        Spark Cluster spark = 3;
    }

    message RayCluster {
        // specific configuration omitted
    }

    message SparkCluster {
        // specific configuration omitted
    }

    /*
     * add new resources to existing entity templates
     */

    message WorkflowTemplate {
        // existing definition omitted

        map<string, Resource> resources = ?;
    }

    TaskTemplate {
        // existing definition omitted

        Resource resource = ?;
    }

## 4 Metrics & Dashboards

To fully integrate with Ray it will be important to investigate the metrics available through [KubeRay](https://github.com/ray-project/kuberay). We want to ensure that these are available to facilitate robust performance monitoring.

Regarding Flyte specific metrics, we think these are twofold. (1) Tracking managed clusters, including the number and duration of running clusters and (2) tracking cluster management operations, specifically the count and latency of cluster startup / shutdown operations.

## 5 Drawbacks

There are no apparent drawbacks in committing to this implementation.

## 6 Alternatives

Dynamic cluster management is a difficult problem. We believe there many alternatives to the design decisions we have proposed:

- _Managing Clusters on a Per-Task Basis:_ Currently, the Spark integration with Flyte involves managing clusters at the per-task basis. We think that this scheme mitigates some of the advantages inherit in reusing long-running clusters among multiple tasks.
- _Using FlyteAdmin to Manage Clusters:_ FlyteAdmin could manage clusters to allow reuse between workflows. We encountered a few major challenges in this paradigm. (1) Cluster lifetimes are not well defined. For example, when should a cluster be shutdown? (2) Fine-grained configuration of cluster reuse between parents and children becomes quite difficult.

## 7 Potential Impact and Dependencies

There should be no impact on backwards compatibility as this feature is entirely new implementation.

This will add additional dependencies on Ray for workflows which leverage this functionality.

## 8 Unresolved questions

1. How should dynamically managed clusters be defined in the FlyteKit API? We have proposed a verbose API that covers the use-cases that we envision. However, this will hopefully be actively discussed and improved.
2. How do we define ray tasks within FlyteKit? The specific API for Ray tasks and been omitted from this document. It should be simple to replicate existing approaches, but further exploration of Ray is required to ensure seamless integration.
3. How are clusters shutdown within FlytePropeller? This is explored in further detail in the implementation section, including a collection of possible approaches. The most promising is using the k8s ownership relationship between FlyteWorkflow CRDs and clusters.

## 9 Conclusion

In this RFC we have proposed an approach to integrate Ray into the Flyte ecosystem. This includes an API and FlytePropeller extensions to dynamically manage cluster resources during Flyte workflow execution and an approach to integrate Ray jobs into Flyte workflow and task definitions. We believe this integration will benefit both communities, easing the management of AI/ML operations at scale while simultaneously improving performance.
