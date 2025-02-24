import datetime
import json
import os
import pathlib
import subprocess
import time
import typing

import joblib
import pytest

from flytekit import LaunchPlan, kwtypes
from flytekit.configuration import Config, ImageConfig, SerializationSettings
from flytekit.exceptions.user import FlyteAssertion, FlyteEntityNotExistException
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.remote.remote import FlyteRemote
from flytekit.types.schema import FlyteSchema

MODULE_PATH = pathlib.Path(__file__).parent / "workflows/basic"
CONFIG = os.environ.get("FLYTECTL_CONFIG", str(pathlib.Path.home() / ".flyte" / "config-sandbox.yaml"))
IMAGE = os.environ.get("FLYTEKIT_IMAGE", "localhost:30000/flytekit:dev")
PROJECT = "flytesnacks"
DOMAIN = "development"
VERSION = f"v{os.getpid()}"


@pytest.fixture(scope="session")
def register():
    subprocess.run(
        [
            "pyflyte",
            "-c",
            CONFIG,
            "register",
            "--image",
            IMAGE,
            "--project",
            PROJECT,
            "--domain",
            DOMAIN,
            "--version",
            VERSION,
            MODULE_PATH,
        ]
    )


def test_fetch_execute_launch_plan(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_launch_plan = remote.fetch_launch_plan(name="basic.hello_world.my_wf", version=VERSION)
    execution = remote.execute(flyte_launch_plan, inputs={}, wait=True)
    assert execution.outputs["o0"] == "hello world"


def fetch_execute_launch_plan_with_args(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_launch_plan = remote.fetch_launch_plan(name="basic.basic_workflow.my_wf", version=VERSION)
    execution = remote.execute(flyte_launch_plan, inputs={"a": 10, "b": "foobar"}, wait=True)
    assert execution.node_executions["n0"].inputs == {"a": 10}
    assert execution.node_executions["n0"].outputs == {"t1_int_output": 12, "c": "world"}
    assert execution.node_executions["n1"].inputs == {"a": "world", "b": "foobar"}
    assert execution.node_executions["n1"].outputs == {"o0": "foobarworld"}
    assert execution.node_executions["n0"].task_executions[0].inputs == {"a": 10}
    assert execution.node_executions["n0"].task_executions[0].outputs == {"t1_int_output": 12, "c": "world"}
    assert execution.node_executions["n1"].task_executions[0].inputs == {"a": "world", "b": "foobar"}
    assert execution.node_executions["n1"].task_executions[0].outputs == {"o0": "foobarworld"}
    assert execution.inputs["a"] == 10
    assert execution.inputs["b"] == "foobar"
    assert execution.outputs["o0"] == 12
    assert execution.outputs["o1"] == "foobarworld"


def test_monitor_workflow_execution(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_launch_plan = remote.fetch_launch_plan(name="basic.hello_world.my_wf", version=VERSION)
    execution = remote.execute(
        flyte_launch_plan,
        inputs={},
    )

    poll_interval = datetime.timedelta(seconds=1)
    time_to_give_up = datetime.datetime.utcnow() + datetime.timedelta(seconds=60)

    execution = remote.sync_execution(execution, sync_nodes=True)
    while datetime.datetime.utcnow() < time_to_give_up:
        if execution.is_done:
            break

        with pytest.raises(
            FlyteAssertion, match="Please wait until the execution has completed before requesting the outputs."
        ):
            execution.outputs

        time.sleep(poll_interval.total_seconds())
        execution = remote.sync_execution(execution, sync_nodes=True)

        if execution.node_executions:
            assert execution.node_executions["start-node"].closure.phase == 3  # SUCCEEDED

    for key in execution.node_executions:
        assert execution.node_executions[key].closure.phase == 3

    assert execution.node_executions["n0"].inputs == {}
    assert execution.node_executions["n0"].outputs["o0"] == "hello world"
    assert execution.node_executions["n0"].task_executions[0].inputs == {}
    assert execution.node_executions["n0"].task_executions[0].outputs["o0"] == "hello world"
    assert execution.inputs == {}
    assert execution.outputs["o0"] == "hello world"


def test_fetch_execute_launch_plan_with_subworkflows(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)

    flyte_launch_plan = remote.fetch_launch_plan(name="basic.subworkflows.parent_wf", version=VERSION)
    execution = remote.execute(flyte_launch_plan, inputs={"a": 101}, wait=True)
    # check node execution inputs and outputs
    assert execution.node_executions["n0"].inputs == {"a": 101}
    assert execution.node_executions["n0"].outputs == {"t1_int_output": 103, "c": "world"}
    assert execution.node_executions["n1"].inputs == {"a": 103}
    assert execution.node_executions["n1"].outputs == {"o0": "world", "o1": "world"}

    # check subworkflow task execution inputs and outputs
    subworkflow_node_executions = execution.node_executions["n1"].subworkflow_node_executions
    subworkflow_node_executions["n1-0-n0"].inputs == {"a": 103}
    subworkflow_node_executions["n1-0-n1"].outputs == {"t1_int_output": 107, "c": "world"}


def test_fetch_execute_launch_plan_with_child_workflows(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)

    flyte_launch_plan = remote.fetch_launch_plan(name="basic.child_workflow.parent_wf", version=VERSION)
    execution = remote.execute(flyte_launch_plan, inputs={"a": 3}, wait=True)

    # check node execution inputs and outputs
    assert execution.node_executions["n0"].inputs == {"a": 3}
    assert execution.node_executions["n0"].outputs["o0"] == 6
    assert execution.node_executions["n1"].inputs == {"a": 6}
    assert execution.node_executions["n1"].outputs["o0"] == 12
    assert execution.node_executions["n2"].inputs == {"a": 6, "b": 12}
    assert execution.node_executions["n2"].outputs["o0"] == 18


def test_fetch_execute_workflow(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_workflow = remote.fetch_workflow(name="basic.hello_world.my_wf", version=VERSION)
    execution = remote.execute(flyte_workflow, inputs={}, wait=True)
    assert execution.outputs["o0"] == "hello world"
    assert isinstance(execution.closure.duration, datetime.timedelta)
    assert execution.closure.duration > datetime.timedelta(seconds=1)

    execution_to_terminate = remote.execute(flyte_workflow, {})
    remote.terminate(execution_to_terminate, cause="just because")


def test_fetch_execute_task(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_task = remote.fetch_task(name="basic.basic_workflow.t1", version=VERSION)
    execution = remote.execute(flyte_task, inputs={"a": 10}, wait=True)
    assert execution.outputs["t1_int_output"] == 12
    assert execution.outputs["c"] == "world"
    assert execution.inputs["a"] == 10
    assert execution.outputs["c"] == "world"


def test_execute_python_task(register):
    """Test execution of a @task-decorated python function that is already registered."""
    from .workflows.basic.basic_workflow import t1

    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    execution = remote.execute(
        t1,
        name="basic.basic_workflow.t1",
        inputs={"a": 10},
        version=VERSION,
        wait=True,
        overwrite_cache=True,
        envs={"foo": "bar"},
        tags=["flyte"],
        cluster_pool="gpu",
    )
    assert execution.outputs["t1_int_output"] == 12
    assert execution.outputs["c"] == "world"
    assert execution.spec.envs.envs == {"foo": "bar"}
    assert execution.spec.tags == ["flyte"]
    assert execution.spec.cluster_assignment.cluster_pool == "gpu"


def test_execute_python_workflow_and_launch_plan(register):
    """Test execution of a @workflow-decorated python function and launchplan that are already registered."""
    from .workflows.basic.basic_workflow import my_wf

    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    execution = remote.execute(
        my_wf, name="basic.basic_workflow.my_wf", inputs={"a": 10, "b": "xyz"}, version=VERSION, wait=True
    )
    assert execution.outputs["o0"] == 12
    assert execution.outputs["o1"] == "xyzworld"

    launch_plan = LaunchPlan.get_or_create(workflow=my_wf, name=my_wf.name)
    execution = remote.execute(
        launch_plan, name="basic.basic_workflow.my_wf", inputs={"a": 14, "b": "foobar"}, version=VERSION, wait=True
    )
    assert execution.outputs["o0"] == 16
    assert execution.outputs["o1"] == "foobarworld"

    flyte_workflow_execution = remote.fetch_execution(name=execution.id.name)
    assert execution.inputs == flyte_workflow_execution.inputs
    assert execution.outputs == flyte_workflow_execution.outputs


def test_fetch_execute_launch_plan_list_of_floats(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_launch_plan = remote.fetch_launch_plan(name="basic.list_float_wf.my_wf", version=VERSION)
    xs: typing.List[float] = [42.24, 999.1, 0.0001]
    execution = remote.execute(flyte_launch_plan, inputs={"xs": xs}, wait=True)
    assert execution.outputs["o0"] == "[42.24, 999.1, 0.0001]"


def test_fetch_execute_task_list_of_floats(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_task = remote.fetch_task(name="basic.list_float_wf.concat_list", version=VERSION)
    xs: typing.List[float] = [0.1, 0.2, 0.3, 0.4, -99999.7]
    execution = remote.execute(flyte_task, inputs={"xs": xs}, wait=True)
    assert execution.outputs["o0"] == "[0.1, 0.2, 0.3, 0.4, -99999.7]"


def test_fetch_execute_task_convert_dict(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_task = remote.fetch_task(name="basic.dict_str_wf.convert_to_string", version=VERSION)
    d: typing.Dict[str, str] = {"key1": "value1", "key2": "value2"}
    execution = remote.execute(flyte_task, inputs={"d": d}, wait=True)
    remote.sync_execution(execution, sync_nodes=True)
    assert json.loads(execution.outputs["o0"]) == {"key1": "value1", "key2": "value2"}


def test_execute_python_workflow_dict_of_string_to_string(register):
    """Test execution of a @workflow-decorated python function and launchplan that are already registered."""
    from .workflows.basic.dict_str_wf import my_wf

    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    d: typing.Dict[str, str] = {"k1": "v1", "k2": "v2"}
    execution = remote.execute(
        my_wf,
        name="basic.dict_str_wf.my_wf",
        inputs={"d": d},
        version=VERSION,
        wait=True,
    )
    assert json.loads(execution.outputs["o0"]) == {"k1": "v1", "k2": "v2"}

    launch_plan = LaunchPlan.get_or_create(workflow=my_wf, name=my_wf.name)
    execution = remote.execute(
        launch_plan,
        name="basic.dict_str_wf.my_wf",
        inputs={"d": {"k2": "vvvv", "abc": "def"}},
        version=VERSION,
        wait=True,
    )
    assert json.loads(execution.outputs["o0"]) == {"k2": "vvvv", "abc": "def"}


def test_execute_python_workflow_list_of_floats(register):
    """Test execution of a @workflow-decorated python function and launchplan that are already registered."""
    from .workflows.basic.list_float_wf import my_wf

    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)

    xs: typing.List[float] = [42.24, 999.1, 0.0001]
    execution = remote.execute(
        my_wf,
        name="basic.list_float_wf.my_wf",
        inputs={"xs": xs},
        version=VERSION,
        wait=True,
    )
    assert execution.outputs["o0"] == "[42.24, 999.1, 0.0001]"

    launch_plan = LaunchPlan.get_or_create(workflow=my_wf, name=my_wf.name)
    execution = remote.execute(
        launch_plan,
        name="basic.list_float_wf.my_wf",
        inputs={"xs": [-1.1, 0.12345]},
        version=VERSION,
        wait=True,
    )
    assert execution.outputs["o0"] == "[-1.1, 0.12345]"


@pytest.mark.skip(reason="Waiting for https://github.com/flyteorg/flytectl/pull/440 to land")
def test_execute_sqlite3_task(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)

    example_db = "https://www.sqlitetutorial.net/wp-content/uploads/2018/03/chinook.zip"
    interactive_sql_task = SQLite3Task(
        "basic_querying",
        query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
        task_config=SQLite3Config(
            uri=example_db,
            compressed=True,
        ),
    )
    registered_sql_task = remote.register_task(
        interactive_sql_task,
        serialization_settings=SerializationSettings(image_config=ImageConfig.auto(img_name=IMAGE)),
        version=VERSION,
    )
    execution = remote.execute(registered_sql_task, inputs={"limit": 10}, wait=True)
    output = execution.outputs["results"]
    result = output.open().all()
    assert result.__class__.__name__ == "DataFrame"
    assert "TrackId" in result
    assert "Name" in result


@pytest.mark.skip(reason="Waiting for https://github.com/flyteorg/flytectl/pull/440 to land")
def test_execute_joblib_workflow(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    flyte_workflow = remote.fetch_workflow(name="basic.joblib.joblib_workflow", version=VERSION)
    input_obj = [1, 2, 3]
    execution = remote.execute(flyte_workflow, inputs={"obj": input_obj}, wait=True)
    joblib_output = execution.outputs["o0"]
    joblib_output.download()
    output_obj = joblib.load(joblib_output.path)
    assert execution.outputs["o0"].extension() == "joblib"
    assert output_obj == input_obj


def test_execute_with_default_launch_plan(register):
    from .workflows.basic.subworkflows import parent_wf

    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    execution = remote.execute(
        parent_wf, inputs={"a": 101}, version=VERSION, wait=True, image_config=ImageConfig.auto(img_name=IMAGE)
    )
    # check node execution inputs and outputs
    assert execution.node_executions["n0"].inputs == {"a": 101}
    assert execution.node_executions["n0"].outputs == {"t1_int_output": 103, "c": "world"}
    assert execution.node_executions["n1"].inputs == {"a": 103}
    assert execution.node_executions["n1"].outputs == {"o0": "world", "o1": "world"}

    # check subworkflow task execution inputs and outputs
    subworkflow_node_executions = execution.node_executions["n1"].subworkflow_node_executions
    subworkflow_node_executions["n1-0-n0"].inputs == {"a": 103}
    subworkflow_node_executions["n1-0-n1"].outputs == {"t1_int_output": 107, "c": "world"}


def test_fetch_not_exist_launch_plan(register):
    remote = FlyteRemote(Config.auto(config_file=CONFIG), PROJECT, DOMAIN)
    with pytest.raises(FlyteEntityNotExistException):
        remote.fetch_launch_plan(name="basic.list_float_wf.fake_wf", version=VERSION)
