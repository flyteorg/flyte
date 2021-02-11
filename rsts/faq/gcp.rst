.. _faq_gcp:

#######################
Google Cloud Platform
#######################

I tried to run examples, but task fails with 401 error?
-------------------------------------------------------
 Steps:
 - Are you using Workload Identity, then you have to pass in the ServiceAccount when you create the launchplan. Docs here https://lyft.github.io/flyte/user/features/roles.html?highlight=serviceaccount#kubernetes-serviceaccount-examples More information about WorkloadIdentity at https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
 - If you are just using a simple Nodepool wide permissions then check the cluster's ServiceACcount for Storage permissions. Do they look fine?

 - If not, then start a dummy pod in the intended namespace and check for ::

  gcloud auth list

.. note::

    FlytePropeller uses Google Application credentials, but gsutil does not use these credentials


