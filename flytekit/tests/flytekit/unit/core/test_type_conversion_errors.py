"""Unit tests for type conversion errors."""

from datetime import timedelta
from string import ascii_lowercase
from typing import Tuple

import pytest
from hypothesis import given, settings
from hypothesis import strategies as st

from flytekit import task, workflow


@task
def int_to_float(n: int) -> float:
    return float(n)


@task
def task_incorrect_output(a: float) -> int:
    return str(a)  # type: ignore [return-value]


@task
def task_correct_output(a: float) -> str:
    return str(a)


@workflow
def wf_with_task_error(a: int) -> str:
    return task_incorrect_output(a=int_to_float(n=a))


@workflow
def wf_with_output_error(a: int) -> int:
    return task_correct_output(a=int_to_float(n=a))


@workflow
def wf_with_multioutput_error0(a: int, b: int) -> Tuple[int, str]:
    out_a = task_correct_output(a=int_to_float(n=a))
    out_b = task_correct_output(a=int_to_float(n=b))
    return out_a, out_b


@workflow
def wf_with_multioutput_error1(a: int, b: int) -> Tuple[str, int]:
    out_a = task_correct_output(a=int_to_float(n=a))
    out_b = task_correct_output(a=int_to_float(n=b))
    return out_a, out_b


@given(st.booleans() | st.integers() | st.text(ascii_lowercase))
@settings(deadline=timedelta(seconds=2))
def test_task_input_error(incorrect_input):
    with pytest.raises(
        TypeError,
        match=(
            r"Failed to convert inputs of task '{}':\n"
            r"  Failed argument 'a': Expected value of type \<class 'float'\> but got .+ of type .+"
        ).format(task_correct_output.name),
    ):
        task_correct_output(a=incorrect_input)


@given(st.floats())
@settings(deadline=timedelta(seconds=2))
def test_task_output_error(correct_input):
    with pytest.raises(
        TypeError,
        match=(
            r"Failed to convert outputs of task '{}' at position 0:\n"
            r"  Expected value of type \<class 'int'\> but got .+ of type .+"
        ).format(task_incorrect_output.name),
    ):
        task_incorrect_output(a=correct_input)


@given(st.integers())
@settings(deadline=timedelta(seconds=2))
def test_workflow_with_task_error(correct_input):
    with pytest.raises(
        TypeError,
        match=(
            r"Error encountered while executing 'wf_with_task_error':\n"
            r"  Failed to convert outputs of task '.+' at position 0:\n"
            r"  Expected value of type \<class 'int'\> but got .+ of type .+"
        ).format(),
    ):
        wf_with_task_error(a=correct_input)


@given(st.booleans() | st.floats() | st.text(ascii_lowercase))
@settings(deadline=timedelta(seconds=2))
def test_workflow_with_input_error(incorrect_input):
    with pytest.raises(
        TypeError,
        match=r"Failed argument".format(),
    ):
        wf_with_output_error(a=incorrect_input)


@given(st.integers())
@settings(deadline=timedelta(seconds=2))
def test_workflow_with_output_error(correct_input):
    with pytest.raises(
        TypeError,
        match=(r"Failed to convert output in position 0 of value .+, expected type \<class 'int'\>"),
    ):
        wf_with_output_error(a=correct_input)


@pytest.mark.parametrize(
    "workflow, position",
    [
        (wf_with_multioutput_error0, 0),
        (wf_with_multioutput_error1, 1),
    ],
)
@given(st.integers())
@settings(deadline=timedelta(seconds=2))
def test_workflow_with_multioutput_error(workflow, position, correct_input):
    with pytest.raises(
        TypeError,
        match=(r"Failed to convert output in position {} of value .+, expected type \<class 'int'\>").format(position),
    ):
        workflow(a=correct_input, b=correct_input)
