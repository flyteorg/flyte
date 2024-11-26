# Flyte 1.12.0 Release Notes

Flyte 1.12.0 brings a host of new features, optimizations, and fixes, enhancing the platform's functionality and user experience. This release also welcomes several new contributors to the Flyte community. Below are the highlights of this release.

## 🚀 New Features & Improvements

1. **Admin & Core Enhancements**  
   - [Implemented the `GetProject` endpoint](https://github.com/flyteorg/flyte/pull/4825) in FlyteAdmin.
   - [Added tracking for active node and task execution counts](https://github.com/flyteorg/flyte/pull/4986) in Propeller.
   - [Set FlyteAdmin gRPC port correctly in config](https://github.com/flyteorg/flyte/pull/5013) and included [separate gRPC Ingress flag](https://github.com/flyteorg/flyte/pull/4946).

2. **Helm Charts & Manifests**  
   - [CI workflow enhancements](https://github.com/flyteorg/flyte/pull/5027) for Helm charts and manifests.
   - [Fixed syntax issues](https://github.com/flyteorg/flyte/pull/5056) and [rendering errors](https://github.com/flyteorg/flyte/pull/5048) in Helm charts.
   - [Updated Spark-on-K8s-operator address in Helm charts](https://github.com/flyteorg/flyte/pull/5198).

3. **Documentation Improvements**  
   - Enhanced documentation across various modules, including [a guide for running the newest Flyteconsole in Flyte sandbox](https://github.com/flyteorg/flyte/pull/5100) and [a troubleshooting guide for Docker errors](https://github.com/flyteorg/flyte/pull/4972).
   - [Updated Propeller architecture documentation](https://github.com/flyteorg/flyte/pull/5117) and [added a guide for enabling/disabling local caching](https://github.com/flyteorg/flyte/pull/5242).

4. **Performance & Bug Fixes**  
   - [Improved lint error detection](https://github.com/flyteorg/flyte/pull/5072), fixing multiple issues across components.
   - [Optimized Flyte components](https://github.com/flyteorg/flyte/pull/5097), including Golang, Protobuf, and GRPC versions.
   - Fixed several bugs related to execution phases, [GPU resource overrides](https://github.com/flyteorg/flyte/pull/4925), and [Databricks errors](https://github.com/flyteorg/flyte/pull/5226).

## 🔧 Housekeeping & Deprecations

1. **Deprecated Configuration**  
   - [Deprecated `MaxDatasetSizeBytes` propeller config](https://github.com/flyteorg/flyte/pull/4852) in favor of `GetLimitMegabytes` storage config.
   - [Removed obsolete Flyte config files](https://github.com/flyteorg/flyte/pull/5196).

2. **Miscellaneous**  
   - [Simplified boilerplate](https://github.com/flyteorg/flyte/pull/5134).
   - [Regenerated Ray pflags](https://github.com/flyteorg/flyte/pull/5149).
   - Upgraded various dependencies, including [cloudevents](https://github.com/flyteorg/flyte/pull/5142), [logrus](https://github.com/flyteorg/flyte/pull/5139), and [go-restful](https://github.com/flyteorg/flyte/pull/5140).

## 🆕 New Contributors

- **@RRap0so** for [implementing the `GetProject` endpoint](https://github.com/flyteorg/flyte/pull/4825).
- **@cjidboon94** for [adding the GKE starter values file](https://github.com/flyteorg/flyte/pull/5026).
- **@ddl-rliu** for [improving audience mismatch debugging](https://github.com/flyteorg/flyte/pull/5078).
- **@pbrogan12** for [fixing Helm chart rendering](https://github.com/flyteorg/flyte/pull/5048).
- **@austin362667** for [showing diff structure when re-registering tasks](https://github.com/flyteorg/flyte/pull/4924).
- **@ssen85** for [updating environment setup documentation](https://github.com/flyteorg/flyte/pull/4963).
- **@dansola** for [changing retry error types](https://github.com/flyteorg/flyte/pull/5128).
- **@noahjax** for [adding identity to task execution metadata](https://github.com/flyteorg/flyte/pull/5105).
- **@ongkong** for [fixing ID bigint conversion issues](https://github.com/flyteorg/flyte/pull/5157).
- **@sshardool** for [tracking active node and task execution counts](https://github.com/flyteorg/flyte/pull/4986).
- **@Jeinhaus** for [adding a missing key in the auth guide](https://github.com/flyteorg/flyte/pull/5169).
- **@yini7777** for [fixing secret mounting issues](https://github.com/flyteorg/flyte/pull/5063).
- **@Sovietaced** for [fixing grammatical errors in the documentation](https://github.com/flyteorg/flyte/pull/5227).
- **@agiron123** for [configuring the RunLLM widget](https://github.com/flyteorg/flyte/pull/5266).
- **@mark-thm** for [removing the upper bound on FlyteIDL's protobuf dependency](https://github.com/flyteorg/flyte/pull/5285).

**[Full Changelog](https://github.com/flyteorg/flyte/compare/v1.11.0...v1.12.0)**

