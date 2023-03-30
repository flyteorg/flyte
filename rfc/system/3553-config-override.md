# Flyte Config Overrides RFC


https://github.com/flyteorg/flyte/issues/475

## Motivation
As a user of Flyte, configuration like "resources (cpu/mem/gpu)", "catalog (enable/disable/version)", "retries", "spark config", "hive cluster" are created only at registration time. The only way to update is to re-register. This is not desirable, as there are times when the user may want to update these parameters for specific launchplans, or executions.

Using with_overrides can provide node-level override for workflow, but this is still at compile time. Users need to use dynamic workflow if they want to pass the override values as inputs to the workflow. Moreover, with_overrides can not be applied to reference workflow because the workflow function body is not exposed.

## Problem Statement
Allow users to override task config at execution time without re-registering the workflows.
- Support overriding config on both flyteconsole and CLI.  

## API Design
- flyteremote

```python
wf_overrides = WFOverride(
    n0=TaskOverride(
       container_image="repo/image:0.0.1"
       limits=Resource(cpu="2")
    ),
    n1=WFOverride( # subwf
        n0=TaskOverride(
            container_image="repo/image:0.0.3"
            limits=Resource(cpu="3")
        )
    )
)

flyteremote.execute(wf, wf_overrides)
```

- flytectl

```python
$ flytectl create execution --execFile execution_spec.yaml
workflow: core.control_flow.merge_sort.merge
wfOverride:
  n0:
    container_image: "repo/image:0.0.1"
    limits:
      cpu: 2
  n1:
    n0:
      container_image: "repo/image:0.0.3"
        limits:
          cpu: 3
```

## UI design

The hardest part. Need suggestion.

## Implementation 

WIP
