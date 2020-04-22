.. _features-roles:

Why roles?
==========
Roles can be applied to launch plans (and the executions they create) to determine execution privileges.
Out of the box, Flyte offers a way to configure permissions via IAM role or Kubernetes service account.

IAM role examples
-----------------

Configure project-wide IAM roles by adding the following to your config:

.. code:: python

   [auth]
   assumable_iam_role=arn:aws:iam::123accountid:role/my-super-cool-role


Alternatively, pass the role as an argument to ``create_launch_plan``:

.. code:: python

   my_lp = MyWorkflow.create_launch_plan(assumable_iam_role='arn:aws:iam::123accountid:role/my-super-cool-role')


Kubernetes serviceaccount examples
----------------------------------

Configure project-wide kubernetes serviceaccounts by adding the following to your config:

.. code:: python

   [auth]
   kubernetes_service_account=my-kube-service-acct


Alternatively, pass the role as an argument to ``create_launch_plan``:

.. code:: python

   my_lp = MyWorkflow.create_launch_plan(kubernetes_service_account='my-kube-service-acct')


