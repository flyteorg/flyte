# Flyte v1.12.0-b0

## What's Changed
* [Docs] Remove kustomize link in secrets.md doc by @lowc1012 in https://github.com/flyteorg/flyte/pull/5043
* CI workflow for helm charts and manifests by @lowc1012 in https://github.com/flyteorg/flyte/pull/5027
* Update spark-on-k8s-operator address in helm charts by @eapolinario in https://github.com/flyteorg/flyte/pull/5057
* Fix wrong syntax for path filtering in validate-helm-charts.yaml by @lowc1012 in https://github.com/flyteorg/flyte/pull/5056
* Fix: flyte-secret-auth secret not mounted properly in flyte-core by @lowc1012 in https://github.com/flyteorg/flyte/pull/5054
* Match flytekit versions used to register and run functional tests by @eapolinario in https://github.com/flyteorg/flyte/pull/5059
* integration test config by @troychiu in https://github.com/flyteorg/flyte/pull/5058
* Add org as an optional request param to dataproxy CreateUploadLocation by @katrogan in https://github.com/flyteorg/flyte/pull/5060
* Add k8s env from by @neilisaur in https://github.com/flyteorg/flyte/pull/4969
* Implement GetProject endpoint in FlyteAdmin by @RRap0so in https://github.com/flyteorg/flyte/pull/4825
* Prepopulate ArrayNode output literals with TaskNode interface output variables by @hamersaw in https://github.com/flyteorg/flyte/pull/5080
* [House Keeping] deprecate MaxDatasetSizeBytes propeller config in favor of GetLimitMegabytes storage config by @pvditt in https://github.com/flyteorg/flyte/pull/4852
* Fix lint errors caught by `chart-testing` by @eapolinario in https://github.com/flyteorg/flyte/pull/5072
* Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/5093
* sagemaker agent backend setup documentation by @samhita-alla in https://github.com/flyteorg/flyte/pull/5064
* add first version of gke-starter values file by @cjidboon94 in https://github.com/flyteorg/flyte/pull/5026
* Fix open ai secret name by @eapolinario in https://github.com/flyteorg/flyte/pull/5098
* Allow setting a ExecutionClusterLabel when triggering a Launchplan/Workflow/Task by @RRap0so in https://github.com/flyteorg/flyte/pull/4998
* Improve audience mismatch debugging by @ddl-rliu in https://github.com/flyteorg/flyte/pull/5078
* docs(sandbox): Add guide for running newest flyteconsole in flyte sandbox by @MortalHappiness in https://github.com/flyteorg/flyte/pull/5100
* Remove unnecessary step and fix numbering in code examples by @eapolinario in https://github.com/flyteorg/flyte/pull/5104
* fix rendering of flyte-core and flyteagent charts by @pbrogan12 in https://github.com/flyteorg/flyte/pull/5048
* Bump golang.org/x/net from 0.3.1-0.20221206200815-1e63c2f08a10 to 0.7.0 in /docker/sandbox-bundled/bootstrap by @dependabot in https://github.com/flyteorg/flyte/pull/3390
* Update repeated value filters with ValueNotIn support by @troychiu in https://github.com/flyteorg/flyte/pull/5110
* Update container builds from go 1.21.5 to 1.21.latest by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5097
* Add optional org param to ProjectGetRequest by @katrogan in https://github.com/flyteorg/flyte/pull/5118
* Stop building read-the-docs for flyteidl by @eapolinario in https://github.com/flyteorg/flyte/pull/5120
* Bump google.golang.org/grpc and otelgrpc by @eapolinario in https://github.com/flyteorg/flyte/pull/5121
* Fix broken link in "Mapping Python to Flyte types" table by @neverett in https://github.com/flyteorg/flyte/pull/5122
* Bump version of otel and grpc in flytestdlib by @eapolinario in https://github.com/flyteorg/flyte/pull/5123
* [flyteadmin] Show diff structure when re-registration two different task with same ids by @austin362667 in https://github.com/flyteorg/flyte/pull/4924
* Set flyteadmin grpc port correctly in config / Flyte-core flyteadmin / datacatalog expose ports by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5013
* Add flyte-core missing priorityClassName to webhook values by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/4987
* Update environment_setup.md by @ssen85 in https://github.com/flyteorg/flyte/pull/4963
* [Docs] Open external links in new tab by @MortalHappiness in https://github.com/flyteorg/flyte/pull/4966
* Fix separateGrpcIngress flag not working in flyte-binary helm chart  by @lowc1012 in https://github.com/flyteorg/flyte/pull/4946
* docs(contribute): Change go mod tidy to make go-tidy by @MortalHappiness in https://github.com/flyteorg/flyte/pull/5131
* Fix execution phase by @troychiu in https://github.com/flyteorg/flyte/pull/5127
* Add trailing slash to compile make target by @eapolinario in https://github.com/flyteorg/flyte/pull/4648
* Change retry error from RuntimeError to FlyteRecoverableException by @dansola in https://github.com/flyteorg/flyte/pull/5128
* Adapt ray flyteplugin to Kuberay 1.1.0 by @ByronHsu in https://github.com/flyteorg/flyte/pull/5067
* Boilerplate simplification by @eapolinario in https://github.com/flyteorg/flyte/pull/5134
* Upgrade cloudevents to v2.15.2 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5142
* update sagemaker agent setup doc as secrets aren't required anymore by @samhita-alla in https://github.com/flyteorg/flyte/pull/5138
* Upgrade logrus to v1.9.3 everywhere by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5139
* Upgrade go-restful to v3.12.0 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5140
* Regenerate ray pflags by @eapolinario in https://github.com/flyteorg/flyte/pull/5149
* Split access token into half and store to avoid "securecookie: the value is too long" error by @yubofredwang in https://github.com/flyteorg/flyte/pull/4863
* Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/5150
* Bump golang.org/x/crypto from 0.11.0 to 0.17.0 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/5148
* use javascript to open new tab for external links by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5159
* Update K8s plugin config docs by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5070
* Update Propeller architecture documentation by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5117
* update protobuf v1.32.0 -> v1.33.0 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5156
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flyte/pull/5143
* Upgrade grpc health probe 0.4.11 -> 0.4.25 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5153
* Upgrade go-jose v3.0.0 -> v3.0.3 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5154
* docs: update agent development documentation by @pingsutw in https://github.com/flyteorg/flyte/pull/5130
* Add variables to ease separate bucket config by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5015
* Bump golang.org/x/net from 0.12.0 to 0.17.0 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/5146
* Bump google.golang.org/protobuf from 1.30.0 to 1.33.0 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/5147
* Upgrade lestrrat-go/jwx to v1.2.29 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5141
* Revert "Pin flyteconsole version in release process (#5037)" by @eapolinario in https://github.com/flyteorg/flyte/pull/5176
* Add identity to task execution metadata by @noahjax in https://github.com/flyteorg/flyte/pull/5105
* Bump k8s.io/client-go from 0.0.0-20210217172142-7279fc64d847 to 0.17.16 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/5145
* Upgrade jackc/pgconn v1.14.1 -> v1.14.3 / pgx/v5 v5.4.3 -> v5.5.5 / pgproto3 v2.3.2 -> v2.3.3 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5155
* fix make link error by @novahow in https://github.com/flyteorg/flyte/pull/5175
* Stop admin launcher copying shard key from parent workflow by @Tom-Newton in https://github.com/flyteorg/flyte/pull/5174
* Fix Id bigint conversation for not yet created table by @ongkong in https://github.com/flyteorg/flyte/pull/5157
* Add tracking for active node and task execution counts in propeller by @sshardool in https://github.com/flyteorg/flyte/pull/4986
* [House keeping] include container statuses for all container exit errors by @pvditt in https://github.com/flyteorg/flyte/pull/5161
* docs: add missing key in auth guide by @Jeinhaus in https://github.com/flyteorg/flyte/pull/5169
* Shallow copying EnvironmentVariables map before injecting ArrayNode env vars by @hamersaw in https://github.com/flyteorg/flyte/pull/5182
* Feature/array node workflow parallelism by @pvditt in https://github.com/flyteorg/flyte/pull/5062
* Fix streak length metric reporting by @Tom-Newton in https://github.com/flyteorg/flyte/pull/5172
* Fix path to AuthMetadataService in flyte-binary chart by @eapolinario in https://github.com/flyteorg/flyte/pull/5185
* Change phase to queue on job submit for webapi plugins by @pingsutw in https://github.com/flyteorg/flyte/pull/5188
* [Docs] Testing agents in the development environment by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5106
* Use ratelimiter config in webapi plugins by @kumare3 in https://github.com/flyteorg/flyte/pull/5190
* docs(ray): Update kuberay documentation by @MortalHappiness in https://github.com/flyteorg/flyte/pull/5179
* Change phase to WaitingForResources when quota exceeded by @pingsutw in https://github.com/flyteorg/flyte/pull/5195
* Fix: Update spark operator helm repository by @fg91 in https://github.com/flyteorg/flyte/pull/5198
* docs(troubleshoot): Add docker error troubleshooting guide by @MortalHappiness in https://github.com/flyteorg/flyte/pull/4972
* add cache client read and write otel tracing by @pvditt in https://github.com/flyteorg/flyte/pull/5184
* Fix FlyteIDL docs link by @neverett in https://github.com/flyteorg/flyte/pull/5199
* [easy] [flyteagent] Add `ExecuteTaskSync` function timeout setting by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5209
* [easy] [flyteagent]  Add `agent-service` endpoint settings for `flyte-core` deployment by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5208
* Update Monitoring documentation by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5206
* chore: remove obsolete flyte config files by @pingsutw in https://github.com/flyteorg/flyte/pull/5196
* Generate rust grpc using tonic by @eapolinario in https://github.com/flyteorg/flyte/pull/5187
* enable parallelism to be set to nil for array node by @pvditt in https://github.com/flyteorg/flyte/pull/5214
* Fix mounting secrets by @yini7777 in https://github.com/flyteorg/flyte/pull/5063
* Update "Creating a Flyte project" with link to new Dockerfile project template by @neverett in https://github.com/flyteorg/flyte/pull/5215
* Re-apply changes  to dataclass docs from flytesnacks#1553 by @neverett in https://github.com/flyteorg/flyte/pull/5211
* feat(ray): Remove initContainers by @MortalHappiness in https://github.com/flyteorg/flyte/pull/5178
* fix(databricks): Check the response body before unmarshal by @pingsutw in https://github.com/flyteorg/flyte/pull/5226
* perf(cache): Use AddRateLimited for batch enqueue by @pingsutw in https://github.com/flyteorg/flyte/pull/5228
* update scroll behavior so that sidebar maintains location by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5229
* propagate dark/light theme to algolia search bar by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5231
* Added additional port configuration for flyte services by @hamersaw in https://github.com/flyteorg/flyte/pull/5233
* [BUG] fix(doc): Wrong configuration in spark plugin with binary chart by @lowc1012 in https://github.com/flyteorg/flyte/pull/5230
* Fix grammatical error in workflow lifecycle docs by @Sovietaced in https://github.com/flyteorg/flyte/pull/5227
* add plugins support for k8s imagepullpolicy by @novahow in https://github.com/flyteorg/flyte/pull/5167
* Added section on overriding lp config in loop by @pryce-turner in https://github.com/flyteorg/flyte/pull/5223
* Added unmarshal attribute for expires_in for device flow auth token by @eapolinario in https://github.com/flyteorg/flyte/pull/4319
* Update template to link issue for closing by @thomasjpfan in https://github.com/flyteorg/flyte/pull/5239
* [Feature] add retries and backoffs for propeller sending events to admin by @pvditt in https://github.com/flyteorg/flyte/pull/5166
* Explain how to enable/disable local caching by @eapolinario in https://github.com/flyteorg/flyte/pull/5242
* [House keeping] remove setting max size bytes in node context  by @pvditt in https://github.com/flyteorg/flyte/pull/5092
* Fix support for limit-namespace in FlytePropeller by @hamersaw in https://github.com/flyteorg/flyte/pull/5238
* [WAIT TO MERGE] Use remoteliteralinclude for code in user guide docs by @neverett in https://github.com/flyteorg/flyte/pull/5207
* refactor(python): Replace os.path with pathlib by @MortalHappiness in https://github.com/flyteorg/flyte/pull/5243
* Containerize documentation build environment and add sphinx-autobuild for hot-reload by @MortalHappiness in https://github.com/flyteorg/flyte/pull/4960
* Finalize flyteidl Rust crate by @austin362667 in https://github.com/flyteorg/flyte/pull/5219
* Use sphinx design instead of sphinx panels by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5254
* [Docs] add envd plugin installation command to the environment setup guide by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5257
* [Docs] add colon rule in version by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5258
* [easy] [Docs] Raw Container Local Execution by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5262
* [Docs] add validation file type by @jasonlai1218 in https://github.com/flyteorg/flyte/pull/5259
* WebAPI plugins optimization by @pingsutw in https://github.com/flyteorg/flyte/pull/5237
* fix(webapi): Ensure cache deletion on abort workflow by @pingsutw in https://github.com/flyteorg/flyte/pull/5235
* RunLLM Widget Configuration by @agiron123 in https://github.com/flyteorg/flyte/pull/5266
* fix dropdown formatting by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5276
* added configuration for arraynode default parallelism behavior by @hamersaw in https://github.com/flyteorg/flyte/pull/5268
* Improve error message for limit exceeded by @pingsutw in https://github.com/flyteorg/flyte/pull/5275
* test sphinx-reredirects by @cosmicBboy in https://github.com/flyteorg/flyte/pull/5281
* enabling parallelism controls on arraynode by @hamersaw in https://github.com/flyteorg/flyte/pull/5284
* Move deprecated integrations docs to flyte by @neverett in https://github.com/flyteorg/flyte/pull/5283
* Fix rli number ranges in Snowflake plugin example doc by @neverett in https://github.com/flyteorg/flyte/pull/5287
* Fix broken gpu resource override when using pod templates by @fg91 in https://github.com/flyteorg/flyte/pull/4925
* Fix order of arguments in copilot log by @eapolinario in https://github.com/flyteorg/flyte/pull/5292
* fix(databricks): Handle FAILED state as retryable error by @pingsutw in https://github.com/flyteorg/flyte/pull/5277

