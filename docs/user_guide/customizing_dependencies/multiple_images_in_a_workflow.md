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

(multi_images)=

# Multiple images in a workflow

```{eval-rst}
.. tags:: Containerization, Intermediate
```

For every {py:class}`flytekit.PythonFunctionTask` task or a task decorated with the `@task` decorator, you can specify rules for binding container images.
By default, flytekit binds a single container image, i.e., the [default Docker image](https://ghcr.io/flyteorg/flytekit), to all tasks.
To modify this behavior, use the `container_image` parameter available in the {py:func}`flytekit.task` decorator.

:::{note}
If the Docker image is not available publicly, refer to {ref}`Pulling Private Images<private_images>`.
:::

```{code-cell}
:lines_to_next_cell: 2

import numpy as np
from flytekit import task, workflow


@task(container_image="{{.image.mindmeld.fqn}}:{{.image.mindmeld.version}}")
def get_data() -> np.ndarray:
    # here we're importing scikit learn within the Flyte task
    from sklearn import datasets

    iris = datasets.load_iris()
    X = iris.data[:, :2]
    return X


@task(container_image="{{.image.borebuster.fqn}}:{{.image.borebuster.version}}")
def normalize(X: np.ndarray) -> np.ndarray:
    return (X - X.mean(axis=0)) / X.std(axis=0)


@workflow
def multi_images_wf() -> np.ndarray:
    X = get_data()
    X = normalize(X=X)
    return X
```

Observe how the `sklearn` library is imported in the context of a Flyte task.
This approach is beneficial when creating tasks in a single module, where some tasks have dependencies that others do not require.

## Configuring image parameters

The following parameters can be used to configure images in the `@task` decorator:

1. `image` refers to the name of the image in the image configuration. The name `default` is a reserved keyword and will automatically apply to the default image name for this repository.
2. `fqn` refers to the fully qualified name of the image. For example, it includes the repository and domain URL of the image. Example: docker.io/my_repo/xyz.
3. `version` refers to the tag of the image. For example: latest, or python-3.9 etc. If `container_image` is not specified, then the default configured image for the project is used.

## Sending images to `pyflyte` command

You can pass Docker images to the `pyflyte run` or `pyflyte register` command.
For instance:

```
pyflyte run --remote --image mindmeld="ghcr.io/flyteorg/flytecookbook:core-latest" --image borebuster="ghcr.io/flyteorg/flytekit:py3.9-latest" multi_images.py multi_images_wf
```

## Configuring images in `$HOME/.flyte/config.yaml`

To specify images in your `$HOME/.flyte/config.yaml` file (or whichever configuration file you are using), include an "images" section in the configuration.
For example:

```{code-block} yaml
:emphasize-lines: 6-8

admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: localhost:30080
  authType: Pkce
  insecure: true
images:
  mindmeld: ghcr.io/flyteorg/flytecookbook:core-latest
  borebuster: ghcr.io/flyteorg/flytekit:py3.9-latest
console:
  endpoint: http://localhost:30080
logger:
  show-source: true
  level: 0
```

Send the name of the configuration file to your `pyflyte run` command as follows:

```
pyflyte --config $HOME/.flyte/config.yaml run --remote multi_images.py multi_images_wf
```
