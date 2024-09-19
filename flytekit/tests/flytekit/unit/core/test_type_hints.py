import dataclasses
import datetime
import functools
import os
import random
import re
import sys
import tempfile
import typing
from collections import OrderedDict
from dataclasses import dataclass
from enum import Enum

import pytest
from dataclasses_json import DataClassJsonMixin
from google.protobuf.struct_pb2 import Struct
from typing_extensions import Annotated, get_origin

import flytekit
import flytekit.configuration
from flytekit import Secret, SQLTask, dynamic, kwtypes, map_task
from flytekit.configuration import FastSerializationSettings, Image, ImageConfig
from flytekit.core import context_manager, launch_plan, promise
from flytekit.core.condition import conditional
from flytekit.core.context_manager import ExecutionState
from flytekit.core.data_persistence import FileAccessProvider, flyte_tmp_dir
from flytekit.core.hash import HashMethod
from flytekit.core.node import Node
from flytekit.core.promise import NodeOutput, Promise, VoidPromise
from flytekit.core.resources import Resources
from flytekit.core.task import TaskMetadata, task
from flytekit.core.testing import patch, task_mock
from flytekit.core.type_engine import RestrictedTypeError, SimpleTransformer, TypeEngine
from flytekit.core.workflow import workflow
from flytekit.exceptions.user import FlyteValidationException
from flytekit.models import literals as _literal_models
from flytekit.models.core import types as _core_types
from flytekit.models.interface import Parameter
from flytekit.models.task import Resources as _resource_models
from flytekit.models.types import LiteralType, SimpleType
from flytekit.tools.translator import get_serializable
from flytekit.types.directory import FlyteDirectory, TensorboardLogs
from flytekit.types.error import FlyteError
from flytekit.types.file import FlyteFile
from flytekit.types.schema import FlyteSchema, SchemaOpenMode
from flytekit.types.structured.structured_dataset import StructuredDataset

serialization_settings = flytekit.configuration.SerializationSettings(
    project="proj",
    domain="dom",
    version="123",
    image_config=ImageConfig(Image(name="name", fqn="asdf/fdsa", tag="123")),
    env={},
)


def test_default_wf_params_works():
    @task
    def my_task(a: int):
        wf_params = flytekit.current_context()
        assert str(wf_params.execution_id) == "ex:local:local:local"
        assert flyte_tmp_dir in wf_params.raw_output_prefix

    my_task(a=3)
    assert context_manager.FlyteContextManager.size() == 1


def test_simple_input_output():
    @task
    def my_task(a: int) -> typing.NamedTuple("OutputsBC", b=int, c=str):
        ctx = flytekit.current_context()
        assert str(ctx.execution_id) == "ex:local:local:local"
        return a + 2, "hello world"

    assert my_task(a=3) == (5, "hello world")
    assert context_manager.FlyteContextManager.size() == 1


def test_forwardref_namedtuple_output():
    # This test case tests typing.NamedTuple outputs for cases where eg.
    # from __future__ import annotations is enabled, such that all type hints become ForwardRef
    @task
    def my_task(a: int) -> typing.NamedTuple("OutputsBC", b=typing.ForwardRef("int"), c=typing.ForwardRef("str")):
        ctx = flytekit.current_context()
        assert str(ctx.execution_id) == "ex:local:local:local"
        return a + 2, "hello world"

    assert my_task(a=3) == (5, "hello world")
    assert context_manager.FlyteContextManager.size() == 1


def test_annotated_namedtuple_output():
    @task
    def my_task(a: int) -> typing.NamedTuple("OutputA", a=Annotated[int, "metadata-a"]):
        return a + 2

    assert my_task(a=9) == (11,)
    assert get_origin(my_task.python_interface.outputs["a"]) is Annotated


def test_simple_input_no_output():
    @task
    def my_task(a: int):
        pass

    assert my_task(a=3) is None

    ctx = context_manager.FlyteContextManager.current_context()
    with context_manager.FlyteContextManager.with_context(ctx.with_new_compilation_state()) as ctx:
        outputs = my_task(a=3)
        assert isinstance(outputs, VoidPromise)

    assert context_manager.FlyteContextManager.size() == 1


def test_single_output():
    @task
    def my_task() -> str:
        return "Hello world"

    assert my_task() == "Hello world"

    ctx = context_manager.FlyteContextManager.current_context()
    with context_manager.FlyteContextManager.with_context(ctx.with_new_compilation_state()) as ctx:
        outputs = my_task()
        assert ctx.compilation_state is not None
        nodes = ctx.compilation_state.nodes
        assert len(nodes) == 1
        assert outputs.is_ready is False
        assert outputs.ref.node is nodes[0]

    assert context_manager.FlyteContextManager.size() == 1


def test_missing_output():
    @workflow
    def wf() -> str:
        return None  # type: ignore

    with pytest.raises(FlyteValidationException, match="Failed to bind output"):
        wf.compile()


def test_engine_file_output():
    basic_blob_type = _core_types.BlobType(
        format="",
        dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
    )

    fs = FileAccessProvider(local_sandbox_dir="/tmp/flytetesting", raw_output_prefix="/tmp/flyteraw")
    ctx = context_manager.FlyteContextManager.current_context()

    with context_manager.FlyteContextManager.with_context(ctx.with_file_access(fs)) as ctx:
        # Write some text to a file not in that directory above
        test_file_location = "/tmp/sample.txt"
        with open(test_file_location, "w") as fh:
            fh.write("Hello World\n")

        lit = TypeEngine.to_literal(ctx, test_file_location, os.PathLike, LiteralType(blob=basic_blob_type))

        # Since we're using local as remote, we should be able to just read the file from the 'remote' location.
        with open(lit.scalar.blob.uri, "r") as fh:
            assert fh.readline() == "Hello World\n"

        # We should also be able to turn the thing back into regular python native thing.
        redownloaded_local_file_location = TypeEngine.to_python_value(ctx, lit, os.PathLike)
        with open(redownloaded_local_file_location, "r") as fh:
            assert fh.readline() == "Hello World\n"

    assert context_manager.FlyteContextManager.size() == 1


