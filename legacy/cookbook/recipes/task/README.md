[Back to Cookbook Menu](../../)

# Writing Tasks

Tasks are the fundamental building blocks in Flyte. To get started, please review the main Flyte docs on [tasks](https://lyft.github.io/flyte/user/concepts/tasks.html) and the various [task types](https://lyft.github.io/flyte/user/tasktypes/index.html) that Flyte currently has.

The most common tasks used today are Python tasks so we should start there. Python tasks begin their lives as functions. Let's say we have a function called `def rotate` that rotates an image.

```python
def rotate(image, angle, out_image):
    ...
```

In order to hook it up to the Flyte system, we have to tell Flyte that we want to make this function a Task via a decorator.

```python
from flytekit.sdk.tasks import python_task
@python_task(cache=True, cache_version="1")
def rotate(wf_params, image, angle, out_image):
    ...
```

Note that this added one more argument in the front of the function. You can think of this as an object that provides some utilities like logging and stats (more on this later). We've also added a caching argument that tells Flyte to [memoize](https://lyft.github.io/flyte/user/concepts/tasks.html#memoization) the result.

This is not enough however, we also need to tell Flyte what the inputs and outputs are. To do this, we can add some more decorators. 

```python
from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types

@inputs(image=Types.Blob, angle=Types.Float)
@outputs(out_image=Types.Blob)
@python_task(cache=True, cache_version="1")
def rotate(wf_params, image, angle, out_image):
    ...
```

Please see the examples in [tasks.py](tasks.py) for more details.
