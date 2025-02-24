import typing
from collections import OrderedDict

import pytest
from flyteidl.admin import launch_plan_pb2 as _launch_plan_idl

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core import launch_plan, notification
from flytekit.core.schedule import CronSchedule
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.models.common import Annotations, AuthRole, Labels, RawOutputDataConfig
from flytekit.models.core import execution as _execution_model
from flytekit.models.core import identifier as identifier_models
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


# This test is used for documentation.
def test_lp_documentation():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def wf(a: int, c: str) -> (int, str):
        x, y = t1(a=a)
        return x, y

    # fixed_and_default_start
    launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="your_lp_name_1", default_inputs={"a": 3}, fixed_inputs={"c": "4"}
    )
    # fixed_and_default_end

    # schedule_start
    sched = CronSchedule("* * ? * * *", kickoff_time_input_arg="abc")
    email_notif = notification.Email(
        phases=[_execution_model.WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@email.com"]
    )
    launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="your_lp_name_2", schedule=sched, notifications=[email_notif]
    )
    # schedule_end

    # auth_role_start
    auth_role_model = AuthRole(assumable_iam_role="my:iam:role")
    launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="your_lp_name_3",
    )

    labels_model = Labels({"label": "foo"})
    annotations_model = Annotations({"annotate": "bar"})
    launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="your_lp_name_4",
        auth_role=auth_role_model,
        labels=labels_model,
        annotations=annotations_model,
    )

    raw_output_data_config = RawOutputDataConfig("s3://foo/output")
    launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="your_lp_name_5", raw_output_data_config=raw_output_data_config
    )
    # auth_role_end


def test_lp():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def wf(a: int) -> (str, str):
        x, y = t1(a=a)
        u, v = t1(a=x)
        return y, v

    lp = launch_plan.LaunchPlan.get_or_create(wf, "get_or_create1")
    lp2 = launch_plan.LaunchPlan.get_or_create(wf, "get_or_create1")
    assert lp.name == "get_or_create1"
    assert lp is lp2

    default_lp = launch_plan.LaunchPlan.get_or_create(wf)
    default_lp2 = launch_plan.LaunchPlan.get_or_create(wf)
    assert default_lp is default_lp2

    with pytest.raises(ValueError):
        launch_plan.LaunchPlan.get_or_create(wf, default_inputs={"a": 3})

    lp_with_defaults = launch_plan.LaunchPlan.create("get_or_create2", wf, default_inputs={"a": 3})
    assert lp_with_defaults.parameters.parameters["a"].default.scalar.primitive.integer == 3


def test_lp_each_parameter():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def wf(a: int, c: str) -> (int, str):
        x, y = t1(a=a)
        return x, y

    # Fixed Inputs Parameter
    launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_fixed", fixed_inputs={"a": 1, "c": "4"})
    with pytest.raises(AssertionError):
        launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_fixed")

    # Schedule Parameter
    obj = CronSchedule("* * ? * * *", kickoff_time_input_arg="abc")
    schedule_lp = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_schedule", schedule=obj)
    schedule_lp1 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_schedule", schedule=obj)

    assert schedule_lp is schedule_lp1

    # Default Inputs Parameter
    launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_schedule_def_inputs", default_inputs={"a": 9})
    with pytest.raises(AssertionError):
        launch_plan.LaunchPlan.get_or_create(
            workflow=wf, name="get_or_create_schedule_def_inputs", default_inputs={"a": 19}
        )

    # Notifications Parameter
    email_notif = notification.Email(
        phases=[_execution_model.WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@email.com"]
    )
    notification_lp = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_notification", notifications=[email_notif]
    )
    notification_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_notification", notifications=[email_notif]
    )

    assert notification_lp is notification_lp1

    # Auth Parameter
    auth_role_model1 = AuthRole(assumable_iam_role="my:iam:role")
    auth_role_model2 = _launch_plan_idl.Auth(kubernetes_service_account="my:service:account")
    launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_auth", auth_role=auth_role_model1)
    with pytest.raises(AssertionError):
        launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_auth", auth_role=auth_role_model2)

    # Labels parameters
    labels_model1 = Labels({"label": "foo"})
    labels_model2 = Labels({"label": "foo"})
    labels_lp1 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model1)
    labels_lp2 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model2)
    assert labels_lp1 is labels_lp2

    # Annotations parameters
    annotations_model1 = Annotations({"anno": "bar"})
    annotations_model2 = Annotations({"anno": "bar"})
    annotations_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model1
    )
    annotations_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model2
    )
    assert annotations_lp1 is annotations_lp2

    # Raw output prefix parameters
    raw_output_data_config1 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config2 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config1
    )
    raw_output_data_config_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config2
    )
    assert raw_output_data_config_lp1 is raw_output_data_config_lp2

    # Max parallelism
    max_parallelism = 100
    max_parallelism_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    max_parallelism_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    assert max_parallelism_lp1 is max_parallelism_lp2

    # Labels parameters
    labels_model1 = Labels({"label": "foo"})
    labels_model2 = Labels({"label": "foo"})
    labels_lp1 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model1)
    labels_lp2 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model2)
    assert labels_lp1 is labels_lp2

    # Annotations parameters
    annotations_model1 = Annotations({"anno": "bar"})
    annotations_model2 = Annotations({"anno": "bar"})
    annotations_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model1
    )
    annotations_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model2
    )
    assert annotations_lp1 is annotations_lp2

    # Raw output prefix parameters
    raw_output_data_config1 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config2 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config1
    )
    raw_output_data_config_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config2
    )
    assert raw_output_data_config_lp1 is raw_output_data_config_lp2

    # Max parallelism
    max_parallelism = 100
    max_parallelism_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    max_parallelism_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    assert max_parallelism_lp1 is max_parallelism_lp2

    # Labels parameters
    labels_model1 = Labels({"label": "foo"})
    labels_model2 = Labels({"label": "foo"})
    labels_lp1 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model1)
    labels_lp2 = launch_plan.LaunchPlan.get_or_create(workflow=wf, name="get_or_create_labels", labels=labels_model2)
    assert labels_lp1 is labels_lp2

    # Annotations parameters
    annotations_model1 = Annotations({"anno": "bar"})
    annotations_model2 = Annotations({"anno": "bar"})
    annotations_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model1
    )
    annotations_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_annotations", annotations=annotations_model2
    )
    assert annotations_lp1 is annotations_lp2

    # Raw output prefix parameters
    raw_output_data_config1 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config2 = RawOutputDataConfig("s3://foo/output")
    raw_output_data_config_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config1
    )
    raw_output_data_config_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf, name="get_or_create_raw_output_prefix", raw_output_data_config=raw_output_data_config2
    )
    assert raw_output_data_config_lp1 is raw_output_data_config_lp2

    # Max parallelism
    max_parallelism = 100
    max_parallelism_lp1 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    max_parallelism_lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create_max_parallelism",
        max_parallelism=max_parallelism,
    )
    assert max_parallelism_lp1 is max_parallelism_lp2

    # Default LaunchPlan
    name_lp = launch_plan.LaunchPlan.get_or_create(workflow=wf)
    name_lp1 = launch_plan.LaunchPlan.get_or_create(workflow=wf)

    assert name_lp is name_lp1


