.. _deployment-plugin-setup-k8s:

K8s Plugins
-----------------------------------------

.. tags:: Kubernetes, Integration, KubernetesOperator, Spark, AWS, GCP, MachineLearning, DistributedComputing, Advanced

This guide gives an overview of setting up the K8s Operator backend plugin in your Flyte deployment.

Add Flyte chart repo to Helm.

.. prompt:: bash $

   helm repo add flyteorg https://flyteorg.github.io/flyte

Set up the cluster
==================

.. tabs::

   .. tab:: Sandbox
   
      Start the sandbox cluster:
   
      .. prompt:: bash $
   
         flytectl demo start
   
      Generate flytectl config:
   
      .. prompt:: bash $
   
         flytectl config init
   
   .. tab:: AWS/GCP
   
      Make sure you have:
   
      * A flyte cluster up and running in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__ / `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__.
      * The right ``kubeconfig`` and Kubernetes context.
      * The right ``flytectl`` config at ``~/.flyte/config.yaml``.


Install the K8S Operator
========================

.. tabs::

  .. group-tab:: PyTorch/TensorFlow/MPI

     Build and apply the training-operator:
   
     .. code-block:: bash
   
        export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
        kustomize build "https://github.com/kubeflow/training-operator.git/manifests/overlays/standalone?ref=v1.5.0" | kubectl apply -f -


    **Optional: Using a Gang Scheduler**

    With the default Kubernetes scheduler, it can happen that some worker pods of distributed training jobs are scheduled
    later than others due to resource constraints. This often causes the job to fail with a timeout error. To avoid
    this you can use a gang scheduler, meaning that the worker pods are only scheduled once all of them can be scheduled at
    the same time.
    
    To `enable gang scheduling for the Kubeflow training-operator <https://www.kubeflow.org/docs/components/training/job-scheduling/>`_,
    you can install the `Kubernetes scheduler plugins <https://github.com/kubernetes-sigs/scheduler-plugins/tree/master>`_ or `Apache YuniKorn scheduler <https://yunikorn.apache.org/>`_.

    1. Install the `scheduler plugin <https://github.com/kubernetes-sigs/scheduler-plugins/tree/master/manifests/install/charts/as-a-second-scheduler>`_ or
       `Apache YuniKorn <https://yunikorn.apache.org/docs/next/#install>`_ as a second scheduler
    2. Configure the Kubeflow training-operator to use the new scheduler:

        Create a manifest called ``kustomization.yaml`` with the following content:

        .. code-block:: yaml

          apiVersion: kustomize.config.k8s.io/v1beta1
          kind: Kustomization

          resources:
          - github.com/kubeflow/training-operator/manifests/overlays/standalone

          patchesStrategicMerge:
          - patch.yaml


        Create a patch file called ``patch.yaml`` with the following content:

        .. code-block:: yaml

          apiVersion: apps/v1
          kind: Deployment
          metadata:
            name: training-operator
          spec:
            template:
              spec:
                containers:
                - name: training-operator
                  command:
                  - /manager
                  - --gang-scheduler-name=<scheduler-plugins/yunikorn>


        Install the patched kustomization with:

        .. code-block:: bash

          kustomize build path/to/overlay/directory | kubectl apply -f -

    3. (Only for Apache YuniKorn) Configure ``template.metadata.annotations.yunikorn.apache.org/task-group-name`` ,
       ``template.metadata.annotations.yunikorn.apache.org/task-groups`` and
       ``template.metadata.annotations.yunikorn.apache.org/schedulingPolicyParameters`` in Flyte pod templates.
       See `Apache YuniKorn Gang-Scheduling <https://yunikorn.apache.org/docs/next/user_guide/gang_scheduling>`_ for more configuration detail.

    4. Use a Flyte pod template with ``template.spec.schedulerName: scheduler-plugins-scheduler``
       to use the new gang scheduler for your tasks.
      
       See the :ref:`using-k8s-podtemplates` section for more information on pod templates in Flyte.
       You can set the scheduler name in the pod template passed to the ``@task`` decorator. However, to prevent the
       two different schedulers from competing for resources, it is recommended to set the scheduler name in the pod template
       in the ``flyte`` namespace which is applied to all tasks. Non distributed training tasks can be scheduled by the
       gang scheduler as well.



  .. group-tab:: Ray
  
    Install the Ray Operator:
  
    .. code-block:: bash
  
        export KUBERAY_VERSION=v0.3.0
        kubectl create -k "github.com/ray-project/kuberay/manifests/cluster-scope-resources?ref=${KUBERAY_VERSION}&timeout=90s"
        kubectl apply -k "github.com/ray-project/kuberay/manifests/base?ref=${KUBERAY_VERSION}&timeout=90s"
  
  .. group-tab:: Spark
  
    Add the Spark repository:
  
    .. code-block:: bash
  
       helm repo add spark-operator https://googlecloudplatform.github.io/spark-on-k8s-operator
  
    Install the Spark Operator:
  
    .. code-block:: bash
  
       helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
  
  
  .. group-tab:: Dask
  
    Add Dask repository
  
    .. code-block:: bash
  
       helm repo add dask https://helm.dask.org
  
    Install Dask Operator
  
    .. code-block:: bash
  
       helm install dask-operator dask/dask-kubernetes-operator --namespace dask-operator --create-namespace


