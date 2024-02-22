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

(remote_launchplan)=

# Running launch plans

## Flytectl

This is multi-steps process where we create an execution spec file, update the spec file and then create the execution.
More details can be found [here](https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html).

**Generate an execution spec file**

```
flytectl get launchplan -p flytesnacks -d development myapp.workflows.example.my_wf  --execFile exec_spec.yaml
```

**Update the input spec file for arguments to the workflow**

```
....
inputs:
    name: "adam"
....
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

A launch plan can be launched via FlyteRemote programmatically.

```python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config
from flytekit import LaunchPlan

# FlyteRemote object is the main entrypoint to API
remote = FlyteRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)

# Fetch launch plan
flyte_lp = remote.fetch_launch_plan(
    name="workflows.example.wf", version="v1", project="flytesnacks", domain="development"
)

# Execute
execution = remote.execute(
    flyte_lp, inputs={"mean": 1}, execution_name="lp-execution", wait=True
)

# Or use execution_name_prefix to avoid repeated execution names
execution = remote.execute(
    flyte_lp, inputs={"mean": 1}, execution_name_prefix="flyte", wait=True
)
```
