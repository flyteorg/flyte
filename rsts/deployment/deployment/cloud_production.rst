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

.. tabbed:: AWS - ``flyte-binary``

   .. literalinclude:: ../../../charts/flyte-binary/eks-production.yaml
      :caption: charts/flyte-binary/eks-production.yaml
      :language: yaml
      :lines: 127-135

.. note::
   
   This section assumes that you're using the NGINX Ingress controller. Instructions and annotations for the ALB controller
   are covered in the `Flyte The Hard Way <https://github.com/davidmirror-ops/flyte-the-hard-way/blob/main/docs/06-intro-to-ingress.md#setting-up-amazons-load-balancer-alb-ingress-controller>`__ tutorial.

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
