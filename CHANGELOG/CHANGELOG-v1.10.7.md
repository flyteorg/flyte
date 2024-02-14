# Flyte 1.10.7 Release Notes

We're excited to share the release of Flyte 1.10.7, featuring a broad spectrum of updates, improvements, and bug fixes across the Flyte ecosystem. This release marks a pivotal shift in our development approach, notably with our adoption of [buf](https://github.com/flyteorg/flyte/pull/4806) for protobuf stub generation. This move optimizes our development workflow and discontinues the automatic creation of Java and C++ stubs, making it easier to adapt the generated code for other languages as needed. Additionally, we've upgraded to gRPC-gateway v2, aligning with the latest advancements and recommendations found in the [v2 migration guide](https://grpc-ecosystem.github.io/grpc-gateway/docs/development/grpc-gateway_v2_migration_guide/).

Our sincere gratitude goes to all contributors for their invaluable efforts towards this release.

## Core Improvements and Bug Fixes

- Improved error handling for transient secret sync issues, enhancing the robustness of secret management. [[PR #4310]](https://github.com/flyteorg/flyte/pull/4310)
- Introduced Sphinx build for Monodocs, improving documentation generation and integration. [[PR #4347]](https://github.com/flyteorg/flyte/pull/4347)
- Enhanced the Spark plugin by fixing the environment variable `ValueFrom` for pod templates, allowing for more dynamic configurations. [[PR #4532]](https://github.com/flyteorg/flyte/pull/4532)
- Optimized fastcache behavior to not cache lookups on node skip, reducing unnecessary cache hits. [[PR #4524]](https://github.com/flyteorg/flyte/pull/4524)
- Removed composition errors from branch nodes, streamlining execution paths. [[PR #4528]](https://github.com/flyteorg/flyte/pull/4528)
- Added support for ignoring warnings related to AWS SageMaker imports, improving integration compatibility. [[PR #4540]](https://github.com/flyteorg/flyte/pull/4540)
- Fixed a bug related to setting the service account from PodTemplate, ensuring correct service account usage. [[PR #4536]](https://github.com/flyteorg/flyte/pull/4536)
- Addressed flaky tests in test_monitor, enhancing test reliability. [[PR #4537]](https://github.com/flyteorg/flyte/pull/4537)
- Updated the boilerplate version and contribution guide, facilitating better community contributions. [[PR #4541]](https://github.com/flyteorg/flyte/pull/4541), [[PR #4501]](https://github.com/flyteorg/flyte/pull/4501)
- Improved documentation build processes by manually creating version files and introducing a conda-lock file for consistent environment setup. [[PR #4556]](https://github.com/flyteorg/flyte/pull/4556), [[PR #4553]](https://github.com/flyteorg/flyte/pull/4553)
- Enhanced array node evaluation frequency optimization by detecting subNode phase updates. [[PR #4535]](https://github.com/flyteorg/flyte/pull/4535)
- Introduced support for failure nodes, allowing workflows to handle failures more gracefully. [[PR #4308]](https://github.com/flyteorg/flyte/pull/4308)
- Made various updates to Go versions, plugin integrations, and GitHub workflows to enhance performance and developer experience. [[PR #4534]](https://github.com/flyteorg/flyte/pull/4534), [[PR #4582]](https://github.com/flyteorg/flyte/pull/4582), [[PR #4589]](https://github.com/flyteorg/flyte/pull/4589)
- Addressed several bugs and made improvements in caching, metadata handling, and task execution, further stabilizing the Flyte platform. [[PR #4594]](https://github.com/flyteorg/flyte/pull/4594), [[PR #4590]](https://github.com/flyteorg/flyte/pull/4590), [[PR #4607]](https://github.com/flyteorg/flyte/pull/4607)
- Streamlined development workflow with the transition to buf for generating protobuf stubs, ceasing the automatic generation of Java and C++ stubs.
- Upgraded to grpc-gateway v2, optimizing API performance and compatibility.

## Plugin and Integration Enhancements

- Added new features and fixed bugs in the Spark plugin, Ray Autoscaler integration, and other areas, expanding Flyte's capabilities and integration ecosystem. [[PR #4363]](https://github.com/flyteorg/flyte/pull/4363)
- Updated various dependencies and configurations, ensuring compatibility and security. [[PR #4571]](https://github.com/flyteorg/flyte/pull/4571), [[PR #4643]](https://github.com/flyteorg/flyte/pull/4643)
- Improved the handling and documentation of plugin secrets management, making it easier for users to manage sensitive information. [[PR #4732]](https://github.com/flyteorg/flyte/pull/4732)

## Documentation and Community

- Updated community meeting cadence and contribution guidelines, fostering a more engaged and welcoming community. [[PR #4699]](https://github.com/flyteorg/flyte/pull/4699)
- Enhanced documentation through various updates, including the introduction of a new architecture image for FlytePlugins and clarification of propeller scaling. [[PR #4661]](https://github.com/flyteorg/flyte/pull/4661), [[PR #4741]](https://github.com/flyteorg/flyte/pull/4741)

## Full Changelog
- Fix transient secret sync error handling by @Tom-Newton in https://github.com/flyteorg/flyte/pull/4310
- Monodocs sphinx build by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4347
- [Spark plugin] Fix environment variable ValueFrom for pod templates by @Tom-Newton in https://github.com/flyteorg/flyte/pull/4532
- fastcache should not cache lookup on node skip by @hamersaw in https://github.com/flyteorg/flyte/pull/4524
- Removed composition error from branch node by @hamersaw in https://github.com/flyteorg/flyte/pull/4528
- ignore warnings related to awssagemaker import by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4540
- [BUG] Fix setting of service_account from PodTemplate by @pvditt in https://github.com/flyteorg/flyte/pull/4536
- Fix flaky test_monitor by @pingsutw in https://github.com/flyteorg/flyte/pull/4537
- Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flyte/pull/4541
- remove hardcoded list of tests by @samhita-alla in https://github.com/flyteorg/flyte/pull/4521
- manually create flytekit/_version.py file in docs build by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4556
- introduce conda-lock file for docs by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4553
- Detect subNode phase updates to reduce evaluation frequency of ArrayNode by @hamersaw in https://github.com/flyteorg/flyte/pull/4535
- Add support failure node by @pingsutw in https://github.com/flyteorg/flyte/pull/4308
- Return InvalidArgument for workflow compilation failures in CreateWorkflow by @katrogan in https://github.com/flyteorg/flyte/pull/4566
- Update to go 1.21 by @eapolinario in https://github.com/flyteorg/flyte/pull/4534
- Update contribution guide by @pingsutw in https://github.com/flyteorg/flyte/pull/4501
- Add flyin plugin to monodocs integrations page by @neverett in https://github.com/flyteorg/flyte/pull/4582
- Use updated cronSchedule in CreateLaunchPlanModel by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/4564
- Writing zero length inputs by @hamersaw in https://github.com/flyteorg/flyte/pull/4594
- Feature/add pod pending timeout config by @pvditt in https://github.com/flyteorg/flyte/pull/4590
- Run single-binary gh workflows on all PRs by @eapolinario in https://github.com/flyteorg/flyte/pull/4589
- auto-generate toctree from flytesnacks index.md docs by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4587
- add repo tag and commit associated with the build by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4571
- monodocs - gracefully handle case when external repo doesn't contain tags: use current commit by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4598
- convert commit to string by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4599
- Bug/abort map task subtasks by @pvditt in https://github.com/flyteorg/flyte/pull/4506
- Supporting parallelized workers in ArrayNode subNodes by @hamersaw in https://github.com/flyteorg/flyte/pull/4567
- Don't use experimental readthedocs build.commands config by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4606
- Ignore cache variables by @hamersaw in https://github.com/flyteorg/flyte/pull/4618
- Feature/add cleanup non recoverable pod statuses by @pvditt in https://github.com/flyteorg/flyte/pull/4607
- Agent Metadata Servicer by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4511
- Add Flyin propeller config by @eapolinario in https://github.com/flyteorg/flyte/pull/4610
- Correctly computing ArrayNode maximum attempts and system failures by @hamersaw in https://github.com/flyteorg/flyte/pull/4627
- Agent Sync Plugin by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4107
- Add github token in buf gh action by @eapolinario in https://github.com/flyteorg/flyte/pull/4626
- Update flyte-binary values by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4604
- Fixing cache overwrite metadata update by @hamersaw in https://github.com/flyteorg/flyte/pull/4617
- Fixing 100 kilobyte max error message size by @hamersaw in https://github.com/flyteorg/flyte/pull/4631
- Add Ray Autoscaler to the Flyte-Ray plugin by @Yicheng-Lu-llll in https://github.com/flyteorg/flyte/pull/4363
- Artifact protos and related changes by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4474
- Remove protoc-gen-validate by @eapolinario in https://github.com/flyteorg/flyte/pull/4643
- Readme update 2023 by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4549
- Fixing ArrayNode integration with backoff controller by @hamersaw in https://github.com/flyteorg/flyte/pull/4640
- Avoid to use the http.DefaultClient by @andresgomezfrr in https://github.com/flyteorg/flyte/pull/4667
- Update dns policy for sandbox buildkit instance to ClusterFirstWithHo… by @jeevb in https://github.com/flyteorg/flyte/pull/4678
- Updating ArrayNode ExternalResourceInfo ID by @hamersaw in https://github.com/flyteorg/flyte/pull/4677
- Feat: Inject user identity as pod label in K8s plugin by @fg91 in https://github.com/flyteorg/flyte/pull/4637
- Artifacts shell 2 by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4649
- Improve Agent Metadata Service Error Message by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4682
- move pod start/end time to a common template vars by @vraiyaninv in https://github.com/flyteorg/flyte/pull/4676
- switch readthedocs config to monodocs build by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4687
- Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4690
- Add GetTaskMetrics and GetTaskLogs to agent by @pingsutw in https://github.com/flyteorg/flyte/pull/4662
- add algolia searchbar by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4696
- monodocs: do not use beta releases when importing projects by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4712
- add cache evicted status by @pvditt in https://github.com/flyteorg/flyte/pull/4705
- Update community meeting cadence by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4699
- Remove dockerfiles from subfolder by @pingsutw in https://github.com/flyteorg/flyte/pull/4715
- Update to artifact idl - Add List Usage endpoint by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4714
- monodocs uses flytekit/flytectl index rst file by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4720
- Replace grpc gateway endpoints with post by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4717
- [BUG] Retry fetching subworkflow output data on failure by @pvditt in https://github.com/flyteorg/flyte/pull/4602
- Option to clear node state on any termination by @Tom-Newton in https://github.com/flyteorg/flyte/pull/4596
- Add org to identifier protos by @katrogan in https://github.com/flyteorg/flyte/pull/4663
- Update docs for plugin secrets management by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4732
- Reintroduce k8s client fallback to cache lookups by @hamersaw in https://github.com/flyteorg/flyte/pull/4733
- Remove unused validate files by @eapolinario in https://github.com/flyteorg/flyte/pull/4644
- delete old docs by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4742
- Docs/Clarify propeller scaling by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4741
- Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4744
- Add org to all flyteadmin endpoints for consistency by @katrogan in https://github.com/flyteorg/flyte/pull/4746
- update conda lock file by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4749
- Use logger with formatter by @andrewwdye in https://github.com/flyteorg/flyte/pull/4747
- [housekeeping] Remove pull_request_template from each subdirectory by @pingsutw in https://github.com/flyteorg/flyte/pull/4753
- [Docs] Reapply Databricks agent docs changes from #4008 by @neverett in https://github.com/flyteorg/flyte/pull/4751
- Fix test get logs template uri test by @eapolinario in https://github.com/flyteorg/flyte/pull/4760
- Small formatting fixes for Databrick agents docs by @neverett in https://github.com/flyteorg/flyte/pull/4758
- [housekeeping] Remove flytearchives by @pingsutw in https://github.com/flyteorg/flyte/pull/4761
- Guard against open redirect URL parameters in login by @katrogan in https://github.com/flyteorg/flyte/pull/4763
- Wrapping k8s client with write filter and cache reader by @hamersaw in https://github.com/flyteorg/flyte/pull/4752
- [BUG] Handle Potential Indefinite Propeller Update Loops by @pvditt in https://github.com/flyteorg/flyte/pull/4755
- Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4768
- Deprecated Agent State to Agent Phase by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4738
- Add docs build process readme by @ppiegaze in https://github.com/flyteorg/flyte/pull/4772
- GetDynamicNodeWorkflow endpoint by @iaroslav-ciupin in https://github.com/flyteorg/flyte/pull/4689
- [BUG] subworkflow timeout propagation by @pvditt in https://github.com/flyteorg/flyte/pull/4766
- Update artifact IDL with new time partition by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4737
- Proto changes by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4778
- Updates for onboarding docs revamp by @neverett in https://github.com/flyteorg/flyte/pull/4548
- [Docs] Fix toctree links to User Guide, Environment Setup, and Contributing sections by @neverett in https://github.com/flyteorg/flyte/pull/4781
- docs: add FlytePlugins architecture image by @jasonlai1218 in https://github.com/flyteorg/flyte/pull/4661
- Fix repeated items in left nav by @ppiegaze in https://github.com/flyteorg/flyte/pull/4783
- [Docs] Remove broken link from Understand How Flyte Handles Data page (for new monodocs site) (second attempt) by @neverett in https://github.com/flyteorg/flyte/pull/4757
- docs: update Flyte sandbox configuration and documentation by @jasonlai1218 in https://github.com/flyteorg/flyte/pull/4729
- Replace `Storage` To `Ephemeral Storage` in Helm Chart by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4782
- Fix webhook typo, add podLabels, add podEnv to flyte-core Helm chart by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4756
- Dynamic log links by @eapolinario in https://github.com/flyteorg/flyte/pull/4774
- Remove storage as a task resource option by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4658
- feat: add apache 2.0 license to python flyteidl by @michaeltinsley in https://github.com/flyteorg/flyte/pull/4786
- Reduce maptask transitions between WaitingForResources and CheckingSubtaskExecutions by @hamersaw in https://github.com/flyteorg/flyte/pull/4790
- propeller gc ttl comparison should allow 23 by @hamersaw in https://github.com/flyteorg/flyte/pull/4791
- Bring Scheme back for backwards compatibility by @eapolinario in https://github.com/flyteorg/flyte/pull/4789
- Pass secret to invocation of go_generate gh workflow by @eapolinario in https://github.com/flyteorg/flyte/pull/4630
- [BUG] handle potential uncaught OOMKilled terminations by @pvditt in https://github.com/flyteorg/flyte/pull/4793
- Update additional bindings for org in path to be consistent by @katrogan in https://github.com/flyteorg/flyte/pull/4795
- Update pyflyte serve into pyflyte serve agent by @chaohengstudent in https://github.com/flyteorg/flyte/pull/4526
- Agent ClientSet by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4718
- Support kuberay v1.0.0 by @Yicheng-Lu-llll in https://github.com/flyteorg/flyte/pull/4656
- Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4803
- install latest flyteidl when building monodocs by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4815
- Rewrite GetExecutionData path additional bindings for org by @katrogan in https://github.com/flyteorg/flyte/pull/4816
- update docs README environment setup by @cosmicBboy in https://github.com/flyteorg/flyte/pull/4819
- Flyte-core add missing nodeSelector values by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4808
- Flyte-core add missing imagePullSecrets support by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4810
- Logger disable HTML escaping by @andrewwdye in https://github.com/flyteorg/flyte/pull/4828
- Move intro docs from flytesnacks to flyte by @ppiegaze in https://github.com/flyteorg/flyte/pull/4814
- Flyte-agent configure pod securityContext by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4785
- [Docs] add sandbox to local cluster resource path by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4837
- Flyte-core add missing podEnv values by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4807
- Flyte-core Expose propeller webhook port 9443 in charts by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4804
- Align dir structure and URL structure with left nav hierarchy by @ppiegaze in https://github.com/flyteorg/flyte/pull/4843
- Generate version with setuptools_scm and migrate to pyproject.toml by @pingsutw in https://github.com/flyteorg/flyte/pull/4799
- MNT Fixes packaging for flyteidl wheel by @thomasjpfan in https://github.com/flyteorg/flyte/pull/4846
- Use buf to generate stubs by @eapolinario in https://github.com/flyteorg/flyte/pull/4806
- [FLYTE-486] Support selecting IDP based on the query parameter by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/4838
- Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4847
- Add plugin_config for agent by @pingsutw in https://github.com/flyteorg/flyte/pull/4848
- Adds MANIFEST.in for flyteidl by @thomasjpfan in https://github.com/flyteorg/flyte/pull/4850
- Verify unbounded inputs for all scheduled launch plan types by @katrogan in https://github.com/flyteorg/flyte/pull/4867
- Fix npm publish of flyteidl package by @eapolinario in https://github.com/flyteorg/flyte/pull/4861
- Remove protoc_gen_swagger by @eapolinario in https://github.com/flyteorg/flyte/pull/4860
- Create CODEOWNERS file and add docs team by @neverett in https://github.com/flyteorg/flyte/pull/4857
- [Docs] update outdated link to on-prem tutorial by @ALMerrill in https://github.com/flyteorg/flyte/pull/4868
- Re-add link to hosted sandbox by @neverett in https://github.com/flyteorg/flyte/pull/4856
- Fix asterisk in cron table being rendered as list item by @neverett in https://github.com/flyteorg/flyte/pull/4836
- Add notes to selfAuth with Azure docs by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4835
- Add protos to support cache overrides by @hamersaw in https://github.com/flyteorg/flyte/pull/4820
- Flyte-core add support for ingressClassName in ingress by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4805
- Update flyte docs build directions by @ppiegaze in https://github.com/flyteorg/flyte/pull/4862
- Replaced deprecated bitnami/bitnami-shell image with bitnami/os-shell by @kamaleybov in https://github.com/flyteorg/flyte/pull/4882
- Flyte-core define pod and container securityContext by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4809
- Leverage KubeRay v1 instead of v1alpha1 for resources by @peterghaddad in https://github.com/flyteorg/flyte/pull/4818

## New Contributors

- A warm welcome to our new contributors: [@pvditt](https://github.com/pvditt), [@ppiegaze](https://github.com/ppiegaze), [@jasonlai1218](https://github.com/jasonlai1218), and [@ddl-ebrown](https://github.com/ddl-ebrown). Thank you for your contributions to the Flyte community!
