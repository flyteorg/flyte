.. _deployment-plugin-setup-k8s-tensorflow-operator:

Kubeflow TensorFlow Operator Plugin Setup
-----------------------------------------

This guide gives an overview of how to set up the Tensorflow operator in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      flytectl sandbox start --source=./flytesnacks
      flytectl config init

3. Install the TensorFlow Operator.

   .. code-block:: bash

      export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
      git clone https://github.com/kubeflow/training-operator.git
      kustomize build training-operator/manifests/overlays/kubeflow | kubectl apply -f -

4. Create a file named ``values-tensorflow.yaml`` and add the following config to it:

   .. code-block::

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

5. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-tensorflow.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

6. (Optional) Build & Serialize the Tensorflow plugin example. (TODO: https://github.com/flyteorg/flyte/issues/1757)

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/kubernetes/kftensorflow serialize

7. Register the TensorFlow plugin example. (TODO: https://github.com/flyteorg/flyte/issues/1757)

   .. code-block:: bash

      TODO: https://github.com/flyteorg/flyte/issues/1757
      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.225/snacks-cookbook-integrations-kubernetes-kftensorflow.tar.gz --archive -p flytesnacks -d development

8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development <TODO: https://github.com/flyteorg/flyte/issues/1757>  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