Specify Plugin Configuration
===============================

Create a file named ``values-override.yaml`` and add the following config to it:

.. tabs::

   .. group-tab:: PyTorch
   
     Enable PyTorch backend plugin:
   
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
   
   .. group-tab:: TensorFlow
   
      Enable the TensorFlow backend plugin:
   
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
                    - tensorflow
                  default-for-task-types:
                    container: container
                    sidecar: sidecar
                    container_array: k8s-array
                    tensorflow: tensorflow
   
   .. group-tab:: MPI
   
      Enable the MPI backend plugin:
   
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
   
   .. group-tab:: Ray
   
      Enable the Ray backend plugin:
   
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
   
   .. group-tab:: Spark
   
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
   
   .. group-tab:: Dask
   
     Enable dask backend plugin
   
     .. code-block:: yaml

        configmap:
          enabled_plugins:
            # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
            tasks:
              # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
              task-plugins:
                # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config).
                # plugins
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - dask
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  dask: dask

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

Upgrade the Flyte Helm release
==============================

.. code-block:: bash

  helm upgrade flyte-core flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml -f values-override.yaml -n flyte

Register the plugin example
===========================

.. tabs::

   .. group-tab:: PyTorch
   
       .. code-block:: bash
   
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kfpytorch.tar.gz --archive -p flytesnacks -d development --version latest
   
   .. group-tab:: TensorFlow
   
       .. code-block:: bash
   
          # TODO: https://github.com/flyteorg/flyte/issues/1757
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kftensorflow.tar.gz --archive -p flytesnacks -d development --version latest
   
   .. group-tab:: MPI
   
       .. code-block:: bash
   
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-kfmpi.tar.gz --archive -p flytesnacks -d development --version latest
   
   .. group-tab:: Ray
   
       .. code-block:: bash
   
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-ray_example.tar.gz --archive -p flytesnacks -d development --version latest
   
   
   .. group-tab:: Spark
   
       .. code-block:: bash
   
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.112/snacks-cookbook-integrations-kubernetes-k8s_spark.tar.gz --archive -p flytesnacks -d development --version latest
   
   .. group-tab:: Dask
   
       .. code-block:: bash
   
          flytectl register files --config ~/.flyte/config.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.3.75/snacks-cookbook-integrations-kubernetes-k8s_dask.tar.gz --archive -p flytesnacks -d development --version latest


Launch an execution
===================

.. tabs::

   .. tab:: Flyte Console
   
      * Navigate to the Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the relevant workflow.
      * Click on `Launch` to open up a launch form.
      * Specify **spark** as the service account if launching a Spark example.
      * Submit the form to launch an execution.
   
   .. tab:: Flytectl

      .. tabs::
   
         .. group-tab:: PyTorch
     
           Retrieve an execution in the form of a YAML file:
     
           .. code-block:: bash
     
              flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development kfpytorch.pytorch_mnist.pytorch_training_wf  --latest --execFile exec_spec.yaml
     
           Launch! 🚀
     
           .. code-block:: bash
     
              flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
     
         .. group-tab:: TensorFlow
     
           Retrieve an execution in the form of a YAML file:
     
           .. code-block:: bash
     
              flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development <TODO: https://github.com/flyteorg/flyte/issues/1757>  --latest --execFile exec_spec.yaml
     
           Launch! 🚀
     
           .. code-block:: bash
     
              flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
     
         .. group-tab:: MPI
     
           Retrieve an execution in the form of a YAML file:
     
           .. code-block:: bash
     
              flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development kfmpi.mpi_mnist.horovod_training_wf  --latest --execFile exec_spec.yaml
     
           Launch! 🚀
     
           .. code-block:: bash
     
              flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
     
         .. group-tab:: Ray
     
           Retrieve an execution in the form of a YAML file:
     
           .. code-block:: bash
     
              flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development ray_example.ray_example.ray_workflow  --latest --execFile exec_spec.yaml
     
           Launch! 🚀
     
           .. code-block:: bash
     
              flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
     
         .. group-tab:: Spark
     
           Retrieve an execution in the form of a YAML file:
     
           .. code-block:: bash
     
              flytectl get launchplan --config ~/.flyte/config.yaml --project flytesnacks --domain development k8s_spark.pyspark_pi.my_spark  --latest --execFile exec_spec.yaml
     
           Fill in the ``kubeServiceAcct`` as **spark** in the ``exec_spec.yaml`` file.
     
           Launch! 🚀
     
           .. code-block:: bash
     
              flytectl --config ~/.flyte/config.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
