. _howto_serviceaccounts:

######################################################################
How do I use Kubernetes ServiceAccounts to access my cloud resources?
######################################################################

Kubernetes serviceaccount examples
----------------------------------

Configure project-wide kubernetes serviceaccounts by adding the following to your config:

.. code:: python

   [auth]
   kubernetes_service_account=my-kube-service-acct


Alternatively, pass the role as an argument to ``create_launch_plan``:

.. code:: python

   my_lp = MyWorkflow.create_launch_plan(kubernetes_service_account='my-kube-service-acct')