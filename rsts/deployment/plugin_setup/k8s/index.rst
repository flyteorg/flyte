.. _deployment-plugin-setup-k8s:


K8s Operator backend Plugin Setup
-----------------------------------------

This guide gives an overview of how to set up the K8s Operator backend plugin in your Flyte deployment.

1. Add the flyte chart repo to helm

  .. code-block::

     helm repo add flyteorg https://flyteorg.github.io/flyte


2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      flytectl sandbox start

3. Install the Kubeflow Operator.

   .. tabbed:: PyTorch Operator

      * Clone PyTorch repository

        .. code-block:: bash

           git clone https://github.com/kubeflow/pytorch-operator.git

      * Build & apply pytorch operator:

        .. code-block:: bash

           export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
           kustomize build pytorch-operator/manifests/overlays/kubeflow | kubectl apply -f -

   .. tabbed:: Tensorflow Operator

      * Clone training-operator repository

        .. code-block:: bash

           git clone https://github.com/kubeflow/training-operator.git

      * Build & apply Tensorflow operator:

        .. code-block:: bash

           export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
           kustomize build training-operator/manifests/overlays/kubeflow | kubectl apply -f -


   .. tabbed:: MPI Operator

      * Clone training-operator repository

        .. code-block:: bash

           git clone https://github.com/kubeflow/mpi-operator.git

      * Build & apply MPI operator:

        .. code-block:: bash

           export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
           kustomize build mpi-operator/manifests/overlays/kubeflow | kubectl apply -f -

   .. tabbed:: Spark Operator

      * Clone training-operator repository

        .. code-block:: bash

           helm repo add incubator https://charts.helm.sh/incubator --force-update

      * Install spark operator:

        .. code-block:: bash

           helm install incubator/sparkoperator --namespace spark-operator --kubeconfig=~/.flyte/k3s/k3s.yaml


4. Create a file named ``values-override.yaml`` and add the following config to it:

   .. tabbed:: PyTorch Operator

      * Enable the pytorch backend plugin:

        .. code-block:: yaml

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
                     - pytorch
                   default-for-task-types:
                     container: container
                     sidecar: sidecar
                     container_array: k8s-array
                     pytorch: pytorch

   .. tabbed:: Tensorflow Operator

      * Enable the Tensorflow backend plugin:

        .. code-block:: yaml

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
                     - tensorflow
                   default-for-task-types:
                     container: container
                     sidecar: sidecar
                     container_array: k8s-array
                     tensorflow: tensorflow

   .. tabbed:: MPI Operator

      * Enable the MPI backend plugin:

        .. code-block:: yaml

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
                     - mpi
                   default-for-task-types:
                     container: container
                     sidecar: sidecar
                     container_array: k8s-array
                     mpi: mpi

   .. tabbed:: Spark Operator

      * Enable the Spark backend plugin:

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

               - apiVersion: rbac.authorization.k8s.io/v1beta1
                 kind: Role
                 metadata:
                     name: spark-role
                     namespace: {{ namespace }}
                 rules:
                   - apiGroups:
                     - ""
                     resources:
                     - pods
                     verbs:
                     - '*'
                   - apiGroups:
                     - ""
                     resources:
                     - services
                     verbs:
                     - '*'
                   - apiGroups:
                     - ""
                     resources:
                     - configmaps
                     verbs:
                     - '*'
               - apiVersion: v1
                 kind: ServiceAccount
                 metadata:
                   name: spark
                   namespace: {{ namespace }}

               - apiVersion: rbac.authorization.k8s.io/v1beta1
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
                     - spark.blacklist.enabled: "true"
                     - spark.blacklist.timeout: "5m"
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

5. Upgrade the Flyte Helm release.

   .. tabbed:: Sandbox

        .. code-block:: bash

           helm upgrade -n flyte -f values-override.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

   .. tabbed:: AWS/GCP

        .. code-block:: bash

            helm upgrade -n flyte -f values-override.yaml flyteorg/flyte-core --kubeconfig=~/.flyte/k3s/k3s.yaml



6. Register the plugin example.

   .. tabbed:: PyTorch Operator

        .. code-block:: bash

           flytectl register files --config ~/.flyte/flytectl.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.2.225/snacks-cookbook-integrations-kubernetes-kfpytorch.tar.gz --archive -p flytesnacks -d development --version latest

   .. tabbed:: Tensorflow Operator

        .. code-block:: bash

           # TODO: https://github.com/flyteorg/flyte/issues/1757
           flytectl register files --config ~/.flyte/flytectl.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.2.225/<TODO>.tar.gz --archive -p flytesnacks -d development --version latest

   .. tabbed:: MPI Operator

        .. code-block:: bash

           flytectl register files --config ~/.flyte/flytectl.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-kubernetes-kfmpi.tar.gz --archive -p flytesnacks -d development --version latest

   .. tabbed:: Spark Operator

        .. code-block:: bash

           flytectl register files --config ~/.flyte/flytectl.yaml https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-kubernetes-k8s_spark.tar.gz --archive -p flytesnacks -d development --version latest


7. Launch an execution

   .. tabbed:: Flyte Console

      * Navigate to Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the workflow.
      * Click on `Launch` to open up the launch form.
      * Submit the form.

   .. tabbed:: PyTorch Operator

      * Retrieve an execution form in the form of a yaml file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development kfpytorch.pytorch_mnist.pytorch_training_wf  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml

      * Launch an execution:

        .. code-block:: bash

           flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

   .. tabbed:: Tensorflow Operator

      * Retrieve an execution form in the form of a yaml file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development <TODO: https://github.com/flyteorg/flyte/issues/1757>  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml

      * Launch an execution:

        .. code-block:: bash

           flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

   .. tabbed:: MPI Operator

      * Retrieve an execution form in the form of a yaml file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development kfmpi.mpi_mnist.horovod_training_wf  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml

      * Launch an execution:

        .. code-block:: bash

           flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml

   .. tabbed:: Spark Operator

      * Retrieve an execution form in the form of a yaml file:

        .. code-block:: bash

           flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development k8s_spark.pyspark_pi.my_spark  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml

      * Launch an execution:

        .. code-block:: bash

           flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
