"""
Using Enum types
----------------

.. tags:: Basic

Sometimes you may want to restrict the set of inputs / outputs to a finite set of acceptable values. This is commonly
achieved using Enum types in programming languages.

Since version 0.15.0, Flyte supports Enum Types with string values. You can create a python Enum type and pass it to
a task or return from a task. Flyte will automatically convert this to and limit the inputs etc to a finite set of
values.

UX: flytectl will allow only the finite set of values to be acceptable and (in progress) UI will provide a drop-down for
the values.

**Caveat:** Only string values are supported as valid enum values. The first value in the list is assumed as the default
and the Enum types are not optional. So when defining enums, design them well to always make the first value as a
valid default.
"""
from enum import Enum
from typing import Tuple

from flytekit import task, workflow


# %%
# Enums are natively supported in flyte's type system. Enum values can only be of type string. At runtime they are
# represented using their string values.
#
# .. note::
#
#   ENUM Values can only be string. Other languages will receive enum as a string.
#
class Color(Enum):
    RED = "red"
    GREEN = "green"
    BLUE = "blue"


# %%
# Enums can be used as a regular type
@task
def enum_stringify(c: Color) -> str:
    return c.value


# %%
# Their values can be accepted as string
@task
def string_to_enum(c: str) -> Color:
    return Color(c)


@workflow
def enum_wf(c: Color = Color.RED) -> Tuple[Color, str]:
    v = enum_stringify(c=c)
    return string_to_enum(c=v), v


if __name__ == "__main__":
    print(enum_wf())
