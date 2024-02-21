import pytest
from flyteidl.core import catalog_pb2, errors_pb2

from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.task import task
from flytekit.core.type_engine import TypeEngine
from flytekit.core.workflow import workflow
from flytekit.models.types import LiteralType, SimpleType


def test_proto():
    @task
    def t1(in1: errors_pb2.ContainerError) -> errors_pb2.ContainerError:
        e2 = errors_pb2.ContainerError(code=in1.code, message=in1.message + "!!!", kind=in1.kind + 1)
        return e2

    @workflow
    def wf(a: errors_pb2.ContainerError) -> errors_pb2.ContainerError:
        return t1(in1=a)

    e1 = errors_pb2.ContainerError(code="test", message="hello world", kind=1)
    e_out = wf(a=e1)
    assert e_out.kind == 2
    assert e_out.message == "hello world!!!"


def test_pb_guess_python_type():
    artifact_tag = catalog_pb2.CatalogArtifactTag(artifact_id="artifact_1", name="artifact_name")

    x = {"a": artifact_tag}
    lt = TypeEngine.to_literal_type(catalog_pb2.CatalogArtifactTag)
    gt = TypeEngine.guess_python_type(lt)
    assert gt == catalog_pb2.CatalogArtifactTag
    ctx = FlyteContextManager.current_context()
    lm = TypeEngine.dict_to_literal_map(ctx, x, {"a": gt})
    pv = TypeEngine.to_python_value(ctx, lm.literals["a"], gt)
    assert pv == artifact_tag


def test_bad_tag():
    # Will not be able to load this
    with pytest.raises(ValueError):
        lt = LiteralType(simple=SimpleType.STRUCT, metadata={"pb_type": "bad.tag"})
        TypeEngine.guess_python_type(lt)

    # Doesn't match pb field key
    with pytest.raises(ValueError):
        lt = LiteralType(simple=SimpleType.STRUCT, metadata={})
        TypeEngine.guess_python_type(lt)


def test_workflow():
    @task
    def grab_catalog_artifact(artifact_id: str, artifact_name: str) -> catalog_pb2.CatalogArtifactTag:
        return catalog_pb2.CatalogArtifactTag(artifact_id=artifact_id, name=artifact_name)

    @workflow
    def wf(artifact_id: str, artifact_name: str) -> catalog_pb2.CatalogArtifactTag:
        return grab_catalog_artifact(artifact_id=artifact_id, artifact_name=artifact_name)

    catalog_artifact = wf(artifact_id="id-1", artifact_name="some-name")
    assert catalog_artifact.artifact_id == "id-1"
    assert catalog_artifact.name == "some-name"
