.. _production-eks:

Using AWS EKS to host Flyte
------------------------------

Illustration
*************

.. note::

   - Flyte needs a prefix in an AWS S3 bucket to store all its metadata. This is where the data about executions, workflows, tasks is stored
     - this S3 bucket/prefix should be accessible to all FlytePropeller, FlyteAdmin, Datacatalog and running executions (user pods)
   - FlyteAdmin can use any RDBMS database but we recommend Postgres. At scale we have used AWS Aurora
   - Datacatalog also uses a postgres database similar to admin. They both could share the same physical instance, but prefer to have 2 logically separate databases
   - If you want to use AWS IAM role for SeviceAccounts, then you have to manage the provisioning of the service account and providing it to Flyte at the time of execution
   - For secrets, you can use Vault, Kube secrets etc, we are working on getting first class support for this

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_single_cluster_eks.png
    :alt: Illustration of setting up Flyte Cluster in a single AWS EKS (or any K8s cluster on AWS)


