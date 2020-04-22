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

We will divide our roadmap into multiple sections. Please refer to each section's description to understand its goal.

Continuous work
=================
This section highlights features and issues that are a continuous area of Focus and will always be in progress. 

1. Documentation 
----------------
   Flyte's aim is to make it absolutely easy for any new user, administrator, contributor to read the documentation and complete their job successfully. We realize That we are not even close. This
   effort will continue on an ongoing basis and we would love feedback on the top areas where we should improve.

2. Bug Fixes & Reliability Improvements
----------------------------------------   
   As mentioned since Flyte is being actively used at Lyft and our scale is always increasing, there are always bugs that we want to squish. So Bugs, are always work in progress and most often then
   not receive highest priority.

3. Observability
----------------
   We are currently working on a one time Observability epic that will enable open source users to use dashboard templates exactly like we use within Lyft (as of April 2020). Going forward,
   Observability is one of the pillars of Flyte. We will continue emphasizing on this aspect and improving it with every iteration.


Living Roadmap
===============
We are trying to maintain a Living roadmap `here <https://docs.google.com/spreadsheets/d/1V8DQfcsX_02Zac5EfAo0UrGJtLwdMPcw3wuuigVIMZU/edit?usp=sharing>`_

We also maintain a raw list of all the ideas `here <https://docs.google.com/document/d/1yq8pIlhlG3gci3GJQNjdAd9bzZ-KYyLfm6I5NVms9-4/edit?usp=sharing>`_


Milestones and Releases
========================
Flyte consists of many components and services. In true Agile fashion each service is independently iterated and co-ordiated by maintaing backwards compatible contracts using protobuf defined in :ref:`flyteidltoc`. Thus components like flytekit, flytepropeller, datacatalog are independently versioned.

We have decided to release a new version of the overall platform in the github.com/lyft/flyte repo every month. Thus we create one milestone for end of every month which points to a new release of
Flyte. This may change in the future, but to match our velocity of development this is our preferred option. Every release will be associated with a CHANGELOG and we will communicate over our
communication medium.

Change management
------------------
- PR templates
- Every PR associated with an issue (automatic searchable documentation)
- Large PR’s associated with Proposals
- Start forming a PMC, to approve large proposals
- Every major change is associated with documentation
- Owner files for all repositories
- Unit test coverage should be above 60% by x
- End to end tests improvement plan

Release Train
--------------
- We will start tagging issues with milestones, every new issue will be associated with the next milestone. We should move it manually to a future milestone or at the time of cutting the milestone that issue may be removed.
- Every new issue has a “untriaged” label associated with, if we remove this label we should add an assignee. If a contributor is working on the issue, please remove this label.
- Release indicates a release for overall flyte - marked mostly by a milestone.
- We can start with monthly cadence
- Of-course we can have patch releases eg. 0.1.x in between the monthly releases.
- Add Golden Test Suite that is tested for every release train (could be very large)