def test_wf1():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a=y, b=b)
        return x, d

    assert len(my_wf.nodes) == 2
    assert my_wf._nodes[0].id == "n0"
    assert my_wf._nodes[1]._upstream_nodes[0] is my_wf._nodes[0]

    assert len(my_wf.output_bindings) == 2
    assert my_wf._output_bindings[0].var == "o0"
    assert my_wf._output_bindings[0].binding.promise.var == "t1_int_output"

    nt = typing.NamedTuple("SingleNT", [("t1_int_output", float)])

    @task
    def t3(a: int) -> nt:
        return nt(a + 2)

    assert t3.python_interface.output_tuple_name == "SingleNT"
    assert t3.interface.outputs["t1_int_output"] is not None
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_run():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a=y, b=b)
        return x, d

    x = my_wf(a=5, b="hello ")
    assert x == (7, "hello world")

    @workflow
    def my_wf2(a: int, b: str) -> (int, str):
        tup = t1(a=a)
        d = t2(a=tup.c, b=b)
        return tup.t1_int_output, d

    x = my_wf2(a=5, b="hello ")
    assert x == (7, "hello world")
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_with_overrides():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a).with_overrides(name="x")
        d = t2(a=y, b=b).with_overrides()
        return x, d

    x = my_wf(a=5, b="hello ")
    assert x == (7, "hello world")
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_with_list_of_inputs():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: typing.List[str]) -> str:
        return " ".join(a)

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        xx, yy = t1(a=a)
        d = t2(a=[b, yy])
        return xx, d

    x = my_wf(a=5, b="hello")
    assert x == (7, "hello world")

    @workflow
    def my_wf2(a: int, b: str) -> int:
        x, y = t1(a=a)
        t2(a=[b, y])
        return x

    x = my_wf2(a=5, b="hello")
    assert x == 7
    assert context_manager.FlyteContextManager.size() == 1


def test_wf_output_mismatch():
    with pytest.raises(AssertionError):

        @workflow
        def my_wf(a: int, b: str) -> (int, str):
            return a

        my_wf()

    with pytest.raises(AssertionError):

        @workflow
        def my_wf2(a: int, b: str) -> int:
            return a, b  # type: ignore

        my_wf2()

    with pytest.raises(AssertionError):

        @workflow
        def my_wf3(a: int, b: str) -> int:
            return (a,)  # type: ignore

        my_wf3()

    assert context_manager.FlyteContextManager.size() == 1


def test_promise_return():
    """
    Testing that when a workflow is local executed but a local wf execution context already exists, Promise objects
    are returned wrapping Flyte literals instead of the unpacked dict.
    """

    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def mimic_sub_wf(a: int) -> (str, str):
        x, y = t1(a=a)
        u, v = t1(a=x)
        return y, v

    ctx = context_manager.FlyteContextManager.current_context()

    with context_manager.FlyteContextManager.with_context(
        ctx.with_execution_state(
            ctx.new_execution_state().with_params(mode=ExecutionState.Mode.LOCAL_WORKFLOW_EXECUTION)
        )
    ) as ctx:
        a, b = mimic_sub_wf(a=3)

    assert isinstance(a, promise.Promise)
    assert isinstance(b, promise.Promise)
    assert a.val.scalar.value.string_value == "world-5"
    assert b.val.scalar.value.string_value == "world-7"
    assert context_manager.FlyteContextManager.size() == 1


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf1_with_sql():
    import pandas as pd

    sql = SQLTask(
        "my-query",
        query_template="SELECT * FROM hive.city.fact_airport_sessions WHERE ds = '{{ .Inputs.ds }}' LIMIT 10",
        inputs=kwtypes(ds=datetime.datetime),
        outputs=kwtypes(results=FlyteSchema),
        metadata=TaskMetadata(retries=2),
    )

    @task
    def t1() -> datetime.datetime:
        return datetime.datetime.now()

    @workflow
    def my_wf() -> FlyteSchema:
        dt = t1()
        return sql(ds=dt)

    with task_mock(sql) as mock:
        mock.return_value = pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})
        assert (my_wf().open().all() == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()
    assert context_manager.FlyteContextManager.size() == 1


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf1_with_sql_with_patch():
    import pandas as pd

    sql = SQLTask(
        "my-query",
        query_template="SELECT * FROM hive.city.fact_airport_sessions WHERE ds = '{{ .Inputs.ds }}' LIMIT 10",
        inputs=kwtypes(ds=datetime.datetime),
        outputs=kwtypes(results=FlyteSchema),
        metadata=TaskMetadata(retries=2),
    )

    @task
    def t1() -> datetime.datetime:
        return datetime.datetime.now()

    @workflow
    def my_wf() -> FlyteSchema:
        dt = t1()
        return sql(ds=dt)

    @patch(sql)
    def test_user_demo_test(mock_sql):
        mock_sql.return_value = pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})
        assert (my_wf().open().all() == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()

    # Have to call because tests inside tests don't run
    test_user_demo_test()
    assert context_manager.FlyteContextManager.size() == 1


def test_flyte_file_in_dataclass():
    @dataclass
    class InnerFileStruct(DataClassJsonMixin):
        a: FlyteFile
        b: FlyteFile

    @dataclass
    class FileStruct(DataClassJsonMixin):
        a: FlyteFile
        b: InnerFileStruct

    @task
    def t1(path: str) -> FileStruct:
        file = FlyteFile(path)
        fs = FileStruct(a=file, b=InnerFileStruct(a=file, b=FlyteFile(path)))
        return fs

    @dynamic
    def dyn(fs: FileStruct):
        t2(fs=fs)
        t3(fs=fs)

    @task
    def t2(fs: FileStruct) -> os.PathLike:
        assert fs.a.remote_source == "s3://somewhere"
        assert fs.b.a.remote_source == "s3://somewhere"
        assert fs.b.b.remote_source == "s3://somewhere"
        assert flyte_tmp_dir in fs.a.path
        assert flyte_tmp_dir in fs.b.a.path
        assert flyte_tmp_dir in fs.b.b.path

        os.makedirs(os.path.dirname(fs.a.path), exist_ok=True)
        with open(fs.a.path, "w") as file1:
            file1.write("hello world")

        return fs.a.path

    @task
    def t3(fs: FileStruct) -> FlyteFile:
        return fs.a

    @workflow
    def wf(path: str) -> (os.PathLike, FlyteFile):
        n1 = t1(path=path)
        dyn(fs=n1)
        return t2(fs=n1), t3(fs=n1)

    assert flyte_tmp_dir in wf(path="s3://somewhere")[0].path
    assert flyte_tmp_dir in wf(path="s3://somewhere")[1].path
    assert "s3://somewhere" == wf(path="s3://somewhere")[1].remote_source


def test_flyte_directory_in_dataclass():
    @dataclass
    class InnerFileStruct(DataClassJsonMixin):
        a: FlyteDirectory
        b: TensorboardLogs

    @dataclass
    class FileStruct(DataClassJsonMixin):
        a: FlyteDirectory
        b: InnerFileStruct

    @task
    def t1(path: str) -> FileStruct:
        dir = FlyteDirectory(path)
        fs = FileStruct(a=dir, b=InnerFileStruct(a=dir, b=TensorboardLogs(path)))
        return fs

    @task
    def t2(fs: FileStruct) -> FlyteDirectory:
        return fs.a.path

    @workflow
    def wf(path: str) -> FlyteDirectory:
        n1 = t1(path=path)
        return t2(fs=n1)

    assert flyte_tmp_dir in wf(path="s3://somewhere").path


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_structured_dataset_in_dataclass():
    import pandas as pd
    from pandas._testing import assert_frame_equal

    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})

    @dataclass
    class InnerDatasetStruct(DataClassJsonMixin):
        a: StructuredDataset

    @dataclass
    class DatasetStruct(DataClassJsonMixin):
        a: StructuredDataset
        b: InnerDatasetStruct

    @task
    def t1(path: str) -> DatasetStruct:
        sd = StructuredDataset(dataframe=df, uri=path)
        return DatasetStruct(a=sd, b=InnerDatasetStruct(a=sd))

    @workflow
    def wf(path: str) -> DatasetStruct:
        return t1(path=path)

    with tempfile.TemporaryDirectory() as tmp_dir:
        fname = os.path.join(tmp_dir, "df_file")
        res = wf(path=fname)
        assert "parquet" == res.a.file_format
        assert "parquet" == res.b.a.file_format
        assert_frame_equal(df, res.a.open(pd.DataFrame).all())
        assert_frame_equal(df, res.b.a.open(pd.DataFrame).all())


