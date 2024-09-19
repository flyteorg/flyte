import datetime
import typing

from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.admin import execution_pb2 as _execution_pb2
from flyteidl.admin import launch_plan_pb2 as _launch_plan_pb2
from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from flyteidl.admin import node_execution_pb2 as _node_execution_pb2
from flyteidl.admin import project_domain_attributes_pb2 as _project_domain_attributes_pb2
from flyteidl.admin import project_pb2 as _project_pb2
from flyteidl.admin import task_execution_pb2 as _task_execution_pb2
from flyteidl.admin import task_pb2 as _task_pb2
from flyteidl.admin import workflow_attributes_pb2 as _workflow_attributes_pb2
from flyteidl.admin import workflow_pb2 as _workflow_pb2
from flyteidl.service import dataproxy_pb2 as _data_proxy_pb2
from google.protobuf.duration_pb2 import Duration

from flytekit.clients.raw import RawSynchronousFlyteClient as _RawSynchronousFlyteClient
from flytekit.models import common as _common
from flytekit.models import execution as _execution
from flytekit.models import filters as _filters
from flytekit.models import launch_plan as _launch_plan
from flytekit.models import node_execution as _node_execution
from flytekit.models import project as _project
from flytekit.models import task as _task
from flytekit.models.admin import common as _admin_common
from flytekit.models.admin import task_execution as _task_execution
from flytekit.models.admin import workflow as _workflow
from flytekit.models.core import identifier as _identifier


