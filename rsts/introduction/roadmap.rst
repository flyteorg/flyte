.. _introduction-roadmap:

###############
Roadmap
###############

How the core team works
========================
Flyte is used actively in production at Lyft and its various subsidiaries. The core team is very customer focused and cares deeply about a high quality customer experience. Thus we always
prioritize stability, reliability, observability and maintainability over raw feature development. Features are developed usually in response to specific use cases and user scenarios. That being said,
we are proactively thinking about the evolution of the system and how we want to keep adapting to changing requirements. Thus most of our changes reflect future development scenarios and in
cases where we feel rapid prototyping would enable us to discover potential pitfalls or uncover hidden use cases we would proactively develop features, behind feature flags.

We want to extend the same sense of customer obsession to our Open Source community. We would love to hear use cases that various teams are solving and see how we could adapt parts of Flyte to meet
those requirements. We welcome collaboration and contributions, but please follow our Contribution Guidelines. 


Milestones and Releases
========================
Flyte consists of many components and services. In true Agile fashion each service is independently iterated and co-ordiated by maintaing backwards compatible contracts using protobuf defined in :ref:`flyteidltoc`. Thus components like flytekit, flytepropeller, datacatalog are independently versioned.

We have decided to release a new version of the overall platform in the github.com/lyft/flyte repo every month. Thus we create one milestone for end of every month which points to a new release of
Flyte. This may change in the future, but to match our velocity of development this is our preferred option. Every release will be associated with a CHANGELOG and we will communicate over our
communication medium.

Change management
------------------
To ensure that changes are trackable and the history is explainable, we use a slightly cumbersome but helpful process as follows. Some of these are our immediate goals
- Every PR associated with an issue (automatic searchable documentation)
- Large PR’s associated with Proposals
- Every major change is associated with documentation
- Owner files for all repositories

Release Train
--------------
- We will start tagging issues with milestones, every new issue will be associated with the next milestone. If the issue is not completed by the milestone, or the contributor feels it may slip the deadline, they should manually move it to the next milestone. Every issue not removed, will be moved to the next milestone.
- Every new issue has a “untriaged” label associated with, if we remove this label we should add an assignee. If a contributor is working on the issue, please remove this label.
- Release indicates a release for overall flyte - marked mostly by a milestone.
- Flyte release are monthly
- We may have patch releases eg. 0.1.x in between the monthly releases.

Upcoming Features
=================

.. note::

   We are curently gathering requirements from our users to prioritize. Please
   checkout :ref:`Link here <https://docs.google.com/spreadsheets/d/1LE6X4Pgf1XWriXkDOYPlL7W0SkHWbz5Fa9XKjnCDTMY/edit#gid=0>`_

#. Flytekit Enhancement proposal: GA
   Thw new version of flytekit is available in alpha - refer to :ref:`Release Notes <https://github.com/lyft/flytekit/releases/tag/v0.16.0a2>`_
   For examples refer to :ref:`FlyteCookbook docs on RTD <https://flytecookbook.readthedocs.io/en/latest/auto_recipes/index.html>`_
   Goal: Make flyte almost invisible to the user.
   - Use python native typing system - 0 ramp to learn types
   - Ability to execute everything locally (in some cases mocking out things)
   - minimal imports - just one import for task and workflow
   - simplified extensibility for types and task-types in flytekit

#. Getting started overhaul. Example driven learning.

#. flytekit java feature complete
   Refer to :ref:`Flyte kit java <https://github.com/spotify/flytekit-java>`_

#. Observability stack open source

#. Flyte workflows and tasks inline documentation support and visualization

#. Visualization of Blobs in UI/Console

#. Performance visualization of Task execution

#. Data Lineage and Provenance visualization

#. Plugins


