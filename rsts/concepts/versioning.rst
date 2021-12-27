.. _divedeep-versioning:

##########
Versions
##########
One of the most important features and reasons for certain design decisions in Flyte is the need for Machine Learning and Data practitioners to experiment.
When users are experimenting, they usually work in isolation and try multiple iterations. Unlike traditional software, it is necessary for the users to be able to change versions and sometimes runs multiple concurrently. This may happen when different data scientists are iterating on the same workflow/pipeline concurrently.
The cost of creating independent infrastructure for each version is enormous and not desirable. Also it is desirable to share the same centralized infrastructure, where the burden of maintaining the infrastructure is with a central infra team, while users can use this infrastructure independently. This also improves the cost of operation, as
it is possible to reuse the same infrastructure for more teams.

Moreover, versioning helps in users being able to quickly reproduce prior results or identify prior experiments which were successful.

Why do you need versioning?
=============================
Thus as explained above versioning is needed to
 - Work on the same project concurrently and yet identify the version / experiment that was successful
 - Capture the environment for a version and independently launch this environment
 - visualize prior runs and tie them to experiment results
 - easily and cleanly roll-back production deployments in case of failures
 - execute multiple experiments in production, which may use different training or data processing algorithms


Why versioning is hard?
=============================
Git has become a defacto-standard for version control for code. It makes it extremely easy to work on branches, merge these branches and revert unwanted changes.
But achieveing this for a live (running) algorithm, usually needs the entire infrastructure to be associated and potentially re-created for every execution. Versioning also helps in understanding how a specific system evolved and
answer questions related to effectiveness of a specific strategy.

How is versioning tied to reproducibility?
==============================================

What is the cost of Versioning & Reproducibility?
==================================================

What is the best way to version your tasks and workflows?
===========================================================

