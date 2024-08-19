# Flyte 1.13.0 Release Notes

Flyte 1.13.0 brings a host of new features, optimizations, and fixes, enhancing the platform's functionality and user experience. This release also welcomes several new contributors to the Flyte community. Below are the highlights of this release.

## 🚀 New Features & Improvements

1. **Features**
   - [Watch agent metadata service](https://github.com/flyteorg/flyte/pull/5017)
   - [Allow controlling in which task phases log links are shown](https://github.com/flyteorg/flyte/pull/4726)
   - [Add executionClusterLabel](https://github.com/flyteorg/flyte/pull/5394)

2. **Helm Charts & Manifests**
   - [Grafana dashboard updates](https://github.com/flyteorg/flyte/pull/5255) for better monitoring and observability.
   - [Core helm chart updates for propeller configuration of agent service](https://github.com/flyteorg/flyte/pull/5402).
   - [Helm chart updates related to Prometheus, Webhook HPA, and Flyteconsole probes](https://github.com/flyteorg/flyte/pull/5508).

3. **Documentation Improvements**
   - Added examples and guides, including [a guide for setting up the OpenAI batch agent backend](https://github.com/flyteorg/flyte/pull/5291) and [updated Optimization Performance docs](https://github.com/flyteorg/flyte/pull/5278).
   - [Fix doc link to testing agent on local cluster](https://github.com/flyteorg/flyte/pull/5398) and [replace Azure AD OIDC URL](https://github.com/flyteorg/flyte/pull/4075).
   - [Clarify networking configurations between data plane propeller and control plane data catalog in multi-cluster deployments](https://github.com/flyteorg/flyte/pull/5345).

4. **Performance & Bug Fixes**
   - [Fix cache and token management issues](https://github.com/flyteorg/flyte/pull/5388), including addressing auto-refresh cache race conditions and improving token refresh mechanisms.
   - [Fix SQL insert operations](https://github.com/flyteorg/flyte/pull/5482) to handle NULL values appropriately.
   - [Keep EnvFrom from pod template](https://github.com/flyteorg/flyte/pull/5423)
   - [Miscellaneous fixes](https://github.com/flyteorg/flyte/pull/5416) for pod templates, Kubeflow webhook errors, and logging configurations.
   
5. **Flytectl**
   - [Add prefetch functionality for paginator](https://github.com/flyteorg/flyte/pull/5310)
   - [Fix `upgrade` and `version` commands](https://github.com/flyteorg/flyte/pull/5470)


## 🔧 Housekeeping & Deprecations

1. **Miscellaneous**
   - [Add OTLP and sampling to otelutils](https://github.com/flyteorg/flyte/pull/5504)
   - [Refactor distributed job using common ReplicaSpec](https://github.com/flyteorg/flyte/pull/5355)
   - [Fix typos using codespell CI job](https://github.com/flyteorg/flyte/pull/5418).
   - [Remove obsolete Flyte config files](https://github.com/flyteorg/flyte/pull/5495).
   - [Upgraded various dependencies](https://github.com/flyteorg/flyte/pull/5313), including Golang, Docker, and others.
   - [Move to upstream mockery](https://github.com/flyteorg/flyte/pull/4937), which simplifies the migration to mockery-v2 
   - [Key-value execution tags](https://github.com/flyteorg/flyte/pull/5453)

## 🆕 New Contributors

- **@zychen5186** for [adding prefetch functionality for paginator](https://github.com/flyteorg/flyte/pull/5310).
- **@EraYaN** for [replacing Azure AD OIDC URL](https://github.com/flyteorg/flyte/pull/4075).
- **@Dlougach** for [making BaseURL insensitive to trailing slashes for metadata endpoint redirect](https://github.com/flyteorg/flyte/pull/5458).
- **@flixr** for [honoring redoc.enabled=false in charts](https://github.com/flyteorg/flyte/pull/5452).
- **@trevormcguire** for [including group in apiVersion in plugin_collector](https://github.com/flyteorg/flyte/pull/5457).
- **@va6996** for [inheriting execution cluster label from source execution](https://github.com/flyteorg/flyte/pull/5431).
- **@mhotan** for [Helm chart updates related to Prometheus, Webhook HPA, and Flyteconsole probes](https://github.com/flyteorg/flyte/pull/5508).
- **@eltociear** for [updating token_source.go](https://github.com/flyteorg/flyte/pull/5396).

**[Full Changelog](https://github.com/flyteorg/flyte/compare/v1.12.0...v1.13.0)**
