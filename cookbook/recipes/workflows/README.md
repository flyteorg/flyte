# Writing A Workflow

Workflows are the next layer of building blocks in Flyte after tasks. After you have multiple tasks you will likely want to assemble them into a graph. Please see the [workflows](https://lyft.github.io/flyte/user/concepts/workflows_nodes.html) documentation as well.

Continuing with the examples from the tasks chapter, let's say we have two tasks

* `def scale(wf_params, image, scale_factor, out_image)`, and
* `def rotate(wf_params, image, angle, fail, out_image)`

that we wish to put together. Workflows are declared using Python classes, instead of functions, the static members of which are analyzed to ascertain task structure.
 

```python
from flytekit.sdk.workflow import workflow_class, Input, Output

@workflow_class
class ScaleAndRotateWorkflow(object):
    ...
```

Please see the [workflows.py](workflows.py) companion file for the full code. Another difference worth pointing out is that inputs and outputs to workflows are declared as static workflow objects.

This for example,
```python
    angle = Input(Types.Float, default=180.0)
    scale = Input(Types.Integer, default=2)
    ...
    out_image = Output(rotate_task.outputs.out_image, sdk_type=Types.Blob)
```

declares an input named `angle` that is a float with a default of 180, and an integer called `scale` with a default of 2. An output, derived from the output named `out_image` of the `rotate_task` task is also declared. 

One important thing to note - logic that is put into the class declaration is only run at compile time, not at run time. We encourage users to not have logic in their workflow declaration, at least initially, as it may lead to confusion.

To execute a workflow, we should first introduce the concept of [Launch Plans](../launchplans/README.md). 
 