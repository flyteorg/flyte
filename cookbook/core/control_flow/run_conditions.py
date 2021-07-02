"""
Conditions
--------------
Flytekit supports conditions as a first class construct in the language. Conditions offer a way to selectively execute
branches of a workflow based on static or dynamic data produced by other tasks or come in as workflow inputs.
Conditions are very performant to be evaluated however, they are limited to certain binary and logical operators and can
only be performed on primitive values.
"""

# %%
# To start off, import `conditional` module
import random

from flytekit import conditional, task, workflow


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


if __name__ == "__main__":
    print(f"Output of multiplier(my_input=3): {multiplier(my_input=3)}")
    print(f"Output of multiplier(my_input=0.5): {multiplier(my_input=0.5)}")


# %%
# Example 2
# ^^^^^^^^^
# In the following example we have an if condition with multiple branches and we fail if no conditions are met. Flyte
# expects any conditional() statement to be _complete_ meaning all possible branches have to be handled.
#
# .. note::
#
#   Notice the use of bitwise (&). Python (PEP-335) does not allow overloading of Logical ``and, or, not`` operators. Flytekit uses bitwise `&` and `|` as logical and and or. This is a common practice in other libraries as well.
#
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


if __name__ == "__main__":
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


if __name__ == "__main__":
    print(f"Output of multiplier_3(my_input=5): {multiplier_3(my_input=5)}")


# %%
# Example 4
# ^^^^^^^^^^
# It is possible to test if a boolean returned from previous tasks is True or False. But, unary operations are not
# supported. Instead, please use the `is_true`, `is_false` or `is_` on the result.
#
# .. note::
#
#    Wondering how output values get these methods. In a workflow no output value is available to access directly. The inputs and outputs are auto-wrapped in a special object called :py:class:`flytekit.extend.Promise`.
#
@task
def coin_toss() -> bool:
    """
    Mimic some condition checking to see if something ran correctly
    """
    if random.random() < 0.5:
        return True
    return False


@task
def failed() -> int:
    """
    Mimic a task that handles a failure case
    """
    return -1


@task
def success() -> int:
    """
    Mimic a task that handles a success case
    """
    return 0


@workflow
def basic_boolean_wf() -> int:
    result = coin_toss()
    return (
        conditional("test").if_(result.is_true()).then(success()).else_().then(failed())
    )


# %%
# Example 5
# ^^^^^^^^^^
# It is also possible to pass a boolean directly to a workflow as follows
#
# .. note::
#
#   Note that the boolean passed in automagically has a method called `is_true`. This is because the boolean is within
#   the workflow context and hence is actually wrapped in a flytekit special object, which allows it to have additional
#   behavior
@workflow
def bool_input_wf(b: bool) -> int:
    return conditional("test").if_(b.is_true()).then(success()).else_().then(failed())


# %%
# these workflows can be locally executed
if __name__ == "__main__":
    print("Running basic_boolean_wf a few times")
    for i in range(0, 5):
        print(f"Basic boolean wf output {basic_boolean_wf()}")
        print(f"Boolean input {True if i < 2 else False}, workflow output {bool_input_wf(b=True if i < 2 else False)}")


# %%
# Example 6
# ^^^^^^^^^
# It is possible to arbitrarily nest conditional sections, inside other
# conditional sections. Remember - conditional sections can only be in the
# `then` part for a previous conditional block
# The follow example shows how you can use float comparisons to create
# a multi-level nested workflow
@workflow
def nested_conditions(my_input: float) -> float:
    return (
        conditional("fractions")
            .if_((my_input > 0.1) & (my_input < 1.0))
            .then(
            conditional("inner_fractions")
                .if_(my_input < 0.5)
                .then(double(n=my_input))
                .elif_((my_input > 0.5) & (my_input < 0.7))
                .then(square(n=my_input))
                .else_()
                .fail("Only <0.7 allowed")
        )
            .elif_((my_input > 1.0) & (my_input < 10.0))
            .then(square(n=my_input))
            .else_()
            .then(double(n=my_input))
    )


# %%
# As usual you can execute nested conditionals locally
if __name__ == "__main__":
    print(f"nested_conditions(0.4) -> {nested_conditions(my_input=0.4)}")


# %%
#  Example 7
# ^^^^^^^^^^^
# It is also possible to consume the outputs from conditional nodes as shown in the following example.
# In the case of conditionals though, the outputs are computed
# to be the subset of outputs that all then-nodes produce. In the following example, we call square() in one condition
# and call double in another.
@task
def sum_diff(a: float, b: float) -> (float, float):
    """
    sum_diff returns the sum and difference between a and b.
    """
    return a + b, a - b


# %%
# Putting it together, the workflow that consumes outputs from conditionals can be constructed as follows.
#
# .. tip::
#
#   A useful mental model for consuming outputs of conditions is to think of them like ternary operators in programming
#   languages. The only difference being they can be n-ary. In python this is equivalent to
#
#       .. code-block:: python
#
#           x = 0 if m < 0 else 1
@workflow
def consume_outputs(my_input: float) -> float:
    is_heads = coin_toss()
    res = (
        conditional("double_or_square")
            .if_(is_heads == True)
            .then(square(n=my_input))
            .else_()
            .then(sum_diff(a=my_input, b=my_input))
    )

    # Regardless of the result, always double before returning
    # the variable `res` in this case will carry the value of either square or double of the variable `my_input`
    return double(n=res)


# %%
# As usual local execution does not change
if __name__ == "__main__":
    print(f"consume_outputs(0.4) => {consume_outputs(my_input=0.4)}")
