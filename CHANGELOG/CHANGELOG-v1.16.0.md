# Flyte 1.16.0

# Flyte 1.16.0 Release Highlights

## Added

- Agent/Connector Functionality: Renamed agent to connector ([#6400](https://github.com/flyteorg/flyte/pull/6400)) with improved task retry support in Flyte Connectors ([#6486](https://github.com/flyteorg/flyte/pull/6486))
- Storage Configuration: Support for configuring storage from secrets to avoid exposing sensitive information ([#6419](https://github.com/flyteorg/flyte/pull/6419))
- Deployment Annotations: Configurable deployment annotations in Helm charts ([#6385](https://github.com/flyteorg/flyte/pull/6385))
- Pod Customization: Support for affinity & topologySpreadConstraints in webhook component ([#6431](https://github.com/flyteorg/flyte/pull/6431)), and configurable rollout strategy ([#6447](https://github.com/flyteorg/flyte/pull/6447))
- Labels and Annotations: Apply labels and annotations set via task decorator to pods and CR objects ([#6421](https://github.com/flyteorg/flyte/pull/6421))
- GRPC Configuration: MaxConcurrentStreams setting for flyteadmin grpc server ([#6448](https://github.com/flyteorg/flyte/pull/6448))
- ArrayNode Configuration: max-delta-timestamp configuration option for FlytePropeller ArrayNode ([#6453](https://github.com/flyteorg/flyte/pull/6453))
- Cloud Events: NATS support as cloud events sender ([#6507](https://github.com/flyteorg/flyte/pull/6507))
- Signal Handling: SignalWatcher functionality in Copilot ([#6501](https://github.com/flyteorg/flyte/pull/6501))
- Workflow Control: Workflow concurrency control feature ([#6475](https://github.com/flyteorg/flyte/pull/6475))
- Monitoring: Service monitor support for flyte scheduler ([#6558](https://github.com/flyteorg/flyte/pull/6558)) and datacatalog ([#6571](https://github.com/flyteorg/flyte/pull/6571))
- OpenTelemetry: OTEL configuration support in helm chart ([#6543](https://github.com/flyteorg/flyte/pull/6543))
- Scaling: HPA (Horizontal Pod Autoscaler) support for flyte admin ([#6615](https://github.com/flyteorg/flyte/pull/6615)) and datacatalog ([#6616](https://github.com/flyteorg/flyte/pull/6616))
- RBAC: Option to configure RBAC rules as namespace scoped ([#6534](https://github.com/flyteorg/flyte/pull/6534))
- Pre-init Containers: preInitContainers value support in Flyte-Binary charts ([#6559](https://github.com/flyteorg/flyte/pull/6559))
- Task Configuration: Task config support in flytectl config file ([#6538](https://github.com/flyteorg/flyte/pull/6538))
- Probes: Overrides for flyteadmin readiness and liveness probes ([#6484](https://github.com/flyteorg/flyte/pull/6484))
- Bug Reporting: Version details to bug report template ([#6479](https://github.com/flyteorg/flyte/pull/6479))
- Documentation: Kafka version field clarification in cloud events ([#6502](https://github.com/flyteorg/flyte/pull/6502))

## Changed

- Container Naming: Renamed container_image to image for improved UX ([#6211](https://github.com/flyteorg/flyte/pull/6211))
- Default Configuration: flyte-core now defaults propeller rawoutput-prefix to use storage.bucketName ([#6433](https://github.com/flyteorg/flyte/pull/6433))
- Log Context: Updated IDL with log context ([#6443](https://github.com/flyteorg/flyte/pull/6443))
- Error Handling: Treat kubelet NodeAffinity status.reason as retryable system error ([#6461](https://github.com/flyteorg/flyte/pull/6461))
- Code Refactoring: Multiple refactoring improvements including replacing HasPrefix+TrimPrefix with CutPrefix ([#6456](https://github.com/flyteorg/flyte/pull/6456)), making createMsgChan private ([#6467](https://github.com/flyteorg/flyte/pull/6467)), and adjusting CoPilot init and Pod status logic ([#6523](https://github.com/flyteorg/flyte/pull/6523))
- Go Version: Upgraded to Go 1.23 ([#6249](https://github.com/flyteorg/flyte/pull/6249))
- Repository Updates: Updated multiple repositories to use mockery ([#6477](https://github.com/flyteorg/flyte/pull/6477), [#6607](https://github.com/flyteorg/flyte/pull/6607), [#6608](https://github.com/flyteorg/flyte/pull/6608))
- CoPilot Improvements: Better folder handling ([#6620](https://github.com/flyteorg/flyte/pull/6620)) and multi-part blob support ([#6617](https://github.com/flyteorg/flyte/pull/6617))
- Service Account: Set service account in base ray pod spec, allow override by pod template ([#6514](https://github.com/flyteorg/flyte/pull/6514))
- Ray Job: Set name max length for rayJob plugin ([#6611](https://github.com/flyteorg/flyte/pull/6611))
- Container Logs: Allow using container name in kubeflow plugin log links ([#6524](https://github.com/flyteorg/flyte/pull/6524))
  Fixed
- GitHub Actions: Updated and fixed "stale" GitHub Action ([#6449](https://github.com/flyteorg/flyte/pull/6449)) and PR stale comments ([#6458](https://github.com/flyteorg/flyte/pull/6458))
- NodeShutdown: Updated DemystifyFailure to respect NodeShutdown ([#6452](https://github.com/flyteorg/flyte/pull/6452))
- Connector Discovery: Resolved conflicts between agent auto discovery and explicit task type mapping ([#6464](https://github.com/flyteorg/flyte/pull/6464))
- Release Scripts: Fixed release script for beta versions ([#6489](https://github.com/flyteorg/flyte/pull/6489)) and corrected file paths in release workflows ([#6487](https://github.com/flyteorg/flyte/pull/6487))
- Resource Management: Fixed PodTemplate resource state pollution between workflow runs ([#6530](https://github.com/flyteorg/flyte/pull/6530)) and compile-time podTemplate resources override issues ([#6483](https://github.com/flyteorg/flyte/pull/6483))
- Workflow Processing: Fixed workflow equality check ([#6521](https://github.com/flyteorg/flyte/pull/6521)) and enqueue correct work item in node execution context ([#6526](https://github.com/flyteorg/flyte/pull/6526))
- Grafana Dashboards: Fixed Grafana dashboard queries for dynamic workflow metrics ([#6546](https://github.com/flyteorg/flyte/pull/6546))
- Scheduler: Fixed invalid cron date schedule creating infinite loop in flytescheduler ([#6555](https://github.com/flyteorg/flyte/pull/6555))
- Plugin Safety: Made plugin metric registration thread safe ([#6532](https://github.com/flyteorg/flyte/pull/6532))
- Server Shutdown: Fixed flyteadmin not shutting down servers gracefully ([#6289](https://github.com/flyteorg/flyte/pull/6289))
- Strict Mode: Disabled strict mode in flytectl ([#6619](https://github.com/flyteorg/flyte/pull/6619))

## Removed

- CI/CD: Removed docs job from GitHub actions workflow ([#6432](https://github.com/flyteorg/flyte/pull/6432))
- Helm Cleanup: Cleaned up helm template since metrics path is not configurable ([#6444](https://github.com/flyteorg/flyte/pull/6444))
- Legacy Plugins: Deleted hive plugin ([#6580](https://github.com/flyteorg/flyte/pull/6580)) and cleaned up flyteplugins ([#6515](https://github.com/flyteorg/flyte/pull/6515))
- Annotations: Removed projectcontour annotations from helm charts by default ([#6564](https://github.com/flyteorg/flyte/pull/6564)) and primary pod annotation from raw container task ([#6618](https://github.com/flyteorg/flyte/pull/6618))
- JWT: Removed usage of golang-jwt v3 ([#6596](https://github.com/flyteorg/flyte/pull/6596))

## Security

- Permissions: Removed SYS_PTRACE and sharenamespace settings for improved security ([#6509](https://github.com/flyteorg/flyte/pull/6509))

## Dependencies

- Dependency Updates: Multiple security and maintenance updates including:
  golang.org/x/net updates ([#6338](https://github.com/flyteorg/flyte/pull/6338), [#6418](https://github.com/flyteorg/flyte/pull/6418), [#6587](https://github.com/flyteorg/flyte/pull/6587), [#6588](https://github.com/flyteorg/flyte/pull/6588), [#6589](https://github.com/flyteorg/flyte/pull/6589), [#6591](https://github.com/flyteorg/flyte/pull/6591), [#6593](https://github.com/flyteorg/flyte/pull/6593))
  golang.org/x/oauth2 updates ([#6535](https://github.com/flyteorg/flyte/pull/6535), [#6594](https://github.com/flyteorg/flyte/pull/6594), [#6595](https://github.com/flyteorg/flyte/pull/6595), [#6597](https://github.com/flyteorg/flyte/pull/6597), [#6598](https://github.com/flyteorg/flyte/pull/6598), [#6599](https://github.com/flyteorg/flyte/pull/6599), [#6601](https://github.com/flyteorg/flyte/pull/6601))
- mapstructure updates ([#6512](https://github.com/flyteorg/flyte/pull/6512), [#6575](https://github.com/flyteorg/flyte/pull/6575))
- Docker and crypto updates ([#6573](https://github.com/flyteorg/flyte/pull/6573), [#6414](https://github.com/flyteorg/flyte/pull/6414))

## Contributors

Special thanks to new contributors: @ppeerttu ([#6385](https://github.com/flyteorg/flyte/pull/6385)), @studystill ([#6456](https://github.com/flyteorg/flyte/pull/6456)), @daadc ([#6502](https://github.com/flyteorg/flyte/pull/6502)), @mattiadevivo ([#6507](https://github.com/flyteorg/flyte/pull/6507)), @gopherorg ([#6511](https://github.com/flyteorg/flyte/pull/6511)), @jingchanglu ([#6528](https://github.com/flyteorg/flyte/pull/6528)), @diranged ([#6534](https://github.com/flyteorg/flyte/pull/6534)), @thomasjhuang ([#6475](https://github.com/flyteorg/flyte/pull/6475)), and @hylje ([#6559](https://github.com/flyteorg/flyte/pull/6559)), along with all returning contributors who made this release possible.

Full Changelog: https://github.com/flyteorg/flyte/compare/v1.15.3...v1.16.0