def test_wf1_with_map():
    @task
    def t1(a: int) -> int:
        a = a + 2
        return a

    @task
    def t2(x: typing.List[int]) -> int:
        return functools.reduce(lambda a, b: a + b, x)

    @workflow
    def my_wf(a: typing.List[int]) -> int:
        x = map_task(t1, metadata=TaskMetadata(retries=1))(a=a)
        return t2(x=x)

    x = my_wf(a=[5, 6])
    assert x == 15
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_compile_time_constant_vars():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a="This is my way", b=b)
        return x, d

    x = my_wf(a=5, b="hello ")
    assert x == (7, "hello This is my way")
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_with_constant_return():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        t2(a="This is my way", b=b)
        return x, "A constant output"

    x = my_wf(a=5, b="hello ")
    assert x == (7, "A constant output")

    @workflow
    def my_wf2(a: int, b: str) -> int:
        t1(a=a)
        t2(a="This is my way", b=b)
        return 10

    assert my_wf2(a=5, b="hello ") == 10
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_with_dynamic():
    @task
    def t1(a: int) -> str:
        a = a + 2
        return "world-" + str(a)

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @dynamic
    def my_subwf(a: int) -> typing.List[str]:
        s = []
        for i in range(a):
            s.append(t1(a=i))
        return s

    @workflow
    def my_wf(a: int, b: str) -> (str, typing.List[str]):
        x = t2(a=b, b=b)
        v = my_subwf(a=a)
        return x, v

    v = 5
    x = my_wf(a=v, b="hello ")
    assert x == ("hello hello ", ["world-" + str(i) for i in range(2, v + 2)])

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            flytekit.configuration.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        new_exc_state = ctx.execution_state.with_params(mode=ExecutionState.Mode.TASK_EXECUTION)
        with context_manager.FlyteContextManager.with_context(ctx.with_execution_state(new_exc_state)) as ctx:
            dynamic_job_spec = my_subwf.compile_into_workflow(ctx, my_subwf._task_function, a=5)
            assert len(dynamic_job_spec._nodes) == 5
            assert len(dynamic_job_spec.tasks) == 1

    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_with_fast_dynamic():
    @task
    def t1(a: int) -> str:
        a = a + 2
        return "fast-" + str(a)

    @dynamic
    def my_subwf(a: int) -> typing.List[str]:
        s = []
        for i in range(a):
            s.append(t1(a=i))
        return s

    @workflow
    def my_wf(a: int) -> typing.List[str]:
        v = my_subwf(a=a)
        return v

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            flytekit.configuration.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
                fast_serialization_settings=FastSerializationSettings(
                    enabled=True,
                    destination_dir="/User/flyte/workflows",
                    distribution_location="s3://my-s3-bucket/fast/123",
                ),
            )
        )
    ) as ctx:
        with context_manager.FlyteContextManager.with_context(
            ctx.with_execution_state(
                ctx.execution_state.with_params(
                    mode=ExecutionState.Mode.TASK_EXECUTION,
                )
            )
        ) as ctx:
            input_literal_map = TypeEngine.dict_to_literal_map(ctx, {"a": 5})

            dynamic_job_spec = my_subwf.dispatch_execute(ctx, input_literal_map)
            assert len(dynamic_job_spec._nodes) == 5
            assert len(dynamic_job_spec.tasks) == 1
            args = " ".join(dynamic_job_spec.tasks[0].container.args)
            assert args.startswith(
                "pyflyte-fast-execute --additional-distribution s3://my-s3-bucket/fast/123 "
                "--dest-dir /User/flyte/workflows"
            )

    assert context_manager.FlyteContextManager.size() == 1


def test_list_output():
    @task
    def t1(a: int) -> str:
        a = a + 2
        return "world-" + str(a)

    @workflow
    def lister() -> typing.List[str]:
        s = []
        # FYI: For users who happen to look at this, keep in mind this is only run once at compile time.
        for i in range(10):
            s.append(t1(a=i))
        return s

    assert len(lister.interface.outputs) == 1
    binding_data = lister.output_bindings[0].binding  # the property should be named binding_data
    assert binding_data.collection is not None
    assert len(binding_data.collection.bindings) == 10


def test_comparison_refs():
    def dummy_node(node_id) -> Node:
        n = Node(
            node_id,
            metadata=None,
            bindings=[],
            upstream_nodes=[],
            flyte_entity=SQLTask(name="x", query_template="x", inputs={}),
        )

        n._id = node_id
        return n

    px = Promise("x", NodeOutput(var="x", node=dummy_node("n1")))
    py = Promise("y", NodeOutput(var="y", node=dummy_node("n2")))

    def print_expr(expr):
        print(f"{expr} is type {type(expr)}")

    print_expr(px == py)
    print_expr(px < py)
    print_expr((px == py) & (px < py))
    print_expr(((px == py) & (px < py)) | (px > py))
    print_expr(px < 5)
    print_expr(px >= 5)
    print_expr(px != 5)


def test_comparison_lits():
    px = Promise("x", TypeEngine.to_literal(None, 5, int, None))
    py = Promise("y", TypeEngine.to_literal(None, 8, int, None))

    def eval_expr(expr, expected: bool):
        print(f"{expr} evals to {expr.eval()}")
        assert expected == expr.eval()

    eval_expr(px == py, False)
    eval_expr(px < py, True)
    eval_expr((px == py) & (px < py), False)
    eval_expr(((px == py) & (px < py)) | (px > py), False)
    eval_expr(px < 5, False)
    eval_expr(px >= 5, True)
    eval_expr(py >= 5, True)
    eval_expr(py != 5, True)
    eval_expr(px != 5, False)


