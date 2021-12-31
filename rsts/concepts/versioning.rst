.. _divedeep-versioning:

##########
Versions
##########
One of the most important features and reasons for certain design decisions in Flyte is the need for Machine Learning and Data practitioners to experiment.
When users are experimenting, they usually work in isolation and try multiple iterations. Unlike traditional software, it is necessary for the users conduct multiple experiments concurrently with different environments/algorithms etc. This may happen when different data scientists are iterating on the same workflow/pipeline concurrently.
The cost of creating independent infrastructure for each version is enormous and not desirable. Also it is desirable to share the same centralized infrastructure, where the burden of maintaining the infrastructure is with a central infra team, while users can use this infrastructure independently. This also improves the cost of operation, as
it is possible to reuse the same infrastructure for more teams.

Moreover, versioned workflows helps in users being able to quickly reproduce prior results or identify the source for prior experiments which were successful.

Why do you need versioning?
=============================
As explained above versioning is needed to
 - Work on the same project concurrently and yet identify the version / experiment that was successful
 - Capture the environment for a version and independently launch this environment
 - visualize prior runs and tie them to experiment results
 - easily and cleanly roll-back production deployments in case of failures
 - execute multiple experiments in production, which may use different training or data processing algorithms


Why versioning is hard?
=============================
Git has become a defacto-standard in version control for code. Git makes it extremely easy to work on branches, merge these branches and revert unwanted changes.
But achieving this for a live (running) algorithm, usually needs the entire infrastructure to be associated and potentially re-created for every execution. Versioning also helps in understanding how a specific system evolved and
answer questions related to effectiveness of a specific strategy.


How is versioning tied to reproducibility?
==============================================
Reproducibility is possible without explicit versioning within the workflow system. To reproduce a past experiment, users need to identify the source code, resurrect any dependencies that this code may have used (for example tensorflow 1.x instead of tensorflow 2.x or specific python libraries).
It is also necessary to instantiate any infrastructure that the previous version may have used and if not already recorded, ensure that previously used dataset can be re-constructed.
From first principles, if you consider reproducibility as one of the most important concerns, then one would capture all these variables and provide then in an easy to use method. This is how Flyte was exactly conceived. Every task is versioned, and it captures it dependency set exactly. For external tasks, it is highly encouraged to use
memoization, so that the constructed dataset is cached on Flyte side and hence, one can comfortable guarantee reproducible behavior from external systems as well. Moreover, every code is registered with the version of the code that was used to create this instant and hence users can easily construct the lineage for all the parts of the workflow.

What is the cost of Versioning & Reproducibility?
==================================================
One of the costs of versioning and allowing on-demand reproducibility, is the need to re-instantiate the infrastructure from scratch. This may sometimes cause an additional overhead. The advent of Docker containers and Kubernetes has made it possible to build a platform that makes it possible to achieve these goals.
(Coming soon: We are working on reducing the penalty of on-demand infrastructure creation, while still maintaining the guarantees. Stay Tuned!)


What is the best way to version your tasks and workflows?
===========================================================
The best way to version tasks and workflows is to independently version every task, with the GIT-SHA or the a hash of the entire code artifact. The workflows are also versioned using the GIT-SHA of the containing repository.
