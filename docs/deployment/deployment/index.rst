.. _deployment-deployment:

###################
Deployment Paths
###################

The articles in this section will guide a new Flyte administrator through deploying Flyte.

The most complex parts of a Flyte deployment are authentication, ingress, DNS, and SSL support. Due to the complexity
introduced by these components, we recommend deploying Flyte without these at first and relying on K8s port forwarding
to test your Flyte cluster. After the base deployment is tested, these additional features can be turned on more
seamlessly.

********************************
Components of a Flyte Deployment
********************************

.. important::

   We recommend working with your infrastructure team to set up the cloud service requirements below.

Relational Database
===================

Two of Flyte's components, :ref:`FlyteAdmin <divedeep-admin>` and :ref:`DataCatalog <divedeep-catalog>`, rely on
PostgreSQL to store persistent records. In the sandbox deployment, a containerized version of Postgres is included but
for a proper Flyte installation, we recommend one of the cloud provided databases.

.. note::

   We recommend the following services, depending on your cloud platform of choice:
   
   - **AWS**: `RDS <https://aws.amazon.com/rds/postgresql/>`__
   - **GCP**: `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__
   - **Azure**: `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__

Production Grade Object Store
=============================

Core Flyte components such as :ref:`FlyteAdmin <divedeep-admin>`, :ref:`FlytePropeller <flytepropeller-architecture>`,
:ref:`DataCatalog <divedeep-catalog>`, and user runtime containers rely on an object store to hold files. The sandbox
deployment comes with a containerized `Minio <https://min.io/>`__, which offers AWS S3 compatibility.

.. note::

   We recommend swapping this out for a production-grade object store, depending on your cloud platform of choice:
   
   - **AWS**: `S3 <https://aws.amazon.com/s3/>`__
   - **GCP**: `GCS <https://cloud.google.com/storage/>`__
   - **Azure**: `Azure Blob Storage <https://azure.microsoft.com/en-us/products/storage/blobs>`__

************************
Flyte Deployment Paths
************************

There are three different paths for deploying a Flyte cluster:

.. list-table::
   :header-rows: 1
   :widths: 25, 65, 20

   * - Deployment Path
     - Description
     - Production-grade?
   * - :ref:`Sandbox <deployment-deployment-sandbox>`
     - This uses portable replacements for the relational database and blob store.
       It's good for testing out and experimenting with Flyte.
     - ❌
   * - :ref:`Single Cluster <deployment-deployment-cloud-simple>`
     - This bundles Flyte as one executable. It runs on a single K8s cluster and
       supports all of Flyte's extensions and plugins. Once the simple deployment
       is established, you can follow steps to :ref:`productionize it <deployment-deployment-cloud-production>`.
     - ✅
   * - :ref:`Multiple Clusters <deployment-deployment-multicluster>`
     - For large-scale deployments that require multiple K8s clusters. Flyte's control
       plane (:ref:`FlyteAdmin <divedeep-admin>`, :ref:`FlyteConsole <ui>`, and :ref:`DataCatalog <divedeep-catalog>`)
       is separated from Flyte's execution engine, :ref:`FlytePropeller <flytepropeller-architecture>`, which runs
       typically once per compute cluster.
     - ✅

.. important::

   We recommend the **Single Cluster** option for a capable though not massively scalable cluster.
   
   This option is appropriate if all your compute can `fit on one EKS cluster <https://docs.aws.amazon.com/eks/latest/userguide/service-quotas.html>`__ .
   As of this writing, a single Flyte cluster can handle more than 13,000 nodes.
  
   Regardless of using single or multiple Kubernetes clusters for Flyte, note that ``FlytePropeller`` -the main data plane component- can be scaled out as well by using ``sharding`` if scale demands require it.
   See `Automatic scale-out <https://docs.flyte.org/en/latest/deployment/configuration/performance.html#automatic-scale-out>`__ to learn more about the sharding mechanism.



Helm
====

Flyte uses `Helm <https://helm.sh/>`__ as the K8s release packaging solution, though you may still see some old
`Kustomize <https://kustomize.io/>`__ artifacts in the `flyte <https://github.com/flyteorg/flyte>`__ repo. The core Flyte
team maintains Helm charts that correspond with the latter two deployment paths.

.. note::

   Technically there is a Helm chart for the sandbox environment as well, but it's been tested only with the Dockerized
   K3s bundled container.

.. dropdown:: ``flyte-binary``: chart for the **Single Cluster** option.
   :title: text-muted

   .. literalinclude:: ../../../charts/flyte-binary/Chart.yaml
      :language: yaml
      :caption: charts/flyte-binary/Chart.yaml

.. dropdown:: ``flyte-core``: chart for the **Multiple Cluster** option.
   :title: text-muted

   .. literalinclude:: ../../../charts/flyte-core/Chart.yaml
      :language: yaml
      :caption: charts/flyte-core/Chart.yaml

.. dropdown:: ``flyte-deps``: chart that installs additional useful dependencies alongside Flyte.
   :title: text-muted
  
   .. literalinclude:: ../../../charts/flyte-deps/Chart.yaml
      :language: yaml
      :caption: charts/flyte-deps/Chart.yaml

.. dropdown:: ``flyte``: chart that depends on ``flyte-core``, installing additional dependencies to Flyte deployment.
   :title: text-muted

   .. literalinclude:: ../../../charts/flyte/Chart.yaml
      :language: yaml
      :caption: charts/flyte/Chart.yaml

**************************************
Deployment Tips and Tricks
**************************************

Due to the many choices and constraints that you may face in your organization, the specific steps for deploying Flyte
can vary significantly. For example, which cloud platform to use is typically a big fork in the road for many, and there
are many choices to make in terms of Ingress controllers, auth providers, and versions of different dependent libraries that
may interact with other parts of your stack.

Considering the above, we recommend checking out the `"Flyte The Hard Way" <https://github.com/davidmirror-ops/flyte-the-hard-way/tree/main#flyte-the-hard-way>`__ set of community-maintained tutorials that can guide you through the process of preparing the infrastructure and
deploying Flyte.

In addition to searching and posting on the `#flyte-deployment Slack channel <https://flyte-org.slack.com/archives/C01P3B761A6>`__,
we have a `Github Discussion <https://github.com/flyteorg/flyte/discussions/categories/deployment-tips-tricks>`__
section dedicated to deploying Flyte. Feel free to submit any hints you've found helpful as a discussion, ask questions,
or simply document what worked or what didn't work for you.


.. toctree::
    :maxdepth: 1
    :name: deployment options toc
    :hidden:

    sandbox
    cloud_simple
    cloud_production
    multicluster
