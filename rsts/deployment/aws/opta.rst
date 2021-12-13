.. _deployment-aws-opta:

AWS (EKS) Automated Setup With Opta
===================================

Several essential tasks need to be taken care of independently from the sandbox deployment to achieve high availability and handle production
load robustly and securely.

* The Kubernetes cluster needs to run securely and robustly
* The sandbox's object store must be replaced with a production-grade storage system
* The sandbox's PostgreSQL database must be replaced with a production-grade deployment of PostgreSQL
* A production-grade task queueing system must be provisioned and configured
* A production-grade notification system must be provisioned and configured
* All the above have to be done in a secure manner
* (Optionally) An official DNS domain must be created
* (Optionally) A production-grade email sending system must be provisioned and configured

A Flyte user may provision and orchestrate this setup by themselves, but the Flyte team has partnered with the
`Opta <https://github.com/run-x/opta>`_ team to create a streamlined production deployment strategy for AWS with
ready-to-use templates provided in the `Flyte repo <https://github.com/flyteorg/flyte/tree/master/opta/aws>`__.

The following demo and documentation specify how to use, and further configure them.

.. youtube:: CMp04-mdtQQ

Deploying Opta Environment and Service for Flyte
------------------------------------------------

1. Environment
**************

To begin using Opta, `download the latest version <https://docs.opta.dev/installation/>`__ and all the listed
prerequisites and make sure that you have
`admin/fullwrite AWS credentials setup on your terminal <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html>`__.
With that prepared,

.. _opta-aws-directory:

* Clone Flyte repo: ``git clone git@github.com:flyteorg/flyte.git``
* Go to the ``flyte/opta/aws`` directory
* Open the ``env.yaml`` file in your editor and replace the following values with your desired values:

  * <account_id>: your AWS account ID
  * <region>: your AWS region
  * <domain>: your desired domain for your Flyte deployment (should be a domain which you own or a subdomain thereof - this environment will promptly take ownership of the domain/subdomain, so make sure it will only be used for this purpose)
  * <env_name>: a name for the new isolated cloud environment which is going to be created (e.g., flyte-prod)
  * <your_company>: your company or organization's name

Once complete, run ``opta apply -c env.yaml`` and follow the prompts.

2. DNS Delegation
*****************
Next, you will need to complete DNS delegation to set up public
traffic access fully. You may find instructions on how to do so `here <https://docs.opta.dev/tutorials/ingress/>`__.

3. Flyte Deployment
*******************
Once DNS deployment delegation is complete, you may deploy the Flyte service and affiliated resources.
Open ``flyte.yaml`` present in ``flyte/opta/aws`` in your editor.
Replace the following values with your desired values:

* <account_id>: your AWS account ID
* <region>: your AWS region

Once complete, run ``opta apply -c flyte.yaml`` and follow the prompts.

Understanding the Opta YAMLs
----------------------------

Production-grade Environment
****************************
The Opta ``env.yaml`` is responsible for setting up the base infrastructure necessary for most cloud resources. The base
module sets up the VPC and subnets (both public and private) used by the environment and the shared KMS keys.
The DNS sets up the hosted zone for domain and SSL certificates. The k8s-cluster creates the
Kubernetes cluster and node pool (with encrypted disk storage). And lastly, the k8s-base module sets up the resources
within Kubernetes like the autoscaler, metrics server, and ingress.

Production-grade Database
*************************
The aws-postgres module in ``flyte.yaml`` creates an Aurora PostgreSQL database with disk encryption and regular snapshot
backups. You can read more about it `here <https://docs.opta.dev/modules-reference/service-modules/aws/#postgres>`__.

Production-grade Object Store
*****************************
The aws-s3 module in ``flyte.yaml`` creates a new S3 bucket for Flyte, including disk encryption. You can read more about it
`here <https://docs.opta.dev/modules-reference/service-modules/aws/#aws-s3>`__.

Production-grade Notification System
************************************
Flyte uses a combination of the AWS' Simple Notification Service (SNS) and Simple Queue Service (SQS) for the notification
system. ``flyte.yaml`` creates both the SNS topic and SQS queue (via the notifcationsQueue and topic modules), which are
encrypted with unique KMS keys and only the Flyte roles can access them. You can read more about the queues
`here <https://docs.opta.dev/modules-reference/service-modules/aws/#aws-sqs>`__ and the topics
`here <https://docs.opta.dev/modules-reference/service-modules/aws/#aws-sns>`__.

Production-grade Queueing System
********************************
Flyte uses SQS to power its task scheduling system, and ``flyte.yaml`` creates said queue (via the schedulesQueue
module) with encryption and principle of least privilege RBAC access like the SQS queue mentioned above.

Secure IAM Roles for Data and Control Planes
********************************************


Flyte Deployment via Helm
*************************
Flyte deployment contains around 50 Kubernetes resources.

Additional Setup
----------------

By now, you should be set up for most production deployments, but there are some extra steps that we recommend that
most users consider.

Email Setup
***********

Flyte has the power to send email notifications, which can be enabled in Opta via
`AWS' Simple Email Service <https://aws.amazon.com/ses/>`__ with a few extra steps (NOTE: make sure to have completed DNS
delegation first):

1. Go to ``env.yaml`` and uncomment the last line ( `- type: aws-ses` )
2. Run ``opta apply -c env.yaml`` (again)

   This will enable SES on your account and environment domain -- you may be prompted to fill in some user-specific input to take your account out of SES sandbox if not done already.
   It may take a day for AWS to enable production SES on your account (you will be kept notified via email addresses inputted on the user
   prompt) but that should not prevent you from moving forward.

3. Lastly, go ahead and uncomment the 'Uncomment out for SES' line in the ``flyte.yaml`` and rerun ``opta apply -c flyte.yaml``.

   You will now be able to receive emails sent by Flyte as soon as AWS approves your account. You may also specify other
   non-default email senders via the Heml chart values.

Flyte RBAC
**********

All Flyte deployments are currently insecure at the application level by default (e.g., open/accessible to everyone),
so we strongly recommend users to add :ref:`add authentication <deployment-cluster-config-auth-setup>`.

Extra Configuration
*******************

It is possible to add extra configuration to your Flyte deployment by modifying the values passed in the Helm chart
used by Opta. Refer to the possible values allowed in `Flyte Helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte>`__
and update the values field of the Flyte module in the ``flyte.yaml`` file accordingly.


Raw Helm Deployment
-------------------
It is certainly possible to deploy a production Flyte cluster directly using Helm chart if a user does not wish to
use Opta. To do so properly, one will need to ensure they have completed the initial security/high-availability/robustness checklist,
and then use `Helm <https://helm.sh/>`__ to deploy the `Flyte Helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte>`__.

.. role:: raw-html-m2r(raw)
   :format: html
