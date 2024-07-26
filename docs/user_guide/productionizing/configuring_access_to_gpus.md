(configure-gpus)=

# Configuring access to GPUs

```{eval-rst}
.. tags:: Deployment, Infrastructure, GPU, Intermediate
```

Along with compute resources like CPU and memory, you may want to configure and access GPU resources. 

This section describes the different ways Flyte provides to request accelerator resources directly from the task decorator. 

>The examples in this section use [ImageSpec](https://docs.flyte.org/en/latest/user_guide/customizing_dependencies/imagespec.html#imagespec), a Flyte feature that builds a custom container image without a Dockerfile. Install it using `pip install flytekitplugins-envd`.

## Requesting a GPU with no device preference
The goal in this example is to run the task on a single available GPU :

```python
from flytekit import ImageSpec, Resources, task

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="default",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"))
def gpu_available() -> bool:
   return torch.cuda.is_available() # returns True if CUDA (provided by a GPU) is available
```
### How it works

![](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/gpus/generic_gpu_access.png)

When this task is evaluated, `flytepropeller` injects a [toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in the pod spec:

```yaml
tolerations:    nvidia.com/gpu:NoSchedule op=Exists
```
The Kubernetes scheduler will admit the pod if there are worker nodes in the cluster with a matching taint and available resources.

The resource `nvidia.com/gpu` key name is not arbitrary though. It corresponds to the [Extended Resource](https://kubernetes.io/docs/tasks/administer-cluster/extended-resource-node/) that the Kubernetes worker nodes advertise to the API server through the [device plugin](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/#using-device-plugins). Using the information provided by the device plugin, the Kubernetes scheduler allocates an available accelerator to the Pod.

>NVIDIA maintains a [GPU operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) that automates the management of all software prerequisites on Kubernetes, including the device plugin.

``flytekit`` assumes by default that `nvidia.com/gpu` is the resource name for your GPUs. If your GPU accelerators expose a different resource name, adjust the following key in the Helm values file:

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
>For the above configuration, your worker nodes should have a  `mykey=myvalue:NoSchedule` configured [taint](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).

## Requesting a specific GPU device

The goal is to run the task on a specific type of accelerator: NVIDIA Tesla V100 in the following example:


```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import V100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="default",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=V100, 
              ) #NVIDIA Tesla V100
def gpu_available() -> bool:
   return torch.cuda.is_available()
```


### How it works

When this task is evaluated, `flytepropeller` injects both a toleration and a [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) for a more flexible scheduling configuration.

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
The `key` that the injected node selector uses corresponds to an arbitrary label that your Kubernetes worker nodes should already have. In the above example it's `cloud.google.com/gke-accelerator` but, depending on your cloud provider it could be any other value. You can inform Flyte about the labels your worker nodes use by adjusting the Helm values:

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

`flytekit` supports [Multi-Instance GPU partitioning](https://developer.nvidia.com/blog/getting-the-most-out-of-the-a100-gpu-with-multi-instance-gpu/#mig_partitioning_and_gpu_instance_profiles) on NVIDIA A100 devices for optimal resource utilization.

Example:
```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import A100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="default",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=A100.partition_2g_10gb, 
              ) # 2 compute instances with 10GB memory slice
def gpu_available() -> bool:
   return torch.cuda.is_available()
```
### How it works
In this case, ``flytepropeller`` injects an additional node selector expression to the resulting pod spec, indicating the partition size:

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
In consequence, your Kubernetes worker nodes should have matching labels so the Kubernetes scheduler can admit the Pods:

Node labels (example):
```yaml
nvidia.com/gpu.partition-size: "2g.10gb"
nvidia.com/gpu.accelerator: "nvidia-tesla-a100"
```

 If you want to better control scheduling, configure your worker nodes with taints that match the tolerations injected to the pods.


In the example the ``nvidia.com/gpu.partition-size`` key is arbitrary and can be controlled from the Helm chart:

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
The ``2g.10gb`` value comes from the [NVIDIA A100 supported instance profiles](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#concepts) and it's controlled from the Task decorator (``accelerator=A100.partition_2g_10gb`` in the above example). Depending on the profile requested in the Task, Flyte will inject the corresponding value for the node selector.

>Learn more about the full list of ``flytekit`` supported partition profiles and task decorator options [here](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.extras.accelerators.A100.html#flytekit.extras.accelerators.A100)

## Additional use cases

### Request an A100 device with no preference for partition configuration

Example:

```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import A100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="default",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=A100, 
              ) 
def gpu_available() -> bool:
   return torch.cuda.is_available()
```

#### How it works?

flytekit uses a default `2g.10gb`partition size and `flytepropeller`  injects the node selector that matches labels on nodes with an `A100` device:

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
```

### Request an unpartitioned A100 device
The goal is to run the task using the resources of the entire A100 GPU:

```python
from flytekit import ImageSpec, Resources, task
from flytekit.extras.accelerators import A100

image = ImageSpec(
    base_image= "ghcr.io/flyteorg/flytekit:py3.10-1.10.2",
     name="pytorch",
     python_version="3.10",
     packages=["torch"],
     builder="default",
     registry="<YOUR_CONTAINER_REGISTRY>",
 )

@task(requests=Resources( gpu="1"),
              accelerator=A100.unpartitioned, 
              ) # request the entire A100 device
def gpu_available() -> bool:
   return torch.cuda.is_available()
```

#### How it works

When this task is evaluated `flytepropeller` injects a node selector expression that only matches nodes where the label specifying a partition size is **not** present:

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
            operator: DoesNotExist
```
The expression can be controlled from the Helm values:


**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        gpu-unpartitioned-node-selector-requirement :
          key: cloud.google.com/gke-gpu-partition-size #change to match your node label configuration
          operator: Equal
          value: DoesNotExist
```
**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
        gpu-unpartitioned-node-selector-requirement:
          key: cloud.google.com/gke-gpu-partition-size #change to match your node label configuration
          operator: Equal
          value: DoesNotExist
```


Scheduling can be further controlled by setting in the Helm chart a toleration that `flytepropeller` injects into the task pods:

**flyte-core**
```yaml
configmap:
  k8s:
    plugins:
      k8s:
        gpu-unpartitioned-toleration:
          effect: NoSchedule
          key: cloud.google.com/gke-gpu-partition-size
          operator: Equal
          value: DoesNotExist
```
**flyte-binary**
```yaml
configuration:
  inline:
    plugins:
      k8s:
        gpu-unpartitioned-toleration:
          effect: NoSchedule
          key: cloud.google.com/gke-gpu-partition-size
          operator: Equal
          value: DoesNotExist
```
In case your Kubernetes worker nodes are using taints, they need to match the above configuration.