## New Contributors
* @RRap0so made their first contribution in https://github.com/flyteorg/flyte/pull/4825
* @cjidboon94 made their first contribution in https://github.com/flyteorg/flyte/pull/5026
* @ddl-rliu made their first contribution in https://github.com/flyteorg/flyte/pull/5078
* @pbrogan12 made their first contribution in https://github.com/flyteorg/flyte/pull/5048
* @austin362667 made their first contribution in https://github.com/flyteorg/flyte/pull/4924
* @ssen85 made their first contribution in https://github.com/flyteorg/flyte/pull/4963
* @dansola made their first contribution in https://github.com/flyteorg/flyte/pull/5128
* @noahjax made their first contribution in https://github.com/flyteorg/flyte/pull/5105
* @ongkong made their first contribution in https://github.com/flyteorg/flyte/pull/5157
* @sshardool made their first contribution in https://github.com/flyteorg/flyte/pull/4986
* @Jeinhaus made their first contribution in https://github.com/flyteorg/flyte/pull/5169
* @yini7777 made their first contribution in https://github.com/flyteorg/flyte/pull/5063
* @Sovietaced made their first contribution in https://github.com/flyteorg/flyte/pull/5227
* @agiron123 made their first contribution in https://github.com/flyteorg/flyte/pull/5266
