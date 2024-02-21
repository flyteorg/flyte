import sys
import typing
from collections import OrderedDict

import pytest

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.base_task import kwtypes
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.task import reference_task, task
from flytekit.core.workflow import ImperativeWorkflow, get_promise, workflow
from flytekit.exceptions.user import FlyteValidationException
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.models import literals as literal_models
from flytekit.tools.translator import get_serializable
from flytekit.types.file import FlyteFile
from flytekit.types.schema import FlyteSchema

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


# This is used for docs
def test_imperative():
    # Re import with alias
    from flytekit.core.workflow import ImperativeWorkflow as Workflow  # noqa

    # docs_tasks_start
    @task
    def t1(a: str) -> str:
        return a + " world"

    @task
    def t2():
        print("side effect")

    # docs_tasks_end

    # docs_start
    # Create the workflow with a name. This needs to be unique within the project and takes the place of the function
    # name that's used for regular decorated function-based workflows.
    wb = Workflow(name="my_workflow")
    # Adds a top level input to the workflow. This is like an input to a workflow function.
    wb.add_workflow_input("in1", str)
    # Call your tasks.
    node = wb.add_entity(t1, a=wb.inputs["in1"])
    wb.add_entity(t2)
    # This is analogous to a return statement
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])
    # docs_end

    assert wb(in1="hello") == "hello world"

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wb)
    assert len(wf_spec.template.nodes) == 2
    assert wf_spec.template.nodes[0].task_node is not None
    assert len(wf_spec.template.outputs) == 1
    assert wf_spec.template.outputs[0].var == "from_n0t1"
    assert len(wf_spec.template.interface.inputs) == 1
    assert len(wf_spec.template.interface.outputs) == 1

    # docs_equivalent_start
    nt = typing.NamedTuple("wf_output", [("from_n0t1", str)])

    @workflow
    def my_workflow(in1: str) -> nt:
        x = t1(a=in1)
        t2()
        return nt(x)

    # docs_equivalent_end

    # Create launch plan from wf, that can also be serialized.
    lp = LaunchPlan.create("test_wb", wb)
    lp_model = get_serializable(OrderedDict(), serialization_settings, lp)
    assert lp_model.spec.workflow_id.name == "my_workflow"

    wb2 = ImperativeWorkflow(name="parent.imperative")
    p_in1 = wb2.add_workflow_input("p_in1", str)
    p_node0 = wb2.add_subwf(wb, in1=p_in1)
    wb2.add_workflow_output("parent_wf_output", p_node0.from_n0t1, str)
    wb2_spec = get_serializable(OrderedDict(), serialization_settings, wb2)
    assert len(wb2_spec.template.nodes) == 1
    assert len(wb2_spec.template.interface.inputs) == 1
    assert wb2_spec.template.interface.inputs["p_in1"].type.simple is not None
    assert len(wb2_spec.template.interface.outputs) == 1
    assert wb2_spec.template.interface.outputs["parent_wf_output"].type.simple is not None
    assert wb2_spec.template.nodes[0].workflow_node.sub_workflow_ref.name == "my_workflow"
    assert len(wb2_spec.sub_workflows) == 1

    wb3 = ImperativeWorkflow(name="parent.imperative")
    p_in1 = wb3.add_workflow_input("p_in1", str)
    p_node0 = wb3.add_launch_plan(lp, in1=p_in1)
    wb3.add_workflow_output("parent_wf_output", p_node0.from_n0t1, str)
    wb3_spec = get_serializable(OrderedDict(), serialization_settings, wb3)
    assert len(wb3_spec.template.nodes) == 1
    assert len(wb3_spec.template.interface.inputs) == 1
    assert wb3_spec.template.interface.inputs["p_in1"].type.simple is not None
    assert len(wb3_spec.template.interface.outputs) == 1
    assert wb3_spec.template.interface.outputs["parent_wf_output"].type.simple is not None
    assert wb3_spec.template.nodes[0].workflow_node.launchplan_ref.name == "test_wb"


def test_imperative_list_bound():
    @task
    def t1(a: typing.List[int]) -> int:
        return sum(a)

    wb = ImperativeWorkflow(name="my.workflow.a")
    wb.add_workflow_input("in1", int)
    wb.add_workflow_input("in2", int)
    node = wb.add_entity(t1, a=[wb.inputs["in1"], wb.inputs["in2"]])
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])

    assert wb(in1=3, in2=4) == 7


