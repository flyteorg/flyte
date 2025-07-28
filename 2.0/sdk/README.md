# Flyte v2 SDK

The next-generation SDK for Flyte.

[![Publish Python Packages and Official Images](https://github.com/unionai/unionv2/actions/workflows/publish.yml/badge.svg)](https://github.com/unionai/unionv2/actions/workflows/publish.yml)

## Quick start

1. Run `uv venv`, and `source .venv/bin/activate` to create a new virtual environment.
2. Install the latest version of the SDK by running the following:

```
uv pip install --no-cache --prerelease=allow --upgrade flyte
```

4. Create the config and point it to your cluster by running the following:

```
flyte create config --endpoint <your-endpoint-url> --project <your-project> --domain <your-domain>
```

This will create a `config.yaml` file in the current directory which will be referenced ahead of any other `config.yaml`s found in your system.

5. Now you can run code with the CLI:

```
flyte run <path-to-your-script> <task-name>
```

## Hello World Example

```python
# hello_world.py

import flyte

env = flyte.TaskEnvironment(name="hello_world")


@env.task
async def say_hello(data: str) -> str:
    return f"Hello {data}"


@env.task
async def say_hello_nested(data: str) -> str:
    return await say_hello.override(resources=flyte.Resources(gpu="A100 80G:4")).execute(data)


if __name__ == "__main__":
    import asyncio

    # to run pure python - the SDK is not invoked at all
    asyncio.run(say_hello_nested("test"))

    # To run locally, but run through type system etc
    flyte.init()
    flyte.run(say_hello_nested, "World")

    # To run remote
    flyte.init(endpoint="dns:///localhost:8090", insecure=True)
    flyte.run(say_hello_nested, "World")
    # It is possible to switch local and remote, but keeping init to have and endpoint, but , changing context during run
    flyte.with_runcontext(mode="local").run(...)  # this will run locally only

    # To run remote with a config
    flyte.init_from_config("config.yaml")
```

## CLI

All commands can be run from any root directory.
For examples, it is not needed to have `__init__.py` in the directory.
If you run from a directory, the code will automatically package and upload all modules that are imported.
You can change the behavior by using `--copy-style` flag.

```bash
flyte run hello_world.py say_hello --data "World"
```

To follow the logs for the `a0` action, you can use the `--follow` flag:

```bash
flyte run --follow hello_world.py say_hello --data "World"
```

Note that `--follow` has to be used with the `run` command.

Change copy style:

```bash
flyte run --copy-style all hello_world.py say_hello_nested --data "World"
```

## Building Images

```python
import flyte

env = flyte.TaskEnvironment(
    name="hello_world",
    image=flyte.Image.from_debian_base().with_apt_packages(...).with_pip_packages(...),
)

```

## Deploying

```bash
flyte deploy hello_world.py say_hello_nested
```

## Get information

Get all runs:

```bash
flyte get run
```

Get a specific run:

```bash
flyte get run "run-name"
```

Get all actions for a run:

```bash
flyte get actions "run-name"
```

Get a specific action for a run:

```bash
flyte get action "run-name" "action-name"
```

Get action logs:

```bash
flyte get logs "run-name" ["action-name"]
```

This defaults to root action if no action name is provided

## Running workflows programmatically in Python

You can run any workflow programmatically within the script module using __main__:

```python
if __name__ == "__main__":
    import flyte
    flyte.init()
    flyte.run(say_hello_nested, "World")
```

## Running scripts with dependencies specified in metadata headers

You can also run a `uv` script with dependencies specified in metadata headers
and build the task image automatically based on those dependencies:

```python
# container_images.py

# /// script
# dependencies = [
#    "polars",
#    "flyte>=0.2.0b12"
# ]
# ///

import polars as pl

import flyte


env = flyte.TaskEnvironment(
    name="polars_image",
    image=flyte.Image.from_uv_script(
        __file__,
        name="flyte",
        registry="ghcr.io/<you-username>"
        arch=("linux/amd64", "linux/arm64"),
    ).with_apt_packages("ca-certificates"),
)


@env.task
async def create_dataframe() -> pl.DataFrame:
    return pl.DataFrame(
        {"name": ["Alice", "Bob", "Charlie"], "age": [25, 32, 37], "city": ["New York", "Paris", "Berlin"]}
    )


@env.task
async def print_dataframe(dataframe: pl.DataFrame):
    print(dataframe)


@env.task
async def workflow():
    df = await create_dataframe()
    await print_dataframe(df)


if __name__ == "__main__":
    flyte.init_from_config("config.yaml")
    run = flyte.run(workflow)
    print(run.name)
    print(run.url)
    run.wait(run)
```

When you execute

```bash
 uv run hello_world.py
```

`uv` will automatically update the local virtual environment with the dependencies specified in the metadata headers.
Then, Flyte will build the task image using those dependencies and push it to the registry you specify.
Flyte will then deploy the tasks to the cluster where the system will pull the image and run the tasks using it.
