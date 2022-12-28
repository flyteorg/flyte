.. _administrator-deployment-cloud-simple:

##################################
Core Cloud-Based Flyte Deployment
##################################
These instructions are suitable for the main cloud providers.

****************
Prerequisites
****************
In order to install Flyte, you will need access to the following:

* A Kubernetes cluster (EKS, GKE, etc.).
* At least one blob storage bucket (S3, GCS, etc.), preferably two.
* A Postgres database (RDS, Aurora, etc.). (In the future MySql support may be added.)
* At least one IAM role for the Flyte backend service to assume. You can provision another role for user code to assume as well.

As Flyte documentation cannot keep up with the pace of changes cloud providers are making, please refer to their official documentation for each of these prerequisites.

By Q2 2023, Union AI should have open-sourced a reference implementation of these requirements for the major cloud providers.

***************
Installation
***************

Helm chart
==========
Flyte is installed via a Helm chart.

#. Add the Flyte chart repo to Helm
    .. code-block::

        helm repo add flyteorg https://flyteorg.github.io/flyte

#. Download and update values files.
    .. code-block:: bash

       curl -sL https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-binary/eks-starter.yaml

#. Install
    .. code-block:: bash

	    helm install flyte-binary $(FLYTE_BINARY_CHART_PATH) \
		    --namespace flyte --values eks-starter.yaml



