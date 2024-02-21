"""
Dynamic Workflows
-----------------------------
Dynamic workflows are one of the powerful aspects of Flyte. Please take a look at the :py:func:`flytekit.dynamic` documentation first to get started.


Caveats when using a dynamic workflow
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Because of the dynamic nature of the workflow generated, it can easily be abused. Keep in mind that the workflow
that's compiled out of the decorated function needs to be processed like any other workflow. It's rare to see a
manually written workflow that has 5000 nodes for instance, but you can easily get there with a loop. Please keep,
dynamic workflows to under fifty tasks. For large-scale identical runs, we recommend the upcoming map task.

"""
import functools

from flytekit.core import task
from flytekit.core.python_function_task import PythonFunctionTask

dynamic = functools.partial(task.task, execution_mode=PythonFunctionTask.ExecutionBehavior.DYNAMIC)  # type: ignore[var-annotated]
dynamic.__doc__ = """
Please first see the comments for :py:func:`flytekit.task` and :py:func:`flytekit.workflow`. This ``dynamic``
concept is an amalgamation of both and enables the user to pursue some :std:ref:`pretty incredible <cookbook:advanced_merge_sort>`
constructs.

In short, a task's function is run at execution time only, and a workflow function is run at compilation time only (local
execution notwithstanding). A dynamic workflow is modeled on the backend as a task, but at execution time, the function
body is run to produce a workflow. It is almost as if the decorator changed from ``@task`` to ``@workflow`` except workflows
cannot make use of their inputs like native Python values whereas dynamic workflows can.
The resulting workflow is passed back to the Flyte engine and is
run as a :std:ref:`subworkflow <cookbook:subworkflows>`.  Simple usage

.. code-block::

    @dynamic
    def my_dynamic_subwf(a: int) -> (typing.List[str], int):
        s = []
        for i in range(a):
            s.append(t1(a=i))
        return s, 5

Note in the code block that we call the Python ``range`` operator on the input. This is typically not allowed in a
workflow but it is here. You can even express dependencies between tasks.

.. code-block::

    @dynamic
    def my_dynamic_subwf(a: int, b: int) -> int:
        x = t1(a=a)
        return t2(b=b, x=x)

See the :std:ref:`cookbook <cookbook:subworkflows>` for a longer discussion.
"""  # noqa: W293
