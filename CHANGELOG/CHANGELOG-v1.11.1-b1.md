# Flyte v1.11.1-b1

## What's Changed
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

## New Contributors
* @cjidboon94 made their first contribution in https://github.com/flyteorg/flyte/pull/5026
* @ddl-rliu made their first contribution in https://github.com/flyteorg/flyte/pull/5078
* @pbrogan12 made their first contribution in https://github.com/flyteorg/flyte/pull/5048
* @austin362667 made their first contribution in https://github.com/flyteorg/flyte/pull/4924
* @ssen85 made their first contribution in https://github.com/flyteorg/flyte/pull/4963
* @dansola made their first contribution in https://github.com/flyteorg/flyte/pull/5128
