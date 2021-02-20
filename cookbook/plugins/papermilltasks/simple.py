"""
Execute a Python Script as a Jupyter notebook
-----------------------------------------------
In this example, we will show how to create a flyte task, that runs a simple notebook, that accepts one input variable, transforms it and produces
one output. This can be generalized to multiple inputs and outputs.

"""
import math
import os
import pathlib

from flytekit import kwtypes, task, workflow
from flytekitplugins.papermill import NotebookTask

#%%
# How to specify inputs and outputs
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# #. After you are satisfied with the notebook, ensure that the first cell is just the input variables for the notebook. Now add a tag ``parameters`` for the first cell.
#
# .. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/papermilltasks/parameters.png
#     :alt: Example of "parameters tag" added to the cell with input variables
#
# #. Now usually as the last cell of the notebook (does not need to be the last cell), add a tag ``outputs`` for the intended cell.
#
# .. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/papermilltasks/outputs.png
#     :alt: Example of "parameters tag" added to the cell with input variables
#
# #. Now in a python file, you have to create a new task at the ``module`` level. Refer to :py:class:`NotebookTask`
#    The example task is shown below
nb = NotebookTask(
    name="simple-nb",
    notebook_path=os.path.join(pathlib.Path(__file__).parent.absolute(), "nb-simple.ipynb"),
    inputs=kwtypes(v=float),
    outputs=kwtypes(square=float),
)
#%%
# .. note::
#
#  - Note the notebook_path. It is the absolute path to the actual notebook.
#  - Note the inputs and outputs. The variable names match the variable names in the jupyter notebook.
#
# Other tasks
# ^^^^^^^^^^^^^^^
# You can ofcourse declare other tasks and seamlessly work with notebook tasks, let us declare a task that accepts the squared value from the notebook and provides a sqrt
@task
def square_root_task(f: float) -> float:
    return math.sqrt(f)

#%%
# You can now treat the notebook task as a regular task
@workflow
def nb_to_python_wf(f: float) -> float:
    out = nb(v=f)
    return square_root_task(f=out.square)

#%%
# You can execute the task locally as well
if __name__ == "__main__":
    print(nb_to_python_wf(f=3.14))

#%%
# 3 outputs - Why?
# ^^^^^^^^^^^^^^^^^^
# On executing you should see 3 outputs instead of the expected one. This is because this task generates 2 implicit outputs.
# One of them is the executed notebook (captured) and a rendered (HTML) of the executed notebook. In this case they are called
# ``nb-simple-out.ipynb`` and ``nb-simple-out.html`` respectively