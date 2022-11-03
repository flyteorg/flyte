.. _deployment-plugin-setup-k8s:


K8s Operator
-----------------------------------------

This guide gives an overview of setting up the K8s Operator backend plugin in your Flyte deployment.

1. Add Flyte chart repo to Helm.

.. code-block::

 helm repo add flyteorg https://flyteorg.github.io/flyte

2. Set up the cluster.

.. tabbed:: Sandbox

  * Start the sandbox cluster:

    .. code-block:: bash

       flytectl sandbox start

  * Generate flytectl sandbox config:

    .. code-block:: bash

       flytectl config init

.. tabbed:: AWS/GCP

  * Ensure you have flyte cluster up and running in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__ / `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__.
  * Ensure you have the right ``kubeconfig`` and Kubernetes context.
  * Ensure you have the right ``flytectl`` config at ``~/.flyte/config.yaml``.


3. Install the K8S Operator.

.. tabbed:: PyTorch Operator

  * Clone the PyTorch repository:

    .. code-block:: bash

       git clone https://github.com/kubeflow/pytorch-operator.git

  * Build and apply the PyTorch operator:

    .. code-block:: bash

       export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
       kustomize build pytorch-operator/manifests/overlays/kubeflow | kubectl apply -f -

.. tabbed:: TensorFlow Operator

  * Clone the training-operator repository:

    .. code-block:: bash

       git clone https://github.com/kubeflow/training-operator.git

  * Build and apply the TensorFlow operator:

    .. code-block:: bash

       export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
       kustomize build training-operator/manifests/overlays/kubeflow | kubectl apply -f -


.. tabbed:: MPI Operator

  * Clone the MPI repository:

    .. code-block:: bash

       git clone https://github.com/kubeflow/mpi-operator.git

  * Build and apply the MPI operator:

    .. code-block:: bash

       export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
       kustomize build mpi-operator/manifests/overlays/kubeflow | kubectl apply -f -

.. tabbed:: Ray Operator

  * Install the Ray Operator:

    .. code-block:: bash

        export KUBERAY_VERSION=v0.3.0
        kubectl create -k "github.com/ray-project/kuberay/manifests/cluster-scope-resources?ref=${KUBERAY_VERSION}&timeout=90s"
        kubectl apply -k "github.com/ray-project/kuberay/manifests/base?ref=${KUBERAY_VERSION}&timeout=90s"

.. tabbed:: Spark Operator

  * Add the Spark repository:

    .. code-block:: bash

       helm repo add spark-operator https://googlecloudplatform.github.io/spark-on-k8s-operator

  * Install the Spark Operator:

    .. code-block:: bash

       helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace


4. Create a file named ``values-override.yaml`` and add the following config to it:

.. tabbed:: PyTorch Operator

  * Enable PyTorch backend plugin:

    .. code-block:: yaml

       configmap:
         enabled_plugins:
           # -- Task specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
           tasks:
             # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
             task-plugins:
               # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable SageMaker*, Athena if you install the backend
               # plugins
               enabled-plugins:
                 - container
                 - sidecar
                 - k8s-array
                 - pytorch
               default-for-task-types:
                 container: container
                 sidecar: sidecar
                 container_array: k8s-array
                 pytorch: pytorch

.. tabbed:: TensorFlow Operator

  * Enable the TensorFlow backend plugin:

    .. code-block:: yaml

       configmap:
         enabled_plugins:
           # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
           tasks:
             # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
             task-plugins:
               # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable SageMaker*, Athena if you install the backend
               # plugins
               enabled-plugins:
                 - container
                 - sidecar
                 - k8s-array
                 - Tensorflow
               default-for-task-types:
                 container: container
                 sidecar: sidecar
                 container_array: k8s-array
                 Tensorflow: Tensorflow

.. tabbed:: MPI Operator

  * Enable the MPI backend plugin:

    .. code-block:: yaml

       configmap:
         enabled_plugins:
           # -- Task specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
           tasks:
             # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
             task-plugins:
               # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable SageMaker*, Athena if you install the backend
               # plugins
               enabled-plugins:
                 - container
                 - sidecar
                 - k8s-array
                 - mpi
               default-for-task-types:
                 container: container
                 sidecar: sidecar
                 container_array: k8s-array
                 mpi: mpi

.. tabbed:: Ray Operator

  * Enable the Ray backend plugin:

    .. code-block:: yaml

       configmap:
         enabled_plugins:
           # -- Task specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
           tasks:
             # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
             task-plugins:
               # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable SageMaker*, Athena if you install the backend
               # plugins
               enabled-plugins:
                 - container
                 - sidecar
                 - k8s-array
                 - ray
               default-for-task-types:
                 container: container
                 sidecar: sidecar
                 container_array: k8s-array
                 ray: ray

.. tabbed:: Spark Operator

  .. tabbed:: Sandbox

    Since sandbox uses minio, it needs additional configuration.

    .. code-block:: yaml

      cluster_resource_manager:
        # -- Enables the Cluster resource manager component
        enabled: true
        # -- Configmap for ClusterResource parameters
        config:
          # -- ClusterResource parameters
          # Refer to the [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#ClusterResourceConfig) to customize.
          cluster_resources:
            refreshInterval: 5m
            templatePath: "/etc/flyte/clusterresource/templates"
            customData:
              - production:
                  - projectQuotaCpu:
                      value: "5"
                  - projectQuotaMemory:
                      value: "4000Mi"
              - staging:
                  - projectQuotaCpu:
                      value: "2"
                  - projectQuotaMemory:
                      value: "3000Mi"
              - development:
                  - projectQuotaCpu:
                      value: "4"
                  - projectQuotaMemory:
                      value: "5000Mi"
            refresh: 5m

        # -- Resource templates to be applied
        templates:
          # -- Template for namespaces resources
          - key: aa_namespace
            value: |
              apiVersion: v1
              kind: Namespace
              metadata:
                name: {{ namespace }}
              spec:
                finalizers:
                - kubernetes

          - key: ab_project_resource_quota
            value: |
              apiVersion: v1
              kind: ResourceQuota
              metadata:
                name: project-quota
                namespace: {{ namespace }}
              spec:
                hard:
                  limits.cpu: {{ projectQuotaCpu }}
                  limits.memory: {{ projectQuotaMemory }}

          - key: ac_spark_role
            value: |
              apiVersion: rbac.authorization.k8s.io/v1beta1
              kind: Role
              metadata:
                name: spark-role
                namespace: {{ namespace }}
              rules:
              - apiGroups: ["*"]
                resources: ["pods"]
                verbs: ["*"]
              - apiGroups: ["*"]
                resources: ["services"]
                verbs: ["*"]
              - apiGroups: ["*"]
                resources: ["configmaps", "persistentvolumeclaims"]
                verbs: ["*"]

          - key: ad_spark_service_account
            value: |
              apiVersion: v1
              kind: ServiceAccount
              metadata:
                name: spark
                namespace: {{ namespace }}

          - key: ae_spark_role_binding
            value: |
              apiVersion: rbac.authorization.k8s.io/v1beta1
              kind: RoleBinding
              metadata:
                name: spark-role-binding
                namespace: {{ namespace }}
              roleRef:
                apiGroup: rbac.authorization.k8s.io
                kind: Role
                name: spark-role
              subjects:
              - kind: ServiceAccount
                name: spark
                namespace: {{ namespace }}

      sparkoperator:
        enabled: true
        plugin_config:
          plugins:
            spark:
              # -- Spark default configuration
              spark-config-default:
                # We override the default credentials chain provider for Hadoop so that
                # it can use the serviceAccount based IAM role or ec2 metadata based.
                # This is more in line with how AWS works
                - spark.hadoop.fs.s3a.aws.credentials.provider: "org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider"
                - spark.hadoop.fs.s3a.endpoint: "http://minio.flyte.svc.cluster.local:9000"
                - spark.hadoop.fs.s3a.access.key: "minio"
                - spark.hadoop.fs.s3a.secret.key: "miniostorage"
                - spark.hadoop.fs.s3a.path.style.access: "true"
                - spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version: "2"
                - spark.kubernetes.allocation.batch.size: "50"
                - spark.hadoop.fs.s3a.acl.default: "BucketOwnerFullControl"
                - spark.hadoop.fs.s3n.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3n.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3a.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3a.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3a.multipart.threshold: "536870912"
                - spark.excludeOnFailure.enabled: "true"
                - spark.excludeOnFailure.timeout: "5m"
                - spark.task.maxfailures: "8"
      configmap:
        enabled_plugins:
          # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
          tasks:
            # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
            task-plugins:
              # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
              # plugins
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - spark
              default-for-task-types:
                container: container
                sidecar: sidecar
                container_array: k8s-array
                spark: spark

  .. tabbed:: AWS

    .. code-block:: yaml

      cluster_resource_manager:
        # -- Enables the Cluster resource manager component
        enabled: true
        # -- Configmap for ClusterResource parameters
        config:
          # -- ClusterResource parameters
          # Refer to the [structure](https://pkg.go.dev/github.com/lyft/flyteadmin@v0.3.37/pkg/runtime/interfaces#ClusterResourceConfig) to customize.
          cluster_resources:
            refreshInterval: 5m
            templatePath: "/etc/flyte/clusterresource/templates"
            customData:
              - production:
                  - projectQuotaCpu:
                      value: "5"
                  - projectQuotaMemory:
                      value: "4000Mi"
              - staging:
                  - projectQuotaCpu:
                      value: "2"
                  - projectQuotaMemory:
                      value: "3000Mi"
              - development:
                  - projectQuotaCpu:
                      value: "4"
                  - projectQuotaMemory:
                      value: "3000Mi"
            refresh: 5m

        # -- Resource templates that should be applied
        templates:
          # -- Template for namespaces resources
          - key: aa_namespace
            value: |
              apiVersion: v1
              kind: Namespace
              metadata:
                name: {{ namespace }}
              spec:
                finalizers:
                - kubernetes

          - key: ab_project_resource_quota
            value: |
              apiVersion: v1
              kind: ResourceQuota
              metadata:
                name: project-quota
                namespace: {{ namespace }}
              spec:
                hard:
                  limits.cpu: {{ projectQuotaCpu }}
                  limits.memory: {{ projectQuotaMemory }}

          - key: ac_spark_role
            value: |
              apiVersion: rbac.authorization.k8s.io/v1beta1
              kind: Role
              metadata:
                name: spark-role
                namespace: {{ namespace }}
              rules:
              - apiGroups: ["*"]
                resources:
                - pods
                verbs:
                - '*'
              - apiGroups: ["*"]
                resources:
                - services
                verbs:
                - '*'
              - apiGroups: ["*"]
                resources:
                - configmaps
                verbs:
                - '*'

          - key: ad_spark_service_account
            value: |
              apiVersion: v1
              kind: ServiceAccount
              metadata:
                name: spark
                namespace: {{ namespace }}

          - key: ae_spark_role_binding
            value: |
              apiVersion: rbac.authorization.k8s.io/v1beta1
              kind: RoleBinding
              metadata:
                name: spark-role-binding
                namespace: {{ namespace }}
              roleRef:
                apiGroup: rbac.authorization.k8s.io
                kind: Role
                name: spark-role
              subjects:
              - kind: ServiceAccount
                name: spark
                namespace: {{ namespace }}

      sparkoperator:
        enabled: true
        plugin_config:
          plugins:
            spark:
              # -- Spark default configuration
              spark-config-default:
                # We override the default credentials chain provider for Hadoop so that
                # it can use the serviceAccount based IAM role or ec2 metadata based.
                # This is more in line with how AWS works
                - spark.hadoop.fs.s3a.aws.credentials.provider: "com.amazonaws.auth.DefaultAWSCredentialsProviderChain"
                - spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version: "2"
                - spark.kubernetes.allocation.batch.size: "50"
                - spark.hadoop.fs.s3a.acl.default: "BucketOwnerFullControl"
                - spark.hadoop.fs.s3n.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3n.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3a.impl: "org.apache.hadoop.fs.s3a.S3AFileSystem"
                - spark.hadoop.fs.AbstractFileSystem.s3a.impl: "org.apache.hadoop.fs.s3a.S3A"
                - spark.hadoop.fs.s3a.multipart.threshold: "536870912"
                - spark.excludeOnFailure.enabled: "true"
                - spark.excludeOnFailure.timeout: "5m"
                - spark.task.maxfailures: "8"
      configmap:
        enabled_plugins:
          # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
          tasks:
            # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
            task-plugins:
              # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
              # plugins
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - spark
              default-for-task-types:
                container: container
                sidecar: sidecar
                container_array: k8s-array
                spark: spark

5. Upgrade the Flyte Helm release:

.. code-block:: bash

  helm upgrade flyte-core flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml -f values-override.yaml -n flyte

6. Register the plugin example:

.. tabbed:: PyTorch Operator

    .. code-block:: bash

       flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kfpytorch.tar.gz --archive -p flytesnacks -d development --version latest

.. tabbed:: TensorFlow Operator

    .. code-block:: bash

       # TODO: https://github.com/flyteorg/flyte/issues/1757
       flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kftensorflow.tar.gz --archive -p flytesnacks -d development --version latest

.. tabbed:: MPI Operator

    .. code-block:: bash

       flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kfmpi.tar.gz --archive -p flytesnacks -d development --version latest

.. tabbed:: Ray Operator

    .. code-block:: bash

       flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-ray_example.tar.gz --archive -p flytesnacks -d development --version latest


.. tabbed:: Spark Operator

    .. code-block:: bash

       flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-k8s_spark.tar.gz --archive -p flytesnacks -d development --version latest


7. Launch an execution.

.. tabbed:: Flyte Console

  * Navigate to the Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the relevant workflow.
  * Click on `Launch` to open up a launch form.
  * Specify **spark** as the service account if launching a Spark example.
  * Submit the form to launch an execution.

.. tabbed:: Flytectl

    .. tabbed:: PyTorch Operator

      * Retrieve an execution in the form of a YAML file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development kfpytorch.pytorch_mnist.pytorch_training_wf  --latest --execFile exec_spec.yaml

      * Launch! 🚀

        .. code-block:: bash

           flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

    .. tabbed:: TensorFlow Operator

      * Retrieve an execution in the form of a YAML file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development <TODO: https://github.com/flyteorg/flyte/issues/1757>  --latest --execFile exec_spec.yaml

      * Launch! 🚀

        .. code-block:: bash

           flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

    .. tabbed:: MPI Operator

      * Retrieve an execution in the form of a YAML file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development kfmpi.mpi_mnist.horovod_training_wf  --latest --execFile exec_spec.yaml

      * Launch! 🚀

        .. code-block:: bash

           flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

    .. tabbed:: Ray Operator

      * Retrieve an execution in the form of a YAML file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development ray_example.ray_example.ray_workflow  --latest --execFile exec_spec.yaml

      * Launch! 🚀

        .. code-block:: bash

           flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

    .. tabbed:: Spark Operator

      * Retrieve an execution in the form of a YAML file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development k8s_spark.pyspark_pi.my_spark  --latest --execFile exec_spec.yaml

      * Fill in the ``kubeServiceAcct`` as **spark** in the ``exec_spec.yaml`` file.

      * Launch! 🚀

        .. code-block:: bash

           flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
