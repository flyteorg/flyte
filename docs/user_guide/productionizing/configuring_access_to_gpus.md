(configure-gpus)=

# Configuring access to GPUs

```{eval-rst}
.. tags:: Deployment, Infrastructure, GPU, Intermediate
```

Along with compute resources like CPU and Memory, you may want to configure and access GPU resources. 

Flyte gives you three main levels of granularity to request accelerator resources from your Task definition.

## Requesting a generic accelerator

Example:

```python
from flytekit import ImageSpec, Resources, task

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="envd",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"))
def gpu_available() -> bool:
   return torch.cuda.is_available()
```
The goal here is to make a simple request of any available GPU device(s).

### How it works?

![](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/gpus/generic_gpu_access.png)

When this task is executed, `flyteproller` injects a toleration to the Pod spec:

```yaml
tolerations:    nvidia.com/gpu:NoSchedule op=Exists
```
The resource `nvidia.com/gpu` key name is not arbitrary. It corresponds to the Extended Resource that the K8s worker nodes have advertised to the API Server. 



### Infrastructure requirements

## Requesting a specific GPU device

### Infrastructure requirements

### How it works?

## Requesting a GPU partition

Flyte
allows you to configure the GPU access policy for your cluster. GPUs are expensive and it would not be ideal to
treat machines with GPUs and machines with CPUs equally. You may want to reserve machines with GPUs for tasks
that explicitly request them. To achieve this, Flyte uses the Kubernetes concept of [taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).



Kubernetes can automatically apply tolerations for extended resources like GPUs using the [ExtendedResourceToleration plugin](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#extendedresourcetoleration), enabled by default in some cloud environments. Make sure the GPU nodes are tainted with a key matching the resource name, i.e., `key: nvidia.com/gpu`.

You can also configure Flyte backend to apply specific tolerations. This configuration is controlled under generic  k8s plugin configuration as can be found [here](https://github.com/flyteorg/flyteplugins/blob/5a00b19d88b93f9636410a41f81a73356a711482/go/tasks/pluginmachinery/flytek8s/config/config.go#L120).

The idea of this configuration is that whenever a task that can execute on Kubernetes requests for GPUs, it automatically
adds the matching toleration for that resource (in this case, `gpu`) to the generated PodSpec.
As it follows here, you can configure it to access specific resources using the tolerations for all resources supported by
Kubernetes.

Here's an example configuration:

```yaml
plugins:
  k8s:
    resource-tolerations:
      - nvidia.com/gpu:
        - key: "key1"
          operator: "Equal"
          value: "value1"
          effect: "NoSchedule"
```

Getting this configuration into your deployment will depend on how Flyte is deployed on your cluster. If you use the default Opta/Helm route, you'll need to amend your Helm chart values ([example](https://github.com/flyteorg/flyte/blob/cc127265aec490ad9537d29bd7baff828043c6f5/charts/flyte-core/values.yaml#L629)) so that they end up [here](https://github.com/flyteorg/flyte/blob/3d265f166fcdd8e20b07ff82b494c0a7f6b7b108/deployment/eks/flyte_helm_generated.yaml#L521).
