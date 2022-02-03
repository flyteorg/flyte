.. _community_roadmap:

###############
Roadmap
###############

How the community works?
=========================
Flyte is used actively in production at multiple companies. We pride ourselves on being extremely customer-focused and care deeply about a high quality customer experience. Thus, we always prioritize stability, reliability, observability, and maintainability over raw feature development. 

The features are developed usually in response to specific use cases and user scenarios. That being said, we are proactively thinking about the evolution of the system and how we want to keep adapting to the changing requirements. Thus, most of our changes reflect future development scenarios, and in cases where we feel rapid prototyping would enable us to discover potential pitfalls or uncover hidden use cases, we would proactively develop features, behind feature flags.

Thus, it is extremely important to let the community know about your use cases and how we could adapt parts of Flyte to meet those requirements. We welcome collaboration and contributions, but please follow our Contribution Guidelines.


Milestones and Releases
========================
Flyte consists of many components and services. In true Agile fashion, each service is independently iterated and coordinated by maintaining backwards-compatible contracts using protobuf defined in `FlyteIDL <https://flyte.readthedocs.io/projects/flyteidl/en/latest/>`__. Thus components like lytekit, Flyte Propeller, and Datacatalog are independently versioned.

We have decided to release a new version of the overall platform in the `Flyte repo <https://github.com/flyteorg/flyte>`_ every month. Thus we create one milestone for end of every month which points to a new release of
Flyte. This may change in the future, but to match our velocity of development this is our preferred option. Every release will be associated with a CHANGELOG (in the repo).


Change management
------------------
To ensure that changes are trackable and the history is explainable, we use a slightly cumbersome but helpful process as follows. Some of these are our immediate goals:

- Every PR is associated with an issue (automatic searchable documentation).
- Large PR’s are associated with Proposals.
- Every major change is associated with documentation.
- Owner files are present for all repositories.

Release Train
--------------
- We tag issues with milestones, every new issue will be associated with the next milestone. If the issue is not completed by the milestone, or the contributor feels it may slip the deadline, they should manually move it to the next milestone. Every issue not removed, will be moved to the next milestone.
- Every new issue has a “untriaged” label associated with it. If we remove this label we should add an assignee. If a contributor is working on the issue, please remove this label.
- Release indicates a release for overall Flyte - marked mostly by a milestone.
- Flyte release are monthly.
- We may have patch releases eg. 0.1.x in between the monthly releases.

Upcoming Features & Issues
==========================

Issues by Theme
----------------

+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Theme       | Description                                                    | Open Issues                                                                           | Comment                                                                                                     |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Bugs        | Currently known and open bugs.                                 | `Bugs <https://github.com/flyteorg/flyte/labels/bug>`_                                | We are always working on bugs. Open a new one `here <https://github.com/flyteorg/flyte/issues/new/choose>`_.|
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Security    | Issues related to security enhancements.                       | `Security issues <https://github.com/flyteorg/flyte/labels/security>`_.               |                                                                                                             |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Docs        | All the issues open with our documentation.                    | `Docs issues <https://github.com/flyteorg/flyte/labels/documentation>`_               | Feb 2021, we are completely overhauling our docs. Your feedback is appreciated!                             |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Features    | All the new features in development.                           | `Features issues <https://github.com/flyteorg/flyte/labels/enhancement>`_             |                                                                                                             |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| plugins     | New capabilities, plugins that are being built into Flyte.     | `Plugins issues <https://github.com/flyteorg/flyte/labels/plugins>`_                  | This is one of the best place to get started contributing to Flyte. Issues with both                        |
|             | These could be hosted services, K8s native execution etc.      |                                                                                       | labels `plugins` and `Flytekit` refer to purely client side plugins and are the fastest to contribute.      |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| scale       | These issues deal with performance,  reliability and           | `Scale issues <https://github.com/flyteorg/flyte/labels/scale>`_                      | We are always working on these issues and we would love to hear feedback about what you                     |
|             | scalability of Flyte.                                          |                                                                                       | would want to change or what should we prioritize.                                                          |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+
| Contribute  | If you are looking to contribute and want a great first issue, | `Contribute issues <https://github.com/flyteorg/flyte/labels/good%20first%20issue>`_  | These are the best issues to get started with.                                                              |
|             | look at these issues.                                          |                                                                                       |                                                                                                             |
+-------------+----------------------------------------------------------------+---------------------------------------------------------------------------------------+-------------------------------------------------------------------------------------------------------------+


Issues by Components
---------------------

+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+--------------------------------------------+
| Theme        | Description                                   | Open Issues                                                                 | Comment                                    |
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+--------------------------------------------+
| flyteconsole | Issues on FlyteConsole (Flytes UI).           | `flyteconsole issues <https://github.com/flyteorg/flyte/labels/ui>`_        | These are great issues to get started with.|
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+--------------------------------------------+
| flytectl     | Issues on FlyteCTL (standalone CLI for Flyte) | `FlyteCTL issues <https://github.com/flyteorg/flyte/labels/flytectl>`_      | Great issues to start with.                |
+--------------+-----------------------------------------------+-----------------------------------------------------------------------------+--------------------------------------------+