# Management of community-contributed plugins

**Authors:**

- @davidmirror-ops


## 1 Executive Summary

The Flyte community as a self-governed and productive collective of individuals, values contributions. This proposal aims to discuss the process community contributors should follow to submit a new `flytekit` plugin, with special attention to mechanisms that ensure stability and maintainability of core flyte code.

## 2 Motivation

- With the current "in-tree" approach, plugins developed by the community land in the `flytekit` repo ([example](https://github.com/flyteorg/flytekit/pull/2537)). It results in Flyte maintainers having to take care of CI test failures due to plugin code or flytekit updates incompatible with plugin code, etc. Flyte maintainers are also expected to provide support about and fix bugs in plugins integrating 3rd party libraries that they might have little knowledge off.

- The goal is to agree on a process for contributors to follow when submitting new integrations in a "out-of-tree" way that clearly communicates that it is a community-contributed -and then- community-supported integration.

## 3 Proposed Implementation

- Create a `community` folder under `flytekit/plugins` and keep releasing the plugins in that folder as separate `pypi` packages.
- Configure CI to only run tests on `plugins/community` when there are changes to a respective plugin.
- Keep releasing community plugins alongside flytekit, even if there are no changes.
- Explicitly mark plugins as community maintained in the import via `import flytekitplugins.contrib.x`
- Plugin authors are responsible for maintaining their plugins. In case there are PRs to change a community plugin, the plugin maintainers review the PR and give a non-binding approval. Once a community plugin maintainer has given a non-binding approval, a `flytekit` maintainer has to give a binding approval in order for the PR to be merged.

This proposal includes agent plugins.
### Promotion process to official plugin

An official plugin is one that is maintained by the core Flyte team and is made part of the official `flytekit` documentation.

- Plugin maintainers or community members can propose the promotion of a plugin to official by creating an Issue on the `flytekit` repo.
- The supermajority of the TSC must approve publicly before promoting a plugin.

To consider it for promotion, a plugin must meet the following criteria:

- Production readiness testing performed by the core Flyte team or documented by plugin users or maintainers
- Evidence of ongoing usage through Github issues or Slack threads
- Documented in flytekit's documentation



## 4 Drawbacks

- Potential overhead: CI configuration changes in flytekit (probably a one-time change) 

## 5 Alternatives

- Maintain community plugins on a separate repo
    - Against the monorepo initiative
-  Have community packages be it's own org
    - Significantly higher management overhead
- `flytekit` plugins built into their own package
    -   Potentially heavier development process

- Adding plugin authors as CODEOWNERS won't be considered due to a [Github permission model](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners) limitation:

>The people you choose as code owners must have write permissions for the repository. 

Getting write permissions in `flytekit` via contributing plugins is not part of the [current Governance model](https://github.com/flyteorg/community/blob/main/GOVERNANCE.md#community-roles-and-path-to-maintainership) for flyte.
