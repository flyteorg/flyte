---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(image_spec_example)=

# ImageSpec

```{eval-rst}
.. tags:: Containerization, Intermediate
```

:::{note}
This is an experimental feature, which is subject to change the API in the future.
:::

`ImageSpec` is a way to specify how to build a container image without a Dockerfile. The `ImageSpec` by default will be
converted to an [Envd](https://envd.tensorchord.ai/) config, and the [Envd builder](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-envd/flytekitplugins/envd/image_builder.py#L12-L34) will build the image for you. However, you can also register your own builder to build
the image using other tools.

For every {py:class}`flytekit.PythonFunctionTask` task or a task decorated with the `@task` decorator,
you can specify rules for binding container images. By default, flytekit binds a single container image, i.e.,
the [default Docker image](https://ghcr.io/flyteorg/flytekit), to all tasks. To modify this behavior,
use the `container_image` parameter available in the {py:func}`flytekit.task` decorator, and pass an
`ImageSpec`.

Before building the image, Flytekit checks the container registry first to see if the image already exists. By doing
so, it avoids having to rebuild the image over and over again. If the image does not exist, flytekit will build the
image before registering the workflow, and replace the image name in the task template with the newly built image name.

```{code-cell}
import typing

import pandas as pd
from flytekit import ImageSpec, Resources, task, workflow
```

:::{admonition} Prerequisites
:class: important

- Install [flytekitplugins-envd](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-envd) to build the `ImageSpec`.
- To build the image on remote machine, check this [doc](https://envd.tensorchord.ai/teams/context.html#start-remote-buildkitd-on-builder-machine).
- When using a registry in ImageSpec, `docker login` is required to push the image
:::

+++ {"lines_to_next_cell": 0}

You can specify python packages, apt packages, and environment variables in the `ImageSpec`.
These specified packages will be added on top of the [default image](https://github.com/flyteorg/flytekit/blob/master/Dockerfile), which can be found in the Flytekit Dockerfile.
More specifically, flytekit invokes [DefaultImages.default_image()](https://github.com/flyteorg/flytekit/blob/f2cfef0ec098d4ae8f042ab915b0b30d524092c6/flytekit/configuration/default_images.py#L26-L27) function.
This function determines and returns the default image based on the Python version and flytekit version. For example, if you are using python 3.8 and flytekit 0.16.0, the default image assigned will be `ghcr.io/flyteorg/flytekit:py3.8-1.6.0`.
If desired, you can also override the default image by providing a custom `base_image` parameter when using the `ImageSpec`.

```{code-cell}
pandas_image_spec = ImageSpec(
    base_image="ghcr.io/flyteorg/flytekit:py3.8-1.6.2",
    packages=["pandas", "numpy"],
    python_version="3.9",
    apt_packages=["git"],
    env={"Debug": "True"},
    registry="ghcr.io/flyteorg",
)

sklearn_image_spec = ImageSpec(
    base_image="ghcr.io/flyteorg/flytekit:py3.8-1.6.2",
    packages=["scikit-learn"],
    registry="ghcr.io/flyteorg",
)
```

+++ {"lines_to_next_cell": 0}

:::{important}
Replace `ghcr.io/flyteorg` with a container registry you've access to publish to.
To upload the image to the local registry in the demo cluster, indicate the registry as `localhost:30000`.
:::

`is_container` is used to determine whether the task is utilizing the image constructed from the `ImageSpec`.
If the task is indeed using the image built from the `ImageSpec`, it will then import Tensorflow.
This approach helps minimize module loading time and prevents unnecessary dependency installation within a single image.

```{code-cell}
if sklearn_image_spec.is_container():
    from sklearn.linear_model import LogisticRegression
```

+++ {"lines_to_next_cell": 0}

To enable tasks to utilize the images built with `ImageSpec`, you can specify the `container_image` parameter for those tasks.

```{code-cell}
@task(container_image=pandas_image_spec)
def get_pandas_dataframe() -> typing.Tuple[pd.DataFrame, pd.Series]:
    df = pd.read_csv("https://storage.googleapis.com/download.tensorflow.org/data/heart.csv")
    print(df.head())
    return df[["age", "thalach", "trestbps", "chol", "oldpeak"]], df.pop("target")


@task(container_image=sklearn_image_spec, requests=Resources(cpu="1", mem="1Gi"))
def get_model(max_iter: int, multi_class: str) -> typing.Any:
    return LogisticRegression(max_iter=max_iter, multi_class=multi_class)


# Get a basic model to train.
@task(container_image=sklearn_image_spec, requests=Resources(cpu="1", mem="1Gi"))
def train_model(model: typing.Any, feature: pd.DataFrame, target: pd.Series) -> typing.Any:
    model.fit(feature, target)
    return model


# Lastly, let's define a workflow to capture the dependencies between the tasks.
@workflow()
def wf():
    feature, target = get_pandas_dataframe()
    model = get_model(max_iter=3000, multi_class="auto")
    train_model(model=model, feature=feature, target=target)


if __name__ == "__main__":
    wf()
```

There exists an option to override the container image by providing an Image Spec YAML file to the `pyflyte run` or `pyflyte register` command.
This allows for greater flexibility in specifying a custom container image. For example:

```yaml
# imageSpec.yaml
python_version: 3.11
registry: pingsutw
packages:
  - sklearn
env:
  Debug: "True"
```

```
# Use pyflyte to register the workflow
pyflyte run --remote --image image.yaml image_spec.py wf
```

+++

If you only want to build the image without registering the workflow, you can use the `pyflyte build` command.

```
pyflyte build --remote image_spec.py wf
```

+++

In some cases, you may want to force an image to rebuild, even if the image spec hasn’t changed. If you want to overwrite an existing image, you can pass the `FLYTE_FORCE_PUSH_IMAGE_SPEC=True` to `pyflyte` command or add `force_push()` to the ImageSpec.

```bash
FLYTE_FORCE_PUSH_IMAGE_SPEC=True pyflyte run --remote image_spec.py wf
```

or

```python
image = ImageSpec(registry="ghcr.io/flyteorg", packages=["pandas"]).force_push()
```
