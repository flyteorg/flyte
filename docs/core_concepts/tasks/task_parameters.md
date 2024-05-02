# Task parameters

```{note}

For more information on task parameters and methods, see the [flytekit documentation][flytekit].
```

You can pass the following parameters to the `@task` decorator:

## `accelerator`

The accelerator to use for this task. For more information, see [Specifying accelerators](https://docs.flyte.org/en/latest/api/flytekit/extras.accelerators.html#specifying-accelerators).

## `cache`

See <ref>`Task caching <task_caching>`.

## `cache_serialize`

See <ref>`Task caching <task_caching>`.

## `cache_version`

See <ref>`Task caching <task_caching>`.

## `cache_ignore_input_vars`

Input variables that should not be included when calculating the hash for the cache.

## `container_image`

See [Imagespec](task-software-environment/imagespec).

## `deprecated`

A string that can be used to provide a warning message for deprecated task. Absence / empty str indicates that the task is active and not deprecated

## `docs`

Documentation about this task.

## `enable_deck`

If true, this task will output a Flyte Deck which can be used to visualize the task execution (see [Decks&#x2B00;](https://docs.flyte.org/en/latest/user_guide/development_lifecycle/decks.html#id1)).

```{code-block} python
@task(enable_deck=True)
def my_task(my_str: str):
    print("hello {my_str}")
```

## `environment`

See [Environment variables](task-software-environment/environment-variables).

## `interruptible`

See [Interruptible instances](task-hardware-environment/interruptible-instances).

## `limits`

See [Customizing task resources](task-hardware-environment/customizing-task-resources).

## `node_dependency_hints`

A list of tasks, launch plans, or workflows that this task depends on. This is only for dynamic tasks/workflows, where Union cannot automatically determine the dependencies prior to runtime. Even on dynamic tasks this is optional, but in some scenarios it will make registering the workflow easier, because it allows registration to be done the same as for static tasks/workflows. For example this is useful to run launch plans dynamically, because launch plans must be registered on Flyteadmin before they can be run. Tasks and workflows do not have this requirement.

```{code-block} python
@workflow
def workflow0():
    launchplan0 = LaunchPlan.get_or_create(workflow0)
    # Specify node_dependency_hints so that launchplan0
    # will be registered on flyteadmin, despite this being a dynamic task.

@dynamic(node_dependency_hints=[launchplan0])
def launch_dynamically():
    # To run a sub-launchplan it must have previously been registered on flyteadmin.
    return [launchplan0]*10
```

## `pod_template`

See [Task hardware environment](task-hardware-environment/index.md#pod_template-and-pod_template_name-task-parameters).

## `pod_template_name`

See [Task hardware environment](task-hardware-environment/index.md#pod_template-and-pod_template_name-task-parameters).

## `requests`

See [Customizing task resources](task-hardware-environment/customizing-task-resources)

## `retries`

Number of times to retry this task during a workflow execution.
Tasks can define a retry strategy to let the system know how to handle
failures (For example: retry 3 times on any kind of error). For more information, see [Interruptible instances](task-hardware-environment/interruptible-instances)

There are two kinds of retries *system retries* and *user retries*.

## `secret_requests`

See [Secrets](task-software-environment/secrets)

## `task_config`

Configuration for a specific task type. See the [Union Agents documentation](../../integrations/agents/index)  and [Flyte plugins documentation](https://docs.flyte.org/en/latest/flytesnacks/integrations.html) for the right object to use.

## `task_resolver`

Provide a custom task resolver.

## `timeout`

The max amount of time for which one execution of this task should be executed for. The execution will be terminated if the runtime exceeds the given timeout (approximately). To ensure that the system is always making progress, tasks must be guaranteed to end gracefully/successfully. The system defines a default timeout period for the tasks. It is possible for task authors to define a timeout period, after which the task is marked as `failure`. Note that a timed-out task will be retried if it has a retry strategy defined. The timeout can be handled in the
[TaskMetadata](https://docs.flyte.org/projects/flytekit/en/latest/generated/flytekit.TaskMetadata.html?highlight=retries.md#flytekit.TaskMetadata).

[flytekit]: https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.task.html
