.. _deployment-deployment-cloud-simple:

#######################################
Single Cluster Simple Cloud Deployment
#######################################

.. tags:: Kubernetes, Infrastructure, Basic

These instructions are suitable for the main cloud providers.

****************
Prerequisites
****************
In order to install Flyte, you will need access to the following:

* A Kubernetes cluster: `EKS <https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html>`__,
  `GKE <https://cloud.google.com/kubernetes-engine/docs/deploy-app-cluster>`__, etc.
* At least one blob storage bucket: `S3 <https://aws.amazon.com/s3/getting-started/>`__,
  `GCS <https://cloud.google.com/storage/docs>`__, etc.
* A Postgres database: `RDS <https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_GettingStarted.html>`__,
  `CloudSQL <https://cloud.google.com/sql/docs/postgres>`__, etc.
* At least one IAM role on `AWS <https://aws.amazon.com/iam/getting-started/>`__,
  `GCP <https://cloud.google.com/iam/docs>`__, etc. This is the role for the Flyte
  backend service to assume. You can provision another role for user code to assume as well.

As Flyte documentation cannot keep up with the pace of change of the cloud
provider APIs, please refer to their official documentation for each of
these prerequisites.

.. note::
   
   `Union.AI <https://www.union.ai/>`__ plans to open-source a reference
   implementation of these requirements for the major cloud providers in early
   2023.

***************
Installation
***************

Flyte is installed via a `Helm <https://helm.sh/>`__ chart. First, add the Flyte
chart repo to Helm:

.. prompt:: bash $

    helm repo add flyteorg https://flyteorg.github.io/flyte

Then download and update the values files:

.. prompt:: bash $

    curl -sL https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-binary/eks-starter.yaml

Finally, install the chart:

.. prompt:: bash $

    helm install flyte-backend flyteorg/flyte-binary \
        --dry-run --namespace flyte --values eks-starter.yaml

When ready to install, remove the ``--dry-run`` switch.

Verify the Installation
=======================

The values supplied by the ``eks-starter.yaml`` file provides only the simplest
installation of Flyte. The core functionality and scalability of Flyte will be
there, but no plugins are included (e.g. Spark tasks will not work), there is no
DNS or SSL, and there is no authentication.

Port Forward Flyte Service
--------------------------

To verify the installation therefore you'll need to port forward the Kubernetes service.

.. prompt:: bash $

    kubectl -n flyte port-forward service/flyte-binary 8088:8088 8089:8089

You should be able to navigate to http://localhost:8088/console.

The Flyte server operates on two different ports, one for HTTP traffic and one for gRPC traffic, which is why we port forward both.

From here, you should be able to run through the :ref:`Getting Started <getting-started>`
examples again. Save a backup copy of your existing configuration if you have one
and generate a new config with ``flytectl``.

.. prompt:: bash $

    mv ~/.flyte/config.yaml ~/.flyte/bak.config.yaml
    flytectl config init --host localhost:8088

This will produce a file like:

.. code-block:: yaml
   :caption: ``~/.flyte/config.yaml``

   admin:
     # For GRPC endpoints you might want to use dns:///flyte.myexample.com
     endpoint: dns:///localhost:8088
     authType: Pkce
     insecure: true
   logger:
     show-source: true
     level: 0

Test Workflow
-------------

You can test a workflow by cloning the ``flytesnacks`` repo and running the
hello world example:

.. prompt:: bash $

   git clone https://github.com/flyteorg/flytesnacks
   cd flytesnacks/examples/basics
   pyflyte run --remote basics/hello_world.py hello_world_wf

***********************************
Flyte in on-premises infrastructure
***********************************

Sometimes, it's also helpful to be able to set up a Flyte environment in an on-premises Kubernetes environment or even on a laptop for testing and development purposes.
Check out `this community-maintained tutorial <https://github.com/davidmirror-ops/flyte-the-hard-way/blob/main/docs/on-premises/single-node/001-configure-single-node-k8s.md>`__ to learn how to setup the required dependencies and deploy the `flyte-binary` chart to a local Kubernetes cluster.


*************
What's Next?
*************

Congratulations ⭐️! Now that you have a Flyte cluster up and running on the cloud,
you can productionize it by following the :ref:`deployment-deployment-cloud-production`
guide.
