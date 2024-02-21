from flyteidl.admin import launch_plan_pb2 as _launch_plan_idl

from flytekit.models import common, interface, launch_plan, literals, schedule, types
from flytekit.models.core import identifier


def test_metadata():
    obj = launch_plan.LaunchPlanMetadata(schedule=None, notifications=[])
    obj2 = launch_plan.LaunchPlanMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_metadata_schedule():
    s = schedule.Schedule("asdf", "1 3 4 5 6 7")
    obj = launch_plan.LaunchPlanMetadata(schedule=s, notifications=[])
    assert obj.schedule == s
    obj2 = launch_plan.LaunchPlanMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.schedule == s


def test_lp_closure():
    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    p = interface.Parameter(var=v)
    parameter_map = interface.ParameterMap({"ppp": p})
    parameter_map.to_flyte_idl()
    variable_map = interface.VariableMap({"vvv": v})
    obj = launch_plan.LaunchPlanClosure(
        state=launch_plan.LaunchPlanState.ACTIVE,
        expected_inputs=parameter_map,
        expected_outputs=variable_map,
    )
    assert obj.expected_inputs == parameter_map
    assert obj.expected_outputs == variable_map

    obj2 = launch_plan.LaunchPlanClosure.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.expected_inputs == parameter_map
    assert obj2.expected_outputs == variable_map


def test_launch_plan_spec():
    identifier_model = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")

    s = schedule.Schedule("asdf", "1 3 4 5 6 7")
    launch_plan_metadata_model = launch_plan.LaunchPlanMetadata(schedule=s, notifications=[])

    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    p = interface.Parameter(var=v)
    parameter_map = interface.ParameterMap({"ppp": p})

    fixed_inputs = literals.LiteralMap(
        {"a": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=1)))}
    )

    labels_model = common.Labels({})
    annotations_model = common.Annotations({"my": "annotation"})

    auth_role_model = common.AuthRole(assumable_iam_role="my:iam:role")
    raw_data_output_config = common.RawOutputDataConfig("s3://bucket")
    empty_raw_data_output_config = common.RawOutputDataConfig("")
    max_parallelism = 100

    lp_spec_raw_output_prefixed = launch_plan.LaunchPlanSpec(
        identifier_model,
        launch_plan_metadata_model,
        parameter_map,
        fixed_inputs,
        labels_model,
        annotations_model,
        auth_role_model,
        raw_data_output_config,
        max_parallelism,
    )

    obj2 = launch_plan.LaunchPlanSpec.from_flyte_idl(lp_spec_raw_output_prefixed.to_flyte_idl())
    assert obj2 == lp_spec_raw_output_prefixed

    lp_spec_no_prefix = launch_plan.LaunchPlanSpec(
        identifier_model,
        launch_plan_metadata_model,
        parameter_map,
        fixed_inputs,
        labels_model,
        annotations_model,
        auth_role_model,
        empty_raw_data_output_config,
        max_parallelism,
    )

    obj2 = launch_plan.LaunchPlanSpec.from_flyte_idl(lp_spec_no_prefix.to_flyte_idl())
    assert obj2 == lp_spec_no_prefix


def test_old_style_role():
    identifier_model = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")

    s = schedule.Schedule("asdf", "1 3 4 5 6 7")
    launch_plan_metadata_model = launch_plan.LaunchPlanMetadata(schedule=s, notifications=[])

    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    p = interface.Parameter(var=v)
    parameter_map = interface.ParameterMap({"ppp": p})

    fixed_inputs = literals.LiteralMap(
        {"a": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=1)))}
    )

    labels_model = common.Labels({})
    annotations_model = common.Annotations({"my": "annotation"})

    raw_data_output_config = common.RawOutputDataConfig("s3://bucket")

    old_role = _launch_plan_idl.Auth(kubernetes_service_account="my:service:account")

    old_style_spec = _launch_plan_idl.LaunchPlanSpec(
        workflow_id=identifier_model.to_flyte_idl(),
        entity_metadata=launch_plan_metadata_model.to_flyte_idl(),
        default_inputs=parameter_map.to_flyte_idl(),
        fixed_inputs=fixed_inputs.to_flyte_idl(),
        labels=labels_model.to_flyte_idl(),
        annotations=annotations_model.to_flyte_idl(),
        raw_output_data_config=raw_data_output_config.to_flyte_idl(),
        auth=old_role,
    )

    lp_spec = launch_plan.LaunchPlanSpec.from_flyte_idl(old_style_spec)

    assert lp_spec.auth_role.assumable_iam_role == "my:service:account"