def test_wf1_branches():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str) -> str:
        return a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = (
            conditional("test1")
            .if_(x == 4)
            .then(t2(a=b))
            .elif_(x >= 5)
            .then(t2(a=y))
            .else_()
            .fail("Unable to choose branch")
        )
        f = conditional("test2").if_(d == "hello ").then(t2(a="It is hello")).else_().then(t2(a="Not Hello!"))
        return x, f

    x = my_wf(a=5, b="hello ")
    assert x == (7, "Not Hello!")

    x = my_wf(a=2, b="hello ")
    assert x == (4, "It is hello")
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_branches_ne():
    @task
    def t1(a: int) -> int:
        return a + 1

    @task
    def t2(a: str) -> str:
        return a

    @workflow
    def my_wf(a: int, b: str) -> str:
        new_a = t1(a=a)
        return conditional("test1").if_(new_a != 5).then(t2(a=b)).else_().fail("Unable to choose branch")

    with pytest.raises(ValueError):
        my_wf(a=4, b="hello")

    x = my_wf(a=5, b="hello")
    assert x == "hello"
    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_branches_no_else_malformed_but_no_error():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str) -> str:
        return a

    with pytest.raises(FlyteValidationException):

        @workflow
        def my_wf(a: int, b: str) -> (int, str):
            x, y = t1(a=a)
            d = conditional("test1").if_(x == 4).then(t2(a=b)).elif_(x >= 5).then(t2(a=y))
            conditional("test2").if_(x == 4).then(t2(a=b)).elif_(x >= 5).then(t2(a=y)).else_().fail("blah")
            return x, d

        my_wf()

    assert context_manager.FlyteContextManager.size() == 1


def test_wf1_branches_failing():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str) -> str:
        return a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = (
            conditional("test1")
            .if_(x == 4)
            .then(t2(a=b))
            .elif_(x >= 5)
            .then(t2(a=y))
            .else_()
            .fail("All Branches failed")
        )
        return x, d

    with pytest.raises(ValueError):
        my_wf(a=1, b="hello ")
    assert context_manager.FlyteContextManager.size() == 1


def test_cant_use_normal_tuples_as_input():
    with pytest.raises(RestrictedTypeError):

        @task
        def t1(a: tuple) -> str:
            return a[0]


def test_cant_use_normal_tuples_as_output():
    with pytest.raises(RestrictedTypeError):

        @task
        def t1(a: str) -> tuple:
            return (a, 3)


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf1_df():
    import pandas as pd

    @task
    def t1(a: int) -> pd.DataFrame:
        return pd.DataFrame(data={"col1": [a, 2], "col2": [a, 4]})

    @task
    def t2(df: pd.DataFrame) -> pd.DataFrame:
        return pd.concat([df, pd.DataFrame(data={"col1": [5, 10], "col2": [5, 10]})])

    @workflow
    def my_wf(a: int) -> pd.DataFrame:
        df = t1(a=a)
        return t2(df=df)

    x = my_wf(a=20)
    assert isinstance(x, pd.DataFrame)
    result_df = x.reset_index(drop=True) == pd.DataFrame(
        data={"col1": [20, 2, 5, 10], "col2": [20, 4, 5, 10]}
    ).reset_index(drop=True)
    assert result_df.all().all()


def test_lp_serialize():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_subwf(a: int) -> typing.Tuple[str, str]:
        x, y = t1(a=a)
        u, v = t1(a=x)
        return y, v

    lp = launch_plan.LaunchPlan.create("serialize_test1", my_subwf)
    lp_with_defaults = launch_plan.LaunchPlan.create("serialize_test2", my_subwf, default_inputs={"a": 3})

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig(Image(name="name", fqn="asdf/fdsa", tag="123")),
        env={},
    )
    lp_model = get_serializable(OrderedDict(), serialization_settings, lp)
    assert len(lp_model.spec.default_inputs.parameters) == 1
    assert lp_model.spec.default_inputs.parameters["a"].required
    assert len(lp_model.spec.fixed_inputs.literals) == 0

    lp_model = get_serializable(OrderedDict(), serialization_settings, lp_with_defaults)
    assert len(lp_model.spec.default_inputs.parameters) == 1
    assert not lp_model.spec.default_inputs.parameters["a"].required
    assert lp_model.spec.default_inputs.parameters["a"].default == _literal_models.Literal(
        scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=3))
    )
    assert len(lp_model.spec.fixed_inputs.literals) == 0

    # Adding a check to make sure oneof is respected. Tricky with booleans... if a default is specified, the
    # required field needs to be None, not False.
    parameter_a = lp_model.spec.default_inputs.parameters["a"]
    parameter_a = Parameter.from_flyte_idl(parameter_a.to_flyte_idl())
    assert parameter_a.default is not None


def test_wf_tuple_fails():
    with pytest.raises(RestrictedTypeError):

        @task
        def t1(a: tuple) -> (int, str):
            return a[0] + 2, str(a) + "-HELLO"


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf_typed_schema():
    import pandas as pd

    from flytekit.types.schema import FlyteSchema

    schema1 = FlyteSchema[kwtypes(x=int, y=str)]

    @task
    def t1() -> schema1:
        s = schema1()
        s.open().write(pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]}))
        return s

    @task
    def t2(s: FlyteSchema[kwtypes(x=int, y=str)]) -> FlyteSchema[kwtypes(x=int)]:
        df = s.open().all()
        return df[s.column_names()[:-1]]

    @workflow
    def wf() -> FlyteSchema[kwtypes(x=int)]:
        return t2(s=t1())

    w = t1()
    assert w is not None
    df = w.open(override_mode=SchemaOpenMode.READ).all()
    result_df = df.reset_index(drop=True) == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]}).reset_index(drop=True)
    assert result_df.all().all()

    df = t2(s=w.as_readonly())
    df = df.open(override_mode=SchemaOpenMode.READ).all()
    result_df = df.reset_index(drop=True) == pd.DataFrame(data={"x": [1, 2]}).reset_index(drop=True)
    assert result_df.all().all()

    x = wf()
    df = x.open().all()
    result_df = df.reset_index(drop=True) == pd.DataFrame(data={"x": [1, 2]}).reset_index(drop=True)
    assert result_df.all().all()


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf_schema_to_df():
    import pandas as pd

    from flytekit.types.schema import FlyteSchema

    schema1 = FlyteSchema[kwtypes(x=int, y=str)]

    @task
    def t1() -> schema1:
        s = schema1()
        s.open().write(pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]}))
        return s

    @task
    def t2(df: pd.DataFrame) -> int:
        return len(df.columns.values)

    @workflow
    def wf() -> int:
        return t2(df=t1())

    x = wf()
    assert x == 2


def test_dict_wf_with_constants():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: typing.Dict[str, str]) -> str:
        return " ".join([v for k, v in a.items()])

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a={"key1": b, "key2": y})
        return x, d

    x = my_wf(a=5, b="hello")
    assert x == (7, "hello world")

    spec = get_serializable(OrderedDict(), serialization_settings, my_wf)
    assert spec.template.nodes[1].upstream_node_ids == ["n0"]


def test_dict_wf_with_conversion():
    @task
    def t1(a: int) -> typing.Dict[str, str]:
        return {"a": str(a)}

    @task
    def t2(a: dict) -> str:
        print(f"HAHAH {a}")
        return " ".join([v for k, v in a.items()])

    @workflow
    def my_wf(a: int) -> str:
        return t2(a=t1(a=a))

    with pytest.raises(TypeError):
        my_wf(a=5)


