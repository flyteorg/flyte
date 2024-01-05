.. _divedeep-versioning:

Versions
========

.. tags:: Basic, Glossary

One of the most important features and reasons for certain design decisions in Flyte is the need for machine learning and data practitioners to experiment.
When users experiment, they do so in isolation and try multiple iterations.
Unlike traditional software, the users must conduct multiple experiments concurrently with different environments, algorithms, etc.
This may happen when multiple data scientists simultaneously iterate on the same workflow/pipeline.

The cost of creating an independent infrastructure for each version is enormous and undesirable.
It is beneficial to share the same centralized infrastructure, where the burden of maintaining the infrastructure is with a central infrastructure team,
while the users can use it independently. This improves the cost of operation since the same infrastructure can be reused by multiple teams.

Versioned workflows help users quickly reproduce prior results or identify the source of previous successful experiments.

Why Do You Need Versioning?
---------------------------

Versioning is required to:

- Work on the same project concurrently and identify the version/experiment that was successful.
- Capture the environment for a version and independently launch it.
- Visualize prior runs and tie them to experiment results.
- Rollback to production deployments in case of failures with ease.
- Execute multiple experiments in production, which may use different training or data processing algorithms.
- Understand how a specific system evolved and answer questions related to the effectiveness of a specific strategy.

Operational Benefits of Completely Versioned Workflows/Pipelines
-------------------------------------------------------------------

The entire workflow in Flyte is versioned and all tasks and entities are immutable which makes it possible to completely change the structure of a workflow between versions, without worrying about the consequences for the pipelines in production. 
This hermetic property makes it effortless to manage and deploy new workflow versions and is important for workflows that are long-running. 
If a workflow execution is in progress and another new workflow version has been activated, Flyte guarantees that the execution of the old version continues unhindered.

Consider a scenario where you need to run all the previous executions if there's a bug to be fixed.
Simply fixing the bug in the task may not solve the problem.
Moreover, fixing bugs involves code changes, which may affect the workflow structure.
Flyte addresses this using two properties:

1. Since the entire workflow is versioned, changing the structure has no impact on the existing execution, and the workflow state won't be corrupted.
2. Flyte provides caching/memoization of outputs. As long as the tasks and their behavior have not changed, it is possible to move them around and still recover their previous outputs, without having to rerun the tasks. This strategy will work even if the workflow changes are in a task.

Let us take a sample workflow:

.. mermaid::

    graph TD;
       A-->B;
       B-->C;
       C-->D;

In the above graph, let us assume that task `C` fails. It is then possible to simply fix `C` and ``relaunch`` the previous execution (maintaining the inputs etc). This will not re-run tasks ``A``, and ``B`` as long as they are marked as `cache=True`.

Now, let us consider that the only solution to fix the bug is to change the graph structure and introduce a new step ``B1`` that short circuits the execution to ``D``:

.. mermaid::

    graph TD;
       A-->B;
       B-->B1;
       B1-->D;
       B1-->C;
       C-->D;

The same ``cache=True`` will handle this complicated situation as well.

Why Is Versioning Hard?
-----------------------

Git has become the defacto-standard in version control for code, making it easy to work on branches, merge them, and revert unwanted changes.
But achieving this for a live (running) algorithm usually requires the entire infrastructure to be associated and potentially re-created for every execution.

How Is Versioning Tied to Reproducibility?
------------------------------------------

Workflows can be reproduced without explicit versioning within the system.
To reproduce a past experiment, users need to identify the source code and resurrect any dependencies that the code may have used (for example, TensorFlow 1.x instead of TensorFlow 2.x, or specific Python libraries).
It is also required to instantiate the infrastructure that the previous version may have used. If not recorded, you'll have to ensure that the previously used dataset (say) can be reconstructed.

This is exactly how Flyte was conceived!

In Flyte, every task is versioned, and it precisely captures the dependency set. For external tasks, memoization is recommended so that the constructed dataset can be cached on the Flyte side. This way, one can guarantee reproducible behavior from the external systems.

Moreover, every piece of code is registered with the version of the code that was used to create the instance.
Therefore, users can easily construct the data lineage for all the parts of the workflow.

What Is the Cost of Versioning & Reproducibility?
-------------------------------------------------

One of the costs of versioning and allowing on-demand reproducibility is the need to re-instantiate the infrastructure from scratch.
This may sometimes result in additional overhead. However, the advent of Docker containers and Kubernetes has made it possible to build a platform to achieve these goals.

.. admonition:: Coming soon!

    We are working on reducing the penalty of on-demand infrastructure creation while still maintaining the guarantees. Stay tuned!

What Is the Best Way to Version Your Tasks and Workflows?
---------------------------------------------------------

The best way to version tasks and workflows is to independently version every task with the GIT-SHA or hash of the entire code artifact.
The workflows are also versioned using the GIT-SHA of the containing repository.
