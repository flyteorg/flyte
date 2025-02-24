import datetime
import os
import shutil
import tempfile
import typing
from unittest import mock

import pandas as pd
from click.testing import CliRunner
from flytekitplugins.awsbatch import AWSBatchConfig
from flytekitplugins.papermill import NotebookTask
from flytekitplugins.pod import Pod
from kubernetes.client import V1Container, V1PodSpec

import flytekit
from flytekit import StructuredDataset, kwtypes, map_task, task, workflow
from flytekit.clients.friendly import SynchronousFlyteClient
from flytekit.clis.sdk_in_container import pyflyte
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.remote import FlyteRemote
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile, PythonNotebook

from .testdata.datatype import X


def _get_nb_path(name: str, suffix: str = "", abs: bool = True, ext: str = ".ipynb") -> str:
    """
    Creates a correct path no matter where the test is run from
    """
    _local_path = os.path.dirname(__file__)
    path = f"{_local_path}/testdata/{name}{suffix}{ext}"
    return os.path.abspath(path) if abs else path


nb_name = "nb-simple"
nb_simple = NotebookTask(
    name="test",
    notebook_path=_get_nb_path(nb_name, abs=False),
    inputs=kwtypes(pi=float),
    outputs=kwtypes(square=float),
)

nb_sub_task = NotebookTask(
    name="test",
    notebook_path=_get_nb_path(nb_name, abs=False),
    inputs=kwtypes(a=float),
    outputs=kwtypes(square=float),
    output_notebooks=False,
)


def test_notebook_task_simple():
    serialization_settings = flytekit.configuration.SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env=None,
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
    )

    sqr, out, render = nb_simple.execute(pi=4)
    assert sqr == 16.0
    assert nb_simple.python_interface.inputs == {"pi": float}
    assert nb_simple.python_interface.outputs.keys() == {"square", "out_nb", "out_rendered_nb"}
    assert nb_simple.output_notebook_path == out == _get_nb_path(nb_name, suffix="-out")
    assert nb_simple.rendered_output_path == render == _get_nb_path(nb_name, suffix="-out", ext=".html")
    assert (
        nb_simple.get_command(settings=serialization_settings)
        == nb_simple.get_container(settings=serialization_settings).args
    )


def test_notebook_task_multi_values():
    nb_name = "nb-multi"
    nb = NotebookTask(
        name="test",
        notebook_path=_get_nb_path(nb_name, abs=False),
        inputs=kwtypes(x=int, y=int, h=str),
        outputs=kwtypes(z=int, m=int, h=str, n=datetime.datetime),
    )
    z, m, h, n, out, render = nb.execute(x=10, y=10, h="blah")
    assert z == 20
    assert m == 100
    assert h == "blah world!"
    assert type(n) == datetime.datetime
    assert nb.python_interface.inputs == {"x": int, "y": int, "h": str}
    assert nb.python_interface.outputs.keys() == {"z", "m", "h", "n", "out_nb", "out_rendered_nb"}
    assert nb.output_notebook_path == out == _get_nb_path(nb_name, suffix="-out")
    assert nb.rendered_output_path == render == _get_nb_path(nb_name, suffix="-out", ext=".html")


def test_notebook_task_complex():
    nb_name = "nb-complex"
    nb = NotebookTask(
        name="test",
        notebook_path=_get_nb_path(nb_name, abs=False),
        inputs=kwtypes(h=str, n=int, w=str),
        outputs=kwtypes(h=str, w=PythonNotebook, x=X),
    )
    h, w, x, out, render = nb.execute(h="blah", n=10, w=_get_nb_path("nb-multi"))
    assert h == "blah world!"
    assert w is not None
    assert x.x == 10
    assert nb.python_interface.inputs == {"n": int, "h": str, "w": str}
    assert nb.python_interface.outputs.keys() == {"h", "w", "x", "out_nb", "out_rendered_nb"}
    assert nb.output_notebook_path == out == _get_nb_path(nb_name, suffix="-out")
    assert nb.rendered_output_path == render == _get_nb_path(nb_name, suffix="-out", ext=".html")


def test_notebook_deck_local_execution_doesnt_fail():
    nb_name = "nb-simple"
    nb = NotebookTask(
        name="test",
        notebook_path=_get_nb_path(nb_name, abs=False),
        render_deck=True,
        inputs=kwtypes(pi=float),
        outputs=kwtypes(square=float),
    )
    sqr, out, render = nb.execute(pi=4)
    # This is largely a no assert test to ensure render_deck never inhibits local execution.
    assert nb._render_deck, "Passing render deck to init should result in private attribute being set"


