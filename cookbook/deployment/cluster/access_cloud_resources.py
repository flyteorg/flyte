"""
Accessing Cloud Resources
-------------------------

.. TODO: add intro paragraph here

**********************************
Kubernetes ServiceAccount Examples
**********************************

Configure project-wide Kubernetes ServiceAccounts by adding the following to your config:

.. code:: python

   [auth]
   kubernetes_service_account=my-kube-service-acct


Alternatively, pass the role as an argument to ``create_launch_plan``:

.. code:: python

   my_lp = MyWorkflow.create_launch_plan(kubernetes_service_account='my-kube-service-acct')

"""