class SynchronousFlyteClient(_RawSynchronousFlyteClient):
    """
    This is a low-level client that users can use to make direct gRPC service calls to the control plane. See the
    :std:doc:`service spec <idl:protos/docs/service/index>`. This is more user-friendly interface than the
    :py:class:`raw client <flytekit.clients.raw.RawSynchronousFlyteClient>` so users should try to use this class
    first. Create a client by ::

        SynchronousFlyteClient("your.domain:port", insecure=True)
        # insecure should be True if your flyteadmin deployment doesn't have SSL enabled

    """

    @property
    def raw(self):
        """
        Gives access to the raw client
        :rtype: flytekit.clients.raw.RawSynchronousFlyteClient
        """
        return super(SynchronousFlyteClient, self)

    ####################################################################################################################
    #
    #  Task Endpoints
    #
    ####################################################################################################################

    def create_task(self, task_identifer, task_spec):
        """
        This will create a task definition in the Admin database. Once successful, the task object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly. Furthermore, as long as the request
            remains identical, calling this method multiple times will result in success.

        :param flytekit.models.core.identifier.Identifier task_identifer: The identifier for this task.
        :param flytekit.models.task.TaskSpec task_spec: This is the actual definition of the task that
            should be created.
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the
            task is found, this exception is raised.  The client might choose to ignore this exception because the
            identical task is already registered.
        :raises grpc.RpcError:
        """
        super(SynchronousFlyteClient, self).create_task(
            _task_pb2.TaskCreateRequest(id=task_identifer.to_flyte_idl(), spec=task_spec.to_flyte_idl())
        )

    def list_task_ids_paginated(self, project, domain, limit=100, token=None, sort_by=None):
        """
        This returns a page of identifiers for the tasks for a given project and domain. Filters can also be
        specified.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param Text project: The namespace of the project to list.
        :param Text domain: The domain space of the project to list.
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param Text token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.common.NamedEntityIdentifier], Text
        """
        identifier_list = super(SynchronousFlyteClient, self).list_task_ids_paginated(
            _common_pb2.NamedEntityIdentifierListRequest(
                project=project,
                domain=domain,
                limit=limit,
                token=token,
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_common.NamedEntityIdentifier.from_flyte_idl(identifier_pb) for identifier_pb in identifier_list.entities],
            str(identifier_list.token),
        )

    def list_tasks_paginated(self, identifier, limit=100, token=None, filters=None, sort_by=None):
        """
        This returns a page of task metadata for tasks in a given project and domain.  Optionally,
        specifying a name will limit the results to only tasks with that name in the given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param flytekit.models.common.NamedEntityIdentifier identifier: NamedEntityIdentifier to list.
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param list[flytekit.models.filters.Filter] filters: [Optional] If specified, the filters will be applied to
            the query.  If the filter is not supported, an exception will be raised.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.task.Task], Text
        """
        task_list = super(SynchronousFlyteClient, self).list_tasks_paginated(
            resource_list_request=_common_pb2.ResourceListRequest(
                id=identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        # TODO: tmp workaround
        for pb in task_list.tasks:
            pb.id.resource_type = _identifier.ResourceType.TASK
        return (
            [_task.Task.from_flyte_idl(task_pb2) for task_pb2 in task_list.tasks],
            str(task_list.token),
        )

    def get_task(self, id):
        """
        This returns a single task for a given identifier.

        :param flytekit.models.core.identifier.Identifier id: The ID representing a given task.
        :raises: TODO
        :rtype: flytekit.models.task.Task
        """
        return _task.Task.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_task(_common_pb2.ObjectGetRequest(id=id.to_flyte_idl()))
        )

    ####################################################################################################################
    #
    #  Workflow Endpoints
    #
    ####################################################################################################################

    def create_workflow(self, workflow_identifier, workflow_spec):
        """
        This will create a workflow definition in the Admin database. Once successful, the workflow object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly. Furthermore, as long as the request
            remains identical, calling this method multiple times will result in success.

        :param: flytekit.models.core.identifier.Identifier workflow_identifier: The identifier for this workflow.
        :param: flytekit.models.admin.workflow.WorkflowSpec workflow_spec: This is the actual definition of the workflow
            that should be created.
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the
            workflow is found, this exception is raised.  The client might choose to ignore this exception because the
            identical workflow is already registered.
        :raises grpc.RpcError:
        """
        super(SynchronousFlyteClient, self).create_workflow(
            _workflow_pb2.WorkflowCreateRequest(
                id=workflow_identifier.to_flyte_idl(), spec=workflow_spec.to_flyte_idl()
            )
        )

    def list_workflow_ids_paginated(self, project, domain, limit=100, token=None, sort_by=None):
        """
        This returns a page of identifiers for the workflows for a given project and domain. Filters can also be
        specified.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: Text project: The namespace of the project to list.
        :param: Text domain: The domain space of the project to list.
        :param: int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param: int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.common.NamedEntityIdentifier], Text
        """
        identifier_list = super(SynchronousFlyteClient, self).list_workflow_ids_paginated(
            _common_pb2.NamedEntityIdentifierListRequest(
                project=project,
                domain=domain,
                limit=limit,
                token=token,
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_common.NamedEntityIdentifier.from_flyte_idl(identifier_pb) for identifier_pb in identifier_list.entities],
            str(identifier_list.token),
        )

    def list_workflows_paginated(self, identifier, limit=100, token=None, filters=None, sort_by=None):
        """
        This returns a page of workflow meta-information for workflows in a given project and domain.  Optionally,
        specifying a name will limit the results to only workflows with that name in the given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param flytekit.models.common.NamedEntityIdentifier identifier: NamedEntityIdentifier to list.
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param list[flytekit.models.filters.Filter] filters: [Optional] If specified, the filters will be applied to
            the query.  If the filter is not supported, an exception will be raised.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.admin.workflow.Workflow], Text
        """
        wf_list = super(SynchronousFlyteClient, self).list_workflows_paginated(
            resource_list_request=_common_pb2.ResourceListRequest(
                id=identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        # TODO: tmp workaround
        for pb in wf_list.workflows:
            pb.id.resource_type = _identifier.ResourceType.WORKFLOW
        return (
            [_workflow.Workflow.from_flyte_idl(wf_pb2) for wf_pb2 in wf_list.workflows],
            str(wf_list.token),
        )

    def get_workflow(self, id):
        """
        This returns a single workflow for a given ID.

        :param flytekit.models.core.identifier.Identifier id: The ID representing a given task.
        :raises: TODO
        :rtype: flytekit.models.admin.workflow.Workflow
        """
        return _workflow.Workflow.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_workflow(_common_pb2.ObjectGetRequest(id=id.to_flyte_idl()))
        )

    ####################################################################################################################
    #
    #  Launch Plan Endpoints
    #
    ####################################################################################################################

    def create_launch_plan(self, launch_plan_identifer, launch_plan_spec):
        """
        This will create a launch plan definition in the Admin database.  Once successful, the launch plan object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly.  This also means that as long as the request
            remains identical, calling this method multiple times will result in success.

        :param: flytekit.models.core.identifier.Identifier launch_plan_identifer: The identifier for this launch plan.
        :param: flytekit.models.launch_plan.LaunchPlanSpec launch_plan_spec: This is the actual definition of the
            launch plan that should be created.
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the
            launch plan is found, this exception is raised.  The client might choose to ignore this exception because
            the identical launch plan is already registered.
        :raises grpc.RpcError:
        """
        super(SynchronousFlyteClient, self).create_launch_plan(
            _launch_plan_pb2.LaunchPlanCreateRequest(
                id=launch_plan_identifer.to_flyte_idl(),
                spec=launch_plan_spec.to_flyte_idl(),
            )
        )

    def get_launch_plan(self, id):
        """
        Retrieves a launch plan entity.

        :param flytekit.models.core.identifier.Identifier id: unique identifier for launch plan to retrieve
        :rtype: flytekit.models.launch_plan.LaunchPlan
        """
        return _launch_plan.LaunchPlan.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_launch_plan(_common_pb2.ObjectGetRequest(id=id.to_flyte_idl()))
        )

    def get_active_launch_plan(self, identifier):
        """
        Retrieves the active launch plan entity given a named entity identifier (project, domain, name).  Raises an
        error if no active launch plan exists.

        :param flytekit.models.common.NamedEntityIdentifier identifier: NamedEntityIdentifier to list.
        :rtype: flytekit.models.launch_plan.LaunchPlan
        """
        return _launch_plan.LaunchPlan.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_active_launch_plan(
                _launch_plan_pb2.ActiveLaunchPlanRequest(id=identifier.to_flyte_idl())
            )
        )

    def list_launch_plan_ids_paginated(self, project, domain, limit=100, token=None, sort_by=None):
        """
        This returns a page of identifiers for the launch plans for a given project and domain. Filters can also be
        specified.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: Text project: The namespace of the project to list.
        :param: Text domain: The domain space of the project to list.
        :param: int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param: int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.common.NamedEntityIdentifier], Text
        """
        identifier_list = super(SynchronousFlyteClient, self).list_launch_plan_ids_paginated(
            _common_pb2.NamedEntityIdentifierListRequest(
                project=project,
                domain=domain,
                limit=limit,
                token=token,
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_common.NamedEntityIdentifier.from_flyte_idl(identifier_pb) for identifier_pb in identifier_list.entities],
            str(identifier_list.token),
        )

    def list_launch_plans_paginated(self, identifier, limit=100, token=None, filters=None, sort_by=None):
        """
        This returns a page of launch plan meta-information for launch plans in a given project and domain.  Optionally,
        specifying a name will limit the results to only workflows with that name in the given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param flytekit.models.common.NamedEntityIdentifier identifier: NamedEntityIdentifier to list.
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param list[flytekit.models.filters.Filter] filters: [Optional] If specified, the filters will be applied to
            the query.  If the filter is not supported, an exception will be raised.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.launch_plan.LaunchPlan], str
        """
        lp_list = super(SynchronousFlyteClient, self).list_launch_plans_paginated(
            resource_list_request=_common_pb2.ResourceListRequest(
                id=identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        # TODO: tmp workaround
        for pb in lp_list.launch_plans:
            pb.id.resource_type = _identifier.ResourceType.LAUNCH_PLAN
        return (
            [_launch_plan.LaunchPlan.from_flyte_idl(pb) for pb in lp_list.launch_plans],
            str(lp_list.token),
        )

    def list_active_launch_plans_paginated(
        self, project, domain, limit=100, token=None, sort_by=None
    ) -> typing.Tuple[typing.List[_launch_plan.LaunchPlan], str]:
        """
        This returns a page of currently active launch plan meta-information for launch plans in a given project and
        domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param Text project:
        :param Text domain:
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param int token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: list[flytekit.models.launch_plan.LaunchPlan], str
        """
        lp_list = super(SynchronousFlyteClient, self).list_active_launch_plans_paginated(
            _launch_plan_pb2.ActiveLaunchPlanListRequest(
                project=project,
                domain=domain,
                limit=limit,
                token=token,
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        # TODO: tmp workaround
        for pb in lp_list.launch_plans:
            pb.id.resource_type = _identifier.ResourceType.LAUNCH_PLAN
        return (
            [_launch_plan.LaunchPlan.from_flyte_idl(pb) for pb in lp_list.launch_plans],
            str(lp_list.token),
        )

    def update_launch_plan(self, id, state):
        """
        Updates a launch plan.  Currently, this can only be used to update a given launch plan's state (ACTIVE v.
        INACTIVE) for schedules.  If a launch plan with a given project, domain, and name is set to ACTIVE,
        then any other launch plan with the same project, domain, and name that was set to ACTIVE will be switched to
        INACTIVE in one transaction.

        :param flytekit.models.core.identifier.Identifier id: identifier for launch plan to update
        :param int state: Enum value from flytekit.models.launch_plan.LaunchPlanState
        """
        super(SynchronousFlyteClient, self).update_launch_plan(
            _launch_plan_pb2.LaunchPlanUpdateRequest(id=id.to_flyte_idl(), state=state)
        )

    ####################################################################################################################
    #
    #  Named Entity Endpoints
    #
    ####################################################################################################################

    def update_named_entity(self, resource_type, id, metadata):
        """
        Updates the metadata associated with a named entity.  A named entity is designated a resource, e.g. a workflow,
        task or launch plan specified by {project, domain, name} across all versions of the resource.

        :param int resource_type: Enum value from flytekit.models.identifier.ResourceType
        :param flytekit.models.admin.named_entity.NamedEntityIdentifier id: identifier for named entity to update
        :param flytekit.models.admin.named_entity.NamedEntityIdentifierMetadata metadata:
        """
        super(SynchronousFlyteClient, self).update_named_entity(
            _common_pb2.NamedEntityUpdateRequest(
                resource_type=resource_type,
                id=id.to_flyte_idl(),
                metadata=metadata.to_flyte_idl(),
            )
        )

    ####################################################################################################################
    #
    #  Execution Endpoints
    #
    ####################################################################################################################

    def create_execution(self, project, domain, name, execution_spec, inputs):
        """
        This will create an execution for the given execution spec.
        :param Text project:
        :param Text domain:
        :param Text name:
        :param flytekit.models.execution.ExecutionSpec execution_spec: This is the specification for the execution.
        :param flytekit.models.literals.LiteralMap inputs: The inputs for the execution
        :returns: The unique identifier for the execution.
        :rtype: flytekit.models.core.identifier.WorkflowExecutionIdentifier
        """
        return _identifier.WorkflowExecutionIdentifier.from_flyte_idl(
            super(SynchronousFlyteClient, self)
            .create_execution(
                _execution_pb2.ExecutionCreateRequest(
                    project=project,
                    domain=domain,
                    name=name,
                    spec=execution_spec.to_flyte_idl(),
                    inputs=inputs.to_flyte_idl(),
                )
            )
            .id
        )

    def recover_execution(self, id, name: str = None):
        """
        Recreates a previously-run workflow execution that will only start executing from the last known failure point.
        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :param name str: Optional name to assign to the newly created execution.
        :rtype: flytekit.models.core.identifier.WorkflowExecutionIdentifier
        """
        return _identifier.WorkflowExecutionIdentifier.from_flyte_idl(
            super(SynchronousFlyteClient, self)
            .recover_execution(_execution_pb2.ExecutionRecoverRequest(id=id.to_flyte_idl(), name=name))
            .id
        )

    def get_execution(self, id):
        """
        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :rtype: flytekit.models.execution.Execution
        """
        return _execution.Execution.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_execution(
                _execution_pb2.WorkflowExecutionGetRequest(id=id.to_flyte_idl())
            )
        )

    def get_execution_data(self, id):
        """
        Returns signed URLs to LiteralMap blobs for an execution's inputs and outputs (when available).

        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :rtype: flytekit.models.execution.WorkflowExecutionGetDataResponse
        """
        return _execution.WorkflowExecutionGetDataResponse.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_execution_data(
                _execution_pb2.WorkflowExecutionGetDataRequest(id=id.to_flyte_idl())
            )
        )

    def list_executions_paginated(self, project, domain, limit=100, token=None, filters=None, sort_by=None):
        """
        This returns a page of executions in a given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param Text project: Project in which to list executions.
        :param Text domain: Project in which to list executions.
        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param Text token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param list[flytekit.models.filters.Filter] filters: [Optional] If specified, the filters will be applied to
            the query.  If the filter is not supported, an exception will be raised.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises: TODO
        :rtype: (list[flytekit.models.execution.Execution], Text)
        """
        exec_list = super(SynchronousFlyteClient, self).list_executions_paginated(
            resource_list_request=_common_pb2.ResourceListRequest(
                id=_common_pb2.NamedEntityIdentifier(project=project, domain=domain),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_execution.Execution.from_flyte_idl(pb) for pb in exec_list.executions],
            str(exec_list.token),
        )

    def terminate_execution(self, id, cause):
        """
        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :param Text cause:
        """
        super(SynchronousFlyteClient, self).terminate_execution(
            _execution_pb2.ExecutionTerminateRequest(id=id.to_flyte_idl(), cause=cause)
        )

    def relaunch_execution(self, id, name=None):
        """
        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier id:
        :param Text name: [Optional] name for the new execution. If not specified, a randomly generated name will be
            used
        :returns: The unique identifier for the new execution.
        :rtype: flytekit.models.core.identifier.WorkflowExecutionIdentifier
        """
        return _identifier.WorkflowExecutionIdentifier.from_flyte_idl(
            super(SynchronousFlyteClient, self)
            .relaunch_execution(_execution_pb2.ExecutionRelaunchRequest(id=id.to_flyte_idl(), name=name))
            .id
        )

    ####################################################################################################################
    #
    #  Node Execution Endpoints
    #
    ####################################################################################################################

    def get_node_execution(self, node_execution_identifier):
        """
        :param flytekit.models.core.identifier.NodeExecutionIdentifier node_execution_identifier:
        :rtype: flytekit.models.node_execution.NodeExecution
        """
        return _node_execution.NodeExecution.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_node_execution(
                _node_execution_pb2.NodeExecutionGetRequest(id=node_execution_identifier.to_flyte_idl())
            )
        )

    def get_node_execution_data(self, node_execution_identifier) -> _execution.NodeExecutionGetDataResponse:
        """
        Returns signed URLs to LiteralMap blobs for a node execution's inputs and outputs (when available).

        :param flytekit.models.core.identifier.NodeExecutionIdentifier node_execution_identifier:
        """
        return _execution.NodeExecutionGetDataResponse.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_node_execution_data(
                _node_execution_pb2.NodeExecutionGetDataRequest(id=node_execution_identifier.to_flyte_idl())
            )
        )

    def list_node_executions(
        self,
        workflow_execution_identifier,
        limit: int = 100,
        token: typing.Optional[str] = None,
        filters: typing.List[_filters.Filter] = None,
        sort_by: _admin_common.Sort = None,
        unique_parent_id: str = None,
    ):
        """Get node executions associated with a given workflow execution.

        :param flytekit.models.core.identifier.WorkflowExecutionIdentifier workflow_execution_identifier:
        :param limit: Limit the number of items returned in the response.
        :param token: If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify ``token="foo"``.
        :param list[flytekit.models.filters.Filter] filters:
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :param unique_parent_id: If specified, returns the node executions for the ``unique_parent_id`` node id.
        :rtype: list[flytekit.models.node_execution.NodeExecution], Text
        """
        exec_list = super(SynchronousFlyteClient, self).list_node_executions_paginated(
            _node_execution_pb2.NodeExecutionListRequest(
                workflow_execution_id=workflow_execution_identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
                unique_parent_id=unique_parent_id,
            )
        )
        return (
            [_node_execution.NodeExecution.from_flyte_idl(e) for e in exec_list.node_executions],
            str(exec_list.token),
        )

    def list_node_executions_for_task_paginated(
        self,
        task_execution_identifier,
        limit=100,
        token=None,
        filters=None,
        sort_by=None,
    ):
        """
        This returns nodes spawned by a specific task execution.  This is generally from things like dynamic tasks.
        :param flytekit.models.core.identifier.TaskExecutionIdentifier task_execution_identifier:
        :param int limit: Number to return per page
        :param Text token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
        If you previously retrieved a page response with token="foo" and you want the next page,
        specify token="foo".
        :param list[flytekit.models.filters.Filter] filters:
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :rtype: list[flytekit.models.node_execution.NodeExecution], Text
        """
        exec_list = self._stub.ListNodeExecutionsForTask(
            _node_execution_pb2.NodeExecutionForTaskListRequest(
                task_execution_id=task_execution_identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_node_execution.NodeExecution.from_flyte_idl(e) for e in exec_list.node_executions],
            str(exec_list.token),
        )

    ####################################################################################################################
    #
    #  Task Execution Endpoints
    #
    ####################################################################################################################

    def get_task_execution(self, id):
        """
        :param flytekit.models.core.identifier.TaskExecutionIdentifier id:
        :rtype: flytekit.models.admin.task_execution.TaskExecution
        """
        return _task_execution.TaskExecution.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_task_execution(
                _task_execution_pb2.TaskExecutionGetRequest(id=id.to_flyte_idl())
            )
        )

    def get_task_execution_data(self, task_execution_identifier):
        """
        Returns signed URLs to LiteralMap blobs for a node execution's inputs and outputs (when available).

        :param flytekit.models.core.identifier.TaskExecutionIdentifier task_execution_identifier:
        :rtype: flytekit.models.execution.NodeExecutionGetDataResponse
        """
        return _execution.TaskExecutionGetDataResponse.from_flyte_idl(
            super(SynchronousFlyteClient, self).get_task_execution_data(
                _task_execution_pb2.TaskExecutionGetDataRequest(id=task_execution_identifier.to_flyte_idl())
            )
        )

    def list_task_executions_paginated(
        self,
        node_execution_identifier,
        limit=100,
        token=None,
        filters=None,
        sort_by=None,
    ):
        """
        :param flytekit.models.core.identifier.NodeExecutionIdentifier node_execution_identifier:
        :param int limit:
        :param Text token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo".
        :param list[flytekit.models.filters.Filter] filters:
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :rtype: (list[flytekit.models.admin.task_execution.TaskExecution], Text)
        """
        exec_list = super(SynchronousFlyteClient, self).list_task_executions_paginated(
            _task_execution_pb2.TaskExecutionListRequest(
                node_execution_id=node_execution_identifier.to_flyte_idl(),
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_task_execution.TaskExecution.from_flyte_idl(e) for e in exec_list.task_executions],
            str(exec_list.token),
        )

    ####################################################################################################################
    #
    #  Project Endpoints
    #
    ####################################################################################################################

    def register_project(self, project):
        """
        Registers a project.
        :param flytekit.models.project.Project project:
        :rtype: flyteidl.admin.project_pb2.ProjectRegisterResponse
        """
        super(SynchronousFlyteClient, self).register_project(
            _project_pb2.ProjectRegisterRequest(
                project=project.to_flyte_idl(),
            )
        )

    def update_project(self, project):
        """
        Update an existing project specified by id.
        :param flytekit.models.project.Project project:
        :rtype: flyteidl.admin.project_pb2.ProjectUpdateResponse
        """
        super(SynchronousFlyteClient, self).update_project(project.to_flyte_idl())

    def list_projects_paginated(self, limit=100, token=None, filters=None, sort_by=None):
        """
        This returns a page of projects.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param int limit: [Optional] The maximum number of entries to return.  Must be greater than 0.  The maximum
            page size is determined by the Flyte Admin Service configuration.  If limit is greater than the maximum
            page size, an exception will be raised.
        :param Text token: [Optional] If specified, this specifies where in the rows of results to skip before reading.
            If you previously retrieved a page response with token="foo" and you want the next page,
            specify token="foo". Please see the notes for this function about the caveats of the paginated API.
        :param list[flytekit.models.filters.Filter] filters: [Optional] If specified, the filters will be applied to
            the query.  If the filter is not supported, an exception will be raised.
        :param flytekit.models.admin.common.Sort sort_by: [Optional] If provided, the results will be sorted.
        :raises grpc.RpcError:
        :rtype: (list[flytekit.models.Project], Text)
        """
        projects = super(SynchronousFlyteClient, self).list_projects(
            project_list_request=_project_pb2.ProjectListRequest(
                limit=limit,
                token=token,
                filters=_filters.FilterList(filters or []).to_flyte_idl(),
                sort_by=None if sort_by is None else sort_by.to_flyte_idl(),
            )
        )
        return (
            [_project.Project.from_flyte_idl(pb) for pb in projects.projects],
            str(projects.token),
        )

    ####################################################################################################################
    #
    #  Matching Attributes Endpoints
    #
    ####################################################################################################################

    def update_project_domain_attributes(self, project, domain, matching_attributes):
        """
        Sets custom attributes for a project and domain combination.
        :param Text project:
        :param Text domain:
        :param flytekit.models.MatchingAttributes matching_attributes:
        :return:
        """
        super(SynchronousFlyteClient, self).update_project_domain_attributes(
            _project_domain_attributes_pb2.ProjectDomainAttributesUpdateRequest(
                attributes=_project_domain_attributes_pb2.ProjectDomainAttributes(
                    project=project,
                    domain=domain,
                    matching_attributes=matching_attributes.to_flyte_idl(),
                )
            )
        )

    def update_workflow_attributes(self, project, domain, workflow, matching_attributes):
        """
        Sets custom attributes for a project, domain, and workflow combination.
        :param Text project:
        :param Text domain:
        :param Text workflow:
        :param flytekit.models.MatchingAttributes matching_attributes:
        :return:
        """
        super(SynchronousFlyteClient, self).update_workflow_attributes(
            _workflow_attributes_pb2.WorkflowAttributesUpdateRequest(
                attributes=_workflow_attributes_pb2.WorkflowAttributes(
                    project=project,
                    domain=domain,
                    workflow=workflow,
                    matching_attributes=matching_attributes.to_flyte_idl(),
                )
            )
        )

    def get_project_domain_attributes(self, project, domain, resource_type):
        """
        Fetches the custom attributes set for a project and domain combination.
        :param Text project:
        :param Text domain:
        :param flytekit.models.MatchableResource resource_type:
        :return:
        """
        return super(SynchronousFlyteClient, self).get_project_domain_attributes(
            _project_domain_attributes_pb2.ProjectDomainAttributesGetRequest(
                project=project,
                domain=domain,
                resource_type=resource_type,
            )
        )

    def get_workflow_attributes(self, project, domain, workflow, resource_type):
        """
        Fetches the custom attributes set for a project, domain, and workflow combination.
        :param Text project:
        :param Text domain:
        :param Text workflow:
        :param flytekit.models.MatchableResource resource_type:
        :return:
        """
        return super(SynchronousFlyteClient, self).get_workflow_attributes(
            _workflow_attributes_pb2.WorkflowAttributesGetRequest(
                project=project,
                domain=domain,
                workflow=workflow,
                resource_type=resource_type,
            )
        )

    def list_matchable_attributes(self, resource_type):
        """
        Fetches all custom attributes for a resource type.
        :param flytekit.models.MatchableResource resource_type:
        :return:
        """
        return super(SynchronousFlyteClient, self).list_matchable_attributes(
            _matchable_resource_pb2.ListMatchableAttributesRequest(
                resource_type=resource_type,
            )
        )

    def get_upload_signed_url(
        self,
        project: str,
        domain: str,
        content_md5: typing.Optional[bytes] = None,
        filename: typing.Optional[str] = None,
        expires_in: typing.Optional[datetime.timedelta] = None,
        filename_root: typing.Optional[str] = None,
    ) -> _data_proxy_pb2.CreateUploadLocationResponse:
        """
        Get a signed url to be used during fast registration

        :param project: Project to create the upload location for
        :param domain: Domain to create the upload location for
        :param content_md5: ContentMD5 restricts the upload location to the specific MD5 provided. The content_md5
            will also appear in the generated path.
        :param filename: If provided this specifies a desired suffix for the generated location
        :param expires_in: If provided this defines a requested expiration duration for
            the generated url
        :param filename_root: If provided will be used as the root of the filename.  If not, Admin will use a hash
          This option is useful when uploading a series of files that you want to be grouped together.
        :rtype: flyteidl.service.dataproxy_pb2.CreateUploadLocationResponse
        """
        expires_in_pb = None
        if expires_in:
            expires_in_pb = Duration()
            expires_in_pb.FromTimedelta(expires_in)
        return super(SynchronousFlyteClient, self).create_upload_location(
            _data_proxy_pb2.CreateUploadLocationRequest(
                project=project,
                domain=domain,
                content_md5=content_md5,
                filename=filename,
                expires_in=expires_in_pb,
                filename_root=filename_root,
            )
        )

    def get_download_signed_url(
        self, native_url: str, expires_in: datetime.timedelta = None
    ) -> _data_proxy_pb2.CreateDownloadLocationResponse:
        expires_in_pb = None
        if expires_in:
            expires_in_pb = Duration()
            expires_in_pb.FromTimedelta(expires_in)
        return super(SynchronousFlyteClient, self).create_download_location(
            _data_proxy_pb2.CreateDownloadLocationRequest(
                native_url=native_url,
                expires_in=expires_in_pb,
            )
        )

    def get_data(self, flyte_uri: str) -> _data_proxy_pb2.GetDataResponse:
        req = _data_proxy_pb2.GetDataRequest(flyte_url=flyte_uri)

        resp = self._dataproxy_stub.GetData(req, metadata=self._metadata)
        return resp
