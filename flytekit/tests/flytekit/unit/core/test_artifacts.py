import datetime
import sys
from collections import OrderedDict

import pytest
from flyteidl.core import artifact_id_pb2 as art_id
from typing_extensions import Annotated, get_args

from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.artifact import Artifact, Inputs
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.interface import detect_artifact
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.exceptions.user import FlyteValidationException
from flytekit.tools.translator import get_serializable

if "pandas" not in sys.modules:
    pytest.skip(reason="Requires pandas", allow_module_level=True)


default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


class CustomReturn(object):
    def __init__(self, data):
        self.data = data


def test_basic_option_a_rev():
    import pandas as pd

    a1_t_ab = Artifact(name="my_data", partition_keys=["a", "b"], time_partitioned=True)

    @task
    def t1(
        b_value: str, dt: datetime.datetime
    ) -> Annotated[pd.DataFrame, a1_t_ab(time_partition=Inputs.dt, b=Inputs.b_value, a="manual")]:
        df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
        return df

    entities = OrderedDict()
    t1_s = get_serializable(entities, serialization_settings, t1)
    assert len(t1_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value) == 2
    p = t1_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.time_partition is not None
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.time_partition.value.input_binding.var == "dt"
    assert p["b"].HasField("input_binding")
    assert p["b"].input_binding.var == "b_value"
    assert p["a"].HasField("static_value")
    assert p["a"].static_value == "manual"
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.version == ""
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.name == "my_data"
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.project == ""


def test_args_getting():
    a1 = Artifact(name="argstst")
    a1_called = a1()
    x = Annotated[int, a1_called]
    gotten = get_args(x)
    assert len(gotten) == 2
    assert gotten[1] is a1_called
    detected = detect_artifact(get_args(int))
    assert detected is None
    detected = detect_artifact(get_args(x))
    assert detected == a1_called.to_partial_artifact_id()


def test_basic_option_no_tp():
    import pandas as pd

    a1_t_ab = Artifact(name="my_data", partition_keys=["a", "b"])
    assert not a1_t_ab.time_partitioned

    # trying to bind to a time partition when not so raises an error.
    with pytest.raises(ValueError):

        @task
        def t1x(
            b_value: str, dt: datetime.datetime
        ) -> Annotated[pd.DataFrame, a1_t_ab(time_partition=Inputs.dt, b=Inputs.b_value, a="manual")]:
            df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
            return df

    @task
    def t1(b_value: str, dt: datetime.datetime) -> Annotated[pd.DataFrame, a1_t_ab(b=Inputs.b_value, a="manual")]:
        df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
        return df

    entities = OrderedDict()
    t1_s = get_serializable(entities, serialization_settings, t1)
    assert len(t1_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value) == 2
    p = t1_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.HasField("time_partition") is False
    assert p["b"].HasField("input_binding")


def test_basic_option_hardcoded_tp():
    a1_t_ab = Artifact(name="my_data", time_partitioned=True)

    dt = datetime.datetime.strptime("04/05/2063", "%m/%d/%Y")

    id_spec = a1_t_ab(time_partition=dt)
    assert id_spec.partitions is None
    assert id_spec.time_partition.value.HasField("time_value")


def test_basic_option_a():
    import pandas as pd

    a1_t_ab = Artifact(name="my_data", partition_keys=["a", "b"], time_partitioned=True)

    @task
    def t1(
        b_value: str, dt: datetime.datetime
    ) -> Annotated[pd.DataFrame, a1_t_ab(b=Inputs.b_value, a="manual", time_partition=Inputs.dt)]:
        df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
        return df

    entities = OrderedDict()
    t1_s = get_serializable(entities, serialization_settings, t1)
    assert len(t1_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value) == 2
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.version == ""
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.name == "my_data"
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.project == ""
    assert t1_s.template.interface.outputs["o0"].artifact_partial_id.time_partition is not None


