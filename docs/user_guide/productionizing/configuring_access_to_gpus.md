(configure-gpus)=

# Configuring access to GPUs

```{eval-rst}
.. tags:: Deployment, Infrastructure, GPU, Intermediate
```

Along with compute resources like CPU and Memory, you may want to configure and access GPU resources. 

Flyte provides different ways to request accelerator resources directly from the task decorator.

>The examples in this section use [ImageSpec](https://docs.flyte.org/en/latest/user_guide/customizing_dependencies/imagespec.html#imagespec), a Flyte feature that builds a custom container image without a Dockerfile. Install it using `pip install flytekitplugins-envd`.

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
The goal here is to request any available GPU device(s).

### How it works?

![](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/gpus/generic_gpu_access.png)

When this task is evaluated, `flyteproller` injects a [toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in the pod spec:

```yaml
tolerations:    nvidia.com/gpu:NoSchedule op=Exists
```
The Kubernetes scheduler will admit the pod if there are worker nodes in the cluster with a matching taint and available resources.

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
>For the above configuration, your worker nodes should be tainted with `mykey=myvalue:NoSchedule`.

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
Leveraging a flytekit feature, you can request a specific accelerator device from the task decorator .

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
While the `key` is arbitrary, the value (`nvidia-tesla-v100`) is not. `flytekit` has a set of [predefined](https://docs.flyte.org/en/latest/api/flytekit/extras.accelerators.html#predefined-accelerator-constants) constants and your node label has to use one of those values. 

## Requesting a GPU partition

Example:
```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import A100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="envd",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=A100.partition_2g_10gb, 
              ) # 2 compute instances with 10GB memory slice
def gpu_available() -> bool:
   return torch.cuda.is_available()
```
### How it works?
In this case, ``flytepropeller`` injects an additional nodeSelector to the resulting Pod spec, indicating the partition size:

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: nvidia.com/gpu.accelerator
            operator: In
            values:
            - nvidia-tesla-a100
          - key: nvidia.com/gpu.partition-size
            operator: In
            values:
            - 2g.10gb
```

Plus and additional toleration:

```yaml
  tolerations:
  - effect: NoSchedule
    key: nvidia.com/gpu.accelerator
    operator: Equal
    value: nvidia-tesla-a100
  - effect: NoSchedule
    key: nvidia.com/gpu.partition-size
    operator: Equal
    value: 2g.10gb
```
In consequence, your nodes should have at least matching labels for the Kubernetes scheduler to admit the Pods. If you want to better control scheduling, add matching taints to the nodes.

>NOTE: currently, ``flytekit`` supports partitions only for NVIDIA A100 devices

In the previous example the ``nvidia.com/gpu.partition-size`` key is arbitrary and can be controlled from the Helm chart:

**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        gpu-partition-size-node-label: "nvidia.com/gpu.partition-size" #change to match your node's config
```
**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
       gpu-partition-size-node-label: "nvidia.com/gpu.partition-size" #change to match your node's config 
```
The ``2g.10gb`` value comes from the [NVIDIA A100 supported instance profiles](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#concepts) and it's controlled from the Task decorator (``accelerator=A100.partition_2g_10gb`` in the above example). Depending on the profile requested in the Task, Flyte will inject the corresponding value for the node selector. Your nodes have to use matching labels:

```yaml
nvidia.com/gpu.partition-size: "2g.10gb"
```

>Learn more about the full list of ``flytekit`` supported partition profiles and task decorator options [here](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.extras.accelerators.A100.html#flytekit.extras.accelerators.A100)
