(configure-gpus)=

# Configuring access to GPUs

```{eval-rst}
.. tags:: Deployment, Infrastructure, GPU, Intermediate
```

Along with compute resources like CPU and Memory, you may want to configure and access GPU resources. 

Flyte gives you multiple levels of granularity to request accelerator resources from your Task definition.

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

When this task is evaluated, `flyteproller` injects a [toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in the pod spec:

```yaml
tolerations:    nvidia.com/gpu:NoSchedule op=Exists
```
The Kubernetes scheduler will admit the pod if there are worker nodes with matching taints and available resources in the cluster.

The resource `nvidia.com/gpu` key name is not arbitrary. It corresponds to the [Extended Resource](https://kubernetes.io/docs/tasks/administer-cluster/extended-resource-node/) that the Kubernetes worker nodes advertise to the API server through the [device plugin](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/#using-device-plugins).

>NVIDIA maintains a [GPU operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) that automates the management of all software prerequisites on Kubernetes.

If your GPU accelerators expose a different resource name, adjust the following key in the Helm values file:

**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        gpu-resource-name: <YOUR_GPU_RESOURCE_NAME>
```

**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
        gpu-resource-name: <YOUR_GPU_RESOURCE_NAME> 
```

If your infrastructure requires additional tolerations for the scheduling of GPU resources to succeed, adjust the following section in the Helm values file:

**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        resource-tolerations:
        - nvidia.com/gpu: 
          - key: "mykey"
            operator: "Equal"
            value: "myvalue"
            effect: "NoSchedule"  
```
**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
        resource-tolerations:
        - nvidia.com/gpu: 
          - key: "mykey"
            operator: "Equal"
            value: "myvalue"
            effect: "NoSchedule" 
```
>For the above configuration, your worker nodes should have a `mykey=myvalue:NoSchedule` matching taint

## Requesting a specific GPU device

Example:
```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import V100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="envd",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=V100, 
              ) #NVIDIA Tesla V100
def gpu_available() -> bool:
   return torch.cuda.is_available()
```
Leveraging a flytekit feature, you can specify the accelerator device in the task decorator .

### How it works?

When this task is evaluated, `flytepropeller` injects both a toleration and a nodeSelector for a more flexible scheduling configuration.

An example pod spec on GKE would include the following:

```yaml
apiVersion: v1
kind: Pod
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: cloud.google.com/gke-accelerator
            operator: In
            values:
            - nvidia-tesla-v100
  containers:
  - resources:
      limits:
        nvidia.com/gpu: 1
  tolerations:
  - key: nvidia.com/gpu  # auto
    operator: Equal
    value: present
    effect: NoSchedule
  - key: cloud.google.com/gke-accelerator
    operator: Equal
    value: nvidia-tesla-v100
    effect: NoSchedule
```
### Configuring the nodeSelector
The `key` that the injected node selector uses corresponds to an arbitrary label that your Kubernetes worker nodes should apply. In the above example it's `cloud.google.com/gke-accelerator` but, depending on your cloud provider it could be any other value. You can inform Flyte about the labels your worker nodes use by adjusting the Helm values:

**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        gpu-device-node-label: "cloud.google.com/gke-accelerator" #change to match your node's config
```
**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
       gpu-device-node-label: "cloud.google.com/gke-accelerator" #change to match your node's config 
```
While the `key` is arbitrary the `value` is not. flytekit has a set of [predefined](https://docs.flyte.org/en/latest/api/flytekit/extras.accelerators.html#predefined-accelerator-constants) constants and your node label has to use one of those keys. 





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