def test_wf_with_empty_dict():
    @task
    def t1() -> typing.Dict:
        return {}

    @task
    def t2(d: typing.Dict):
        assert d == {}

    @workflow
    def wf():
        d = t1()
        t2(d=d)

    wf()


def test_wf_with_catching_no_return():
    @task
    def t1() -> typing.Dict:
        return {}

    @task
    def t2(d: typing.Dict):
        assert d == {}

    @task
    def t3(s: str):
        pass

    with pytest.raises(AssertionError):

        @workflow
        def wf():
            d = t1()
            # The following statement is wrong, this should not be allowed to pass to another task
            x = t2(d=d)
            # Passing x is wrong in this case
            t3(s=x)

        wf()


def test_wf_custom_types_missing_dataclass_json():
    with pytest.raises(AssertionError):

        @dataclass
        class MyCustomType(object):
            pass

        @task
        def t1(a: int) -> MyCustomType:
            return MyCustomType()


def test_wf_custom_types():
    @dataclass
    class MyCustomType(DataClassJsonMixin):
        x: int
        y: str

    @task
    def t1(a: int) -> MyCustomType:
        return MyCustomType(x=a, y="t1")

    @task
    def t2(a: MyCustomType, b: str) -> (MyCustomType, int):
        return MyCustomType(x=a.x, y=f"{a.y} {b}"), 5

    @workflow
    def my_wf(a: int, b: str) -> (MyCustomType, int):
        return t2(a=t1(a=a), b=b)

    c, v = my_wf(a=10, b="hello")
    assert v == 5
    assert c.x == 10
    assert c.y == "t1 hello"


def test_arbit_class():
    class Foo(object):
        def __init__(self, number: int):
            self.number = number

    @task
    def t1(a: int) -> Foo:
        return Foo(number=a)

    @task
    def t2(a: Foo) -> typing.List[Foo]:
        return [a, a]

    @task
    def t3(a: typing.List[Foo]) -> typing.Dict[str, Foo]:
        return {"hello": a[0]}

    def wf(a: int) -> typing.Dict[str, Foo]:
        o1 = t1(a=a)
        o2 = t2(a=o1)
        return t3(a=o2)

    assert wf(1)["hello"].number == 1


def test_dataclass_more():
    @dataclass
    class Datum(DataClassJsonMixin):
        x: int
        y: str
        z: typing.Dict[int, str]

    @task
    def stringify(x: int) -> Datum:
        return Datum(x=x, y=str(x), z={x: str(x)})

    @task
    def add(x: Datum, y: Datum) -> Datum:
        x.z.update(y.z)
        return Datum(x=x.x + y.x, y=x.y + y.y, z=x.z)

    @workflow
    def wf(x: int, y: int) -> Datum:
        return add(x=stringify(x=x), y=stringify(x=y))

    wf(x=10, y=20)


def test_enum_in_dataclass():
    class Color(Enum):
        RED = "red"
        GREEN = "green"
        BLUE = "blue"

    @dataclass
    class Datum(DataClassJsonMixin):
        x: int
        y: Color

    @task
    def t1(x: int) -> Datum:
        return Datum(x=x, y=Color.RED)

    @workflow
    def wf(x: int) -> Datum:
        return t1(x=x)

    assert wf(x=10) == Datum(10, Color.RED)


def test_flyte_schema_dataclass():
    TestSchema = FlyteSchema[kwtypes(some_str=str)]

    @dataclass
    class InnerResult(DataClassJsonMixin):
        number: int
        schema: TestSchema

    @dataclass
    class Result(DataClassJsonMixin):
        result: InnerResult
        schema: TestSchema

    schema = TestSchema()

    @task
    def t1(x: int) -> Result:
        return Result(result=InnerResult(number=x, schema=schema), schema=schema)

    @workflow
    def wf(x: int) -> Result:
        return t1(x=x)

    assert wf(x=10) == Result(result=InnerResult(number=10, schema=schema), schema=schema)


def test_environment():
    @task(environment={"FOO": "foofoo", "BAZ": "baz"})
    def t1(a: int) -> str:
        a = a + 2
        return "now it's " + str(a)

    @workflow
    def my_wf(a: int) -> str:
        x = t1(a=a)
        return x

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={"FOO": "foo", "BAR": "bar"},
    )
    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_new_compilation_state()
    ):
        task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
        assert task_spec.template.container.env == {"FOO": "foofoo", "BAR": "bar", "BAZ": "baz"}


def test_resources():
    @task(
        requests=Resources(cpu="1", ephemeral_storage="500Mi"),
        limits=Resources(cpu="2", mem="400M", ephemeral_storage="501Mi"),
    )
    def t1(a: int) -> str:
        a = a + 2
        return "now it's " + str(a)

    @task(requests=Resources(cpu="3"))
    def t2(a: int) -> str:
        a = a + 200
        return "now it's " + str(a)

    @workflow
    def my_wf(a: int) -> str:
        x = t1(a=a)
        return x

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )
    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_new_compilation_state()
    ):
        task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
        assert task_spec.template.container.resources.requests == [
            _resource_models.ResourceEntry(_resource_models.ResourceName.EPHEMERAL_STORAGE, "500Mi"),
            _resource_models.ResourceEntry(_resource_models.ResourceName.CPU, "1"),
        ]
        assert task_spec.template.container.resources.limits == [
            _resource_models.ResourceEntry(_resource_models.ResourceName.EPHEMERAL_STORAGE, "501Mi"),
            _resource_models.ResourceEntry(_resource_models.ResourceName.CPU, "2"),
            _resource_models.ResourceEntry(_resource_models.ResourceName.MEMORY, "400M"),
        ]

        task_spec2 = get_serializable(OrderedDict(), serialization_settings, t2)
        assert task_spec2.template.container.resources.requests == [
            _resource_models.ResourceEntry(_resource_models.ResourceName.CPU, "3")
        ]
        assert task_spec2.template.container.resources.limits == []


def test_wf_explicitly_returning_empty_task():
    @task
    def t1():
        ...

    @workflow
    def my_subwf():
        return t1()  # This forces the wf local_execute to handle VoidPromises

    assert my_subwf() is None


def test_nested_dict():
    @task(cache=True, cache_version="1.0.0")
    def squared(value: int) -> typing.Dict[str, int]:
        return {"value:": value**2}

    @workflow
    def compute_square_wf(input_integer: int) -> typing.Dict[str, int]:
        compute_square_result = squared(value=input_integer)
        return compute_square_result

    compute_square_wf(input_integer=5)


def test_nested_dict2():
    @task(cache=True, cache_version="1.0.0")
    def squared(value: int) -> typing.List[typing.Dict[str, int]]:
        return [
            {"squared_value": value**2},
        ]

    @workflow
    def compute_square_wf(input_integer: int) -> typing.List[typing.Dict[str, int]]:
        compute_square_result = squared(value=input_integer)
        return compute_square_result


