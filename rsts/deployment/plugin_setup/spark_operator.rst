.. _deployment-plugin-setup-spark-operator:

Kubeflow Spark Operator Plugin Setup
------------------------------------

This guide gives an overview of how to set up the Spark operator in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      // NOTE: MPI plugin is only available in v0.18.0+ flyte release.
      flytectl sandbox start --source=./flytesnacks

3. Install the Spark Operator.

   .. code-block:: bash

      helm repo add incubator https://charts.helm.sh/incubator --force-update
      helm install incubator/sparkoperator --namespace spark-operator

4. Create a file named ``values-spark.yaml`` and add the following config to it:

.. code-block::

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

   .. code-block:: bash

      helm upgrade -n flyte -f values-spark.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

6. Build & Serialize the Spark plugin example(Optional).

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/kubernetes/k8s_spark serialize

7. Register the Spark plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-kubernetes-k8s_spark.tar.gz --archive -p flytesnacks -d development

8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development k8s_spark.pyspark_pi.my_spark --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