def test_lp_all_parameters():
    nt = typing.NamedTuple("OutputsBC", [("t1_int_output", int), ("c", str)])

    @task
    def t1(a: int) -> nt:
        a = a + 2
        return nt(a, "world-" + str(a))

    @task
    def t2(a: str, b: str, c: str) -> str:
        return b + a + c

    @workflow
    def wf(a: int, c: str) -> str:
        x, y = t1(a=a)
        u = t2(a=x, b=y, c=c)
        return u

    obj = CronSchedule("* * ? * * *", kickoff_time_input_arg="abc")
    obj1 = CronSchedule("10 * ? * * *", kickoff_time_input_arg="abc")
    slack_notif = notification.Slack(
        phases=[_execution_model.WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@slack.com"]
    )
    auth_role_model = AuthRole(assumable_iam_role="my:iam:role")
    labels = Labels({"label": "foo"})
    annotations = Annotations({"anno": "bar"})
    raw_output_data_config = RawOutputDataConfig("s3://foo/output")

    lp = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create",
        default_inputs={"a": 3},
        fixed_inputs={"c": "4"},
        schedule=obj,
        notifications=[slack_notif],
        auth_role=auth_role_model,
        labels=labels,
        annotations=annotations,
        raw_output_data_config=raw_output_data_config,
    )
    lp2 = launch_plan.LaunchPlan.get_or_create(
        workflow=wf,
        name="get_or_create",
        default_inputs={"a": 3},
        fixed_inputs={"c": "4"},
        schedule=obj,
        notifications=[slack_notif],
        auth_role=auth_role_model,
        labels=labels,
        annotations=annotations,
        raw_output_data_config=raw_output_data_config,
    )

    assert lp is lp2

    with pytest.raises(AssertionError):
        launch_plan.LaunchPlan.get_or_create(
            workflow=wf,
            name="get_or_create",
            default_inputs={"a": 3},
            fixed_inputs={"c": "4"},
            schedule=obj1,
            notifications=[slack_notif],
            auth_role=auth_role_model,
            labels=labels,
            annotations=annotations,
            raw_output_data_config=raw_output_data_config,
        )


def test_lp_nodes():
    @task
    def t1(a: int) -> int:
        return a + 2

    @workflow
    def my_sub_wf(a: int) -> int:
        return t1(a=a)

    lp = launch_plan.LaunchPlan.get_or_create(my_sub_wf, "my_sub_wf_lp1")

    @workflow
    def my_wf(a: int) -> typing.Tuple[int, int]:
        t = t1(a=a)
        w = lp(a=a)
        return t, w

    all_entities = OrderedDict()
    wf_spec = get_serializable(all_entities, serialization_settings, my_wf)
    assert wf_spec.template.nodes[1].workflow_node is not None
    assert (
        wf_spec.template.nodes[1].workflow_node.launchplan_ref.resource_type
        == identifier_models.ResourceType.LAUNCH_PLAN
    )
    assert wf_spec.template.nodes[1].workflow_node.launchplan_ref.name == "my_sub_wf_lp1"


def test_lp_default_paramaters_work_when_called_from_another_workflow():
    @task
    def t1(a: int, b: int) -> int:
        return a + b

    @workflow
    def my_sub_wf(a: int, b: int = 5) -> int:
        return t1(a=a, b=b)

    lp = launch_plan.LaunchPlan.get_or_create(my_sub_wf)

    @workflow
    def my_wf(a: int) -> int:
        return lp(a=a)

    assert my_wf(a=8) == 13


def test_lp_with_docstring():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def wf_with_docstring(a: int) -> (str, str):
        """
        sample workflow

        Parameters
        ----------
        a : str
            foo

        Returns
        -------
        out1, out2 : tuple
            ramen
        """
        x, y = t1(a=a)
        return x, y

    lp = launch_plan.LaunchPlan.get_or_create(workflow=wf_with_docstring)
    assert lp.parameters.parameters["a"].var.description == "foo"