def test_secrets():
    @task(secret_requests=[Secret("my_group", "my_key")])
    def foo() -> str:
        return flytekit.current_context().secrets.get("my_group", "")

    with pytest.raises(ValueError):
        foo()

    @task(secret_requests=[Secret("group", group_version="v1", key="key")])
    def foo2() -> str:
        return flytekit.current_context().secrets.get("group", "key")

    os.environ[flytekit.current_context().secrets.get_secrets_env_var("group", "key")] = "super-secret-value2"
    assert foo2() == "super-secret-value2"

    with pytest.raises(AssertionError):

        @task(secret_requests=["test"])
        def foo() -> str:
            pass


def test_nested_dynamic():
    @task
    def t1(a: int) -> str:
        a = a + 2
        return "world-" + str(a)

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> typing.Tuple[str, typing.List[str]]:
        @dynamic
        def my_subwf(a: int) -> typing.List[str]:
            s = []
            for i in range(a):
                s.append(t1(a=i))
            return s

        x = t2(a=b, b=b)
        v = my_subwf(a=a)
        return x, v

    v = 5
    x = my_wf(a=v, b="hello ")
    assert x == ("hello hello ", ["world-" + str(i) for i in range(2, v + 2)])

    settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )

    nested_my_subwf = my_wf.get_all_tasks()[0]

    ctx = context_manager.FlyteContextManager.current_context().with_serialization_settings(settings)
    with context_manager.FlyteContextManager.with_context(ctx) as ctx:
        es = ctx.new_execution_state().with_params(mode=ExecutionState.Mode.TASK_EXECUTION)
        with context_manager.FlyteContextManager.with_context(ctx.with_execution_state(es)) as ctx:
            dynamic_job_spec = nested_my_subwf.compile_into_workflow(ctx, nested_my_subwf._task_function, a=5)
            assert len(dynamic_job_spec._nodes) == 5


def test_workflow_named_tuple():
    @task
    def t1() -> str:
        return "Hello"

    @workflow
    def wf() -> typing.NamedTuple("OP", [("a", str), ("b", str)]):  # type: ignore
        return t1(), t1()

    assert wf() == ("Hello", "Hello")


def test_conditional_asymmetric_return():
    @task
    def square(n: int) -> int:
        """
        Parameters:
            n (float): name of the parameter for the task will be derived from the name of the input variable
                   the type will be automatically deduced to be Types.Integer
        Return:
            float: The label for the output will be automatically assigned and type will be deduced from the annotation
        """
        return n * n

    @task
    def double(n: int) -> int:
        """
        Parameters:
            n (float): name of the parameter for the task will be derived from the name of the input variable
                   the type will be automatically deduced to be Types.Integer
        Return:
            float: The label for the output will be automatically assigned and type will be deduced from the annotation
        """
        return 2 * n

    @task
    def coin_toss(seed: int) -> bool:
        """
        Mimic some condition checking to see if something ran correctly
        """
        r = random.Random(seed)
        if r.random() < 0.5:
            return True
        return False

    @task
    def sum_diff(a: int, b: int) -> typing.Tuple[int, int]:
        """
        sum_diff returns the sum and difference between a and b.
        """
        return a + b, a - b

    @workflow
    def consume_outputs(my_input: int, seed: int = 5) -> int:
        is_heads = coin_toss(seed=seed)
        res = (
            conditional("double_or_square")
            .if_(is_heads.is_true())
            .then(square(n=my_input))
            .else_()
            .then(sum_diff(a=my_input, b=my_input))
        )

        # Regardless of the result, always double before returning
        # the variable `res` in this case will carry the value of either square or double of the variable `my_input`
        return double(n=res)

    assert consume_outputs(my_input=4, seed=7) == 32
    assert consume_outputs(my_input=4) == 16


def test_guess_dict():
    @task
    def t2(a: dict) -> str:
        return ", ".join([f"K: {k} V: {v}" for k, v in a.items()])

    task_spec = get_serializable(OrderedDict(), serialization_settings, t2)
    assert task_spec.template.interface.inputs["a"].type.simple == SimpleType.STRUCT

    pt = TypeEngine.guess_python_type(task_spec.template.interface.inputs["a"].type)
    assert pt is dict

    input_map = {"a": {"k1": "v1", "k2": "2"}}
    guessed_types = {"a": pt}
    ctx = context_manager.FlyteContext.current_context()
    lm = TypeEngine.dict_to_literal_map(ctx, d=input_map, type_hints=guessed_types)
    assert isinstance(lm.literals["a"].scalar.generic, Struct)

    output_lm = t2.dispatch_execute(ctx, lm)
    str_value = output_lm.literals["o0"].scalar.primitive.string_value
    assert str_value == "K: k2 V: 2, K: k1 V: v1" or str_value == "K: k1 V: v1, K: k2 V: 2"


def test_guess_dict2():
    @task
    def t2(a: typing.List[dict]) -> str:
        strs = []
        for input_dict in a:
            strs.append(", ".join([f"K: {k} V: {v}" for k, v in input_dict.items()]))
        return " ".join(strs)

    task_spec = get_serializable(OrderedDict(), serialization_settings, t2)
    assert task_spec.template.interface.inputs["a"].type.collection_type.simple == SimpleType.STRUCT
    pt_map = TypeEngine.guess_python_types(task_spec.template.interface.inputs)
    assert pt_map == {"a": typing.List[dict]}


def test_guess_dict3():
    @task
    def t2() -> dict:
        return {"k1": "v1", "k2": 3, 4: {"one": [1, "two", [3]]}}

    task_spec = get_serializable(OrderedDict(), serialization_settings, t2)

    pt_map = TypeEngine.guess_python_types(task_spec.template.interface.outputs)
    assert pt_map["o0"] is dict

    ctx = context_manager.FlyteContextManager.current_context()
    output_lm = t2.dispatch_execute(ctx, _literal_models.LiteralMap(literals={}))
    expected_struct = Struct()
    expected_struct.update({"k1": "v1", "k2": 3, "4": {"one": [1, "two", [3]]}})
    assert output_lm.literals["o0"].scalar.generic == expected_struct


