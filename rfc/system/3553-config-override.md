# Flyte Config Overrides RFC


https://github.com/flyteorg/flyte/issues/475

## Motivation
As a user of Flyte, configuration like "resources (cpu/mem/gpu)", "catalog (enable/disable/version)", "retries", "spark config", "hive cluster" are created only at registration time. The only way to update is to re-register. This is not desirable, as there are times when the user may want to update these parameters for specific launchplans, or executions.

Using `with_overrides` can provide node-level override for workflow, but this is still at compile time. Users need to use dynamic workflow if they want to pass the override values as inputs to the workflow. Moreover, with_overrides can not be applied to reference workflow because the workflow function body is not exposed.

## Design

By default, all task nodes are not overridable at runtime. However, users can call `.with_runtime_override(name: str)` on task nodes to register a "hook" on tasks, which allows the task to be identified by `name` and overridden with new config at runtime.

One main motivation for introducing such a "hook mechanism" based on identifiers/names is that we couldn't come up with a good answer how a good UX can be achieved when having to specify overrides in a nested workflow graph in the UI, in code, or via the CLI since all approaches would require replicating the workflow graph structure in the overrides config. This way, the overrides can be specified in a simple map and the structure of the workflow graph does not matter/does not have to be shown.

```python
@task
def t1():
  ...

@task
def t2():
  ...

@workflow
def wf():
  t1() # this task node cannot be overridden
  t1().with_runtime_override("task-yee") # can be overridden under the name "task-yee"
  t2().with_runtime_override("task-ketan") # can be overridden under the name "task-ketan"
  t3() # this task node cannot be overridden
```

We will have reuse proto `taskNodeConfigOverride`, which contains all the overridable fields.

```proto
// Optional task node overrides that will be applied at task execution time.
message TaskNodeOverrides {
    // A customizable interface to convey resources requested for a task container. 
    Resources resources = 1;

    // Boolean that indicates if caching should be enabled
    bool cache = 2;
    
    // Boolean that indicates if identical (ie. same inputs) instances of this task should be
    // executed in serial when caching is enabled.
    bool cache_serialize = 3;

    // Cache version to use
    string cache_version = 4;

    //  Number of times to retry this task during a workflow execution
    int32 retries = 5;

    // Boolean that indicates that this task can be interrupted and/or scheduled on nodes with lower QoS guarantees
    bool interruptible = 6;

    // Container image to use
    string container_image = 7;
    
    // Environment variables that should be added for this tasks execution
    map<string, string> environment = 9;

    // This argument provides configuration for a specific task types.
    google.protobuf.Struct task_config = 10;
}
```

We provide multiple ways to override the task nodes at runtime, including UI, workflow decorator, launch plan, etc.

### 1. UI

The registered task nodes will prompt on UI for users to assign config override values.

<img src="../images/config-override-ui.png" style="width: 50%; height: 50%">

### 2. Inside workflow

Users can override values inside workflow. The value will become the default values on UI.

```python
@workflow
def wf():
  t0().with_runtime_override("model_1_resources", runtime_override_config(cpu=1, mem="1Gi"))
  sub_wf().with_runtime_override(...).with_runtime_override(...).with_runtime_override(...)
```

### 3. launch plan

Users can also provide overrides to `LaunchPlan.get_or_create`. If users provide both workflow decorator and launch plan predefined values, launch plan one will override workflow decorator one.

```python
launch_plan.LaunchPlan.get_or_create(
    workflow=wf, 
    name="your_lp_name_5", 
    runtime_override={
      "task-yee": TaskNodeConfigOverride(...),
      "task-ketan": TaskNodeConfigOverride(...)
    }
)
```

## Concerns

1. How to resolve when two hooks have the same names?
- The flytekit compiler should error out