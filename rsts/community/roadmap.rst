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

It is extremely important to let the community know about your use cases, so that we adapt parts of Flyte to meet those requirements. We welcome collaboration and contributions, but please follow our `Contribution Guidelines <https://docs.flyte.org/en/latest/community/contribute.html>`_. 


Milestones and Releases
========================

Flyte consists of many components and services. In true Agile fashion, each service is independently iterated and coordinated by maintaing backwards compatible contracts using protobuf defined in `FlyteIDL <https://flyte.readthedocs.io/projects/flyteidl/en/latest/>`__. Therefore, components like FlyteKit, FlytePropeller, and datacatalog are independently versioned.

We have decided to release a new version of the overall platform in the `Flyte repo <https://github.com/flyteorg/flyte>`_ every month. We create one milestone at the end of every month, which points to a new release of
Flyte. This may change in the future, but to match the rate of our development, this is our preferred option. Every release will be associated with a CHANGELOG (in the repo).

Change Management
------------------
To ensure that changes are trackable and the history is explainable, we use a slightly cumbersome but helpful process, with the following immediate goals:
- Every PR is associated with an issue (automatic searchable documentation)
- Large PRs are associated with Proposals
- Every major change is associated with documentation
- Owner files exist for all repositories

Release Train
--------------
- We tag issues with milestones, where every new issue will be associated with the next milestone. If an issue is not complete by the upcoming milestone, or the contributor expects that it may miss the deadline, it should be manually moved to the following milestone. Every issue not removed will be moved to the following milestone.
- Every new issue has a “untriaged” label associated with it, and if removed, an assignee must be added. If a contributor is working on the issue, please remove this "untriaged" label.
- "Release" indicates a release for Flyte overall, usually marked by a milestone.
- Flyte release are monthly.
- We may have patch releases eg. 0.1.x in between the monthly releases.


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

+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+---------------------------------------------+
| Theme        | Description                                   | Open Issues                                                                 | Comment                                     |
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+---------------------------------------------+
| flyteconsole | Issues on FlyteConsole (Flyte's UI)           | `flyteconsole issues <https://github.com/flyteorg/flyte/labels/ui>`_        | These are great issues to get started with. |
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+---------------------------------------------+
| flytectl     | Issues on Flytectl (standalone CLI for Flyte) | `flytectl issues <https://github.com/flyteorg/flyte/labels/flytectl>`_      | Great issues to start with.                 |
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+---------------------------------------------+

For an overview of what we're currently working on, check out our `live roadmap <https://github.com/orgs/flyteorg/projects/3>`__.

