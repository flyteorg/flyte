import tempfile

from flytekit import task, workflow
from flytekit.configuration import ImageConfig, SerializationSettings
from flytekit.sensor.file_sensor import FileSensor
from tests.flytekit.unit.test_translator import default_img


def test_sensor_task():
    sensor = FileSensor(name="test_sensor")
    assert sensor.task_type == "sensor"
    settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )
    assert sensor.get_custom(settings) == {"sensor_module": "flytekit.sensor.file_sensor", "sensor_name": "FileSensor"}
    tmp_file = tempfile.NamedTemporaryFile()

    @task()
    def t1():
        print("flyte")

    @workflow
    def wf():
        sensor(tmp_file.name) >> t1()

    if __name__ == "__main__":
        wf()
