.. _divedeep-versioning:

Versions
========

One of the most important features and reasons for certain design decisions in Flyte is the need for machine learning and data practitioners to experiment.
When users experiment, they usually work in isolation and try multiple iterations.
Unlike traditional software, the users must conduct multiple experiments concurrently with different environments, algorithms, etc.
This may happen when different data scientists simultaneously iterate on the same workflow/pipeline.

The cost of creating an independent infrastructure for each version is enormous and not desirable.
Also, it is desirable to share the same centralized infrastructure, where the burden of maintaining the infrastructure is with a central infra team,
while users can use it independently. This also improves the cost of operation, as it is possible to reuse the same infrastructure for more teams.

Moreover, versioned workflows help users quickly reproduce prior results or identify the source of previous successful experiments.

Why Do You Need Versioning?
---------------------------

Versioning is needed to:

- Work on the same project concurrently and yet identify the version/experiment that was successful
- Capture the environment for a version and independently launch this environment
- Visualize prior runs and tie them to experiment results
- Easily and cleanly roll-back production deployments in case of failures
- Execute multiple experiments in production, which may use different training or data processing algorithms
- Understand how a specific system evolved and answer questions related to the effectiveness of a specific strategy

Why Versioning Is Hard?
-----------------------

Git has become the defacto-standard in version control for code. Git makes it extremely easy to work on branches, merge them, and revert unwanted changes.
But achieving this for a live (running) algorithm usually needs the entire infrastructure to be associated and potentially re-created for every execution.

How Is Versioning Tied to Reproducibility?
------------------------------------------

Reproducibility is possible without explicit versioning within the workflow system.
To reproduce a past experiment, users need to identify the source code, resurrect any dependencies that this code may have used (for example, TensorFlow 1.x instead of TensorFlow 2.x, or specific Python libraries).
It is also necessary to instantiate any infrastructure that the previous version may have used and, if not already recorded, ensure that the previously used dataset (say) can be reconstructed.

From the first principles, if reproducibility is considered to be one of the most important concerns, then one would capture all these variables and provide them in an easy-to-use method.
This is how Flyte was exactly conceived. Every task is versioned, and Flyte captures its dependency set precisely. For external tasks, it is highly encouraged to use
memoization so that the constructed dataset is cached on Flyte side, and hence, one can comfortably guarantee reproducible behavior from external systems.
Moreover, every piece of code is registered with the version of the code that was used to create the instance.
Hence, users can easily construct the lineage for all the parts of the workflow.

What Is the Cost of Versioning & Reproducibility?
-------------------------------------------------

One of the costs of versioning and allowing on-demand reproducibility is the need to re-instantiate the infrastructure from scratch.
This may sometimes cause additional overhead. However, the advent of Docker containers and Kubernetes has made it possible to build a platform to achieve these goals.

.. admonition:: Coming soon!

    We are working on reducing the penalty of on-demand infrastructure creation while still maintaining the guarantees. Stay tuned!

What Is the Best Way to Version Your Tasks and Workflows?
---------------------------------------------------------

The best way to version tasks and workflows is to independently version every task with the GIT-SHA or hash of the entire code artifact.
The workflows are also versioned using the GIT-SHA of the containing repository.
