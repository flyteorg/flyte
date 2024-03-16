import datetime
import os
import sys
import tempfile
import typing
from dataclasses import dataclass

import pytest
from dataclasses_json import DataClassJsonMixin

import flytekit
from flytekit import kwtypes
from flytekit.exceptions.user import FlyteRecoverableException
from flytekit.extras.tasks.shell import (
    OutputLocation,
    RawShellTask,
    ShellTask,
    get_raw_shell_task,
    subproc_execute,
)
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import CSVFile, FlyteFile

test_file_path = os.path.dirname(os.path.realpath(__file__))
testdata = os.path.join(test_file_path, "testdata")
test_csv = os.path.join(testdata, "test.csv")
if os.name == "nt":
    script_sh = os.path.join(testdata, "script.exe")
    script_sh_2 = None
else:
    script_sh = os.path.join(testdata, "script.sh")
    script_sh_2 = os.path.join(testdata, "script_args_env.sh")


def test_shell_task_no_io():
    t = ShellTask(
        name="test",
        script="""
        echo "Hello World!"
        """,
        shell="/bin/bash",
    )

    t()


def test_shell_task_fail():
    t = ShellTask(
        name="test",
        script="""
            non-existent blah
            """,
    )

    with pytest.raises(Exception):
        t()


def test_input_substitution_primitive():
    t = ShellTask(
        name="test",
        script="""
            set -ex
            cat {inputs.f}
            echo "Hello World {inputs.y} on  {inputs.j}"
            """,
        inputs=kwtypes(f=str, y=int, j=datetime.datetime),
    )

    if os.name == "nt":
        t._script = t._script.replace("set -ex", "").replace("cat", "type")

    t(
        f=os.path.join(test_file_path, "__init__.py"),
        y=5,
        j=datetime.datetime(2021, 11, 10, 12, 15, 0),
    )
    t(
        f=os.path.join(test_file_path, "test_shell.py"),
        y=5,
        j=datetime.datetime(2021, 11, 10, 12, 15, 0),
    )
    with pytest.raises(FlyteRecoverableException):
        t(f="non_exist.py", y=5, j=datetime.datetime(2021, 11, 10, 12, 15, 0))


def test_input_substitution_files():
    t = ShellTask(
        name="test",
        script="""
            cat {inputs.f}
            echo "Hello World {inputs.y} on  {inputs.j}"
            """,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
    )

    if os.name == "nt":
        t._script = t._script.replace("cat", "type")

    assert t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0)) is None


def test_input_substitution_files_ctx():
    sec = flytekit.current_context().secrets
    envvar = sec.get_secrets_env_var("group", "key")
    os.environ[envvar] = "value"
    assert sec.get("group", "key") == "value"

    t = ShellTask(
        name="test",
        script="""
            export EXEC={ctx.execution_id}
            export SECRET={ctx.secrets.group.key}
            cat {inputs.f}
            echo "Hello World {inputs.y} on  {inputs.j}"
            """,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
        debug=True,
    )

    if os.name == "nt":
        t._script = t._script.replace("cat", "type").replace("export", "set")

    assert t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0)) is None
    del os.environ[envvar]


def test_input_output_substitution_files():
    script = "cat {inputs.f} > {outputs.y}"
    if os.name == "nt":
        script = script.replace("cat", "type")
    t = ShellTask(
        name="test",
        debug=True,
        script=script,
        inputs=kwtypes(f=FlyteFile),
        output_locs=[
            OutputLocation(var="y", var_type=FlyteFile, location="{inputs.f}.mod"),
        ],
    )

    assert t.script == script

    contents = "1,2,3,4\n"
    with tempfile.TemporaryDirectory() as tmp:
        test_data = os.path.join(tmp, "abc.txt")
        with open(test_data, "w") as f:
            f.write(contents)
        y = t(f=test_data)
        assert y.path[-4:] == ".mod"
        assert os.path.exists(y.path)
        with open(y.path) as f:
            s = f.read()
        assert s == contents


def test_input_single_output_substitution_files():
    script = """
            cat {inputs.f} >> {outputs.z}
            echo "Hello World {inputs.y} on  {inputs.j}"
            """
    if os.name == "nt":
        script = script.replace("cat", "type")

    t = ShellTask(
        name="test",
        debug=True,
        script=script,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
        output_locs=[OutputLocation(var="z", var_type=FlyteFile, location="{inputs.f}.pyc")],
    )

    assert t.script == script
    y = t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0))
    assert y.path[-4:] == ".pyc"


