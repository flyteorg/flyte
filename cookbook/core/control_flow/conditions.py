"""
Conditions
----------

.. tags:: Intermediate

Flytekit supports conditions as a first class construct in the language. Conditions offer a way to selectively execute
branches of a workflow based on static or dynamic data produced by other tasks or come in as workflow inputs.
Conditions are very performant to be evaluated. However, they are limited to certain binary and logical operators and can
only be performed on primitive values.
"""

# %%
# Import the necessary modules.
import random

from flytekit import conditional, task, workflow


# %%
# Example 1
# ^^^^^^^^^
# In this example, we define two tasks `square` and `double`. Depending on whether the workflow input is a
# fraction (0-1) or not, the respective task is executed.
@task
def square(n: float) -> float:
    """
    Parameters:
        n (float): name of the parameter for the task is derived from the name of the input variable, and
               the type is automatically mapped to Types.Integer
    Return:
        float: The label for the output is automatically assigned and the type is deduced from the annotation
    """
    return n * n


@task
def double(n: float) -> float:
    """
    Parameters:
        n (float): name of the parameter for the task is derived from the name of the input variable
               and the type is mapped to ``Types.Integer``
    Return:
        float: The label for the output is auto-assigned and the type is deduced from the annotation
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
    print(f"Output of multiplier(my_input=3.0): {multiplier(my_input=3.0)}")
    print(f"Output of multiplier(my_input=0.5): {multiplier(my_input=0.5)}")


# %%
# Example 2
# ^^^^^^^^^
# In this example, we define an ``if`` condition with multiple branches. It fails if none of the conditions is met. Flyte
# expects any ``conditional()`` statement to be **complete**. This means all possible branches should be handled.
#
# .. note::
#
#   Notice the use of bitwise (&). Python (PEP-335) doesn't allow overloading of the logical ``and``, ``or``, and ``not`` operators. Flytekit uses bitwise `&` and `|` as logical ``and`` and ``or`` operators. This is a common practice in other libraries too.
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
    print(f"Output of multiplier_2(my_input=10.0): {multiplier_2(my_input=10.0)}")


# %%
# Example 3
# ^^^^^^^^^
# In this example, we consume the output returned by the ``conditional()`` in the subsequent task.
@workflow
def multiplier_3(my_input: float) -> float:
    result = (
        conditional("fractions")
        .if_((my_input > 0.1) & (my_input < 1.0))
        .then(double(n=my_input))
        .elif_((my_input > 1.0) & (my_input < 10.0))
        .then(square(n=my_input))
        .else_()
        .fail("The input must be between 0 and 10")
    )

    # the 'result' will either be the output of `double` or `square`. If none of the conditions is true,
    # it gives a failure message.
    return double(n=result)


if __name__ == "__main__":
    print(f"Output of multiplier_3(my_input=5.0): {multiplier_3(my_input=5.0)}")


# %%
# Example 4
# ^^^^^^^^^^
# It is possible to test if a boolean returned from the previous task is True. But unary operations are not
# supported. Use the `is_true`, `is_false` or `is_` on the result instead.
#
# .. note::
#
#    How do output values get these methods?
#    In a workflow, no output can be accessed directly. The inputs and outputs are auto-wrapped in a special object called :py:class:`flytekit.extend.Promise`.
#
# In this example, we create a biased coin whose seed can be controlled.
@task
def coin_toss(seed: int) -> bool:
    """
    Mimic some condition to check if the operation was successfully executed.
    """
    r = random.Random(seed)
    if r.random() < 0.5:
        return True
    return False


@task
def failed() -> int:
    """
    Mimic a task that handles failure
    """
    return -1


@task
def success() -> int:
    """
    Mimic a task that handles success
    """
    return 0


@workflow
def basic_boolean_wf(seed: int = 5) -> int:
    result = coin_toss(seed=seed)
    return (
        conditional("test").if_(result.is_true()).then(success()).else_().then(failed())
    )


# %%
# Example 5
# ^^^^^^^^^^
# It is possible to pass a boolean directly to a workflow.
#
# .. note::
#
#   Note that the boolean passed has a method named `is_true`. This boolean is present within
#   the workflow context and is wrapped in a Flytekit special object. This special object allows it to have the additional
#   behavior.
@workflow
def bool_input_wf(b: bool) -> int:
    return conditional("test").if_(b.is_true()).then(success()).else_().then(failed())


# %%
# The workflow can be executed locally.
if __name__ == "__main__":
    print("Running basic_boolean_wf a few times")
    for i in range(0, 5):
        print(f"Basic boolean wf output {basic_boolean_wf()}")
        print(
            f"Boolean input {True if i < 2 else False}, workflow output {bool_input_wf(b=True if i < 2 else False)}"
        )


# %%
# Example 6
# ^^^^^^^^^
# It is possible to arbitrarily nest conditional sections inside other conditional sections. The conditional sections can only be in the
# ``then`` part of the previous conditional block.
# This example shows how float comparisons can be used to create a multi-level nested workflow.
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
# The nested conditionals can be executed locally.
if __name__ == "__main__":
    print(f"nested_conditions(0.4) -> {nested_conditions(my_input=0.4)}")


# %%
# Example 7
# ^^^^^^^^^^
# It is possible to consume the outputs from conditional nodes.
# In the case of conditionals, the outputs are computed
# as a subset of outputs produced by ``then`` nodes. In this example, we call ``square()`` in one condition
# and ``double()`` in another.
@task
def calc_sum(a: float, b: float) -> float:
    """
    returns the sum of a and b.
    """
    return a + b


# %%
# Altogether, the workflow that consumes outputs from conditionals can be constructed as shown.
#
# .. tip::
#
#   A useful mental model to consume outputs of conditions is to think of them as ternary operators in programming
#   languages. The only difference is that they can be n-ary. In Python, this is equivalent to
#
#   .. code-block:: python
#
#      x = 0 if m < 0 else 1
@workflow
def consume_outputs(my_input: float, seed: int = 5) -> float:
    is_heads = coin_toss(seed=seed)
    res = (
        conditional("double_or_square")
        .if_(is_heads.is_true())
        .then(square(n=my_input))
        .else_()
        .then(calc_sum(a=my_input, b=my_input))
    )

    # Regardless of the result, call ``double`` before
    # the variable `res` is returned. In this case, ``res`` stores the value of the ``square`` or ``double`` of the variable `my_input`
    return double(n=res)


# %%
# The workflow can be executed locally.
if __name__ == "__main__":
    print(
        f"consume_outputs(0.4) with default seed=5. This should return output of calc_sum => {consume_outputs(my_input=0.4)}"
    )
    print(
        f"consume_outputs(0.4, seed=7), this should return output of square => {consume_outputs(my_input=0.4, seed=7)}"
    )
