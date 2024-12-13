# Flyte v0.14.0

## Platform
- Update to the Flyte Compiler, to support nested branches and more. Better
  regression tests
- support for iam roles and k8s serviceaccounts
- customizable pod specs for pod tasks (add labels and annotations)
- copilot improvements
- Bug-fixes, visibility improvements
- support for non Cloud provider emailers - like Sendgrid
- performance improvement for dynamic workflows

## Flyteconsole
 - Bug fixes
 - More updates coming soon

## Flytekit
 - Support for nested conditionals
 - DoltTable and Dolt plugins and integration
 - Better context management (ready for more work)
 - Support for pre-built container plugins in flytekit. this makes it possible
   to library plugins and users do not need to build containers
 - More control plane class features
 - See [full release notes](https://github.com/flyteorg/flytekit/releases/tag/v0.19.0)
 Coming soon:
     - Great Expectations integration
     - More use case driven examples in flytesnacks

## flytectl
 - flytectl is ready for BETA. check it out - https://docs.flyte.org/en/latest/flytectl/overview.html

Please see the [flytekit release](https://github.com/flyteorg/flytekit/releases/tag/v0.18.0) for the full list and more details.
