.. _deployment-gcp-opta:

GCP (GKE) Automated Setup with Opta
-----------------------------------

In order to handle production load robustly, securely and with high availability, there are a number of important tasks that need to
be done independently from the sandbox deployment:

* The kubernetes cluster needs to run securely and robustly
* The sandbox's object store must be replaced by a production grade storage system
* The sandbox's PostgreSQL database must be replaced by a production grade deployment of postgres
* A production grade task queueing system must be provisioned and configured
* A production grade notification system must be provisioned and configured
* All the above must be done in a secure fashion
* (Optionally) An official dns domain must be created
* (Optionally) A production grade email sending system must be provisioned and configured

A Flyte user may provision and orchestrate this setup by themselves, but the Flyte team has partnered with the
`Opta <https://github.com/run-x/opta>`_ team to create a streamlined production deployment strategy for GCP with
ready-to-use templates provided in the `Flyte repo <https://github.com/flyteorg/flyte/tree/master/opta/gcp>`_.
The following demo and documentation specifies how to use and further configure them for AWS (GCP has basically the same steps).

.. youtube:: CMp04-mdtQQ

Deploying Opta Environment and Service for Flyte
************************************************
**The Environment**
To begin using Opta, please first `download the latest version <https://docs.opta.dev/installation/>`_ and all the listed
prerequisites and make sure that you have setup gcloud credentials locally as someone with admin privileges (run ``gcloud auth application-default login`` and ``gcloud auth login``).
With that prepared, go to the `Opta gcp subdirectory <https://github.com/flyteorg/flyte/tree/master/opta/gcp>`_ in the Flyte repo, and open up env.yaml in your editor. Please find and
replace the following values with your desired ones:

* <project_id>: your GCP project id
* <region>: your GCP region
* <domain>: your desired domain for your Flyte deployment (should be a domain which you own or a subdomain thereof - this environment will promptyly take ownership of the domain/subdomain so make sure it will only be used for this purpose)
* <env_name>: a name for the new isolated cloud environment which is going to be created (e.g. flyte-prod)
* <your_company>: your company or organization's name

Once complete please run ``opta apply -c env.yaml`` and follow the prompts.

**DNS Delegation**
Once Opta's apply for the environment is completed, you will need to complete dns delegation to fully setup public
traffic access. You may find instructions on `how to do so here <https://docs.opta.dev/tutorials/ingress/>`__.

**The Flyte Deployment**
Once dns deployment delegation is complete, you may deploy the Flyte service and affiliated resources. Go to the Opta
subdirectory in the Flyte repo, and open up flyte.yaml in your editor. Please find and replace the following values with
your desired ones:

* <project_id>: your GCP project id
* <region>: your GCP region

Once complete please run ``opta apply -c flyte.yaml`` and follow the prompts.

Understanding the Opta Yamls
****************************
The Opta yaml files

**Production Grade Environment**
The Opta env.yaml is responsible for setting up the base infrastructure necessary for most cloud resources. The base
module sets up the VPC and subnets (both public and private) used by the environment as well as the shared KMS keys.
The dns sets up the hosted zone for domain and ssl certificates once completed. The k8s-cluster creates the
Kubernetes cluster and node pool (with encrypted disk storage). And lastly the k8s-base module sets up the resources
within Kubernetes like the autoscaler, metrics server, and ingress.

**Production Grade Database**
The gcp-postgres module in flyte.yaml creates a google Postgresql database with disk encryption and regular snapshot
backups. You can read more about it `here <https://docs.opta.dev/reference/google/service_modules/gcp-postgres>`__

**Production Grade Object Store**
The gcp-gcs module in flyte.yaml creates a new GCS bucket for Flyte, including disk encryption. You can read more about it
`here <https://docs.opta.dev/reference/google/service_modules/gcp-gcs/>`__

**Secure IAM Roles for Data and Control Planes**


**Flyte Deployment via Helm**
A Flyte deployment contains around 50 kubernetes resources.

Additional Setup
****************
By now you should be set up for most production deployments, but there are some extra steps which we recommend that
most users consider.

**Flyte Rbac**
All Flyte deployments are currently insecure on the application level by default (e.g. open/accessible to everyone) so it
is strongly recommended that users `add authentication <https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/cluster/auth_setup.html#authentication-setup>`_.

**Extra configuration**
It is possible to add extra configuration to your Flyte deployment by modifying the values passed in the helm chart
used by Opta. Please refer to the possible values allowed from the `Flyte helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte>`_
and update the values field of the Flyte module in the flyte.yaml file accordingly.


Raw Helm Deployment
*******************
It is certainly possible to deploy a production Flyte cluster directly using the helm chart if a user does not wish to
use Opta. To do so properly, one will need to ensure they have completed the initial security/ha/robustness checklist
from above, and then use `helm <https://helm.sh/>`_ to deploy the `Flyte helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte>`_.

.. role:: raw-html-m2r(raw)
   :format: html
