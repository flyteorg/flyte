# Flyte v1.13.0-rc0 Release Notes

## Major Features and Improvements

- **Key-value execution tags**: Enhanced execution tagging capabilities. (#5453)
- **Watch agent metadata service**: Improved monitoring of agent metadata. (#5017)
- **Distributed job refactoring**: Refactored distributed job using common ReplicaSpec. (#5355)
- **Execution cluster label inheritance**: Executions now inherit cluster labels from source executions. (#5431)
- **Domain API**: Added new API to retrieve domain information. (#5443)
- **Execution environment versioning**: Added version to ExecutionEnv proto message. (#5506)
- **OTLP and sampling in otelutils**: Enhanced observability with OpenTelemetry Protocol (OTLP) support. (#5504)

## Notable Changes

- Removed mmcloud plugin. (#5468)
- Updated k3s version to 1.29.0. (#5475)
- Improved auth flow to support custom base URLs in deployments. (#5192)
- Added flyteconsole URL to FlyteWorkflow CRD. (#5449)
- Enhanced 'flytectl compile' to consider launch plans within workflows. (#5463)
- Introduced control over task phases for log link display. (#4726)
- Improved Helm chart configurations for Prometheus, Webhook HPA, and Flyteconsole probes. (#5508)

## Bug Fixes and Optimizations

- Fixed flaky auto_refresh_test. (#5438)
- Corrected NULL to empty string in SQL insert for migrations. (#5482)
- Resolved issues with broken mermaid diagrams in documentation. (#5498)
- Fixed Ray plugin to use default service account if not set in task metadata. (#5499)

## Documentation and Usability

- Updated community page. (#5496)
- Improved documentation on logging link lifetime configuration. (#5503)
- Replaced 'uctl' with 'flytectl' in OSS docs for consistency. (#5501)

## New Contributors

- @Dlougach, @flixr, @trevormcguire, @va6996, @mhotan, and @eltociear made their first contributions to the project.

For a complete list of changes, please refer to the [full changelog](https://github.com/flyteorg/flyte/compare/flytectl/v0.8.24...v1.13.0-rc0).
