from flytekitplugins.spark import Spark
from mock import MagicMock

from flytekit import task
from flytekit.configuration import Config, SerializationSettings
from flytekit.remote.remote import FlyteRemote


def test_spark_template_with_remote():
    @task(task_config=Spark(spark_conf={"spark": "1"}))
    def my_spark(a: str) -> int:
        return 10

    @task
    def my_python_task(a: str) -> int:
        return 10

    remote = FlyteRemote(
        config=Config.for_endpoint(endpoint="localhost", insecure=True), default_project="p1", default_domain="d1"
    )

    mock_client = MagicMock()
    remote._client = mock_client
    remote._client_initialized = True

    remote.register_task(
        my_spark,
        serialization_settings=SerializationSettings(
            image_config=MagicMock(),
        ),
        version="v1",
    )
    serialized_spec = mock_client.create_task.call_args.kwargs["task_spec"]

    print(serialized_spec)
    # Check if the serialized spark task has mainApplicaitonFile field set.
    assert serialized_spec.template.custom["mainApplicationFile"]
    assert serialized_spec.template.custom["sparkConf"]

    remote.register_task(
        my_python_task, serialization_settings=SerializationSettings(image_config=MagicMock()), version="v1"
    )
    serialized_spec = mock_client.create_task.call_args.kwargs["task_spec"]

    # Check if the serialized python task has no mainApplicaitonFile field set by default.
    assert serialized_spec.template.custom is None
