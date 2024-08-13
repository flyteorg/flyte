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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/customizing_dependencies/customizing_dependencies/multi_images.py
:caption: customizing_dependencies/multi_images.py
:lines: 1-24
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

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/customizing_dependencies/
