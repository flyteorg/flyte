import typing

from flyteidl.admin import launch_plan_pb2 as _launch_plan
from google.protobuf.any_pb2 import Any

from flytekit.models import common as _common
from flytekit.models import interface as _interface
from flytekit.models import literals as _literals
from flytekit.models import schedule as _schedule
from flytekit.models import security
from flytekit.models.core import identifier as _identifier


class LaunchPlanMetadata(_common.FlyteIdlEntity):
    def __init__(self, schedule, notifications, launch_conditions=None):
        """

        :param flytekit.models.schedule.Schedule schedule: Schedule to execute the Launch Plan
        :param list[flytekit.models.common.Notification] notifications: List of notifications based on
            execution status transitions
        :param launch_conditions: Additional metadata for launching
        """
        self._schedule = schedule
        self._notifications = notifications
        self._launch_conditions = launch_conditions

    @property
    def schedule(self):
        """
        Schedule to execute the Launch Plan
        :rtype: flytekit.models.schedule.Schedule
        """
        return self._schedule

    @property
    def notifications(self):
        """
        List of notifications based on Execution status transitions
        :rtype: list[flytekit.models.common.Notification]
        """
        return self._notifications

    @property
    def launch_conditions(self):
        return self._launch_conditions

    def to_flyte_idl(self):
        """
        List of notifications based on Execution status transitions
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanMetadata
        """
        if self.launch_conditions:
            a = Any()
            a.Pack(self.launch_conditions)
        else:
            a = None
        return _launch_plan.LaunchPlanMetadata(
            schedule=self.schedule.to_flyte_idl() if self.schedule is not None else None,
            notifications=[n.to_flyte_idl() for n in self.notifications],
            launch_conditions=a,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.launch_plan_pb2.LaunchPlanMetadata pb2_object:
        :rtype: LaunchPlanMetadata
        """
        return cls(
            schedule=_schedule.Schedule.from_flyte_idl(pb2_object.schedule)
            if pb2_object.HasField("schedule")
            else None,
            notifications=[_common.Notification.from_flyte_idl(n) for n in pb2_object.notifications],
            launch_conditions=pb2_object.launch_conditions if pb2_object.HasField("launch_conditions") else None,
        )


class Auth(_common.FlyteIdlEntity):
    def __init__(self, assumable_iam_role=None, kubernetes_service_account=None):
        """
        DEPRECATED. Do not use. Use flytekit.models.common.AuthRole instead
        At most one of assumable_iam_role or kubernetes_service_account can be set.
        :param Text assumable_iam_role: IAM identity with set permissions policies.
        :param Text kubernetes_service_account: Provides an identity for workflow execution resources. Flyte deployment
            administrators are responsible for handling permissions as they relate to the service account.
        """
        self._assumable_iam_role = assumable_iam_role
        self._kubernetes_service_account = kubernetes_service_account

    @property
    def assumable_iam_role(self):
        """
        The IAM role to execute the workflow with
        :rtype: Text
        """
        return self._assumable_iam_role

    @property
    def kubernetes_service_account(self):
        """
        The kubernetes service account to execute the workflow with
        :rtype: Text
        """
        return self._kubernetes_service_account

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.launch_plan_pb2.Auth
        """
        return _launch_plan.Auth(
            assumable_iam_role=self.assumable_iam_role if self.assumable_iam_role else None,
            kubernetes_service_account=self.kubernetes_service_account if self.kubernetes_service_account else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.launch_plan_pb2.Auth pb2_object:
        :rtype: Auth
        """
        return cls(
            assumable_iam_role=pb2_object.assumable_iam_role,
            kubernetes_service_account=pb2_object.kubernetes_service_account,
        )


class LaunchPlanSpec(_common.FlyteIdlEntity):
    def __init__(
        self,
        workflow_id,
        entity_metadata,
        default_inputs,
        fixed_inputs,
        labels: _common.Labels,
        annotations: _common.Annotations,
        auth_role: _common.AuthRole,
        raw_output_data_config: _common.RawOutputDataConfig,
        max_parallelism: typing.Optional[int] = None,
        security_context: typing.Optional[security.SecurityContext] = None,
    ):
        """
        The spec for a Launch Plan.

        :param flytekit.models.core.identifier.Identifier workflow_id: Unique identifier for the workflow in question
        :param LaunchPlanMetadata entity_metadata: Metadata
        :param flytekit.models.interface.ParameterMap default_inputs: Input values to be passed for the execution
        :param flytekit.models.literals.LiteralMap fixed_inputs: Fixed, non-overridable inputs for the Launch Plan
        :param flytekit.models.common.Labels:
            Any custom kubernetes labels to apply to workflows executed by this launch plan.
        :param flytekit.models.common.Annotations annotations:
            Any custom kubernetes annotations to apply to workflows executed by this launch plan.
        :param flytekit.models.common.AuthRole auth_role: The auth method with which to execute the workflow.
        :param flytekit.models.common.RawOutputDataConfig raw_output_data_config: Value for where to store offloaded
            data like Blobs and Schemas.
        :param max_parallelism int: Controls the maximum number of tasknodes that can be run in parallel for the entire
            workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and
            parallelism/concurrency of MapTasks is independent from this.
        :param security_context: This can be used to add security information to a LaunchPlan, which will be used by
                                 every execution
        """
        self._workflow_id = workflow_id
        self._entity_metadata = entity_metadata
        self._default_inputs = default_inputs
        self._fixed_inputs = fixed_inputs
        self._labels = labels
        self._annotations = annotations
        self._auth_role = auth_role
        self._raw_output_data_config = raw_output_data_config
        self._max_parallelism = max_parallelism
        self._security_context = security_context

    @property
    def workflow_id(self):
        """
        Unique identifier for the workflow in question
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._workflow_id

    @property
    def entity_metadata(self):
        """
        :rtype: LaunchPlanMetadata
        """
        return self._entity_metadata

    @property
    def default_inputs(self):
        """
        Input values to be passed for the execution
        :rtype: flytekit.models.interface.ParameterMap
        """
        return self._default_inputs

    @property
    def fixed_inputs(self):
        """
        Fixed, non-overridable inputs for the Launch Plan
        :rtype: flytekit.models.literals.LiteralMap
        """
        return self._fixed_inputs

    @property
    def labels(self) -> _common.Labels:
        """
        The labels to execute the workflow with
        :rtype: flytekit.models.common.Labels
        """
        return self._labels

    @property
    def annotations(self) -> _common.Annotations:
        """
        The annotations to execute the workflow with
        :rtype: flytekit.models.common.Annotations
        """
        return self._annotations

    @property
    def auth_role(self):
        """
        The authorization method with which to execute the workflow.
        :rtype: flytekit.models.common.AuthRole
        """
        return self._auth_role

    @property
    def raw_output_data_config(self):
        """
        Where to store offloaded data like Blobs and Schemas
        :rtype: flytekit.models.common.RawOutputDataConfig
        """
        return self._raw_output_data_config

    @property
    def max_parallelism(self) -> typing.Optional[int]:
        return self._max_parallelism

    @property
    def security_context(self) -> typing.Optional[security.SecurityContext]:
        return self._security_context

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanSpec
        """
        return _launch_plan.LaunchPlanSpec(
            workflow_id=self.workflow_id.to_flyte_idl(),
            entity_metadata=self.entity_metadata.to_flyte_idl(),
            default_inputs=self.default_inputs.to_flyte_idl(),
            fixed_inputs=self.fixed_inputs.to_flyte_idl(),
            labels=self.labels.to_flyte_idl(),
            annotations=self.annotations.to_flyte_idl(),
            auth_role=self.auth_role.to_flyte_idl() if self.auth_role else None,
            raw_output_data_config=self.raw_output_data_config.to_flyte_idl(),
            max_parallelism=self.max_parallelism,
            security_context=self.security_context.to_flyte_idl() if self.security_context else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2):
        """
        :param flyteidl.admin.launch_plan_pb2.LaunchPlanSpec pb2:
        :rtype: LaunchPlanSpec
        """
        auth_role = None
        # First check the newer field, auth_role.
        if pb2.auth_role is not None and (pb2.auth_role.assumable_iam_role or pb2.auth_role.kubernetes_service_account):
            auth_role = _common.AuthRole.from_flyte_idl(pb2.auth_role)
        # Fallback to the deprecated field.
        elif pb2.auth is not None:
            if pb2.auth.assumable_iam_role:
                auth_role = _common.AuthRole(assumable_iam_role=pb2.auth.assumable_iam_role)
            else:
                auth_role = _common.AuthRole(assumable_iam_role=pb2.auth.kubernetes_service_account)

        return cls(
            workflow_id=_identifier.Identifier.from_flyte_idl(pb2.workflow_id),
            entity_metadata=LaunchPlanMetadata.from_flyte_idl(pb2.entity_metadata),
            default_inputs=_interface.ParameterMap.from_flyte_idl(pb2.default_inputs),
            fixed_inputs=_literals.LiteralMap.from_flyte_idl(pb2.fixed_inputs),
            labels=_common.Labels.from_flyte_idl(pb2.labels),
            annotations=_common.Annotations.from_flyte_idl(pb2.annotations),
            auth_role=auth_role,
            raw_output_data_config=_common.RawOutputDataConfig.from_flyte_idl(pb2.raw_output_data_config),
            max_parallelism=pb2.max_parallelism,
            security_context=security.SecurityContext.from_flyte_idl(pb2.security_context)
            if pb2.security_context
            else None,
        )


class LaunchPlanState(object):
    INACTIVE = _launch_plan.INACTIVE
    ACTIVE = _launch_plan.ACTIVE

    @classmethod
    def enum_to_string(cls, val):
        """
        :param int val:
        :rtype: Text
        """
        if val == cls.INACTIVE:
            return "INACTIVE"
        elif val == cls.ACTIVE:
            return "ACTIVE"
        else:
            return "<UNKNOWN>"


class LaunchPlanClosure(_common.FlyteIdlEntity):
    def __init__(self, state, expected_inputs, expected_outputs):
        """
        :param LaunchPlanState state: Indicate the Launch plan phase
        :param flytekit.models.interface.ParameterMap expected_inputs: Indicates the set of inputs to execute
            the Launch plan
        :param flytekit.models.interface.VariableMap expected_outputs: Indicates the set of outputs from the Launch plan
        """
        self._state = state
        self._expected_inputs = expected_inputs
        self._expected_outputs = expected_outputs

    @property
    def state(self):
        """
        :rtype: LaunchPlanState
        """
        return self._state

    @property
    def expected_inputs(self):
        """
        :rtype: flytekit.models.interface.ParameterMap
        """
        return self._expected_inputs

    @property
    def expected_outputs(self):
        """
        :rtype: flytekit.models.interface.VariableMap
        """
        return self._expected_outputs

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanClosure
        """
        return _launch_plan.LaunchPlanClosure(
            state=self.state,
            expected_inputs=self.expected_inputs.to_flyte_idl(),
            expected_outputs=self.expected_outputs.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.launch_plan_pb2.LaunchPlanClosure pb2_object:
        :rtype: LaunchPlanClosure
        """
        return cls(
            pb2_object.state,
            _interface.ParameterMap.from_flyte_idl(pb2_object.expected_inputs),
            _interface.VariableMap.from_flyte_idl(pb2_object.expected_outputs),
        )


class LaunchPlan(_common.FlyteIdlEntity):
    def __init__(self, id, spec, closure):
        """
        :param flytekit.models.core.identifier.Identifier id:
        :param LaunchPlanSpec spec:
        :param LaunchPlanClosure closure:
        """
        self._id = id
        self._spec = spec
        self._closure = closure

    @property
    def id(self):
        """
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def spec(self):
        """
        :rtype: LaunchPlanSpec
        """
        return self._spec

    @property
    def closure(self):
        """
        :rtype: LaunchPlanClosure
        """
        return self._closure

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlan
        """
        identifier = (
            self.id
            if self.id is not None
            else _identifier.Identifier(_identifier.ResourceType.LAUNCH_PLAN, None, None, None, None)
        )
        return _launch_plan.LaunchPlan(
            id=identifier.to_flyte_idl(),
            spec=self.spec.to_flyte_idl(),
            closure=self.closure.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.launch_plan_pb2.LaunchPlan pb2_object:
        :rtype: LaunchPlan
        """
        return cls(
            id=_identifier.Identifier.from_flyte_idl(pb2_object.id),
            spec=LaunchPlanSpec.from_flyte_idl(pb2_object.spec),
            closure=LaunchPlanClosure.from_flyte_idl(pb2_object.closure),
        )