def test_imperative_map_bound():
    @task
    def t1(a: typing.Dict[str, typing.List[int]]) -> typing.Dict[str, int]:
        return {k: sum(v) for k, v in a.items()}

    wb = ImperativeWorkflow(name="my.workflow.a")
    in1 = wb.add_workflow_input("in1", int)
    wb.add_workflow_input("in2", int)
    in3 = wb.add_workflow_input("in3", int)
    node = wb.add_entity(t1, a={"a": [in1, wb.inputs["in2"]], "b": [wb.inputs["in2"], in3]})
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])

    assert wb(in1=3, in2=4, in3=5) == {"a": 7, "b": 9}


def test_imperative_with_list_io():
    @task
    def t1(a: int) -> typing.List[int]:
        return [1, a, 3]

    @task
    def t2(a: typing.List[int]) -> int:
        return sum(a)

    wb = ImperativeWorkflow(name="my.workflow.a")
    t1_node = wb.add_entity(t1, a=2)
    t2_node = wb.add_entity(t2, a=t1_node.outputs["o0"])
    wb.add_workflow_output("from_n0t2", t2_node.outputs["o0"])

    assert wb() == 6


def test_imperative_wf_list_input():
    @task
    def t1(a: int) -> typing.List[int]:
        return [1, a, 3]

    @task
    def t2(a: typing.List[int], b: typing.List[int]) -> int:
        return sum(a) + sum(b)

    wb = ImperativeWorkflow(name="my.workflow.a")
    wf_in1 = wb.add_workflow_input("in1", typing.List[int])
    t1_node = wb.add_entity(t1, a=2)
    t2_node = wb.add_entity(t2, a=t1_node.outputs["o0"], b=wf_in1)
    wb.add_workflow_output("from_n0t2", t2_node.outputs["o0"])

    assert wb(in1=[5, 6, 7]) == 24

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wb)
    assert len(wf_spec.template.nodes) == 2
    assert wf_spec.template.nodes[0].task_node is not None


def test_imperative_scalar_bindings():
    @task
    def t1(a: typing.Dict[str, typing.List[int]]) -> typing.Dict[str, int]:
        return {k: sum(v) for k, v in a.items()}

    wb = ImperativeWorkflow(name="my.workflow.a")
    node = wb.add_entity(t1, a={"a": [3, 4], "b": [5, 6]})
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])

    assert wb() == {"a": 7, "b": 11}

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wb)
    assert len(wf_spec.template.nodes) == 1
    assert wf_spec.template.nodes[0].task_node is not None


def test_imperative_list_bound_output():
    @task
    def t1() -> int:
        return 3

    @task
    def t2(a: typing.List[int]) -> int:
        return sum(a)

    wb = ImperativeWorkflow(name="my.workflow.a")
    t1_node = wb.add_entity(t1)
    t2_node = wb.add_entity(t2, a=[1, 2, 3])
    wb.add_workflow_output("wf0", [t1_node.outputs["o0"], t2_node.outputs["o0"]], python_type=typing.List[int])

    assert wb() == [3, 6]


def test_imperative_tuples():
    @task
    def t1() -> (int, str):
        return 3, "three"

    @task
    def t3(a: int, b: str) -> typing.Tuple[int, str]:
        return a + 2, "world" + b

    wb = ImperativeWorkflow(name="my.workflow.a")
    t1_node = wb.add_entity(t1)
    t3_node = wb.add_entity(t3, a=t1_node.outputs["o0"], b=t1_node.outputs["o1"])
    wb.add_workflow_output("wf0", t3_node.outputs["o0"], python_type=int)
    wb.add_workflow_output("wf1", t3_node.outputs["o1"], python_type=str)
    res = wb()
    assert res == (5, "worldthree")

    with pytest.raises(KeyError):
        wb = ImperativeWorkflow(name="my.workflow.b")
        t1_node = wb.add_entity(t1)
        wb.add_entity(t3, a=t1_node.outputs["bad"], b=t1_node.outputs["o2"])


