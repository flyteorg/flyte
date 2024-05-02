# Accelerators

:::{admonition} *Accelerators* and *Accelerated datasets* are entirely different things
An accelerator, in Union, is a specialized hardware device that is used to accelerate the execution of a task.
[Accelerated datasets](../../../data-input-output/accelerated-datasets), on the other hand, is a Union feature that enables quick access to large datasets from within a task.
These concepts are entirely different and should not be confused.
:::

Union allows you to specify [requests and limits](customizing-task-resources) for the number of GPUs available for a given task.
However, in some cases, you may want to be more specific about the type of GPU or other specialized device to be used.

You can use the `accelerator` parameter to specify specific GPU types, variations of GPU types, fractional GPUs. or other specialized hardware devices such as TPUs.

Your Union installation will come pre-configured with the GPUs and other hardware that you requested during onboarding.
Each device type has a constant name that you can use to specify the device in the `accelerator` parameter.
For example:


```{code-block} python
from flytekit.extras.accelerators import A100

    @task(
        limits=Resources(gpu="1"),
        accelerator=A100,
    )
    def my_task() -> None:
        ...
```

## Finding your available accelerators

You can find the accelerators available in your Union installation by going to the **Usage > Compute** dashboard in the Union Console.
In the **Accelerators** section, you will see a list of available accelerators and the the named constants to be used in code to refer to them.

## Requesting the provisioning of accelerators

If you need a specific accelerator that is not available in your Union installation, you can request it by contacting the Union team.
Just click on the **Adjust Configuration** button under **Usage** in the Union Console (or go [here](https://get.support.union.ai/servicedesk/customer/portal/1/group/6/create/30)).

## Using predefined accelerator constants

There are a number of predefined accelerator constants available in the `flytekit.extras.accelerators` module.

The predefined list is not exhaustive, but it includes the most common accelerators.
If you know the name of the accelerator, but there is no predefined constant for it, you can simply pass the string name to the task decorator directly.

Note that in order for a specific accelerator to be available in your Union installation, it must have been provisioned by the Union team.

If using the constants, you can import them directly from the module, e.g.:

```{code-block} python
    from flytekit.extras.accelerators import T4

    @task(
        limits=Resources(gpu="1"),
        accelerator=T4,
    )
    def my_task() -> None:
        ...
```

if you want to use a fractional GPU, you can use the `partitioned` method on the accelerator constant, e.g.:

```{code-block} python
    from flytekit.extras.accelerators import A100

    @task(
        limits=Resources(gpu="1"),
        accelerator=A100.partition_2g_10gb,
    )
    def my_task() -> None:
        ...
```

## List of predefined accelerator constants

* `A10G`: [NVIDIA A10 Tensor Core GPU](https://www.nvidia.com/en-us/data-center/a10-tensor-core-gpu/)
* `L4`: [NVIDIA L4 Tensor Core GPU](https://www.nvidia.com/en-us/data-center/l4/)
* `K80`: [NVIDIA Tesla K80 GPU](https://www.nvidia.com/en-gb/data-center/tesla-k80/)
* `M60`: [NVIDIA Tesla M60 GPU](https://www.nvidia.com/en-us/data-center/tesla-m60/)
* `P4`: [NVIDIA Tesla P4 GPU](https://www.nvidia.com/en-us/data-center/tesla-p4/)
* `P100`: [NVIDIA Tesla P100 GPU](https://www.nvidia.com/en-us/data-center/tesla-p100/)
* `T4`: [NVIDIA T4 Tensor Core GPU](https://www.nvidia.com/en-us/data-center/t4/)
* `V100` [NVIDIA Tesla V100 GPU](https://www.nvidia.com/en-us/data-center/tesla-v100/)
* `A100`: An entire [NVIDIA A100 GPU](https://www.nvidia.com/en-us/data-center/a100/). Fractional partitions are also available:
    * `A100.partition_1g_5gb`: 5GB partition of an A100 GPU.
    * `A100.partition_2g_10gb`: 10GB partition of an A100 GPU - 2x5GB slices with 2/7th of the SM (streaming multiprocessor).
    * `A100.partition_3g_20gb`: 20GB partition of an A100 GPU - 4x5GB slices, with 3/7th fraction of the SM.
    * `A100.partition_4g_20gb`: 20GB partition of an A100 GPU - 4x5GB slices, with 4/7th fraction of the SM.
    * `A100.partition_7g_40gb`: 40GB partition of an A100 GPU - 8x5GB slices, with 7/7th fraction of the SM.
* `A100_80GB`: An entire [NVIDIA A100 80GB GPU](https://www.nvidia.com/en-us/data-center/a100/). Fractional partitions are also available:
    *  `A100_80GB.partition_1g_10gb`: 10GB partition of an A100 80GB GPU - 2x5GB slices with 1/7th of the SM (streaming multiprocessor).
    * `A100_80GB.partition_2g_20gb`: 2GB partition of an A100 80GB GPU - 4x5GB slices with 2/7th of the SM.
    * `A100_80GB.partition_3g_40gb`: 3GB partition of an A100 80GB GPU - 8x5GB slices with 3/7th of the SM.
    * `A100_80GB.partition_4g_40gb`: 4GB partition of an A100 80GB GPU - 8x5GB slices with 4/7th of the SM.
    * `A100_80GB.partition_7g_80gb`: 7GB partition of an A100 80GB GPU - 16x5GB slices with 7/7th of the SM.

For more information see [Specifying Accelerators](https://docs.flyte.org/en/latest/api/flytekit/extras.accelerators.html) in the Flyte documentation.
For more information on partitioning, see [Partitioned GPUs](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#partitioning).
