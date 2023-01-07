.. _administrator-deployment-cloud-production:

#######################################
Production Cloud-Based Flyte Deployment
#######################################
The following instructions assume you have Flyte set up as in the :ref:`administrator-deployment-cloud-simple`.

The items here describe additional setup steps to productionalize your Flyte installation. While not strictly required, many administrators will choose to incorporate these.

***********
Ingress/DNS
***********
With a proper Ingress in place, assuming your cluster has an existing Ingress controller, Flyte will be accessible without port forwarding. The base chart installed in the last section already has the ingress rules, but they are not enabled by default.

To turn on ingress, update your ``values.yaml`` file to include the following block.

.. tabbed:: AWS - ``flyte-binary``

    .. literalinclude:: ../../../charts/flyte-binary/eks-production.yaml
       :lines: 123-131

This currently assumes that you have nginx ingress. We'll be updating these in the near future to use the ALB ingress controller instead.

***************
Authentication
***************
Authentication comes stock with Flyte in the form of OAuth 2. Please see the `authentication guide <administrator-configuration-auth-setup>`__ for instructions.

Authorization is not supported in Flyte. The breadth, depth, and variety of requirements we've heard for authorization schemes, and the undertaking any implementation would represent, push authorization beyond the scope of the Flyte project for the forseeable future.

***************
Upgrade Path
***************
Flyte is released regularly and uses semantic versioning. However since Flyte major version 1 will be with us for a while, expect large changes in minor version bumps. These minor version releases will be backwards compatible however for the most part. If using the multi-cluster deployment model for Flyte however, components should be upgraded together.

Expect to see minor version releases roughly 4-6 times a year - we aim to release quarterly, or whenever there is a large enough set of features to warrant a release. Expect to see patch releases at more regular intervals.  Flytekit especially, the Python SDK, will see consistent updates.

To upgrade, simply ``helm upgrade`` your relevant chart.
