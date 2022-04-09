.. _ideal-flow:

How to Streamline Your Flyte Workflows
--------------------------------------

Flyte could be useful for many use cases, be it model training, data processing, ELT/ETL, or bioinformatics.
When workflows are built and deployed, be it in any domain, we may need automation to reduce human-in-the-loop to some extent;
essentially, to automate serializing and registering workflows, creating docker containers, etc.

Before diving into an example use case that explains how DevOps could power Flyte pipelines, let's look at some DevOps-style features Flyte ships with.

- Every DAG can be versioned with Git commit's SHA or the hash of the code artifact; to know more about versioning, see :ref:`divedeep-versioning`
- Tasks and workflows are immutable, i.e., a version is immutable; any mutation results in a new version
- In the case of rapid iteration of code with no dependency modification, :ref:`fast registration <Fast Registration>` can be used
- Multi-tenancy support through projects and domains
- Executions can be monitored using logs on :ref:`Flyte UI <ui>`

Case Study: MLOps at Lyft
=========================

Flyte powers more than 1 million model training and data processing executions per month at Lyft.
DevOps automation has been made part of Flyte pipelines to handle such a massive number of executions.

- Every time a new PR is created:

  - A Docker container is built from the PR
  - Tasks and workflows are registered with Flyte
- If dependencies are modified, PR securely builds the container again
- For just code changes, fast registration is used

  - Code can be pushed from a local machine where users are not permitted to push Docker containers
- Once a PR is merged, the user simulates a deployment through a deployment pipeline
- At each stage of the deployment, tasks and workflows are registered with a specific domain in Flyte,
  where each domain may change the data directory, associated roles, or some meta attributes like labels and annotations
- For production deployments, logs and metrics are automatically tracked, and models are promoted to serving infrastructure
- Users can use interactive notebooks to retrieve intermediate or final outputs, analyze data and automate various monitoring tasks

Thus, everything is automatically tracked. In fact, multiple users can create isolated PRs and test independently.

Implementing DevOps when running Flyte pipelines could be essential when Flyte is deployed to production.
You could have a team of engineers working on various pipelines with continuous code iteration and deployment.
In such a case, we highly encourage you to incorporate automation to build Flyte pipelines.
This could help reduce the time consumed by development, iteration, and deployment!