@pytest.mark.parametrize(
    "script",
    [
        (
            """
            cat {missing} >> {outputs.z}
            echo "Hello World {inputs.y} on  {inputs.j} - output {outputs.x}"
            """
        ),
        (
            """
            cat {inputs.f} {missing} >> {outputs.z}
            echo "Hello World {inputs.y} on  {inputs.j} - output {outputs.x}"
            """
        ),
    ],
)
def test_input_output_extra_and_missing_variables(script):
    if os.name == "nt":
        script = script.replace("cat", "type")
    t = ShellTask(
        name="test",
        debug=True,
        script=script,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
        output_locs=[
            OutputLocation(var="x", var_type=FlyteDirectory, location="{inputs.y}"),
            OutputLocation(var="z", var_type=FlyteFile, location="{inputs.f}.pyc"),
        ],
    )

    with pytest.raises(ValueError, match="missing"):
        t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0))


def test_reuse_variables_for_both_inputs_and_outputs():
    t = ShellTask(
        name="test",
        debug=True,
        script="""
        cat {inputs.f} >> {outputs.y}
        echo "Hello World {inputs.y} on  {inputs.j}"
        """,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
        output_locs=[
            OutputLocation(var="y", var_type=FlyteFile, location="{inputs.f}.pyc"),
        ],
    )
    if os.name == "nt":
        t._script = t._script.replace("cat", "type")

    t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0))


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_can_use_complex_types_for_inputs_to_f_string_template():
    @dataclass
    class InputArgs(DataClassJsonMixin):
        in_file: CSVFile

    t = ShellTask(
        name="test",
        debug=True,
        script="""cat {inputs.input_args.in_file} >> {inputs.input_args.in_file}.tmp""",
        inputs=kwtypes(input_args=InputArgs),
        output_locs=[
            OutputLocation(var="x", var_type=FlyteFile, location="{inputs.input_args.in_file}.tmp"),
        ],
    )

    if os.name == "nt":
        t._script = t._script.replace("cat", "type")

    input_args = InputArgs(FlyteFile(path=test_csv))
    x = t(input_args=input_args)
    assert x.path[-4:] == ".tmp"


def test_shell_script():
    t = ShellTask(
        name="test2",
        debug=True,
        script_file=script_sh,
        inputs=kwtypes(f=CSVFile, y=FlyteDirectory, j=datetime.datetime),
        output_locs=[
            OutputLocation(var="x", var_type=FlyteDirectory, location="{inputs.y}"),
            OutputLocation(var="z", var_type=FlyteFile, location="{inputs.f}.pyc"),
        ],
    )

    assert t.script_file == script_sh
    t(f=test_csv, y=testdata, j=datetime.datetime(2021, 11, 10, 12, 15, 0))


def test_raw_shell_task_with_args(capfd):
    if script_sh_2 is None:
        return
    pst = get_raw_shell_task(name="test")
    pst(script_file=script_sh_2, script_args="first_arg second_arg", env={})
    cap = capfd.readouterr()
    assert "first_arg" in cap.out
    assert "second_arg" in cap.out


def test_raw_shell_task_with_env(capfd):
    if script_sh_2 is None:
        return
    pst = get_raw_shell_task(name="test")
    pst(script_file=script_sh_2, env={"A": "AAAA", "B": "BBBB"}, script_args="")
    cap = capfd.readouterr()
    assert "AAAA" in cap.out
    assert "BBBB" in cap.out


def test_raw_shell_task_properly_restores_env_after_execution():
    if script_sh_2 is None:
        return
    env_as_dict = os.environ.copy()
    pst = get_raw_shell_task(name="test")
    pst(script_file=script_sh_2, env={"A": "AAAA", "B": "BBBB"}, script_args="")
    env_as_dict_after = os.environ.copy()
    assert env_as_dict == env_as_dict_after


def test_raw_shell_task_instantiation(capfd):
    if script_sh_2 is None:
        return
    pst = RawShellTask(
        name="test",
        debug=True,
        inputs=flytekit.kwtypes(env=typing.Dict[str, str], script_args=str, script_file=str),
        output_locs=[
            OutputLocation(
                var="out",
                var_type=FlyteDirectory,
                location="{ctx.working_directory}",
            )
        ],
        script="""
#!/bin/bash

set -uex

cd {ctx.working_directory}

{inputs.export_env}

bash {inputs.script_file} {inputs.script_args}
""",
    )
    pst(script_file=script_sh_2, script_args="first_arg second_arg", env={})
    cap = capfd.readouterr()
    assert "first_arg" in cap.out
    assert "second_arg" in cap.out


@pytest.mark.timeout(20)
def test_long_run_script():
    script = os.path.join(testdata, "long-running.sh")
    ShellTask(
        name="long-running",
        script=script,
    )()


def test_subproc_execute():
    cmd = ["echo", "hello"]
    o, e = subproc_execute(cmd)
    assert o == "hello\n"
    assert e == ""


def test_subproc_execute_with_shell():
    with tempfile.TemporaryDirectory() as tmp:
        opth = os.path.join(tmp, "test.txt")
        cmd = f"echo hello > {opth}"
        subproc_execute(cmd, shell=True)
        cont = open(opth).read()
        assert "hello" in cont
