"""
.. _flyte_pickle:

Using Flyte Pickle
----------------------------

Flyte enforces type safety by leveraging type information to be able to compile
tasks/workflows, which enables all sorts of nice features (like static analysis of tasks/workflows, conditional branching, etc.)

However, we do also want to provide enough flexibility to end-users so that they don't have to put in a lot of up front
investment figuring out all the types of their data structures before experiencing the value that flyte has to offer.

Flyte supports FlytePickle transformer which will convert any unrecognized type in type hint to
FlytePickle, and serialize / deserialize the python value to / from a pickle file.

Caveats
=======
Pickle can only be used to send objects between the exact same version of Python,
and we strongly recommend to use python type that flyte support or register a custom transformer

This example shows how users can custom object without register a transformer.
"""
from flytekit import task, workflow


# %%
# This People is a user defined complex type, which can be used to pass complex data between tasks.
# We will serialize this class to a pickle file and pass it between different tasks.
#
# .. Note::
#
#   Here we can also `turn this object to dataclass <custom_objects.html>`_ to have better performance.
#   We use simple object here for demo purpose.
#   You may have some object that can't turn into a dataclass, e.g. NumPy, Tensor.
class People:
    def __init__(self, name):
        self.name = name


# %%
# Object can be returned as outputs or accepted as inputs
@task
def greet(name: str) -> People:
    return People(name)


@workflow
def welcome(name: str) -> People:
    return greet(name=name)


if __name__ == "__main__":
    """
    This workflow can be run locally. During local execution also,
    the custom object (People) will be marshalled to and from python pickle.
    """
    welcome(name='Foo')
