# Flyte v0.11.0

## Flyte Platform
* New to flyte? https://start.flyte.org takes you through first run experience. (Thanks to @jeevb)
* [Grafana templates](https://docs.flyte.org/en/latest/howto/monitoring/index.html) for monitoring Flyte System and User Workflows.
* [Extend Flyte](https://docs.flyte.org/en/latest/plugins/index.html) docs.
* [FlyteIdl Docs](https://docs.flyte.org/en/latest/reference_flyteidl.html) are published! You can learn about the core language that makes it all work.
* [Additional knob](https://github.com/flyteorg/flytepropeller/pull/219/files#diff-91657d6448dfbf87f4cecf126ad02bd668ea233edcf74e860ef4f54bdd4cb552R78) for fine tuning flyte propeller performance that speeds up executions drastically.
* OidC support for Google Idp (And other OidC compliant Idps)
* Various stabilization bugs.

## Flytekit
Since v0.16.0a2, the last flytekit milestone release, all effort has been towards stabilizing the new API. Please see the individual [releases](https://github.com/flyteorg/flytekit/releases) for detailed information. The highlights are

* Serialization/registration processes have been firmed up and utilities to ease that process introduced (not having to build a container to serialize for instance).
* Plugins structure revamped (eventually we'll move to a separate new repo entirely)
* User-facing imports have been organized into three top-level subpackages (`flytekit`, `flytekit.testing`, and `flytekit.extend`)
* Retries added to read-only Admin calls in client
* Lots of cleanup and additions to the [cookbook](https://flytecookbook.readthedocs.io/en/latest/) and documentation generally.
* Bug fixes.

