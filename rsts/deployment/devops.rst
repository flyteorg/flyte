.. _devops-ideal-flow:

Ideal Flow for DevOps
---------------------

Flyte could be useful for many use cases, be it model training, data processing, ELT/ETL, or bioinformatics.
When workflows are built and deployed, be it in any domain, we may need automation to reduce human-in-the-loop to some extent;
essentially, to automate serializing and registering workflows, creating docker containers, etc.

Before diving into an example use case that explains how DevOps could power Flyte pipelines, let's look at some DevOps-style features Flyte ships with.

- Every DAG is versioned with Git commit's SHA, which is modifiable
- Tasks and workflows are immutable
- In the case of rapid iteration of code with no dependency modification, :ref:`deployment-fast-registration` can be used
- Multi-tenancy support through projects and domains
- Executions can be monitored using logs on :ref:`Flyte UI <ui>`

Lyft's Use Case
===============

Flyte powers more than 1 million model training and data processing executions per month at Lyft.
DevOps automation has been made part of Flyte pipelines to handle such a massive number of executions.

- Every time a new PR is created:

  - A Docker container is built from the PR
  - Tasks and workflows are registered with Flyte
- If dependencies are modified, PR securely builds the container again
- For just code changes, fast registration is used
- Once a PR is merged, the user simulates a deployment through a deployment pipeline
- At each stage of the deployment, tasks and workflows are registered with a specific domain in Flyte,
  where each domain may change the data directory, associated roles, or some meta attributes like labels and annotations
- For production deployments, logs and metrics are automatically tracked, and models are promoted to serving infrastructure
- Users can use interactive notebooks to retrieve intermediate or final outputs, analyze data and automate various monitoring tasks

Thus, everything is automatically tracked. In fact, multiple users can create isolated PRs and test independently.

Implementing DevOps when running Flyte pipelines is essential when Flyte is deployed to production.
You could have a team of engineers working on various pipelines with continuous code iteration and deployment.
In such a case, we highly encourage you to incorporate automation to build Flyte pipelines.
This could help speed up your development, iteration, and deployment time!