"""
08: Conditions
--------------

Flytekit supports conditions as a first class construct in the language. Conditions offer a way to selectively execute
branches of a workflow based on static or dynamic data produced by other tasks or come in as workflow inputs.
Conditions are very performant to be evaluated however, they are limited to certain binary and logical operators and can
only be performed on primitive values.
"""

# %%
# To start off, import `conditional` module
import typing

from flytekit import task, workflow
from flytekit.annotated.condition import conditional


# %%
# Example 1
# ^^^^^^^^^
# In the following example we define two tasks `square` and `double` and depending on whether the workflow input is a
# fraction (0-1) or not, it decided which to execute.
@task
def square(n: float) -> float:
    """
    Parameters:
        n (float): name of the parameter for the task will be derived from the name of the input variable
               the type will be automatically deduced to be Types.Integer

    Return:
        float: The label for the output will be automatically assigned and type will be deduced from the annotation

    """
    return n * n


@task
def double(n: float) -> float:
    """
    Parameters:
        n (float): name of the parameter for the task will be derived from the name of the input variable
               the type will be automatically deduced to be Types.Integer

    Return:
        float: The label for the output will be automatically assigned and type will be deduced from the annotation

    """
    return 2 * n


@workflow
def multiplier(my_input: float) -> float:
    return (
        conditional("fractions")
        .if_((my_input >= 0.1) & (my_input <= 1.0))
        .then(double(n=my_input))
        .else_()
        .then(square(n=my_input))
    )


print(f"Output of multiplier(my_input=3): {multiplier(my_input=3)}")
print(f"Output of multiplier(my_input=0.5): {multiplier(my_input=0.5)}")


# %%
# Example 2
# ^^^^^^^^^
# In the following example we have an if condition with multiple branches and we fail if no conditions are met. Flyte
# expects any conditional() statement to be _complete_ meaning all possible branches have to be handled.
@workflow
def multiplier_2(my_input: float) -> float:
    return (
        conditional("fractions")
        .if_((my_input > 0.1) & (my_input < 1.0))
        .then(double(n=my_input))
        .elif_((my_input > 1.0) & (my_input <= 10.0))
        .then(square(n=my_input))
        .else_()
        .fail("The input must be between 0 and 10")
    )


print(f"Output of multiplier_2(my_input=10): {multiplier_2(my_input=10)}")


# %%
# Example 3
# ^^^^^^^^^
# In the following example we consume the output returned by the conditional() in a subsequent task.
@workflow
def multiplier_3(my_input: float) -> float:
    d = (
        conditional("fractions")
        .if_((my_input > 0.1) & (my_input < 1.0))
        .then(double(n=my_input))
        .elif_((my_input > 1.0) & (my_input < 10.0))
        .then(square(n=my_input))
        .else_()
        .fail("The input must be between 0 and 10")
    )

    # d will be either the output of `double` or t he output of `square`. If the conditional() falls through the fail
    # branch, execution will not reach here.
    return double(n=d)


print(f"Output of multiplier_3(my_input=5): {multiplier_3(my_input=5)}")
