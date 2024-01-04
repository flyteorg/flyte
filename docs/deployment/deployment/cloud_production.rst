.. _deployment-deployment-cloud-production:

#################################################
Single Cluster Production-grade Cloud Deployment
#################################################

.. tags:: Kubernetes, Infrastructure, Advanced

The following guide assumes you've successfully set up a
:ref:`Single Cluster Simple Cloud Deployment <deployment-deployment-cloud-simple>`.

This guide describes additional setup steps to productionize your Flyte
deployment. While not strictly required, we recommend that you incorporate these
changes.

***********
Ingress/DNS
***********

Assuming your cluster has an existing Ingress controller, Flyte will be
accessible without port forwarding. The base chart installed in the previous
guide already contains the ingress rules, but they are not enabled by default.

To turn on ingress, update your ``values.yaml`` file to include the following block.

.. tabs:: 
   
   .. group-tab:: ``flyte-binary`` on EKS using NGINX

      .. literalinclude:: ../../../charts/flyte-binary/eks-starter.yaml
         :caption: charts/flyte-binary/eks-starter.yaml
         :language: yaml
         :lines: 94-102 

   .. group-tab:: ``flyte-binary``/ on EKS using ALB 

      .. literalinclude:: ../../../charts/flyte-binary/eks-starter.yaml
         :caption: charts/flyte-binary/eks-starter.yaml
         :language: yaml
         :lines: 106-118 
     

   .. group-tab:: ``flyte-core`` on GCP using NGINX  

      .. literalinclude:: ../../../charts/flyte-core/values-gcp.yaml        
         :caption: charts/flyte-core/values-gcp.yaml
         :language: yaml
         :lines: 156-164


***************
Authentication
***************

Authentication comes with Flyte in the form of OAuth 2.0. Please see the
`authentication guide <deployment-configuration-auth-setup>`__ for instructions.

.. note::

   Authorization is not supported out-of-the-box in Flyte. This is due to the
   wide and variety of authorization requirements that different organizations use.

***************
Upgrade Path
***************

To upgrade, simply ``helm upgrade`` your relevant chart.

One thing to keep in mind during upgrades is that Flyte is released regularly
using semantic versioning. Since Flyte ``1.0.0`` will be with us for a while,
you should expect large changes in minor version bumps, which backwards
compatibility being maintained, for the most part.

If you're using the :ref:`multi-cluster <deployment-deployment-multicluster>`
deployment model for Flyte, components should be upgraded together.
