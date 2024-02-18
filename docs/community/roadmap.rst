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
Flyte consists of many components and services. Each service is independently iterated and coordinated by maintaining backwards compatible contracts using Protobuf messages defined in `FlyteIDL <https://docs.flyte.org/en/latest/reference_flyteidl.html>`__.

Release Cadence
---------------
We aim to release Flyte quarterly, with the understanding that rather than being tied strictly to the calendar, we aim to have substantial features, improvements, and bug fixes at each quarter. If features slated for a given release are delayed, then the release will be delayed as well. The increased time will also give the Flyte development team more time to beta test each feature and release.

Versioning Scheme
-----------------
*Please keep in mind the CI work to implement this scheme is still in progress*

At each quarterly release, major components of Flyte and the Flyte repository itself will be released with an incremented minor version number and the version number will be aligned across those components. The major version number will remain ``1`` for the foreseeable future. That is, if the current version of Flyte is ``1.2.x``, the next release will be ``1.3.0`` for Flyte and the major components.

After each version is released, merges to master will be assigned beta releases of the next release version. That is, if ``flytepropeller`` version ``v1.2.0`` was just released, the next merge to master will be tagged ``v1.3.0b0``.

Not strictly forcing a time-constraint on the Flyte release cycle means that if a substantial number of changes is merged, perhaps due to a security issue or just a rapid pace of feature development, we can always bring up the timeline of the release.

Components with versions aligned
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
* Propeller
* Admin
* Console
* datacatalog
* flytectl
* flytesnacks
* Flytekit
* flytekit-java

The last two we are going to tie together for now, but realize that we may want to unpin in the future.

Components versioned independently
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
* flyteidl
* flytestdlib
* flyteplugins
* flytecopilot

Helm Charts
^^^^^^^^^^^
Helm charts deserve a special mention here. Unlike the other components which will have patch versions that differ, the Flyte release version and the Helm chart version will always be identical down to the patch. That is, a Flyte release is a Helm release and vice-versa.

Release Branches and Patching
-----------------------------
After each minor release, a release branch will be created. There will be no alignment of patch versions across the components. That is, by the end of the ``1.3.x`` release cycle, ``flyteadmin`` may be on ``1.3.8`` and ``flytepropeller`` may be on ``1.3.2``.

When developing bug fixes, by default we will continue to develop off of master, which will not be the stable branch. After such bug fixes are merged, it will be the responsibility of the developer to ensure that the patches are also applied to prior releases. At the current time, we propose only supporting one release back (two for security patches). That is, if ``flytepropeller`` has a bug fix that results in ``v1.3.0b0`` that patch will be applied to the ``v1.2.x`` release, but not the ``v1.1.x`` release.

Beta Patch Releases
^^^^^^^^^^^^^^^^^^^
We also propose that beta patch versions be merged into the release branch when patching prior releases. For example, assuming no patches have yet to be made to the ``v1.2.0`` release, when porting a bug fix that resulted in ``v1.3.0b0`` onto the ``release-v1.2`` branch, the developer can first release ``v1.2.1b0`` for testing into ``release-v1.2`` before releasing the ``v1.2.1`` release. Such beta releases should be made at the discretion of the developer.

Whether or not a patch version of any of the Flyte components also creates a Flyte patch release shall also be left to the discretion of the developer.

Documentation Versioning
------------------------
We also currently have an issue with our documentation versioning. While our readthedocs page does have versioning enabled and we publish the [docs version](https://github.com/flyteorg/flyte/blob/80c098f10334b1c916d1e4274ab9f204152d9d80/rsts/conf.py#L33), all the [intersphinx mappings](https://github.com/flyteorg/flyte/blob/80c098f10334b1c916d1e4274ab9f204152d9d80/rsts/conf.py#L219) just point to `latest`. Keep in mind that this mapping not only exists in this `flyte` repo, but also in all the other repos that that mapping points to. That is, to maintain an accurate mapping of different versions of documentation, we'll need to update the mapping in all the repos.

To remediate this, we propose the following:

* Documentation should be pinned only to Major.Minor on all the repos that have their versions "aligned".

    * This means that as we release patch versions of Admin, Propeller, etc., if we're on v1.1 for instance, as Admin code/auto-generated documentation changes, the v1.1 listing of readthedocs will automatically pick it up.
* Repos that are not aligned will just default to the "latest" documentation version.

Planning Process
================

Quarterly Planning
------------------
Members of the community should feel free to join these! Core members of the Flyte team will come prepared with general initiatives in mind. We will use these meetings to prioritize these ideas, assess community interest and impact, and decide what goes into the GitHub milestone for the next release. Members of the community looking to contribute should also join. Please look for this meeting invite on the calendar - it may not be set up as a recurring meeting simply because it will likely change by a few days each quarter.

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

Browse Features and Issues
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