def test_basic_no_call():
    import pandas as pd

    a1_t_ab = Artifact(name="my_data", partition_keys=["a", "b"], time_partitioned=True)

    # raise an error because the user hasn't () the artifact
    with pytest.raises(ValueError):

        @task
        def t1(b_value: str, dt: datetime.datetime) -> Annotated[pd.DataFrame, a1_t_ab]:
            df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
            return df


def test_basic_option_a2():
    import pandas as pd

    a2_ab = Artifact(name="my_data2", partition_keys=["a", "b"])

    with pytest.raises(ValueError):

        @task
        def t2x(b_value: str) -> Annotated[pd.DataFrame, a2_ab(a=Inputs.b_value)]:
            ...

    @task
    def t2(b_value: str) -> Annotated[pd.DataFrame, a2_ab(a=Inputs.b_value, b="manualval")]:
        ...

    entities = OrderedDict()
    t2_s = get_serializable(entities, serialization_settings, t2)
    assert len(t2_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value) == 2
    assert t2_s.template.interface.outputs["o0"].artifact_partial_id.version == ""
    assert t2_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.name == "my_data2"
    assert t2_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.project == ""


def test_basic_option_a3():
    import pandas as pd

    a3 = Artifact(name="my_data3")

    @task
    def t3(b_value: str) -> Annotated[pd.DataFrame, a3]:
        ...

    entities = OrderedDict()
    t3_s = get_serializable(entities, serialization_settings, t3)
    assert len(t3_s.template.interface.outputs["o0"].artifact_partial_id.partitions.value) == 0
    assert t3_s.template.interface.outputs["o0"].artifact_partial_id.artifact_key.name == "my_data3"


def test_query_basic():
    aa = Artifact(
        name="ride_count_data",
        time_partitioned=True,
        partition_keys=["region"],
    )
    data_query = aa.query(time_partition=Inputs.dt, region=Inputs.blah)
    assert data_query.bindings == []
    assert data_query.artifact is aa
    dq_idl = data_query.to_flyte_idl()
    assert dq_idl.HasField("artifact_id")
    assert dq_idl.artifact_id.artifact_key.name == "ride_count_data"
    assert len(dq_idl.artifact_id.partitions.value) == 1
    assert dq_idl.artifact_id.partitions.value["region"].HasField("input_binding")
    assert dq_idl.artifact_id.partitions.value["region"].input_binding.var == "blah"
    assert dq_idl.artifact_id.time_partition.value.input_binding.var == "dt"


def test_not_specified_behavior():
    wf_artifact_no_tag = Artifact(project="project1", domain="dev", name="wf_artifact", version="1", partitions=None)
    aq = wf_artifact_no_tag.query("pr", "dom").to_flyte_idl()
    assert aq.artifact_id.HasField("partitions") is False
    assert aq.artifact_id.artifact_key.project == "pr"
    assert aq.artifact_id.artifact_key.domain == "dom"

    assert wf_artifact_no_tag.concrete_artifact_id.HasField("partitions") is False

    wf_artifact_no_tag = Artifact(project="project1", domain="dev", name="wf_artifact", partitions={})
    assert wf_artifact_no_tag.partitions is None
    aq = wf_artifact_no_tag.query().to_flyte_idl()
    assert aq.artifact_id.HasField("partitions") is False
    assert aq.artifact_id.HasField("time_partition") is False


def test_artifact_as_promise_query():
    # when artifact is partially specified, can be used as a query input
    wf_artifact = Artifact(project="project1", domain="dev", name="wf_artifact")

    @task
    def t1(a: CustomReturn) -> CustomReturn:
        print(a)
        return CustomReturn({"name": ["Tom", "Joseph"], "age": [20, 22]})

    @workflow
    def wf(a: CustomReturn = wf_artifact.query()):
        u = t1(a=a)
        return u

    ctx = FlyteContextManager.current_context()
    lp = LaunchPlan.get_default_launch_plan(ctx, wf)
    entities = OrderedDict()
    spec = get_serializable(entities, serialization_settings, lp)
    assert spec.spec.default_inputs.parameters["a"].artifact_query.artifact_id.artifact_key.project == "project1"
    assert spec.spec.default_inputs.parameters["a"].artifact_query.artifact_id.artifact_key.domain == "dev"
    assert spec.spec.default_inputs.parameters["a"].artifact_query.artifact_id.artifact_key.name == "wf_artifact"


