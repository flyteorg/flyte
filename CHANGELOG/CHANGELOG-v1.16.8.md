# Flyte 1.16.8

## Documentation
- Document each skip pattern and whitelisted word ([#7463](https://github.com/flyteorg/flyte/pull/7463))

## Changed
- Load datacatalog config into propeller even when datacatalog deployment is disabled ([#7468](https://github.com/flyteorg/flyte/pull/7468))
- Set priorityClassName for Spark tasks ([#7476](https://github.com/flyteorg/flyte/pull/7476))

## Fixed
- Don't panic when receiving status in k8s resolver ([#7478](https://github.com/flyteorg/flyte/pull/7478))
- Fix non-deterministic workflow digest and oversized error messages ([#7211](https://github.com/flyteorg/flyte/pull/7211))
- Apply pod template labels and annotations to Spark applications ([#7514](https://github.com/flyteorg/flyte/pull/7514))
- Allow disabling or scoping the self-registered MutatingWebhookConfiguration ([#7619](https://github.com/flyteorg/flyte/pull/7619))
- Update install script to paginate releases ([#7633](https://github.com/flyteorg/flyte/pull/7633))
- Fix multipart blob dimensionality for directory uploads ([#7621](https://github.com/flyteorg/flyte/pull/7621))
- Log links with showWhilePending: true are not shown while pending ([#7554](https://github.com/flyteorg/flyte/pull/7554))
- Populate ErrorKind/ErrorCode columns for accept-time execution failures ([#7496](https://github.com/flyteorg/flyte/pull/7496))

## Dependencies
- Bump golang.org/x/sync from 0.20.0 to 0.21.0 ([#7493](https://github.com/flyteorg/flyte/pull/7493))
- Use improved flytectl-setup-action ([#7632](https://github.com/flyteorg/flyte/pull/7632))
- Bump golang.org/x/sync from 0.21.0 to 0.22.0 ([#7649](https://github.com/flyteorg/flyte/pull/7649))

## Removed
- Remove deprecated finalizers ([#7537](https://github.com/flyteorg/flyte/pull/7537))
- Remove dead code in preparation for read replicas ([#7612](https://github.com/flyteorg/flyte/pull/7612))
- Remove unused repo methods ([#7637](https://github.com/flyteorg/flyte/pull/7637))
- Cleanup unused flyteadmin config ([#7644](https://github.com/flyteorg/flyte/pull/7644))

## Security
- Update golang.org/x/crypto to resolve CVEs ([#7638](https://github.com/flyteorg/flyte/pull/7638))

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.7...v1.16.8
