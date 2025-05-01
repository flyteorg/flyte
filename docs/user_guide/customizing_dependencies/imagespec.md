(image_spec_example)=

# ImageSpec

```{eval-rst}
.. tags:: Containerization, Intermediate
```


`ImageSpec` allows you to customize the container image for your Flyte tasks without a Dockerfile.
`ImageSpec` speeds up the build process by allowing you to reuse previously downloaded packages from the PyPI and APT caches.

By default, the `ImageSpec` will be built using the `default` builder associated with Flytekit, but you can register your own builder.

For example, [flytekitplugins-envd](https://github.com/flyteorg/flytekit/blob/c06ef30518dec2057e554fbed375dfa43b985c60/plugins/flytekit-envd/flytekitplugins/envd/image_builder.py#L25) is another image builder that uses envd to build the ImageSpec.

For every {py:class}`flytekit.PythonFunctionTask` task or a task decorated with the `@task` decorator,
you can specify rules for binding container images. By default, flytekit binds a single container image, i.e.,
the [default Docker image](https://ghcr.io/flyteorg/flytekit), to all tasks. To modify this behavior,
use the `image` parameter available in the {py:func}`flytekit.task` decorator, and pass an
`ImageSpec`.

Before building the image, Flytekit checks the container registry to see if the image already exists.
If the image does not exist,
Flytekit will build the image before registering the workflow and replace the image name in the task template with the newly built image name.

:::{admonition} Prerequisites
:class: important

- Make sure `docker` is running on your local machine.
- When using a registry in ImageSpec, `docker login` is required to push the image
:::

## Install Python or APT packages
You can specify Python packages and APT packages in the `ImageSpec`.
These specified packages will be added on top of the [default image](https://github.com/flyteorg/flytekit/blob/master/Dockerfile), which can be found in the Flytekit Dockerfile.
More specifically, flytekit invokes [DefaultImages.default_image()](https://github.com/flyteorg/flytekit/blob/f2cfef0ec098d4ae8f042ab915b0b30d524092c6/flytekit/configuration/default_images.py#L26-L27) function.
This function determines and returns the default image based on the Python version and flytekit version.
For example, if you are using Python 3.8 and flytekit 1.6.0, the default image assigned will be `ghcr.io/flyteorg/flytekit:py3.8-1.6.0`.

:::{important}
Replace `ghcr.io/flyteorg` with a container registry you can publish to.
To upload the image to the local registry in the demo cluster, indicate the registry as `localhost:30000`.
:::

```python
from flytekit import ImageSpec

sklearn_image_spec = ImageSpec(
  packages=["scikit-learn", "tensorflow==2.5.0"],
  apt_packages=["curl", "wget"],
  registry="ghcr.io/flyteorg",
)
```

## Install Conda packages
Define the ImageSpec to install packages from a specific conda channel.
```python
image_spec = ImageSpec(
  conda_packages=["langchain"],
  conda_channels=["conda-forge"],  # List of channels to pull packages from.
  registry="ghcr.io/flyteorg",
)
```

## Use different Python versions in the image
You can specify the Python version in the `ImageSpec` to build the image with a different Python version.

```python
image_spec = ImageSpec(
  packages=["pandas"],
  python_version="3.9",
  registry="ghcr.io/flyteorg",
)
```

## Import modules only in a specific imageSpec environment

`is_container()` is used to determine whether the task is utilizing the image constructed from the `ImageSpec`.
If the task is indeed using the image built from the `ImageSpec`, it will return true.
This approach helps minimize module loading time and prevents unnecessary dependency installation within a single image.

In the following example, both `task1` and `task2` will import the `pandas` module. However, `Tensorflow` will only be imported in `task2`.

```python
from flytekit import ImageSpec, task
import pandas as pd

pandas_image_spec = ImageSpec(
  packages=["pandas"],
  registry="ghcr.io/flyteorg",
)

tensorflow_image_spec = ImageSpec(
  packages=["tensorflow", "pandas"],
  registry="ghcr.io/flyteorg",
)

# Return if and only if the task is using the image built from tensorflow_image_spec.
if tensorflow_image_spec.is_container(): 
  import tensorflow as tf

@task(image=pandas_image_spec)
def task1() -> pd.DataFrame:
  return pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [1, 22]})


@task(image=tensorflow_image_spec)
def task2() -> int:
  num_gpus = len(tf.config.list_physical_devices('GPU'))
  print("Num GPUs Available: ", num_gpus)
  return num_gpus
```

## Install CUDA in the image
There are few ways to install CUDA in the image.

### Use Nvidia docker image
CUDA is pre-installed in the Nvidia docker image. You can specify the base image in the `ImageSpec`.
```python
image_spec = ImageSpec(
  base_image="nvidia/cuda:12.6.1-cudnn-devel-ubuntu22.04",
  packages=["tensorflow", "pandas"],
  python_version="3.9",
  registry="ghcr.io/flyteorg",
)
```

### Install packages from extra index
CUDA can be installed by specifying the `pip_extra_index_url` in the `ImageSpec`.
```python
image_spec = ImageSpec(
  name="pytorch-mnist",
  packages=["torch", "torchvision", "flytekitplugins-kfpytorch"],
  pip_extra_index_url=["https://download.pytorch.org/whl/cu118"],
  registry="ghcr.io/flyteorg",
)
```

## Build an image in different architecture
You can specify the platform in the `ImageSpec` to build the image in a different architecture, such as `linux/arm64` or `darwin/arm64`.
```python
image_spec = ImageSpec(
  packages=["pandas"],
  platform="linux/arm64",
  registry="ghcr.io/flyteorg",
)
```

## Install flytekit from GitHub
When you update the flytekit, you may want to test the changes with your tasks.
You can install the flytekit from a specific commit hash in the `ImageSpec`.

```python
new_flytekit = "git+https://github.com/flyteorg/flytekit@90a4455c2cc2b3e171dfff69f605f47d48ea1ff1"
new_spark_plugins = f"git+https://github.com/flyteorg/flytekit.git@90a4455c2cc2b3e171dfff69f605f47d48ea1ff1#subdirectory=plugins/flytekit-spark"

image_spec = ImageSpec(
  apt_packages=["git"],
  packages=[new_flytekit, new_spark_plugins],
  registry="ghcr.io/flyteorg",
)
```

## Customize the tag of the image
You can customize the tag of the image by specifying the `tag_format` in the `ImageSpec`.
In the following example, the full qualified image name will be `ghcr.io/flyteorg/my-image:<spec_hash>-dev`.

```python
image_spec = ImageSpec(
  name="my-image",
  packages=["pandas"],
  tag_format="{spec_hash}-dev",
  registry="ghcr.io/flyteorg",
)
```

## Copy additional files or directories
You can specify files or directories to be copied into the container `/root`, allowing users to access the required files. The directory structure will match the relative path. Since Docker only supports relative paths, absolute paths and paths outside the current working directory (e.g., paths with "../") are not allowed.

```py
from flytekit.image_spec import ImageSpec
from flytekit import task, workflow

image_spec = ImageSpec(
    name="image_with_copy",
    registry="localhost:30000",
    builder="default",
    copy=["files/input.txt"],
)

@task(image=image_spec)
def my_task() -> str:
    with open("/root/files/input.txt", "r") as f:
        return f.read()
```

## Define ImageSpec in a YAML File

You can override the container image by providing an ImageSpec YAML file to the  `pyflyte run` or `pyflyte register` command.
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

## Build the image without registering the workflow

If you only want to build the image without registering the workflow, you can use the `pyflyte build` command.

```
pyflyte build --remote image_spec.py wf
```

## Force push an image

In some cases, you may want to force an image to rebuild, even if the ImageSpec hasn’t changed.
To overwrite an existing image, pass the `FLYTE_FORCE_PUSH_IMAGE_SPEC=True` to the `pyflyte` command.

```bash
FLYTE_FORCE_PUSH_IMAGE_SPEC=True pyflyte run --remote image_spec.py wf
```

You can also force push an image in the Python code by calling the `force_push()` method.

```python
image = ImageSpec(registry="ghcr.io/flyteorg", packages=["pandas"]).force_push()
```
[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/customizing_dependencies/

## Getting source files into ImageSpec
Typically, getting source code files into a task's image at run time on a live Flyte backend is done through the fast registration mechanism.

However, if your `ImageSpec` constructor specifies a `source_root` and the `copy` argument is set to something other than `CopyFileDetection.NO_COPY`, then files will be copied regardless of fast registration status.
If the `source_root` and `copy` fields to an `ImageSpec` are left blank, then whether or not your source files are copied into the built `ImageSpec` image depends on whether or not you use fast registration. Please see [registering workflows](https://docs.flyte.org/en/latest/flyte_fundamentals/registering_workflows.html#containerizing-your-project) for the full explanation.

Since files are sometimes copied into the built image, the tag that is published for an ImageSpec will change based on whether fast register is enabled, and the contents of any files copied.
