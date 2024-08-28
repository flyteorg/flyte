# Management of community-contributed plugins

**Authors:**

- @davidmirror-ops


## 1 Executive Summary

The Flyte community as a self-governed and productive collective of individuals, values contributions. This proposal aims to discuss the process community contributors should follow to submit a new `flytekit` plugin, with special attention to mechanisms that ensure stability and maintainability of core flyte code.

## 2 Motivation

- With the current "in-tree" approach, plugins developed by the community land on the `flytekit` repo ([example](https://github.com/flyteorg/flytekit/pull/2537)). It results in Flyte maintainers having to take care of CI test failures due to plugin code or flytekit updates incompatible with plugin code, etc.

- The goal is to agree on a process for contributors to follow when submitting new integrations in a "out-of-tree" way that clearly communicates that it is a community-contributed -and then- community-supported integration.

## 3 Proposed Implementation

- Create a `community` folder under `flytekit/plugins` and keep releasing the plugins in that folder as separate `pypi` packages.
- Configure CI to only run tests on `plugins/community` when there are changes.
- Keep releasing plugins alongside flytekit.
- Update [CODEOWNERS](https://github.com/flyteorg/flytekit/blob/master/CODEOWNERS) for the `plugins/community` folder, enabling contributors to merge community plugins (it requires granting Write permissions to contributors on `flytekit`).
- Explicitly mark them as community maintained in the import via `import flytekitplugins.contrib.x`

### Promotion process to official plugin

An official plugin is one that is maintained by the core Flyte team and is made part of the official `flytekit` documentation.

- Plugins maintainers or community members can propose the promotion of a plugin to official by creating an Issue on the `flytekit` repo.
- The supermajority of the TSC must approve publicly before promoting a plugin.

To consider it for promotion, a plugin must meet the following criteria:

- Production readiness testing performed by the core Flyte team or documented by plugin users or maintainers
- Evidence of ongoing usage through Github issues or Slack threads



## 4 Drawbacks

- There's a management overhead associated with this initiative:
    - For every new plugin contributor, we may need to update `CODEOWNERS`
    - CI configuration changes in flytekit (probably a one-time change) 

## 5 Alternatives

- Maintain community plugins on a separate repo
    - Against the monorepo initiative
-  Have community packages be it's own org
    - Significantly higher management overhead
- `flytekit` plugins built into their own package
    -   Potentially heavier development process

## 6 Open questions

- Does this apply to Agents?