def generate_por_spec_for_task():
    primary_container = V1Container(name="primary")
    pod_spec = V1PodSpec(containers=[primary_container])

    return pod_spec


nb = NotebookTask(
    name="test",
    task_config=Pod(pod_spec=generate_por_spec_for_task(), primary_container_name="primary"),
    notebook_path=_get_nb_path("nb-simple", abs=False),
    inputs=kwtypes(h=str, n=int, w=str),
    outputs=kwtypes(h=str, w=PythonNotebook, x=X),
)


def test_notebook_pod_task():
    serialization_settings = flytekit.configuration.SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env=None,
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
    )

    assert nb.get_container(serialization_settings) is None
    assert nb.get_config(serialization_settings)["primary_container_name"] == "primary"
    assert (
        nb.get_command(serialization_settings)
        == nb.get_k8s_pod(serialization_settings).pod_spec["containers"][0]["args"]
    )


nb_batch = NotebookTask(
    name="simple-nb",
    task_config=AWSBatchConfig(platformCapabilities="EC2"),
    notebook_path=_get_nb_path("nb-simple", abs=False),
    inputs=kwtypes(h=str, n=int, w=str),
    outputs=kwtypes(h=str, w=PythonNotebook, x=X),
)


def test_notebook_batch_task():
    serialization_settings = flytekit.configuration.SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env=None,
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
    )

    assert nb_batch.get_container(serialization_settings) is not None
    assert nb_batch.get_container(serialization_settings).args == [
        "pyflyte-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}/0",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "--",
        "task-module",
        "tests.test_task",
        "task-name",
        "nb_batch",
    ]


def test_flyte_types():
    @task
    def create_file() -> FlyteFile:
        tmp_file = tempfile.mktemp()
        with open(tmp_file, "w") as f:
            f.write("abc")
        return FlyteFile(path=tmp_file)

    @task
    def create_dir() -> FlyteDirectory:
        tmp_dir = tempfile.mkdtemp()
        with open(os.path.join(tmp_dir, "file.txt"), "w") as f:
            f.write("abc")
        return FlyteDirectory(path=tmp_dir)

    @task
    def create_sd() -> StructuredDataset:
        df = pd.DataFrame({"a": [1, 2], "b": [3, 4]})
        return StructuredDataset(dataframe=df)

    ff = create_file()
    fd = create_dir()
    sd = create_sd()

    nb_name = "nb-types"
    nb_types = NotebookTask(
        name="test",
        notebook_path=_get_nb_path(nb_name, abs=False),
        inputs=kwtypes(ff=FlyteFile, fd=FlyteDirectory, sd=StructuredDataset),
        outputs=kwtypes(success=bool),
    )
    success, out, render = nb_types.execute(ff=ff, fd=fd, sd=sd)
    assert success is True, "Notebook execution failed"


def test_map_over_notebook_task():
    @workflow
    def wf(a: float) -> typing.List[float]:
        return map_task(nb_sub_task)(a=[a, a])

    assert wf(a=3.14) == [9.8596, 9.8596]


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient", spec=SynchronousFlyteClient)
def test_register_notebook_task(mock_client, mock_remote):
    mock_remote._client = mock_client
    mock_remote.return_value._version_from_hash.return_value = "dummy_version_from_hash"
    mock_remote.return_value.fast_package.return_value = "dummy_md5_bytes", "dummy_native_url"
    runner = CliRunner()
    context_manager.FlyteEntities.entities.clear()
    notebook_task = """
from flytekitplugins.papermill import NotebookTask

nb_simple = NotebookTask(
    name="test",
    notebook_path="./core/notebook.ipython",
)
"""
    with runner.isolated_filesystem():
        os.makedirs("core", exist_ok=True)
        with open(os.path.join("core", "notebook.ipython"), "w") as f:
            f.write("notebook.ipython")
            f.close()
        with open(os.path.join("core", "notebook_task.py"), "w") as f:
            f.write(notebook_task)
            f.close()
        result = runner.invoke(pyflyte.main, ["register", "core"])
        assert "Successfully registered 2 entities" in result.output
        shutil.rmtree("core")
