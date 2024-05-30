(failure_node)=
# Failure node

```{eval-rst}
 .. tags:: FailureNode, Intermediate
```

The failure node feature enables you to designate a specific node to execute in the event of a failure within your workflow.

For example, a workflow involves creating a cluster at the beginning, followed by the execution of tasks, and concludes with the deletion of the cluster once all tasks are completed. However, if any task within the workflow encounters an error, flyte will abort the entire workflow and won’t delete the cluster. This poses a challenge if you still need to clean up the cluster even in a task failure.

To address this issue, you can add a failure node into your workflow. This ensures that critical actions, such as deleting the cluster, are executed even in the event of failures occurring throughout the workflow execution

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:lines: 1-6
```

Create a task that will fail during execution:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:lines: 10-18
```

Create a task that will be executed if any of the tasks in the workflow fail:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:pyobject: clean_up
```

Specify the `on_failure` to a cleanup task. This task will be executed if any of the tasks in the workflow fail:

:::{note}
The input of `clean_up` should be the exact same as the input of the workflow.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:pyobject: subwf
```

By setting the failure policy to `FAIL_AFTER_EXECUTABLE_NODES_COMPLETE` to ensure that the `wf1` is executed even if the subworkflow fails. In this case, both parent and child workflows will fail, resulting in the `clean_up` task being executed twice:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:lines: 42-53
```

You can also set the `on_failure` to a workflow. This workflow will be executed if any of the tasks in the workflow fail:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/failure_node.py
:caption: development_lifecycle/failure_node.py
:pyobject: wf2
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/development_lifecycle/
