[Back to Cookbook Menu](../..)
# Containers and Tasks

Every task in Flyte can be associated with a container (This is not required). Every task in workflow can also have a **different container** - YES!

## Why this is not easily supported in flytekit
This is not easily supported in Flytekit, because Flytekit has a Workflow DSL, which allows importing tasks into a python module. So lets take an example
```python
# Let us assume we have a file called tasks.py, which has one task that performs a function on tensorflow.
import tensorflow as tf

@python_task
def tensorflow_train(wf_params):
  tf.x(...)   
```

Now consider a different module that builds the workflow using *@workflow_class* decorator.

```python
from tasks import tensorflow_train

@workflow_class
class Test(obj):
 ...
 t = tensorflow_train()
 ...

```
This module will fail during registration time, as the python module will be loaded and as the imports are loaded, *import tensorflow as tf* will fail unless tensorflow is installed in the current
modules environment. Thus if one wanted to have a workflow in a module with a container and a task in another module with a different container, then it automatically means (in most cases, unless
imports are postponed to runtime) that we will need the top level module to have all the dependencies installed. This defeats the purpose of multiple containers.


## Option I: Raw Container tasks
One solution is to use Raw Container Tasks. Flyte allows some tasks to be raw containers, some to be regular python tasks. 
**TODO Examples**

## Option II: Remote task fetch
One solution is to fetch a pre-registered task.
**TODO Examples**

## Option III: intelligent importing? :construction:
The issue of unwanted imports can be resolved by changing the task declaration

```python
@python_task
def tensorflow_train(wf_params):
   import tensorflow as tf
   ....
```

But, Currently Flytekit has no intelligent way of associating a container with the task. **Keep a tab on this page**

