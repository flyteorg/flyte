# Flytekit Memory Machine Cloud Plugin

Flyte Agent plugin to allow executing Flyte tasks using MemVerge Memory Machine Cloud.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-mmcloud
```

To get started with MMCloud, refer to the [MMCloud User Guide](https://docs.memverge.com/mmce/current/userguide/olh/index.html).

## Getting Started

This plugin allows executing `PythonFunctionTask` using MMCloud without changing any function code.

[Resource](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/productionizing/customizing_resources.html) (cpu and mem) requests and limits, [container](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/customizing_dependencies/multi_images.html) images, and [environment](https://docs.flyte.org/projects/flytekit/en/latest/generated/flytekit.task.html) variable specifications are supported.

[ImageSpec](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/customizing_dependencies/image_spec.html) may be used to define images to run tasks.

### Credentials

The following [secrets](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/productionizing/use_secrets.html) are required to be defined for the agent server:
* `mmc_address`: MMCloud OpCenter address
* `mmc_username`: MMCloud OpCenter username
* `mmc_password`: MMCloud OpCenter password

### Defaults

Compute resources:
* If only requests are specified, there are no limits.
* If only limits are specified, the requests are equal to the limits.
* If neither resource requests nor limits are specified, the default requests used for job submission are `cpu="1"` and `mem="1Gi"`, and there are no limits.

### Example

`example.py` workflow example:
```python
import pandas as pd
from flytekit import ImageSpec, Resources, task, workflow
from sklearn.datasets import load_wine
from sklearn.linear_model import LogisticRegression

from flytekitplugins.mmcloud import MMCloudConfig

image_spec = ImageSpec(packages=["scikit-learn"], registry="docker.io/memverge")


@task
def get_data() -> pd.DataFrame:
    """Get the wine dataset."""
    return load_wine(as_frame=True).frame


@task(task_config=MMCloudConfig(), container_image=image_spec)  # Task will be submitted as MMCloud job
def process_data(data: pd.DataFrame) -> pd.DataFrame:
    """Simplify the task from a 3-class to a binary classification problem."""
    return data.assign(target=lambda x: x["target"].where(x["target"] == 0, 1))


@task(
    task_config=MMCloudConfig(submit_extra="--migratePolicy [enable=true]"),
    requests=Resources(cpu="1", mem="1Gi"),
    limits=Resources(cpu="2", mem="4Gi"),
    container_image=image_spec,
    environment={"KEY": "value"},
)
def train_model(data: pd.DataFrame, hyperparameters: dict) -> LogisticRegression:
    """Train a model on the wine dataset."""
    features = data.drop("target", axis="columns")
    target = data["target"]
    return LogisticRegression(max_iter=3000, **hyperparameters).fit(features, target)


@workflow
def training_workflow(hyperparameters: dict) -> LogisticRegression:
    """Put all of the steps together into a single workflow."""
    data = get_data()
    processed_data = process_data(data=data)
    return train_model(
        data=processed_data,
        hyperparameters=hyperparameters,
    )
```

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

CMD pyflyte serve agent --port 8000
```
