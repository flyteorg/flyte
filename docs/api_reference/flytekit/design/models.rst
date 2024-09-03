.. _design-models:

###########
Model Files
###########

.. tags:: Design, Basic

***********
Description
***********
This section deals with the files in `models <https://github.com/flyteorg/flytekit/tree/master/flytekit/models>`__ folder.
These files are better formatted versions of the generated Python code from `Flyte IDL <https://github.com/flyteorg/flyteidl>`__. In the future, we hope to be able to improve the Python code generator sufficiently to avoid this manual work.

The __only__ reason these files exist is only because the Protobuf generated Python code doesn't work well with IDEs. It doesn't
offer code completion, doesn't offer argument completion, docstrings, etc.

The structure of the code in the models folder should mirror the folder structure of the IDL. There are a few instances
where this is not the case but this is incidental and we'll work on resolving these.

*********
Structure
*********

It may be helpful to take a look at an example like the ``TaskTemplate`` `here <https://github.com/flyteorg/flytekit/blob/b6f81d3724787640db6ef99ecfddcdab074d2a83/flytekit/models/task.py#L293>`__

Constructor
===========
The constructor should allow specification of all the elements of the Protobuf object. They should be stored as an attribute with a leading underscore.

Properties
==========
Each element of the IDL message should be exposed as a property.

IDL Interaction
===============
Each Python object should have a ``to_flyte_idl`` and a ``from_flyte_idl`` function that converts between the Python model class and the Protobuf-generated Python class.

*********
Testing
*********
Please add unit tests to the ``tests/flytekit/unit/models`` folder to test the conversion logic.
