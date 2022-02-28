.. _divedeep-versioning:

Versions
========

One of the most important features and reasons for certain design decisions in Flyte is the need for machine learning and data practitioners to experiment.
When users experiment, they usually work in isolation and try multiple iterations.
Unlike traditional software, the users must conduct multiple experiments concurrently with different environments, algorithms, etc.
This may happen when different data scientists simultaneously iterate on the same workflow/pipeline.

The cost of creating an independent infrastructure for each version is enormous and not desirable.
Moreover, it is desirable to share the same centralized infrastructure, where the burden of maintaining the infrastructure is with a central infrastructure team,
while users can use it independently. This also improves the cost of operation, since it is possible to reuse the same infrastructure for multiple teams.

Moreover, versioned workflows help users quickly reproduce prior results or identify the source of previous successful experiments.

Why Do You Need Versioning?
---------------------------

Versioning is required to:

- Work on the same project concurrently yet identify the version/experiment that was successful.
- Capture the environment for a version and independently launch this environment.
- Visualize prior runs and tie them to experiment results.
- Easily and cleanly roll-back production deployments in case of failures.
- Execute multiple experiments in production, which may use different training or data processing algorithms.
- Understand how a specific system evolved and answer questions related to the effectiveness of a specific strategy.

Operational Benefits of Completely Versioned Workflows / Pipelines
-------------------------------------------------------------------

Since the entire workflow in Flyte is completely versioned and all tasks and entities are immutable, it is possible to completely change
the structure of a workflow between versions, without worrying about consequences for the pipelines in production. This hermetic property makes it extremely
easy to manage and deploy new workflow versions. This is especially important for workflows that are long-running. Flyte guarantees, that if a workflow execution is in progress
and even if a new workflow version has been activated the execution using the old version, will continue unhindered.

The astute may question, but what if, I had a bug in the previous version and I want to just fix the bug and run all previous executions.
Before we understand how Flyte tackles this, let us analyze the problem further - fixing a bug will need a code change and it is possible
that the bug may actually affect the structure of the workflow. Simply fixing the bug in the task may not solve the problem.

Flyte solves the above problem using 2 properties:

1. Since the workflow is completely versioned, changing the structure has no impact on an existing execution, and the workflow state will not be corrupted.
2. Flyte provides a concept of memoization. As long as the tasks have not changed and their behavior has not changed, it is possible to move them around and their previous outputs will be recovered, without having to rerun these tasks. And if the workflow changes were simply in a task this strategy will still work.

Let us take an example of a workflow:

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

Git has become the defacto-standard in version control for code. Git makes it extremely easy to work on branches, merge them, and revert unwanted changes.
But achieving this for a live (running) algorithm usually needs the entire infrastructure to be associated and potentially re-created for every execution.

How Is Versioning Tied to Reproducibility?
------------------------------------------

Reproducibility is possible without explicit versioning within the workflow system.
To reproduce a past experiment, users need to identify the source code, and resurrect any dependencies that this code may have used (for example, TensorFlow 1.x instead of TensorFlow 2.x, or specific Python libraries).
It is also necessary to instantiate any infrastructure that the previous version may have used and, if not already recorded, ensure that the previously used dataset (say) can be reconstructed.
From the first principles, if reproducibility is considered to be one of the most important concerns, then one would capture all these variables and provide them in an easy-to-use method.

This is exactly how Flyte was conceived!

Every task is versioned, and Flyte precisely captures its dependency set. For external tasks, it is highly encouraged to use
memoization so that the constructed dataset is cached on the Flyte side, and hence, one can comfortably guarantee reproducible behavior from the external systems.
Moreover, every piece of code is registered with the version of the code that was used to create the instance.
Users can therefore easily construct the lineage for all the parts of the workflow.

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
