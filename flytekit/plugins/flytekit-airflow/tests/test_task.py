import jsonpickle
from airflow.operators.bash import BashOperator
from airflow.providers.apache.beam.operators.beam import BeamRunJavaPipelineOperator, BeamRunPythonPipelineOperator
from airflow.providers.google.cloud.operators.dataproc import DataprocCreateClusterOperator
from airflow.sensors.bash import BashSensor
from airflow.sensors.time_sensor import TimeSensor
from airflow.utils.context import Context
from flytekitplugins.airflow.task import (
    AirflowContainerTask,
    AirflowObj,
    AirflowTask,
    _flyte_operator,
    _is_deferrable,
    airflow_task_resolver,
)
from mock import mock

from flytekit import FlyteContextManager
from flytekit.configuration import ImageConfig, SerializationSettings
from flytekit.core import context_manager


def test_xcom_push():
    ctx = FlyteContextManager.current_context()
    ctx.user_space_params._attrs = {}

    execution_state = ctx.execution_state.with_params(
        user_space_params=ctx.user_space_params.new_builder()
        .add_attr("GET_ORIGINAL_TASK", True)
        .add_attr("XCOM_DATA", {})
        .build()
    )

    with FlyteContextManager.with_context(ctx.with_execution_state(execution_state)) as child_ctx:
        print(child_ctx.user_space_params.get_original_task)
        op = BashSensor(task_id="Sensor_succeeds", bash_command="exit 0")
        op.xcom_push(Context(), "key", "value")
        assert child_ctx.user_space_params.xcom_data[1] == "key"
        assert child_ctx.user_space_params.xcom_data[2] == "value"


def test_is_deferrable():
    assert _is_deferrable(BeamRunJavaPipelineOperator) is False
    assert _is_deferrable(BashSensor) is False
    assert _is_deferrable(DataprocCreateClusterOperator) is True


def test_airflow_task():
    cfg = AirflowObj(
        module="airflow.operators.bash",
        name="BashOperator",
        parameters={"task_id": "id", "bash_command": "echo 'hello world'"},
    )
    t = AirflowTask(name="test_bash_operator", task_config=cfg)
    serialization_settings = SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig.auto(),
        env={},
    )
    t.get_custom(serialization_settings)["task_config_pkl"] = jsonpickle.encode(cfg)
    t.execute()


def test_airflow_container_task():
    cfg = AirflowObj(
        module="airflow.providers.apache.beam.operators.beam",
        name="BeamRunJavaPipelineOperator",
        parameters={"task_id": "id", "job_class": "org.apache.beam.examples.WordCount"},
    )
    t = AirflowContainerTask(name="test_dataflow_operator", task_config=cfg)
    serialization_settings = SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig.auto(),
        env={},
    )
    assert t.task_resolver.name() == "AirflowTaskResolver"
    assert t.task_resolver.loader_args(serialization_settings, t) == [
        "task-module",
        "flytekitplugins.airflow.task",
        "task-name",
        "AirflowContainerTask",
        "task-config",
        '{"py/object": "flytekitplugins.airflow.task.AirflowObj", "module": '
        '"airflow.providers.apache.beam.operators.beam", "name": '
        '"BeamRunJavaPipelineOperator", "parameters": {"task_id": "id", "job_class": '
        '"org.apache.beam.examples.WordCount"}}',
    ]
    assert isinstance(
        airflow_task_resolver.load_task(t.task_resolver.loader_args(serialization_settings, t)), AirflowContainerTask
    )


@mock.patch("flytekitplugins.airflow.task.AirflowContainerTask")
@mock.patch("flytekitplugins.airflow.task.AirflowTask")
def test_flyte_operator(airflow_task, airflow_container_task):
    ctx = context_manager.FlyteContext.current_context()
    with context_manager.FlyteContextManager.with_context(ctx.new_builder()):
        params = FlyteContextManager.current_context().user_space_params
        params.builder().add_attr("GET_ORIGINAL_TASK", False).add_attr("XCOM_DATA", {}).build()
        _flyte_operator(BashOperator, task_id="BashOperator")
        airflow_task.assert_called_once()
        _flyte_operator(BeamRunJavaPipelineOperator, task_id="BeamRunJavaPipelineOperator")
        airflow_container_task.assert_called_once()

        airflow_task.reset_mock()
        airflow_container_task.reset_mock()

        _flyte_operator(TimeSensor, task_id="TimeSensor")
        airflow_task.assert_called_once()

        _flyte_operator(BeamRunPythonPipelineOperator, task_id="BeamRunPythonPipelineOperator")
        airflow_container_task.assert_called_once()

        airflow_task.reset_mock()
        airflow_container_task.reset_mock()

        _flyte_operator(DataprocCreateClusterOperator, task_id="DataprocCreateClusterOperator")
        airflow_task.assert_called_once()
