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

(remote_task)=

# Running tasks

## Flytectl

This is a multi-step process where we create an execution spec file, update the spec file, and then create the execution.
More details can be found in the [Flytectl API reference](https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html).

**Generate execution spec file**

```
flytectl get tasks -d development -p flytesnacks workflows.example.generate_normal_df  --latest --execFile exec_spec.yaml
```

**Update the input spec file for arguments to the workflow**

```
iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
inputs:
  n: 200
  mean: 0.0
  sigma: 1.0
kubeServiceAcct: ""
targetDomain: ""
targetProject: ""
task: workflows.example.generate_normal_df
version: "v1"
```

**Create execution using the exec spec file**

```
flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml
```

**Monitor the execution by providing the execution id from create command**

```
flytectl get execution -p flytesnacks -d development <execid>
```

## FlyteRemote

A task can be launched via FlyteRemote programmatically.

```python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config, SerializationSettings

# FlyteRemote object is the main entrypoint to API
remote = FlyteRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)

# Get Task
flyte_task = remote.fetch_task(name="workflows.example.generate_normal_df", version="v1")

flyte_task = remote.register_task(
    entity=flyte_task,
    serialization_settings=SerializationSettings(image_config=None),
    version="v2",
)

# Run Task
execution = remote.execute(
     flyte_task, inputs={"n": 200, "mean": 0.0, "sigma": 1.0}, execution_name="task-execution", wait=True
)

# Or use execution_name_prefix to avoid repeated execution names
execution = remote.execute(
     flyte_task, inputs={"n": 200, "mean": 0.0, "sigma": 1.0}, execution_name_prefix="flyte", wait=True
)

# Inspecting execution
# The 'inputs' and 'outputs' correspond to the task execution.
input_keys = execution.inputs.keys()
output_keys = execution.outputs.keys()
```
