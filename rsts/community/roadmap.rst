.. _community_roadmap:

###############
Roadmap
###############

How the Community Works
=======================
Flyte is actively used in production at multiple companies. We pride ourselves on being extremely customer-focused, and care about providing a high quality customer experience. We therefore always
prioritize stability, reliability, observability and maintainability over raw feature development. 

Features are usually developed in response to specific use cases and user scenarios. That being said, we are proactively thinking about the evolution of the system and how we want to keep adapting to changing requirements. Thus most of our changes reflect future development scenarios, and in
cases where we feel rapid prototyping would enable us to discover potential pitfalls or uncover hidden use cases, we would proactively develop features behind feature flags.

It is extremely important to let the community know about your use cases, so that we adapt parts of Flyte to meet those requirements. We welcome collaboration and contributions, but please follow our `Contribution Guidelines <https://docs.flyte.org/en/latest/community/contribute.html>`_. The quarterly planning meeting is also hosted publicly, please see more below.


Milestones and Release Processes
================================
Flyte consists of many components and services. Each service is independently iterated and coordinated by maintaining backwards compatible contracts using Protobuf messages defined in `FlyteIDL <https://flyte.readthedocs.io/projects/flyteidl/en/latest/>`__.

Release Cadence

Versioning Scheme

At each quarterly release, major components of Flyte and the Flyte repository itself will be released with an incremented minor version number and the version number will be aligned across those components. The major version number will remain `1` for the foreseeable future. That is, if the current version of Flyte is `1.2.x`, the next quarter's release will be `1.3.0` for Flyte and the major components.

After each version is released, merges to master will be assigned beta releases of the next quarter's release. That is, if `flytepropeller` version `v1.2.0` was just released, the next merge to master will be tagged `v1.3.0b0`.

Not strictly forcing a time-constraint on the Flyte release cycle means that if a substantial number of changes is merged, perhaps due to a security issue or just a rapid pace of feature development, we can always bring up the timeline of the release.


Change Management
------------------
To ensure that changes are trackable and the history is explainable, we use a slightly cumbersome but helpful process, with the following immediate goals:
- Every PR is associated with an issue (automatic searchable documentation)
- Large PRs are associated with Proposals
- Every major change is associated with documentation
- Owner files exist for all repositories

Issue Lifecycle
---------------
- Incoming issues are tagged automatically as untriaged.
- Periodically, members of the Flyte community will meet to triage incoming issues. We aim to do this on a weekly basis.
- During this meeting we'll attempt to assign each issue to a milestone. Some issues however will need to be investigated before we can fully assess.
- Once an issue is assigned to a milestone, this means we are committed to delivering it that release. This means the burden for adding something to the milestone is relatively high. Issues that slip should only slip for good reason.



Upcoming Features and Issues
============================

Issues by Theme
----------------

+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Theme       | Description                                                    | Open Issues                                                                           | Comment                                                                                                      |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Bugs        | Currently known and open bugs.                                 | `Bugs <https://github.com/flyteorg/flyte/labels/bug>`_                                | We are always working on bugs. Open a new one `here <https://github.com/flyteorg/flyte/issues/new/choose>`_. |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Security    | Issues related to security enhancements.                       | `Security issues <https://github.com/flyteorg/flyte/labels/security>`_                |                                                                                                              |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Docs        | All issues open with our documentation                         | `Docs issues <https://github.com/flyteorg/flyte/labels/documentation>`_               | Starting Feb 2021, we will be completely overhauling our docs. Feedback appreciated!                         |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Features    | All new features in development                                | `Features issues <https://github.com/flyteorg/flyte/labels/enhancement>`_             |                                                                                                              |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Plugins     | New capabilities and plugins that are built into Flyte.        | `Plugins issues <https://github.com/flyteorg/flyte/labels/plugins>`_                  | This is one of the best places to get started contributing to Flyte. Issues with both                        |
|             | These could be hosted services, K8s native execution, etc.     |                                                                                       | `plugins` and `flytekit` labels refer to purely client-side plugins and are the fastest to contribute to.    |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Scale       | These issues deal with performance,  reliability and           | `Scale issues <https://github.com/flyteorg/flyte/labels/scale>`_                      | We are always working on these issues and we would love to hear feedback about what you                      |
|             | scalability of Flyte                                           |                                                                                       | would want to change or what we should prioritize.                                                           |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+
| Contribute  | If you are looking to contribute and want a great first issue, | `Contribute issues <https://github.com/flyteorg/flyte/labels/good%20first%20issue>`_  | These are the best issues to get started with.                                                               |
|             | check out these issues                                         |                                                                                       |                                                                                                              |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------+


Issues by Components
---------------------

+---------------+---------------------------------------+------------------------------------------------------------------------+
| Theme         | Description                           | Open Issues                                                            |
+===============+=======================================+========================================================================+
| Flyte Console | Issues concerning our web UI.         | `Flyte Console issues <https://github.com/flyteorg/flyte/labels/ui>`_  |
+---------------+---------------------------------------+------------------------------------------------------------------------+
| Flytectl      | Issues concerning our standalone CLI. | `Flytectl issues <https://github.com/flyteorg/flyte/labels/flytectl>`_ |
+---------------+---------------------------------------+------------------------------------------------------------------------+

For an overview of what we're currently working on, check out our `live roadmap <https://github.com/orgs/flyteorg/projects/3>`__.