def test_call_normal():
    @task
    def t1(a: int) -> (int, str):
        return a + 2, "world"

    @workflow
    def my_functional_wf(a: int) -> (int, str):
        return t1(a=a)

    my_functional_lp = LaunchPlan.create("my_functional_wf.lp0", my_functional_wf, default_inputs={"a": 3})

    wb = ImperativeWorkflow(name="imperio")
    node = wb.add_entity(my_functional_wf, a=3)
    wb.add_workflow_output("from_n0_1", node.outputs["o0"])
    wb.add_workflow_output("from_n0_2", node.outputs["o1"])

    assert wb() == (5, "world")

    wb_lp = ImperativeWorkflow(name="imperio")
    node = wb_lp.add_entity(my_functional_lp)
    wb_lp.add_workflow_output("from_n0_1", node.outputs["o0"])
    wb_lp.add_workflow_output("from_n0_2", node.outputs["o1"])

    assert wb_lp() == (5, "world")


def test_imperative_call_from_normal():
    @task
    def t1(a: str) -> str:
        return a + " world"

    wb = ImperativeWorkflow(name="my.workflow")
    wb.add_workflow_input("in1", str)
    node = wb.add_entity(t1, a=wb.inputs["in1"])
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])

    assert wb(in1="hello") == "hello world"

    @workflow
    def my_functional_wf(a: str) -> str:
        x = wb(in1=a)
        return x

    assert my_functional_wf(a="hello") == "hello world"

    # Create launch plan from wf
    lp = LaunchPlan.create("test_wb_2", wb, fixed_inputs={"in1": "hello"})

    @workflow
    def my_functional_wf_lp() -> str:
        x = lp()
        return x

    assert my_functional_wf_lp() == "hello world"


def test_codecov():
    with pytest.raises(FlyteValidationException):
        get_promise(literal_models.BindingData(), {})

    with pytest.raises(FlyteValidationException):
        get_promise(literal_models.BindingData(promise=3), {})

    @task
    def t1(a: str) -> str:
        return a + " world"

    wb = ImperativeWorkflow(name="my.workflow")
    wb.add_workflow_input("in1", str)
    node = wb.add_entity(t1, a=wb.inputs["in1"])
    wb.add_workflow_output("from_n0t1", node.outputs["o0"])

    assert wb(in1="hello") == "hello world"

    with pytest.raises(AssertionError):
        wb(3)

    with pytest.raises(ValueError):
        wb(in2="hello")


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_non_function_task_and_df_input():
    import pandas as pd

    @reference_task(
        project="flytesnacks",
        domain="development",
        name="ref_t1",
        version="fast56d8ce2e373baf011f4d3532e45f0a9b",
    )
    def ref_t1(
        dataframe: pd.DataFrame,
        imputation_method: str = "median",
    ) -> pd.DataFrame:
        ...

    @reference_task(
        project="flytesnacks",
        domain="development",
        name="ref_t2",
        version="aedbd6fe44051c171fd966c280c5c3036f658831",
    )
    def ref_t2(
        dataframe: pd.DataFrame,
        split_mask: int,
        num_features: int,
    ) -> pd.DataFrame:
        ...

    wb = ImperativeWorkflow(name="core.feature_engineering.workflow.fe_wf")
    wb.add_workflow_input("sqlite_archive", FlyteFile[typing.TypeVar("sqlite")])
    sql_task = SQLite3Task(
        name="dummy.sqlite.task",
        query_template="select * from data",
        inputs=kwtypes(),
        output_schema_type=FlyteSchema,
        task_config=SQLite3Config(
            uri="https://sample/data",
            compressed=True,
        ),
    )
    node_sql = wb.add_task(sql_task)
    node_t1 = wb.add_task(ref_t1, dataframe=node_sql.outputs["results"], imputation_method="mean")

    node_t2 = wb.add_task(
        ref_t2,
        dataframe=node_t1.outputs["o0"],
        split_mask=24,
        num_features=15,
    )
    wb.add_workflow_output("output_from_t3", node_t2.outputs["o0"], python_type=pd.DataFrame)

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wb)
    assert len(wf_spec.template.nodes) == 3

    assert len(wf_spec.template.interface.inputs) == 1
    assert wf_spec.template.interface.inputs["sqlite_archive"].type.blob is not None

    assert len(wf_spec.template.interface.outputs) == 1
    assert wf_spec.template.interface.outputs["output_from_t3"].type.structured_dataset_type is not None
    assert wf_spec.template.interface.outputs["output_from_t3"].type.structured_dataset_type.format == ""
