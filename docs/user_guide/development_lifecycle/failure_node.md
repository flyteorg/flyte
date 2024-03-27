(failure_node)=
# Failure node

```{eval-rst}
 .. tags:: FailureNode, Intermediate
```

The failure node feature enables you to designate a specific node to execute in the event of a failure within your workflow.

For example, a workflow involves creating a cluster at the beginning, followed by the execution of tasks, and concludes with the deletion of the cluster once all tasks are completed. However, if any task within the workflow encounters an error, flyte will abort the entire workflow and won’t delete the cluster. This poses a challenge if you still need to clean up the cluster even in a task failure.

To address this issue, you can add a failure node into your workflow. This ensures that critical actions, such as deleting the cluster, are executed even in the event of failures occurring throughout the workflow execution:

```python
from flytekit import WorkflowFailurePolicy, task, workflow


@task
def create_cluster(name: str):
    print(f"Creating cluster: {name}")

```

Create a task that will fail during execution:

```python
@task
def t1(a: int, b: str):
    print(f"{a} {b}")
    raise ValueError("Fail!")


@task
def delete_cluster(name: str):
    print(f"Deleting cluster {name}")
```

Create a task that will be executed if any of the tasks in the workflow fail:

```python
@task
def clean_up(name: str):
    print(f"Cleaning up cluster {name}")

```

Specify the `on_failure` to a cleanup task. This task will be executed if any of the tasks in the workflow fail:


:::{note}
The input of `clean_up` should be the exact same as the input of the workflow.
:::

```python
@workflow(on_failure=clean_up)
def subwf(name: str):
    c = create_cluster(name=name)
    t = t1(a=1, b="2")
    d = delete_cluster(name=name)
    c >> t >> d
```

By setting the failure policy to `FAIL_AFTER_EXECUTABLE_NODES_COMPLETE` to ensure that the `wf1` is executed even if the subworkflow fails. In this case, both parent and child workflows will fail, resulting in the `clean_up` task being executed twice:

```python
@workflow(on_failure=clean_up, failure_policy=WorkflowFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE)
def wf1(name: str = "my_cluster"):
    c = create_cluster(name=name)
    subwf(name="another_cluster")
    t = t1(a=1, b="2")
    d = delete_cluster(name=name)
    c >> t >> d


@workflow
def clean_up_wf(name: str):
    return clean_up(name=name)
```

You can also set the `on_failure` to a workflow. This workflow will be executed if any of the tasks in the workflow fail:

```python
@workflow(on_failure=clean_up_wf)
def wf2(name: str = "my_cluster"):
    c = create_cluster(name=name)
    t = t1(a=1, b="2")
    d = delete_cluster(name=name)
    c >> t >> d
```