def test_guess_dict4():
    @dataclass
    class Foo(DataClassJsonMixin):
        x: int
        y: str
        z: typing.Dict[str, str]

    @dataclass
    class Bar(DataClassJsonMixin):
        x: int
        y: dict
        z: Foo

    @task
    def t1() -> Foo:
        return Foo(x=1, y="foo", z={"hello": "world"})

    task_spec = get_serializable(OrderedDict(), serialization_settings, t1)
    pt_map = TypeEngine.guess_python_types(task_spec.template.interface.outputs)
    assert dataclasses.is_dataclass(pt_map["o0"])

    ctx = context_manager.FlyteContextManager.current_context()
    output_lm = t1.dispatch_execute(ctx, _literal_models.LiteralMap(literals={}))
    expected_struct = Struct()
    expected_struct.update({"x": 1, "y": "foo", "z": {"hello": "world"}})
    assert output_lm.literals["o0"].scalar.generic == expected_struct

    @task
    def t2() -> Bar:
        return Bar(x=1, y={"hello": "world"}, z=Foo(x=1, y="foo", z={"hello": "world"}))

    task_spec = get_serializable(OrderedDict(), serialization_settings, t2)
    pt_map = TypeEngine.guess_python_types(task_spec.template.interface.outputs)
    assert dataclasses.is_dataclass(pt_map["o0"])

    output_lm = t2.dispatch_execute(ctx, _literal_models.LiteralMap(literals={}))
    expected_struct.update({"x": 1, "y": {"hello": "world"}, "z": {"x": 1, "y": "foo", "z": {"hello": "world"}}})
    assert output_lm.literals["o0"].scalar.generic == expected_struct


def test_error_messages():
    @task
    def foo(a: int, b: str) -> typing.Tuple[int, str]:
        return 10, "hello"

    @task
    def foo2(a: int, b: str) -> typing.Tuple[int, str]:
        return "hello", 10  # type: ignore

    @task
    def foo3(a: typing.Dict) -> typing.Dict:
        return a

    # pytest-xdist uses `__channelexec__` as the top-level module
    running_xdist = os.environ.get("PYTEST_XDIST_WORKER") is not None
    prefix = "__channelexec__." if running_xdist else ""

    with pytest.raises(
        TypeError,
        match=(
            f"Failed to convert inputs of task '{prefix}tests.flytekit.unit.core.test_type_hints.foo':\n"
            "  Failed argument 'a': Expected value of type <class 'int'> but got 'hello' of type <class 'str'>"
        ),
    ):
        foo(a="hello", b=10)  # type: ignore

    with pytest.raises(
        TypeError,
        match=(
            f"Failed to convert outputs of task '{prefix}tests.flytekit.unit.core.test_type_hints.foo2' "
            "at position 0:\n"
            "  Expected value of type <class 'int'> but got 'hello' of type <class 'str'>"
        ),
    ):
        foo2(a=10, b="hello")

    with pytest.raises(
        TypeError,
        match=f"Failed to convert inputs of task '{prefix}tests.flytekit.unit.core.test_type_hints.foo3':\n  "
        f"Failed argument 'a': Expected a dict",
    ):
        foo3(a=[{"hello": 2}])


def test_failure_node():
    @task
    def run(a: int, b: str) -> typing.Tuple[int, str]:
        return a + 1, b

    @task
    def fail(a: int, b: str) -> typing.Tuple[int, str]:
        raise ValueError("Fail!")

    @task
    def failure_handler(a: int, b: str, err: typing.Optional[FlyteError]) -> typing.Tuple[int, str]:
        print(f"Handling error: {err}")
        return a + 1, b

    @workflow(on_failure=failure_handler)
    def subwf(a: int, b: str) -> typing.Tuple[int, str]:
        x, y = run(a=a, b=b)
        return fail(a=x, b=y)

    @workflow(on_failure=failure_handler)
    def wf1(a: int, b: str) -> typing.Tuple[int, str]:
        x, y = run(a=a, b=b)
        return fail(a=x, b=y)

    @workflow(on_failure=failure_handler)
    def wf2(a: int, b: str) -> typing.Tuple[int, str]:
        x, y = run(a=a, b=b)
        return subwf(a=x, b=y)

    with pytest.raises(
        ValueError,
        match="Error encountered while executing",
    ):
        v, s = wf1(a=10, b="hello")
        assert v == 11
        assert "hello" in s
        assert wf1.failure_node is not None
        assert wf1.failure_node.flyte_entity == failure_handler

    with pytest.raises(
        ValueError,
        match="Error encountered while executing",
    ):
        v, s = wf2(a=10, b="hello")
        assert v == 11
        assert "hello" in s
        assert wf2.failure_node is not None
        assert wf2.failure_node.flyte_entity == failure_handler


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_union_type():
    import pandas as pd

    from flytekit.types.schema import FlyteSchema

    ut = typing.Union[int, str, float, FlyteFile, FlyteSchema, typing.List[int], typing.Dict[str, int]]

    @task
    def t1(a: ut) -> ut:
        return a

    @workflow
    def wf(a: ut) -> ut:
        return t1(a=a)

    assert wf(a=2) == 2
    assert wf(a="2") == "2"
    assert wf(a=2.0) == 2.0
    file = tempfile.NamedTemporaryFile(delete=False)
    assert isinstance(wf(a=FlyteFile(file.name)), FlyteFile)
    flyteSchema = FlyteSchema()
    flyteSchema.open().write(pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [1, 22]}))
    assert isinstance(wf(a=flyteSchema), FlyteSchema)
    assert wf(a=[1, 2, 3]) == [1, 2, 3]
    assert wf(a={"a": 1}) == {"a": 1}

    @task
    def t2(a: typing.Union[float, dict]) -> typing.Union[float, dict]:
        return a

    @workflow
    def wf2(a: typing.Union[int, str]) -> typing.Union[int, str]:
        return t2(a=a)

    # pytest-xdist uses `__channelexec__` as the top-level module
    running_xdist = os.environ.get("PYTEST_XDIST_WORKER") is not None
    prefix = "__channelexec__." if running_xdist else ""

    with pytest.raises(
        TypeError,
        match=re.escape(
            "Error encountered while executing 'wf2':\n"
            f"  Failed to convert inputs of task '{prefix}tests.flytekit.unit.core.test_type_hints.t2':\n"
            '  Cannot convert from <FlyteLiteral scalar { union { value { scalar { primitive { string_value: "2" } } } '
            'type { simple: STRING structure { tag: "str" } } } }> to typing.Union[float, dict] (using tag str)'
        ),
    ):
        assert wf2(a="2") == "2"


def test_optional_type():
    @task
    def t1(a: typing.Optional[int]) -> typing.Optional[int]:
        return a

    @workflow
    def wf(a: typing.Optional[int]) -> typing.Optional[int]:
        return t1(a=a)

    assert wf(a=2) == 2
    assert wf(a=None) is None


def test_optional_type_implicit_wrapping():
    @task
    def t1(a: int) -> typing.Optional[int]:
        return a if a > 0 else None

    @workflow
    def wf(a: int) -> typing.Optional[int]:
        return t1(a=a)

    assert wf(a=2) == 2
    assert wf(a=-10) is None


def test_union_type_implicit_wrapping():
    @task
    def t1(a: int) -> typing.Union[int, str]:
        return a if a > 0 else str(a)

    @workflow
    def wf(a: int) -> typing.Union[int, str]:
        return t1(a=a)

    assert wf(a=2) == 2
    assert wf(a=-10) == "-10"


