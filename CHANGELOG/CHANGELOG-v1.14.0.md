# Flyte 1.14.0 Release Notes

## Added
- Support for FlyteDirectory as input to ContainerTask (#5715)

A significant update to the flytekit storage interface enables downloading multi-part blobs. This allows Flyte to copy a FlyteDirectory as an input to ContainerTasks, enabling higher flexibility in workflow development with raw containers.
Fixed

- Better handling of CORS in TLS connections (#5855)

When using Flyte with TLS certificates, CORS options were not enabled causing issues like the Flyte console UI not showing any project when multiple were created. This scenario came with relative frequency among users evaluating Flyte in non-production environments. Now, CORS will be enabled on TLS connections too.


## Changed
- Enhanced flexibility for Ray plugin configurations (#5933)

This release makes the configuration of RayJobs more flexible, letting you pass Pod specs independently for each Ray node type: Worker, Head, and Submitter. This enables you to declare configuration for each group to better align with your infrastructure requirements:
```yaml
ray_config = RayJobConfig(
    head_node_config=HeadNodeConfig(
        requests=Resources(mem="64Gi", cpu="4"),
        limits=Resources(mem="64Gi", cpu="4")
        pod_template_name = "ray_head_nodeÄ
    ),
    worker_node_config=[
        WorkerNodeConfig(
            group_name="V100-group",
            replicas=4,
            requests=Resources(mem="256Gi", cpu="64",  gpu="1"),
            limits=Resources(mem="256Gi", cpu="64",  gpu="1"),
            pod_template = V1PodSpec(node_selector={"node_group": "V100"}),
        ),
        WorkerNodeConfig(
            group_name="A100-group",
            replicas=2,
            requests=Resources(mem="480Gi", cpu="60", gpu="2"),
            limits=Resources(mem="480Gi", cpu="60", gpu="2")
            pod_template = V1PodSpec(node_selector={"node_group": "A100"}),
        )
    ],
)
```

## Breaking
As python 3.8 hit the End of Life period in October 2024, starting with this release, flytekit requires Python >=3.9.


## Full changelog
* Flyte docs overhaul (phase 1) by @neverett in https://github.com/flyteorg/flyte/pull/5772
* [Flyte][3][flytepropeller][Attribute Access][flytectl] Binary IDL With MessagePack by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5763
* Update aws-go-sdk to v1.47.11 to support EKS Pod Identity by @mthemis-provenir in https://github.com/flyteorg/flyte/pull/5796
* Add minimal version to failure nodes docs by @eapolinario in https://github.com/flyteorg/flyte/pull/5812
* Add finalizer to avoid premature CRD Deletion by @RRap0so in https://github.com/flyteorg/flyte/pull/5788
* [Docs] Use Pure Dataclass In Example by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5829
* Link to github for flytekit docs by @thomasjpfan in https://github.com/flyteorg/flyte/pull/5831
* Added the documentation about uniqueness of execution IDs by @Murdock9803 in https://github.com/flyteorg/flyte/pull/5828
* DOC-648 Add wandb and neptune dependencies by @neverett in https://github.com/flyteorg/flyte/pull/5842
* Update contributing docs by @neverett in https://github.com/flyteorg/flyte/pull/5830
* Update schedules.md by @RaghavMangla in https://github.com/flyteorg/flyte/pull/5826
* [flytectl] Use Protobuf Struct as dataclass Input for backward compatibility by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5840
* Add David as codeowner of deployment docs by @neverett in https://github.com/flyteorg/flyte/pull/5841
* Add an error for file size exceeded to prevent system retries by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5725
* Add perian dependency for PERIAN plugin by @neverett in https://github.com/flyteorg/flyte/pull/5848
* [FlyteCopilot] Binary IDL Attribute Access Primitive Input by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5850
* [Docs] Update for image spec/fast register notes by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5726
* [Docs] add practical example to improve data management doc by @DenChenn in https://github.com/flyteorg/flyte/pull/5844
* Handle CORS in secure connections by @eapolinario in https://github.com/flyteorg/flyte/pull/5855
* Update Grafana User dashboard by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5703
* RFC: Community plugins by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5610
* Fix link to contributing docs by @Sovietaced in https://github.com/flyteorg/flyte/pull/5869
* [Docs] add streaming support example for file and directory by @DenChenn in https://github.com/flyteorg/flyte/pull/5879
* [Docs] Align Code lines of StructuredDataset with Flytesnacks Example by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/5874
* Sync client should call CloseSend when done sending by @RRap0so in https://github.com/flyteorg/flyte/pull/5884
* Upstream contributions from Union.ai by @andrewwdye in https://github.com/flyteorg/flyte/pull/5769
* Remove duplicate recovery interceptor by @Sovietaced in https://github.com/flyteorg/flyte/pull/5870
* fix: failed to make scheduler under flyteadmin by @lowc1012 in https://github.com/flyteorg/flyte/pull/5866
* [Docs] Recover the expected behaviors of example workflows by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/5880
* Use child context in watchAgents to avoid goroutine leak by @pingsutw in https://github.com/flyteorg/flyte/pull/5888
* [docs]add cache information in raw containers doc by @popojk in https://github.com/flyteorg/flyte/pull/5876
* Clarify behavior of interruptible map tasks by @RaghavMangla in https://github.com/flyteorg/flyte/pull/5845
* [Docs] Simplifying for better user understanding by @10sharmashivam in https://github.com/flyteorg/flyte/pull/5878
* Update ray.go not to fail when going suspend state. by @aminmaghsodi in https://github.com/flyteorg/flyte/pull/5816
* Add dependency review gh workflow by @eapolinario in https://github.com/flyteorg/flyte/pull/5902
* [Docs] improve contribute docs by @DenChenn in https://github.com/flyteorg/flyte/pull/5862
* Fix CONTRIBUTING.md and update by @taieeuu in https://github.com/flyteorg/flyte/pull/5873
* Add pod template support for init containers by @Sovietaced in https://github.com/flyteorg/flyte/pull/5750
* Update monitoring docs by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5903
* feat: add support for seedProjects with descriptions via seedProjectsWithDetails by @Terryhung in https://github.com/flyteorg/flyte/pull/5864
* Quick fix for monitoring docs YAML by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5917
* Added documentation about Fixed and unique domains by @Murdock9803 in https://github.com/flyteorg/flyte/pull/5923
* Remove unnecessary joins for list node and task execution entities in flyteadmin db queries by @katrogan in https://github.com/flyteorg/flyte/pull/5935
* Binary IDL Attribute Access for Map Task by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5942
* Improve Documentation for Registering Workflows to avoid confusion  by @sumana-2705 in https://github.com/flyteorg/flyte/pull/5941
* Added some CICD best practices to the documentation by @Murdock9803 in https://github.com/flyteorg/flyte/pull/5827
* [Docs]Document clarifying notes about the data lifecycle by @popojk in https://github.com/flyteorg/flyte/pull/5922
* fixed [Docs] Spot/interruptible docs imply retries come from the user… Closes #3956 by @ap0calypse8 in https://github.com/flyteorg/flyte/pull/5938
* Added new page under Data types and IO section for tensorflow_types in flyte documentation by @sumana-2705 in https://github.com/flyteorg/flyte/pull/5807
* Add basic SASL and TLS support for Kafka cloud events by @Sovietaced in https://github.com/flyteorg/flyte/pull/5814
* Fix broken links by @ppiegaze in https://github.com/flyteorg/flyte/pull/5946
* feat: improve registration patterns docs by @siiddhantt in https://github.com/flyteorg/flyte/pull/5808
* Improve literal type string representation handling by @pingsutw in https://github.com/flyteorg/flyte/pull/5932
* Update propeller sharding docs - types needs to be capitalized by @cpaulik in https://github.com/flyteorg/flyte/pull/5860
* fix: align the default config output by @Terryhung in https://github.com/flyteorg/flyte/pull/5947
* [Docs] Fix doc links to blob literal and type by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/5952
* Fix remaining misuses of capturing the default file descriptors in flytectl unit tests by @eapolinario in https://github.com/flyteorg/flyte/pull/5950
* Reduce where clause fanout when updating workflow, node & task executions by @katrogan in https://github.com/flyteorg/flyte/pull/5953
* [flyteadmin][API] get control plane version by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5934
* Add multi file error aggregation strategy by @bgedik in https://github.com/flyteorg/flyte/pull/5795
* Fix: avoid log spam for log links generated during the pod's pending phase by @fg91 in https://github.com/flyteorg/flyte/pull/5945
* [Docs] Align Note with the Output Naming Convention by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/5919
* [DOC] add copy command examples and description by @mao3267 in https://github.com/flyteorg/flyte/pull/5782
* Hide generated launch plans starting with .flytegen in the UI by @troychiu in https://github.com/flyteorg/flyte/pull/5949
* Fix link in flyteidl README.md by @amitani in https://github.com/flyteorg/flyte/pull/5957
* Fix indentation for security block in auth_setup.rst by @jkhales in https://github.com/flyteorg/flyte/pull/5968
* [copilot][flytedirectory] multipart blob download by @wayner0628 in https://github.com/flyteorg/flyte/pull/5715
* Mark NamedEntityState reserved enum values by @katrogan in https://github.com/flyteorg/flyte/pull/5975
* [flytepropeller] Add tests in v1alpha by @DenChenn in https://github.com/flyteorg/flyte/pull/5896
* fix(admin): validate cron expression in launch plan schedule by @peterxcli in https://github.com/flyteorg/flyte/pull/5951
* [flytectl][MSGPACK IDL] Gate feature by setting ENV by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5976
* Ray Plugin - Use serviceAccount from Config if set by @peterghaddad in https://github.com/flyteorg/flyte/pull/5928
* DOC-666 Change "Accessing attributes" to "Accessing attributes in workflows" by @ppiegaze in https://github.com/flyteorg/flyte/pull/5886
* [Docs] Align code lines of TensorFlow types with flytesnacks by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/5983
* Decouple ray submitter, worker, and head resources by @Sovietaced in https://github.com/flyteorg/flyte/pull/5933
* [Upstream] [COR-2297/] Fix nested offloaded type validation (#552) by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5996
* docs(contribute): rewrite flyte contribute docs based on 4960 by @bearomorphism in https://github.com/flyteorg/flyte/pull/5260
* Upstream changes to fix token validity and utilizing inmemory creds source by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6001
* [Housekeeping] Enable lint flytecopilot by @wayner0628 in https://github.com/flyteorg/flyte/pull/6003
* Version flyte-binary helm chart and use flyte-binary-release docker images in releases by @eapolinario in https://github.com/flyteorg/flyte/pull/6010
* Docs rli bug by @cosmicBboy in https://github.com/flyteorg/flyte/pull/6008
* Remove unused environment variable by @bgedik in https://github.com/flyteorg/flyte/pull/5987
* Correct stow listing function by @bgedik in https://github.com/flyteorg/flyte/pull/6013
* Replace other instances of rli by @eapolinario in https://github.com/flyteorg/flyte/pull/6014
* Set HttpOnly and Secure flags in session cookies by @eapolinario in https://github.com/flyteorg/flyte/pull/5911
* Run CI in release branches by @eapolinario in https://github.com/flyteorg/flyte/pull/5901
* Update CODEOWNERS - remove nikki by @eapolinario in https://github.com/flyteorg/flyte/pull/6015
* Adopt protogetter by @eapolinario in https://github.com/flyteorg/flyte/pull/5981
* Add visibility control to dynamic log links by @pingsutw in https://github.com/flyteorg/flyte/pull/6000
* Bump github.com/golang-jwt/jwt/v4 from 4.5.0 to 4.5.1 by @dependabot in https://github.com/flyteorg/flyte/pull/6017
* Bump github.com/golang-jwt/jwt/v4 from 4.5.0 to 4.5.1 in /flyteadmin by @dependabot in https://github.com/flyteorg/flyte/pull/6020
* fix: return the config file not found error by @Terryhung in https://github.com/flyteorg/flyte/pull/5972
* Add option to install CRD as a part of chart install in flyte-binary by @marrrcin in https://github.com/flyteorg/flyte/pull/5967
* Deterministic error propagation for distributed (training) tasks by @fg91 in https://github.com/flyteorg/flyte/pull/5598
* Add DedupingBucketRateLimiter by @andrewwdye in https://github.com/flyteorg/flyte/pull/6025
* docs: update command for running the single binary by @machichima in https://github.com/flyteorg/flyte/pull/6031
* Inject offloading literal env vars by @eapolinario in https://github.com/flyteorg/flyte/pull/6027
* Fix Union type with dataclass ambiguous error and support superset comparison by @mao3267 in https://github.com/flyteorg/flyte/pull/5858
* [Fix] Add logger for compiler and marshal while comparing union by @mao3267 in https://github.com/flyteorg/flyte/pull/6034
* How to verify that the grpc service of flyteadmin works as expected by @popojk in https://github.com/flyteorg/flyte/pull/5958
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/3601
* [Docs] MessagePack IDL, Pydantic Support, and Attribute Access by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6022
* Revert "[COR-2297/] Fix nested offloaded type validation (#552) (#5996)" by @eapolinario in https://github.com/flyteorg/flyte/pull/6045
* Bump docker/build-push-action to v6 by @lowc1012 in https://github.com/flyteorg/flyte/pull/5747
* Add a new AgentError field in the agent protos by @RRap0so in https://github.com/flyteorg/flyte/pull/5916
* Update contribute_code.rst by @davidlin20dev in https://github.com/flyteorg/flyte/pull/6009
* Mark execution mode enum reserved by @katrogan in https://github.com/flyteorg/flyte/pull/6040
* Add is_eager bit to indicate eager tasks in flyte system by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6041
* Fix: Make appProtocols optional in flyteadmin and flyteconsole services in helm chart by @fg91 in https://github.com/flyteorg/flyte/pull/5944
* Fix: customize demo cluster port by @vincent0426 in https://github.com/flyteorg/flyte/pull/5969
* helm: Add support for passing env variables to flyteadmin using envFrom by @ctso in https://github.com/flyteorg/flyte/pull/5216
* Adding support for downloading offloaded literal in copilot by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6048
* Add tolerations for extended resources by @troychiu in https://github.com/flyteorg/flyte/pull/6033
* [Docs][flyteagent] Added description of exception deletion cases. by @SZL741023 in https://github.com/flyteorg/flyte/pull/6039
* Update note about Terraform reference implementations by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6056
* Minor fixes to single-cloud page by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6059
* [Docs] Align sd code lines with Flytesnacks example  by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/6063
* [copilot] rename sidecar to uploader by @wayner0628 in https://github.com/flyteorg/flyte/pull/6043
* add sub_node_interface field to array node by @pvditt in https://github.com/flyteorg/flyte/pull/6018
* [demo sandbox] Add env vars to plugin config by @wild-endeavor in https://github.com/flyteorg/flyte/pull/6072
* Fix inconsistent code examples in caching.md documentation by @davidlin20dev in https://github.com/flyteorg/flyte/pull/6077
* Support ArrayNode subNode timeouts by @hamersaw in https://github.com/flyteorg/flyte/pull/6054
* [WIP] docs: add raise user error section by @machichima in https://github.com/flyteorg/flyte/pull/6084
* Enable literal offloading in flyte-binary and flyte-core by @eapolinario in https://github.com/flyteorg/flyte/pull/6087
* Special-case type annotations to force cache invalidations for msgpack-binary literals by @eapolinario in https://github.com/flyteorg/flyte/pull/6078
* Regenerate flytectl docs by @neverett in https://github.com/flyteorg/flyte/pull/5914
* [ Docs ] fix: contribute link  by @taieeuu in https://github.com/flyteorg/flyte/pull/6076
* Update single-binary.yml end2end_execute by @taieeuu in https://github.com/flyteorg/flyte/pull/6061
* fix: modify deprecated functions by @machichima in https://github.com/flyteorg/flyte/pull/6052
* Use an empty config file to produce docs by @eapolinario in https://github.com/flyteorg/flyte/pull/6092

## New Contributors
* @mthemis-provenir made their first contribution in https://github.com/flyteorg/flyte/pull/5796
* @Murdock9803 made their first contribution in https://github.com/flyteorg/flyte/pull/5828
* @RaghavMangla made their first contribution in https://github.com/flyteorg/flyte/pull/5826
* @DenChenn made their first contribution in https://github.com/flyteorg/flyte/pull/5844
* @JiangJiaWei1103 made their first contribution in https://github.com/flyteorg/flyte/pull/5874
* @popojk made their first contribution in https://github.com/flyteorg/flyte/pull/5876
* @10sharmashivam made their first contribution in https://github.com/flyteorg/flyte/pull/5878
* @aminmaghsodi made their first contribution in https://github.com/flyteorg/flyte/pull/5816
* @taieeuu made their first contribution in https://github.com/flyteorg/flyte/pull/5873
* @Terryhung made their first contribution in https://github.com/flyteorg/flyte/pull/5864
* @sumana-2705 made their first contribution in https://github.com/flyteorg/flyte/pull/5941
* @ap0calypse8 made their first contribution in https://github.com/flyteorg/flyte/pull/5938
* @siiddhantt made their first contribution in https://github.com/flyteorg/flyte/pull/5808
* @mao3267 made their first contribution in https://github.com/flyteorg/flyte/pull/5782
* @amitani made their first contribution in https://github.com/flyteorg/flyte/pull/5957
* @jkhales made their first contribution in https://github.com/flyteorg/flyte/pull/5968
* @peterxcli made their first contribution in https://github.com/flyteorg/flyte/pull/5951
* @bearomorphism made their first contribution in https://github.com/flyteorg/flyte/pull/5260
* @marrrcin made their first contribution in https://github.com/flyteorg/flyte/pull/5967
* @machichima made their first contribution in https://github.com/flyteorg/flyte/pull/6031
* @davidlin20dev made their first contribution in https://github.com/flyteorg/flyte/pull/6009
* @vincent0426 made their first contribution in https://github.com/flyteorg/flyte/pull/5969
* @ctso made their first contribution in https://github.com/flyteorg/flyte/pull/5216
* @SZL741023 made their first contribution in https://github.com/flyteorg/flyte/pull/6039

**Link to full changelog**: https://github.com/flyteorg/flyte/compare/v1.13.3...v1.14.0

