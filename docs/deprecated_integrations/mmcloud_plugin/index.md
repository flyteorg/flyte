```{eval-rst}
.. tags:: AWS, GCP, AliCloud, Integration, Advanced
```

(mmcloud_plugin)=

# Memory Machine Cloud plugin

```{note}

This is a legacy implementation of the Memory Machine Cloud integration. We recommend using the {ref}`Memory Machine Cloud agent <mmcloud_agent>` instead.

```

[MemVerge](https://memverge.com/) [Memory Machine Cloud](https://www.mmcloud.io/) (MMCloud)—available on AWS, GCP, and AliCloud—empowers users to continuously optimize cloud resources during runtime, safely execute stateful tasks on spot instances, and monitor resource usage in real time. These capabilities make it an excellent fit for long-running batch workloads. Flyte can be integrated with MMCloud, allowing you to execute Flyte tasks using MMCloud.

## Installation

To install the plugin, run the following command:

```{eval-rst}
.. prompt:: bash

    pip install flytekitplugins-mmcloud
```

To get started with MMCloud, see the [MMCloud user guide](https://docs.memverge.com/mmce/current/userguide/olh/index.html).

## Flyte deployment configuration

The MMCloud plugin is [enabled in FlytePropeller's configuration](https://docs.flyte.org/en/latest/deployment/plugins/memverge/mmcloud.html).

## Getting Started

This plugin allows executing `PythonFunctionTask` using MMCloud without changing any function code.

```{eval-rst}
.. testcode:: awsbatch-quickstart
    from flytekitplugins.mmcloud import MMCloudConfig

    @task(task_config=MMCloudConfig())
    def to_str(i: int) -> str:
        return str(i)
```

[Resource](https://docs.flyte.org/en/latest/user_guide/productionizing/customizing_task_resources.html) (cpu and mem) requests and limits, [container](https://docs.flyte.org/en/latest/user_guide/customizing_dependencies/multiple_images_in_a_workflow.html) images, and environment variable specifications are supported.

[ImageSpec](https://docs.flyte.org/en/latest/user_guide/customizing_dependencies/imagespec.html#image-spec-example) may be used to define images to run tasks.

### Credentials

The following [secrets](https://docs.flyte.org/en/latest/user_guide/productionizing/secrets.html) are required to be defined for the agent server:
* `mmc_address`: MMCloud OpCenter address
* `mmc_username`: MMCloud OpCenter username
* `mmc_password`: MMCloud OpCenter password

### Defaults

Compute resources:
* If only requests are specified, there are no limits.
* If only limits are specified, the requests are equal to the limits.
* If neither resource requests nor limits are specified, the default requests used for job submission are `cpu="1"` and `mem="1Gi"`, and there are no limits.

### Agent Image

Install `flytekitplugins-mmcloud` in the agent image.

A `float` binary (obtainable via the OpCenter) is required. Copy it to the agent image `PATH`.

Sample `Dockerfile` for building an agent image:
```dockerfile
FROM python:3.11-slim-bookworm

WORKDIR /root
ENV PYTHONPATH /root

# flytekit will autoload the agent if package is installed.
RUN pip install flytekitplugins-mmcloud
COPY float /usr/local/bin/float

# For flytekit versions <= v1.10.2, use pyflyte serve.
# CMD pyflyte serve --port 8000
# For flytekit versions > v1.10.2, use pyflyte serve agent.
CMD pyflyte serve agent --port 8000
```

## Example usage

For a usage example, see the {ref}`Memory Machine Cloud example <mmcloud_plugin_example>` page.



```{toctree}
:maxdepth: -1
:hidden:

mmcloud_plugin_example
```