def test_union_type_ambiguity_checking():
    class MyInt:
        def __init__(self, x: int):
            self.val = x

        def __eq__(self, other):
            if not isinstance(other, MyInt):
                return False
            return other.val == self.val

    TypeEngine.register(
        SimpleTransformer(
            "MyInt",
            MyInt,
            LiteralType(simple=SimpleType.INTEGER),
            lambda x: _literal_models.Literal(
                scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=x.val))
            ),
            lambda x: MyInt(x.scalar.primitive.integer),
        )
    )

    @task
    def t1(a: typing.Union[int, MyInt]) -> int:
        if isinstance(a, MyInt):
            return a.val
        return a

    @workflow
    def wf(a: int) -> int:
        return t1(a=a)

    with pytest.raises(
        TypeError, match="Ambiguous choice of variant for union type. Both int and MyInt transformers match"
    ):
        assert wf(a=10) == 10

    del TypeEngine._REGISTRY[MyInt]


def test_union_type_ambiguity_resolution():
    class MyInt:
        def __init__(self, x: int):
            self.val = x

        def __eq__(self, other):
            if not isinstance(other, MyInt):
                return False
            return other.val == self.val

    TypeEngine.register(
        SimpleTransformer(
            "MyInt",
            MyInt,
            LiteralType(simple=SimpleType.INTEGER),
            lambda x: _literal_models.Literal(
                scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=x.val))
            ),
            lambda x: MyInt(x.scalar.primitive.integer),
        )
    )

    @task
    def t1(a: typing.Union[int, MyInt]) -> str:
        if isinstance(a, MyInt):
            return f"MyInt {str(a.val)}"
        return str(a)

    @task
    def t2(a: int) -> typing.Union[int, MyInt]:
        if a < 0:
            return MyInt(a)
        return a

    @workflow
    def wf(a: int) -> str:
        return t1(a=t2(a=a))

    assert wf(a=10) == "10"
    assert wf(a=-10) == "MyInt -10"

    del TypeEngine._REGISTRY[MyInt]


def test_task_annotate_primitive_type_is_allowed():
    @task
    def plus_two(
        a: int,
    ) -> Annotated[int, HashMethod(lambda x: str(x + 1))]:
        return a + 2

    assert plus_two(a=1) == 3

    ctx = context_manager.FlyteContextManager.current_context()
    output_lm = plus_two.dispatch_execute(
        ctx,
        _literal_models.LiteralMap(
            literals={
                "a": _literal_models.Literal(
                    scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=3))
                )
            }
        ),
    )
    assert output_lm.literals["o0"].scalar.primitive.integer == 5
    assert output_lm.literals["o0"].hash == "6"


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_task_hash_return_pandas_dataframe():
    import pandas as pd

    constant_value = "road-hash"

    def constant_function(df: pd.DataFrame) -> str:
        return constant_value

    @task
    def t0() -> Annotated[pd.DataFrame, HashMethod(constant_function)]:
        return pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]})

    ctx = context_manager.FlyteContextManager.current_context()
    output_lm = t0.dispatch_execute(ctx, _literal_models.LiteralMap(literals={}))
    assert output_lm.literals["o0"].hash == constant_value

    # Confirm that the literal containing a hash does not have any effect on the scalar.
    df = TypeEngine.to_python_value(ctx, output_lm.literals["o0"], pd.DataFrame)
    expected_df = pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]})
    assert df.equals(expected_df)


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_workflow_containing_multiple_annotated_tasks():
    import pandas as pd

    def hash_function_t0(df: pd.DataFrame) -> str:
        return "hash-0"

    @task
    def t0() -> Annotated[pd.DataFrame, HashMethod(hash_function_t0)]:
        return pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]})

    def hash_function_t1(df: pd.DataFrame) -> str:
        return "hash-1"

    @task
    def t1() -> Annotated[pd.DataFrame, HashMethod(hash_function_t1)]:
        return pd.DataFrame(data={"col1": [10, 20], "col2": [30, 40]})

    @task
    def t2() -> pd.DataFrame:
        return pd.DataFrame(data={"col1": [100, 200], "col2": [300, 400]})

    # Auxiliary task used to sum up the dataframes. It demonstrates that the use of `Annotated` does not
    # have any impact in the definition and execution of cached or uncached downstream tasks
    @task
    def sum_dataframes(df0: pd.DataFrame, df1: pd.DataFrame, df2: pd.DataFrame) -> pd.DataFrame:
        return df0 + df1 + df2

    @workflow
    def wf() -> pd.DataFrame:
        df0 = t0()
        df1 = t1()
        df2 = t2()
        return sum_dataframes(df0=df0, df1=df1, df2=df2)

    df = wf()

    expected_df = pd.DataFrame(data={"col1": [1 + 10 + 100, 2 + 20 + 200], "col2": [3 + 30 + 300, 4 + 40 + 400]})
    assert expected_df.equals(df)


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_list_containing_multiple_annotated_pandas_dataframes():
    import pandas as pd

    def hash_pandas_dataframe(df: pd.DataFrame) -> str:
        return str(pd.util.hash_pandas_object(df))

    @task
    def produce_list_of_annotated_dataframes() -> (
        typing.List[Annotated[pd.DataFrame, HashMethod(hash_pandas_dataframe)]]
    ):
        return [pd.DataFrame({"column_1": [1, 2, 3]}), pd.DataFrame({"column_1": [4, 5, 6]})]

    @task(cache=True, cache_version="v0")
    def sum_list_of_pandas_dataframes(lst: typing.List[pd.DataFrame]) -> pd.DataFrame:
        return sum(lst)

    @workflow
    def wf() -> pd.DataFrame:
        lst = produce_list_of_annotated_dataframes()
        return sum_list_of_pandas_dataframes(lst=lst)

    df = wf()

    expected_df = pd.DataFrame({"column_1": [5, 7, 9]})
    assert expected_df.equals(df)


def test_ref_as_key_name():
    class MyOutput(typing.NamedTuple):
        # to make sure flytekit itself doesn't use this string
        ref: str

    @task
    def produce_things() -> MyOutput:
        return MyOutput(ref="ref")

    @workflow
    def run_things() -> MyOutput:
        return produce_things()

    assert run_things().ref == "ref"


def test_promise_not_allowed_in_overrides():
    @task
    def t1(a: int) -> int:
        return a + 1

    @workflow
    def my_wf(a: int, cpu: str) -> int:
        return t1(a=a).with_overrides(requests=Resources(cpu=cpu))

    with pytest.raises(AssertionError):
        my_wf(a=1, cpu=1)


def test_promise_illegal_resources():
    @task
    def t1(a: int) -> int:
        return a + 1

    @workflow
    def my_wf(a: int) -> int:
        return t1(a=a).with_overrides(requests=Resources(cpu=1))  # type: ignore

    with pytest.raises(AssertionError):
        my_wf(a=1)


def test_promise_illegal_retries():
    @task
    def t1(a: int) -> int:
        return a + 1

    @workflow
    def my_wf(a: int, retries: int) -> int:
        return t1(a=a).with_overrides(retries=retries)

    with pytest.raises(AssertionError):
        my_wf(a=1, retries=1)