def test_artifact_as_promise():
    # when the full artifact is specified, the artifact should be bindable as a literal
    wf_artifact = Artifact(project="pro", domain="dom", name="key", version="v0.1.0", partitions={"region": "LAX"})

    @task
    def t1(a: CustomReturn) -> CustomReturn:
        print(a)
        return CustomReturn({"name": ["Tom", "Joseph"], "age": [20, 22]})

    @workflow
    def wf2(a: CustomReturn = wf_artifact):
        u = t1(a=a)
        return u

    ctx = FlyteContextManager.current_context()
    lp = LaunchPlan.get_default_launch_plan(ctx, wf2)
    entities = OrderedDict()
    spec = get_serializable(entities, serialization_settings, lp)
    assert spec.spec.default_inputs.parameters["a"].artifact_id.artifact_key.project == "pro"
    assert spec.spec.default_inputs.parameters["a"].artifact_id.artifact_key.domain == "dom"
    assert spec.spec.default_inputs.parameters["a"].artifact_id.artifact_key.name == "key"

    aq = wf_artifact.query().to_flyte_idl()
    assert aq.artifact_id.HasField("partitions") is True
    assert aq.artifact_id.partitions.value["region"].static_value == "LAX"


def test_partition_none():
    # confirm that we can distinguish between partitions being set to empty, and not being set
    # though this is not currently used.
    ak = art_id.ArtifactKey(project="p", domain="d", name="name")
    no_partition = art_id.ArtifactID(artifact_key=ak, version="without_p")
    assert not no_partition.HasField("partitions")

    p = art_id.Partitions()
    with_partition = art_id.ArtifactID(artifact_key=ak, version="without_p", partitions=p)
    assert with_partition.HasField("partitions")


def test_as_artf_no_partitions():
    int_artf = Artifact(name="important_int")

    @task
    def greet(day_of_week: str, number: int, am: bool) -> str:
        greeting = "Have a great " + day_of_week + " "
        greeting += "morning" if am else "evening"
        return greeting + "!" * number

    @workflow
    def go_greet(day_of_week: str, number: int = int_artf.query(), am: bool = False) -> str:
        return greet(day_of_week=day_of_week, number=number, am=am)

    tst_lp = LaunchPlan.create(
        "morning_lp",
        go_greet,
        fixed_inputs={"am": True},
        default_inputs={"day_of_week": "monday"},
    )

    entities = OrderedDict()
    spec = get_serializable(entities, serialization_settings, tst_lp)
    aq = spec.spec.default_inputs.parameters["number"].artifact_query
    assert aq.artifact_id.artifact_key.name == "important_int"
    assert not aq.artifact_id.HasField("partitions")
    assert not aq.artifact_id.HasField("time_partition")


def test_check_input_binding():
    import pandas as pd

    a1_t_ab = Artifact(name="my_data", partition_keys=["a", "b"], time_partitioned=True)

    with pytest.raises(FlyteValidationException):

        @task
        def t1(
            b_value: str, dt: datetime.datetime
        ) -> Annotated[pd.DataFrame, a1_t_ab(time_partition=Inputs.dt, b=Inputs.xyz, a="manual")]:
            df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
            return df

    with pytest.raises(FlyteValidationException):

        @task
        def t2(
            b_value: str, dt: datetime.datetime
        ) -> Annotated[pd.DataFrame, a1_t_ab(time_partition=Inputs.dtt, b=Inputs.b_value, a="manual")]:
            df = pd.DataFrame({"a": [1, 2, 3], "b": [b_value, b_value, b_value]})
            return df
