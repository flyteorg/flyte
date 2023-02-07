/*
Package manager introduces a FlytePropeller Manager implementation that enables horizontal scaling of FlytePropeller by sharding FlyteWorkflows.

The FlytePropeller Manager manages a collection of FlytePropeller instances to effectively distribute load. Each managed FlytePropller instance is created as a k8s pod using a configurable k8s PodTemplate resource. The FlytePropeller Manager use a control loop to periodically check the status of managed FlytePropeller instances and creates, updates, or deletes pods as required. It is important to note that if the FlytePropeller Manager fails, managed instances are left running. This is in effort to ensure progress continues in evaluating FlyteWorkflow CRDs.

FlytePropeller Manager is configured at the root of the FlytePropeller configurtion. Below is an example of the variety of configuration options along with succinct associated descriptions for each field:

	manager:
	  pod-application: "flytepropeller"             # application name for managed pods
	  pod-template-container-name: "flytepropeller" # the container name within the K8s PodTemplate name used to set FlyteWorkflow CRD labels selectors
	  pod-template-name: "flytepropeller-template"  # k8s PodTemplate name to use for starting FlytePropeller pods
	  pod-template-namespace: "flyte"               # namespace where the k8s PodTemplate is located
	  scan-interval: 10s                            # frequency to scan FlytePropeller pods and start / restart if necessary
	  shard:                                        # configure sharding strategy
	    # shard configuration redacted

FlytePropeller Manager handles dynamic updates to both the k8s PodTemplate and shard configuration. The k8s PodTemplate resource has an associated resource version which uniquely identifies changes. Additionally, shard configuration modifications may be tracked using a simple hash. Flyte stores these values as annotations on managed FlytePropeller instances. Therefore, if either of there values change the FlytePropeller Manager instance will detect it and perform the necessary deployment updates.

# Shard Strategies

Flyte defines a variety of Shard Strategies for configuring how FlyteWorkflows are sharded. These options may include the shard type (ex. hash, project, or domain) along with the number of shards or the distribution of project / domain IDs over shards.

Internally, FlyteWorkflow CRDs are initialized with k8s labels for project, domain, and a shard-key. The project and domain label values are associated with the environment of the registered workflow. The shard-key value is a range-bounded hash over various components of the FlyteWorkflow metadata, currently the keyspace range is defined as [0,32). A sharded Flyte deployment ensures deterministic FlyteWorkflow evalutions by setting disjoint k8s label selectors, based on the aforementioned labels, on each managed FlytePropeller instance. This ensures that only a single FlytePropeller instance is responsible for processing each FlyteWorkflow.

The Hash Shard Strategy, denoted by "type: hash" in the configuration below, uses consistent hashing to evenly distribute FlyteWorkflows over managed FlytePropeller instances. This is achieved by partitioning the keyspace (i.e. [0,32)) into a collection of disjoint ranges and using label selectors to assign those ranges to managed FlytePropeller instances. For example, with "shard-count: 4" the first instance is responsible for FlyteWorkflows with "shard-keys" in the range [0,8), the second [8,16), the third [16,24), and the fourth [24,32). It may be useful to note that the default shard type is "hash", so it will be implicitly defined if otherwise left out of the configuration. An example configuration for the Hash Shard Strategy is provided below:

	# a configuration example using the "hash" shard type
	manager:
	  # pod and scanning configuration redacted
	  shard:
	    type: hash     # use the "hash" shard strategy
	    shard-count: 4 # the total number of shards

The Project and Domain Shard Strategies, denoted by "type: project" and "type: domain" respectively, use the FlyteWorkflow project and domain metadata to distributed FlyteWorkflows over managed FlytePropeller instances. These Shard Strategies are configured using a "per-shard-mapping" option, which is a list of ID lists. Each element in the "per-shard-mapping" list defines a new shard and the ID list assigns responsibility for the specified IDs to that shard. The assignment is performed using k8s label selectors, where each managed FlytePropeller instance includes FlyteWorkflows with the specified project or domain labels.

A shard configured as a single wildcard ID (i.e. "*") is responsible for all IDs that are not covered by other shards. Only a single shard may be configured with a wildcard ID and on that shard their must be only one ID, namely the wildcard. In this case, the managed FlytePropeller instance uses k8s label selectors to exclude FlyteWorkflows with project or domain IDs from other shards.

	# a configuration example using the "project" shard type
	manager:
	  # pod and scanning configuration redacted
	  shard:
	    type: project       # use the "project" shard strategy
	    per-shard-mapping:  # a list of per shard mappings - one shard is created for each element
	      - ids:            # the list of ids to be managed by the first shard
	        - flytesnacks
	      - ids:            # the list of ids to be managed by the second shard
	        - flyteexamples
	        - flytelabs
	      - ids:            # the list of ids to be managed by the third shard
	        - "*"           # use the wildcard to manage all ids not managed by other shards

	# a configuration example using the "domain" shard type
	manager:
	  # pod and scanning configuration redacted
	  shard:
	    type: domain        # use the "domain" shard strategy
	    per-shard-mapping:  # a list of per shard mappings - one shard is created for each element
	      - ids:            # the list of ids to be managed by the first shard
	        - production
	      - ids:            # the list of ids to be managed by the second shard
	        - "*"           # use the wildcard to manage all ids not managed by other shards
*/
package manager
