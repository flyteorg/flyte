"""
.. _configure-gpus:

Configuring Flyte to Access GPUs
--------------------------------

.. tags:: Deployment, Infrastructure, GPU, Intermediate

Along with the simpler resources like CPU/Memory, you may want to configure and access GPU resources. Flyte
allows you to configure the GPU access poilcy for your cluster. GPUs are expensive and it would not be ideal to
treat machines with GPUs and machines with CPUs equally. You may want to reserve machines with GPUs for tasks
that explicitly request GPUs. To achieve this, Flyte uses the Kubernetes concept of `taints and tolerations <https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/>`__.

Kubernetes can automatically apply tolerations for extended resources like GPUs using the `ExtendedResourceToleration plugin <https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#extendedresourcetoleration>`__, enabled by default in some cloud environments. Make sure the GPU nodes are tainted with a key matching the resource name, i.e., ``key: nvidia.com/gpu``.

You can also configure Flyte backend to apply specific tolerations. This configuration is controlled under generic  k8s plugin configuration as can be found `here <https://github.com/flyteorg/flyteplugins/blob/5a00b19d88b93f9636410a41f81a73356a711482/go/tasks/pluginmachinery/flytek8s/config/config.go#L120>`__.

The idea of this configuration is that whenever a task that can execute on Kubernetes requests for GPUs, it automatically
adds the matching toleration for that resource (in this case, ``gpu``) to the generated PodSpec.
As it follows here, you can configure it to access specific resources using the tolerations for all resources supported by
Kubernetes.

Here's an example configuration:

.. code-block:: yaml

    plugins:
      k8s:
        resource-tolerations:
          - nvidia.com/gpu:
            - key: "key1"
              operator: "Equal"
              value: "value1"
              effect: "NoSchedule"

Getting this configuration into your deployment will depend on how Flyte is deployed on your cluster. If you use the default Opta/Helm route, you'll need to amend your Helm chart values (`example <https://github.com/flyteorg/flyte/blob/cc127265aec490ad9537d29bd7baff828043c6f5/charts/flyte-core/values.yaml#L629>`__) so that they end up `here <https://github.com/flyteorg/flyte/blob/3d265f166fcdd8e20b07ff82b494c0a7f6b7b108/deployment/eks/flyte_helm_generated.yaml#L521>`__.
"""
