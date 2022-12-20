"""
Named Outputs
-------------

.. tags:: Basic

By default, Flyte names the outputs of a task or workflow using a standardized convention. All the outputs are named
as ``o1, o2, o3, ... o<n>.`` where ``o`` is the standard prefix and ``1, 2, .. <n>`` is the index position within the return values.

It is possible to give custom names to the outputs of a task or a workflow, so it's easier to refer to them while
debugging or visualising them in the UI. This is not possible natively in Python, so Flytekit provides an
alternative using ``typing.NamedTuple``.

The following example shows how to name outputs of a task and a workflow.
"""
import typing

from flytekit import task, workflow

# %%
# Named outputs can be declared inline as in the following task signature.
#
# .. note::
#
#  Note that the name of the NamedTuple does not matter, but the names and types of the variables do. We used a
#  a default name like ``OP``. NamedTuples can be inline, but by convention we prefer to declare them, as pypy
#  linter errors can be avoided this way.
#
#  .. code-block::
#
#     def say_hello() -> typing.NamedTuple("OP", greet=str):
#         pass
#
hello_output = typing.NamedTuple("OP", greet=str)


@task
def say_hello() -> hello_output:
    return hello_output("hello world")


# %%
# You can also declare the NamedTuple ahead of time and then use it in the signature as follows:
wf_outputs = typing.NamedTuple("OP2", greet1=str, greet2=str)


# %%
# As shown in this example, you can now refer to the declared NamedTuple.
# As seen in the workflow, ``say_hello`` returns a tuple. Like other tuples, you can simply unbox
# it inline. Also the workflow itself returns a tuple. You can also construct the tuple as you return the output.
#
# .. note::
#
#    Note that we are de-referencing the individual task execution outputs because named-outputs use NamedTuple
#    which are tuples that need to be de-referenced.


@workflow
def my_wf() -> wf_outputs:
    return wf_outputs(say_hello().greet, say_hello().greet)


# %%
# The workflow can be executed as usual.
if __name__ == "__main__":
    print(f"Running my_wf() {my_wf()}")
