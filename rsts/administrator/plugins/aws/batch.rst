.. _deployment-plugin-setup-aws-array:

AWS Batch Setup
---------------

This setup document applies to both :std:doc:`map tasks <generated/flytekit.map_task>` and single tasks running on AWS Batch. (For single [non-map] task use, please take note of the additional code when updating Propeller config.)

AWS Batch enables developers, scientists, and engineers to easily and efficiently run hundreds of thousands of batch
computing jobs on AWS.

Flyte abstracts away the complexity of integrating AWS Batch into users' workflows. It takes care of packaging inputs,
reading outputs, scheduling map tasks, leveraging AWS Batch Job Queues to distribute the load and coordinate priorities.

1. Set-up AWS Batch

   At the end of this step, the AWS Account should have a configured Compute Environment and one or more AWS Batch Job Queues.
   Follow the guide `Running batch jobs at scale for less <https://aws.amazon.com/getting-started/hands-on/run-batch-jobs-at-scale-with-ec2-spot/>`_.

2. Modify Users' AWS IAM Role trust policy document

   When running Workflows in Flyte, users have the option to specify a K8s Account and/or an IAM Role to run as. For AWS Batch, an IAM Role must be specified.
   For every one of these IAM Roles, modify the trust policy to allow ECS to assume the role.
   Follow the guide `AWS Batch Execution IAM role <https://docs.aws.amazon.com/batch/latest/userguide/execution-IAM-role.html>`_.

3. Modify System's AWS IAM Role policies

   The recommended way of assigning permissions to flyte components is using `OIDC <https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html>`_.
   This involves assigning an IAM Role for every service account used. You will need to find the IAM Role assigned to the flytepropeller's kubernetes service account.
   Then modify the policy document to allow the role to pass other roles to AWS Batch.
   Follow the guide: `Granting a user permissions to pass a role to an AWS service <https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_passrole.html>`_.

4. Update FlyteAdmin's Config

   FlyteAdmin needs to be made aware of all the AWS Batch Job Queues and how the system should distribute the load onto them.
   The simplest setup looks something like this:

   .. code-block:: yaml

      flyteadmin:
        roleNameKey: "eks.amazonaws.com/role-arn"
      queues:
        # A list of items, one per AWS Batch Job Queue.
        executionQueues:
          # The name of the job queue from AWS Batch
          - dynamic: "tutorial"
            # A list of tags/attributes that can be used to match workflows to this queue.
            attributes:
              - default
        # A list of configs to match project and/or domain and/or workflows to job queues using tags.
        workflowConfigs:
          # An empty rule to match any workflow to the queue tagged as "default"
          - tags:
              - default

   If you are using Helm, this block can be added under `configMaps.adminServer` section `here <https://github.com/flyteorg/flyte/blob/master/charts/flyte/values.yaml#L526-L527>`_.

   An example of a more complex matching config below defines 3 different queues with separate attributes and matching 
   logic based on project/domain/workflowName.

   .. code-block:: yaml

        queues:
          executionQueues:
            - dynamic: "gpu_dynamic"
              attributes:
              - gpu
            - dynamic: "critical"
              attributes:
              - critical
            - dynamic: "default"
              attributes:
              - default
          workflowConfigs:
            - project: "my_queue_1"
              domain: "production"
              workflowName: "my_workflow_1"
              tags:
              - critical
            - project: "production"
              workflowName: "my_workflow_2"
              tags:
              - gpu
            - project: "my_queue_3"
              domain: "production"
              workflowName: "my_workflow_3"
              tags:
              - critical
            - tags:
              - default
    
   These settings can also be dynamically altered through `flytectl` (or flyteAdmin API).
   Read about the :ref:`core concept here <divedeep-admin-matchable-resources>`. Then visit :ref:`flytectl docs <flytectl:flytectl_update_execution-queue-attribute>` for a guide on how to dynamically
   update these configs.

5. Update Flyte Propeller's Config

   AWS Array Plugin requires some configurations to correctly communicate with AWS Batch Service.

   These configurations live within flytepropeller's configMap. The config should be modified to set the following keys:

   .. code-block:: yaml

      plugins:
        aws:
          batch:
            # Must match that set in flyteAdmin's configMap flyteadmin.roleNameKey
            roleAnnotationKey: eks.amazonaws.com/role-arn
          # Must match the desired region to launch these tasks.
          region: us-east-2
      tasks:
        task-plugins:
          enabled-plugins:
            # Enable aws_array task plugin.
            - aws_array
          default-for-task-types:
            # Set it as the default handler for array/map tasks.
            container_array: aws_array
            # Make sure to add this line to enable single (non-map) AWS Batch tasks
            aws-batch: aws_array

Let's now look at how to launch an execution to leverage AWS Batch to execute jobs:

1. Follow `this guide <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/map_task.html#sphx-glr-auto-core-control-flow-map-task-py>`_ to
   write a workflow with a Map Task.

2. Serialize and Register the workflow/task to a Flyte backend.

3. Launch an execution

   .. tabbed:: Flyte Console

      * Navigate to Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the workflow.
      * Click on `Launch` to open up the launch form.
      * Select `IAM Role` and enter the full `AWS Arn` of an IAM Role configured according to the above guide.
      * Submit the form.

   .. tabbed:: Flytectl

      * Retrieve an execution form in the form of a yaml file:

        .. code-block:: bash
     
           flytectl --config ~/.flyte/flytectl.yaml get launchplan -p <project> -d <domain> <workflow full name> --version <version> --execFile ~/map_wf.yaml

      * Fill in `iamRole` field (and optionally `kubeServiceAcct` if required in the deployment)

      * Launch an execution:

        .. code-block:: bash

           flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/map_wf.yaml

As soon as the task starts executing, a link for the AWS Array Job will appear in the log links section in Flyte Console. 
As individual jobs start getting scheduled, links to their individual cloudWatch log streams will also appear in the UI.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/aws_plugin_setup/map_task_success.png
    :alt: A screenshot of Flyte Console displaying log links for a successful array job.

A screenshot of Flyte Console displaying log links for a successful array job.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/aws_plugin_setup/map_task_failure.png
    :alt: A screenshot of Flyte Console displaying log links for a failed array job.

A screenshot of Flyte Console displaying log links for a failed array job.
