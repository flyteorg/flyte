.. _administrator-deployment-cloud-production:

#######################################
Production Cloud-Based Flyte Deployment
#######################################
The following instructions assume you have Flyte set up as in the :ref:`administrator-deployment-cloud-simple`.

The items here describe additional setup steps to productionalize your Flyte installation. While not strictly required, many administrators will choose to incorporate these.

***********
Ingress/DNS
***********
With a proper Ingress in place, assuming your cluster has an existing Ingress controller, Flyte will be accessible without the port forward. The base chart installed in the basic installation already has the ingress rules, but they are not enabled by default.

To turn on ingress, update your ``values.yaml`` file to include the following block.

AWS:



This assumes that you've followed official instructions for EKS, which would provision you an ALB ingress controller.

***************
Authentication
***************
Authentication comes stock with Flyte in the form of OAuth 2. Please see the `authentication guide <administrator-configuration-auth-setup>`__ for instructions.

Authorization is not supported in Flyte. The breadth, depth, and variety of requirements we've heard for authorization schemes, and the undertaking any implementation would represent, push authorization beyond the scope of the Flyte project for the forseeable future.


***************
Upgrade Path
***************
