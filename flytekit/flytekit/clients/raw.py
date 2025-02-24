from __future__ import annotations

import typing

import grpc
from flyteidl.admin.project_pb2 import ProjectListRequest
from flyteidl.admin.signal_pb2 import SignalList, SignalListRequest, SignalSetRequest, SignalSetResponse
from flyteidl.service import admin_pb2_grpc as _admin_service
from flyteidl.service import dataproxy_pb2 as _dataproxy_pb2
from flyteidl.service import dataproxy_pb2_grpc as dataproxy_service
from flyteidl.service import signal_pb2_grpc as signal_service
from flyteidl.service.dataproxy_pb2_grpc import DataProxyServiceStub

from flytekit.clients.auth_helper import (
    get_channel,
    upgrade_channel_to_authenticated,
    upgrade_channel_to_proxy_authenticated,
    wrap_exceptions_channel,
)
from flytekit.configuration import PlatformConfig
from flytekit.loggers import logger


class RawSynchronousFlyteClient(object):
    """
    This is a thin synchronous wrapper around the auto-generated GRPC stubs for communicating with the admin service.

    This client should be usable regardless of environment in which this is used. In other words, configurations should
    be explicit as opposed to inferred from the environment or a configuration file. To create a client,

    .. code-block:: python

        from flytekit.configuration import PlatformConfig
        RawSynchronousFlyteClient(PlatformConfig(endpoint="a.b.com", insecure=True))  # or
        SynchronousFlyteClient(PlatformConfig(endpoint="a.b.com", insecure=True))
    """

    _dataproxy_stub: DataProxyServiceStub

    def __init__(self, cfg: PlatformConfig, **kwargs):
        """
        Initializes a gRPC channel to the given Flyte Admin service.

        Args:
          url: The server address.
          insecure: if insecure is desired
        """
        self._cfg = cfg
        self._channel = wrap_exceptions_channel(
            cfg, upgrade_channel_to_authenticated(cfg, upgrade_channel_to_proxy_authenticated(cfg, get_channel(cfg)))
        )
        self._stub = _admin_service.AdminServiceStub(self._channel)
        self._signal = signal_service.SignalServiceStub(self._channel)
        self._dataproxy_stub = dataproxy_service.DataProxyServiceStub(self._channel)

        logger.info(
            f"Flyte Client configured -> {self._cfg.endpoint} in {'insecure' if self._cfg.insecure else 'secure'} mode."
        )
        # metadata will hold the value of the token to send to the various endpoints.
        self._metadata = None

    @classmethod
    def with_root_certificate(cls, cfg: PlatformConfig, root_cert_file: str) -> RawSynchronousFlyteClient:
        b = None
        with open(root_cert_file, "rb") as fp:
            b = fp.read()
        return RawSynchronousFlyteClient(cfg, credentials=grpc.ssl_channel_credentials(root_certificates=b))

    @property
    def url(self) -> str:
        return self._cfg.endpoint

    ####################################################################################################################
    #
    #  Task Endpoints
    #
    ####################################################################################################################

    def create_task(self, task_create_request):
        """
        This will create a task definition in the Admin database. Once successful, the task object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly. This also means that as long as the request
            remains identical, calling this method multiple times will result in success.

        :param: flyteidl.admin.task_pb2.TaskCreateRequest task_create_request: The request protobuf object.
        :rtype: flyteidl.admin.task_pb2.TaskCreateResponse
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the task
            is found, this exception is raised.  The client might choose to ignore this exception because the identical
            task is already registered.
        :raises grpc.RpcError:
        """
        return self._stub.CreateTask(task_create_request, metadata=self._metadata)

    def list_task_ids_paginated(self, identifier_list_request):
        """
        This returns a page of identifiers for the tasks for a given project and domain. Filters can also be
        specified.

        .. note ::

            The name field in the TaskListRequest is ignored.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: flyteidl.admin.common_pb2.NamedEntityIdentifierListRequest identifier_list_request:
        :rtype: flyteidl.admin.common_pb2.NamedEntityIdentifierList
        :raises: TODO
        """
        return self._stub.ListTaskIds(identifier_list_request, metadata=self._metadata)

    def list_tasks_paginated(self, resource_list_request):
        """
        This returns a page of task metadata for tasks in a given project and domain.  Optionally,
        specifying a name will limit the results to only tasks with that name in the given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: flyteidl.admin.common_pb2.ResourceListRequest resource_list_request:
        :rtype: flyteidl.admin.task_pb2.TaskList
        :raises: TODO
        """
        return self._stub.ListTasks(resource_list_request, metadata=self._metadata)

    def get_task(self, get_object_request):
        """
        This returns a single task for a given identifier.

        :param: flyteidl.admin.common_pb2.ObjectGetRequest get_object_request:
        :rtype: flyteidl.admin.task_pb2.Task
        :raises: TODO
        """
        return self._stub.GetTask(get_object_request, metadata=self._metadata)

    def set_signal(self, signal_set_request: SignalSetRequest) -> SignalSetResponse:
        """
        This sets a signal
        """
        return self._signal.SetSignal(signal_set_request, metadata=self._metadata)

    def list_signals(self, signal_list_request: SignalListRequest) -> SignalList:
        """
        This lists signals
        """
        return self._signal.ListSignals(signal_list_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Workflow Endpoints
    #
    ####################################################################################################################

    def create_workflow(self, workflow_create_request):
        """
        This will create a workflow definition in the Admin database.  Once successful, the workflow object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly.  This also means that as long as the request
            remains identical, calling this method multiple times will result in success.

        :param: flyteidl.admin.workflow_pb2.WorkflowCreateRequest workflow_create_request:
        :rtype: flyteidl.admin.workflow_pb2.WorkflowCreateResponse
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the
            workflow is found, this exception is raised.  The client might choose to ignore this exception because the
            identical workflow is already registered.
        :raises grpc.RpcError:
        """
        return self._stub.CreateWorkflow(workflow_create_request, metadata=self._metadata)

    def list_workflow_ids_paginated(self, identifier_list_request):
        """
        This returns a page of identifiers for the workflows for a given project and domain. Filters can also be
        specified.

        .. note ::

            The name field in the WorkflowListRequest is ignored.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: flyteidl.admin.common_pb2.NamedEntityIdentifierListRequest identifier_list_request:
        :rtype: flyteidl.admin.common_pb2.NamedEntityIdentifierList
        :raises: TODO
        """
        return self._stub.ListWorkflowIds(identifier_list_request, metadata=self._metadata)

    def list_workflows_paginated(self, resource_list_request):
        """
        This returns a page of workflow meta-information for workflows in a given project and domain.  Optionally,
        specifying a name will limit the results to only workflows with that name in the given project and domain.

        .. note ::

            This is a paginated API.  Use the token field in the request to specify a page offset token.
            The user of the API is responsible for providing this token.

        .. note ::

            If entries are added to the database between requests for different pages, it is possible to receive
            entries on the second page that also appeared on the first.

        :param: flyteidl.admin.common_pb2.ResourceListRequest resource_list_request:
        :rtype: flyteidl.admin.workflow_pb2.WorkflowList
        :raises: TODO
        """
        return self._stub.ListWorkflows(resource_list_request, metadata=self._metadata)

    def get_workflow(self, get_object_request):
        """
        This returns a single workflow for a given identifier.

        :param: flyteidl.admin.common_pb2.ObjectGetRequest get_object_request:
        :rtype: flyteidl.admin.workflow_pb2.Workflow
        :raises: TODO
        """
        return self._stub.GetWorkflow(get_object_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Launch Plan Endpoints
    #
    ####################################################################################################################

    def create_launch_plan(self, launch_plan_create_request):
        """
        This will create a launch plan definition in the Admin database.  Once successful, the launch plan object can be
        retrieved via the client or viewed via the UI or command-line interfaces.

        .. note ::

            Overwrites are not supported so any request for a given project, domain, name, and version that exists in
            the database must match the existing definition exactly.  This also means that as long as the request
            remains identical, calling this method multiple times will result in success.

        :param: flyteidl.admin.launch_plan_pb2.LaunchPlanCreateRequest launch_plan_create_request:  The request
            protobuf object
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanCreateResponse
        :raises flytekit.common.exceptions.user.FlyteEntityAlreadyExistsException: If an identical version of the
            launch plan is found, this exception is raised.  The client might choose to ignore this exception because
            the identical launch plan is already registered.
        :raises grpc.RpcError:
        """
        return self._stub.CreateLaunchPlan(launch_plan_create_request, metadata=self._metadata)

    # TODO: List endpoints when they come in

    def get_launch_plan(self, object_get_request):
        """
        Retrieves a launch plan entity.

        :param flyteidl.admin.common_pb2.ObjectGetRequest object_get_request:
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlan
        """
        return self._stub.GetLaunchPlan(object_get_request, metadata=self._metadata)

    def get_active_launch_plan(self, active_launch_plan_request):
        """
        Retrieves a launch plan entity.

        :param flyteidl.admin.common_pb2.ActiveLaunchPlanRequest active_launch_plan_request:
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlan
        """
        return self._stub.GetActiveLaunchPlan(active_launch_plan_request, metadata=self._metadata)

    def update_launch_plan(self, update_request):
        """
        Allows updates to a launch plan at a given identifier.  Currently, a launch plan may only have it's state
        switched between ACTIVE and INACTIVE.

        :param flyteidl.admin.launch_plan_pb2.LaunchPlanUpdateRequest update_request:
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanUpdateResponse
        """
        return self._stub.UpdateLaunchPlan(update_request, metadata=self._metadata)

    def list_launch_plan_ids_paginated(self, identifier_list_request):
        """
        Lists launch plan named identifiers for a given project and domain.

        :param: flyteidl.admin.common_pb2.NamedEntityIdentifierListRequest identifier_list_request:
        :rtype: flyteidl.admin.common_pb2.NamedEntityIdentifierList
        """
        return self._stub.ListLaunchPlanIds(identifier_list_request, metadata=self._metadata)

    def list_launch_plans_paginated(self, resource_list_request):
        """
        Lists Launch Plans for a given Identifier (project, domain, name)

        :param: flyteidl.admin.common_pb2.ResourceListRequest resource_list_request:
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanList
        """
        return self._stub.ListLaunchPlans(resource_list_request, metadata=self._metadata)

    def list_active_launch_plans_paginated(self, active_launch_plan_list_request):
        """
        Lists Active Launch Plans for a given (project, domain)

        :param: flyteidl.admin.common_pb2.ActiveLaunchPlanListRequest active_launch_plan_list_request:
        :rtype: flyteidl.admin.launch_plan_pb2.LaunchPlanList
        """
        return self._stub.ListActiveLaunchPlans(active_launch_plan_list_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Named Entity Endpoints
    #
    ####################################################################################################################

    def update_named_entity(self, update_named_entity_request):
        """
        :param flyteidl.admin.common_pb2.NamedEntityUpdateRequest update_named_entity_request:
        :rtype: flyteidl.admin.common_pb2.NamedEntityUpdateResponse
        """
        return self._stub.UpdateNamedEntity(update_named_entity_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Workflow Execution Endpoints
    #
    ####################################################################################################################

    def create_execution(self, create_execution_request):
        """
        This will create an execution for the given execution spec.
        :param flyteidl.admin.execution_pb2.ExecutionCreateRequest create_execution_request:
        :rtype: flyteidl.admin.execution_pb2.ExecutionCreateResponse
        """
        return self._stub.CreateExecution(create_execution_request, metadata=self._metadata)

    def recover_execution(self, recover_execution_request):
        """
        This will recreate an execution with the same spec as the one belonging to the given execution identifier.
        :param flyteidl.admin.execution_pb2.ExecutionRecoverRequest recover_execution_request:
        :rtype: flyteidl.admin.execution_pb2.ExecutionRecoverResponse
        """
        return self._stub.RecoverExecution(recover_execution_request, metadata=self._metadata)

    def get_execution(self, get_object_request):
        """
        Returns an execution of a workflow entity.

        :param flyteidl.admin.execution_pb2.WorkflowExecutionGetRequest get_object_request:
        :rtype: flyteidl.admin.execution_pb2.Execution
        """
        return self._stub.GetExecution(get_object_request, metadata=self._metadata)

    def get_execution_data(self, get_execution_data_request):
        """
        Returns signed URLs to LiteralMap blobs for an execution's inputs and outputs (when available).

        :param flyteidl.admin.execution_pb2.WorkflowExecutionGetRequest get_execution_data_request:
        :rtype: flyteidl.admin.execution_pb2.WorkflowExecutionGetDataResponse
        """
        return self._stub.GetExecutionData(get_execution_data_request, metadata=self._metadata)

    def get_execution_metrics(self, get_execution_metrics_request):
        """
        Returns metrics partitioning and categorizing the workflow execution time-series.

        :param flyteidl.admin.execution_pb2.WorkflowExecutionGetMetricsRequest get_execution_metrics_request:
        :rtype: flyteidl.admin.execution_pb2.WorkflowExecutionGetMetricsResponse
        """
        return self._stub.GetExecutionMetrics(get_execution_metrics_request, metadata=self._metadata)

    def list_executions_paginated(self, resource_list_request):
        """
        Lists the executions for a given identifier.

        :param flyteidl.admin.common_pb2.ResourceListRequest resource_list_request:
        :rtype: flyteidl.admin.execution_pb2.ExecutionList
        """
        return self._stub.ListExecutions(resource_list_request, metadata=self._metadata)

    def terminate_execution(self, terminate_execution_request):
        """
        :param flyteidl.admin.execution_pb2.TerminateExecutionRequest terminate_execution_request:
        :rtype: flyteidl.admin.execution_pb2.TerminateExecutionResponse
        """
        return self._stub.TerminateExecution(terminate_execution_request, metadata=self._metadata)

    def relaunch_execution(self, relaunch_execution_request):
        """
        :param flyteidl.admin.execution_pb2.ExecutionRelaunchRequest relaunch_execution_request:
        :rtype: flyteidl.admin.execution_pb2.ExecutionCreateResponse
        """
        return self._stub.RelaunchExecution(relaunch_execution_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Node Execution Endpoints
    #
    ####################################################################################################################

    def get_node_execution(self, node_execution_request):
        """
        :param flyteidl.admin.node_execution_pb2.NodeExecutionGetRequest node_execution_request:
        :rtype: flyteidl.admin.node_execution_pb2.NodeExecution
        """
        return self._stub.GetNodeExecution(node_execution_request, metadata=self._metadata)

    def get_node_execution_data(self, get_node_execution_data_request):
        """
        Returns signed URLs to LiteralMap blobs for a node execution's inputs and outputs (when available).

        :param flyteidl.admin.node_execution_pb2.NodeExecutionGetDataRequest get_node_execution_data_request:
        :rtype: flyteidl.admin.node_execution_pb2.NodeExecutionGetDataResponse
        """
        return self._stub.GetNodeExecutionData(get_node_execution_data_request, metadata=self._metadata)

    def list_node_executions_paginated(self, node_execution_list_request):
        """
        :param flyteidl.admin.node_execution_pb2.NodeExecutionListRequest node_execution_list_request:
        :rtype: flyteidl.admin.node_execution_pb2.NodeExecutionList
        """
        return self._stub.ListNodeExecutions(node_execution_list_request, metadata=self._metadata)

    def list_node_executions_for_task_paginated(self, node_execution_for_task_list_request):
        """
        :param flyteidl.admin.node_execution_pb2.NodeExecutionListRequest node_execution_for_task_list_request:
        :rtype: flyteidl.admin.node_execution_pb2.NodeExecutionList
        """
        return self._stub.ListNodeExecutionsForTask(node_execution_for_task_list_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Task Execution Endpoints
    #
    ####################################################################################################################

    def get_task_execution(self, task_execution_request):
        """
        :param flyteidl.admin.task_execution_pb2.TaskExecutionGetRequest task_execution_request:
        :rtype: flyteidl.admin.task_execution_pb2.TaskExecution
        """
        return self._stub.GetTaskExecution(task_execution_request, metadata=self._metadata)

    def get_task_execution_data(self, get_task_execution_data_request):
        """
        Returns signed URLs to LiteralMap blobs for a task execution's inputs and outputs (when available).

        :param flyteidl.admin.task_execution_pb2.TaskExecutionGetDataRequest get_task_execution_data_request:
        :rtype: flyteidl.admin.task_execution_pb2.TaskExecutionGetDataResponse
        """
        return self._stub.GetTaskExecutionData(get_task_execution_data_request, metadata=self._metadata)

    def list_task_executions_paginated(self, task_execution_list_request):
        """
        :param flyteidl.admin.task_execution_pb2.TaskExecutionListRequest task_execution_list_request:
        :rtype: flyteidl.admin.task_execution_pb2.TaskExecutionList
        """
        return self._stub.ListTaskExecutions(task_execution_list_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Project Endpoints
    #
    ####################################################################################################################

    def list_projects(self, project_list_request: typing.Optional[ProjectListRequest] = None):
        """
        This will return a list of the projects registered with the Flyte Admin Service
        :param flyteidl.admin.project_pb2.ProjectListRequest project_list_request:
        :rtype: flyteidl.admin.project_pb2.Projects
        """
        if project_list_request is None:
            project_list_request = ProjectListRequest()
        return self._stub.ListProjects(project_list_request, metadata=self._metadata)

    def register_project(self, project_register_request):
        """
        Registers a project along with a set of domains.
        :param flyteidl.admin.project_pb2.ProjectRegisterRequest project_register_request:
        :rtype: flyteidl.admin.project_pb2.ProjectRegisterResponse
        """
        return self._stub.RegisterProject(project_register_request, metadata=self._metadata)

    def update_project(self, project):
        """
        Update an existing project specified by id.
        :param flyteidl.admin.project_pb2.Project project:
        :rtype: flyteidl.admin.project_pb2.ProjectUpdateResponse
        """
        return self._stub.UpdateProject(project, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Matching Attributes Endpoints
    #
    ####################################################################################################################
    def update_project_domain_attributes(self, project_domain_attributes_update_request):
        """
        This updates the attributes for a project and domain registered with the Flyte Admin Service
        :param flyteidl.admin.ProjectDomainAttributesUpdateRequest project_domain_attributes_update_request:
        :rtype: flyteidl.admin.ProjectDomainAttributesUpdateResponse
        """
        return self._stub.UpdateProjectDomainAttributes(
            project_domain_attributes_update_request, metadata=self._metadata
        )

    def update_workflow_attributes(self, workflow_attributes_update_request):
        """
        This updates the attributes for a project, domain, and workflow registered with the Flyte Admin Service
        :param flyteidl.admin.UpdateWorkflowAttributesRequest workflow_attributes_update_request:
        :rtype: flyteidl.admin.WorkflowAttributesUpdateResponse
        """
        return self._stub.UpdateWorkflowAttributes(workflow_attributes_update_request, metadata=self._metadata)

    def get_project_domain_attributes(self, project_domain_attributes_get_request):
        """
        This fetches the attributes for a project and domain registered with the Flyte Admin Service
        :param flyteidl.admin.ProjectDomainAttributesGetRequest project_domain_attributes_get_request:
        :rtype: flyteidl.admin.ProjectDomainAttributesGetResponse
        """
        return self._stub.GetProjectDomainAttributes(project_domain_attributes_get_request, metadata=self._metadata)

    def get_workflow_attributes(self, workflow_attributes_get_request):
        """
        This fetches the attributes for a project, domain, and workflow registered with the Flyte Admin Service
        :param flyteidl.admin.GetWorkflowAttributesAttributesRequest workflow_attributes_get_request:
        :rtype: flyteidl.admin.WorkflowAttributesGetResponse
        """
        return self._stub.GetWorkflowAttributes(workflow_attributes_get_request, metadata=self._metadata)

    def list_matchable_attributes(self, matchable_attributes_list_request):
        """
        This fetches the attributes for a specific resource type registered with the Flyte Admin Service
        :param flyteidl.admin.ListMatchableAttributesRequest matchable_attributes_list_request:
        :rtype: flyteidl.admin.ListMatchableAttributesResponse
        """
        return self._stub.ListMatchableAttributes(matchable_attributes_list_request, metadata=self._metadata)

    ####################################################################################################################
    #
    #  Event Endpoints
    #
    ####################################################################################################################

    # TODO: (P2) Implement the event endpoints in case there becomes a use-case for third-parties to submit events
    # through the client in Python.

    ####################################################################################################################
    #
    #  Data proxy endpoints
    #
    ####################################################################################################################
    def create_upload_location(
        self, create_upload_location_request: _dataproxy_pb2.CreateUploadLocationRequest
    ) -> _dataproxy_pb2.CreateUploadLocationResponse:
        """
        Get a signed url to be used during fast registration
        :param flyteidl.service.dataproxy_pb2.CreateUploadLocationRequest create_upload_location_request:
        :rtype: flyteidl.service.dataproxy_pb2.CreateUploadLocationResponse
        """
        return self._dataproxy_stub.CreateUploadLocation(create_upload_location_request, metadata=self._metadata)

    def create_download_location(
        self, create_download_location_request: _dataproxy_pb2.CreateDownloadLocationRequest
    ) -> _dataproxy_pb2.CreateDownloadLocationResponse:
        return self._dataproxy_stub.CreateDownloadLocation(create_download_location_request, metadata=self._metadata)

    def create_download_link(
        self, create_download_link_request: _dataproxy_pb2.CreateDownloadLinkRequest
    ) -> _dataproxy_pb2.CreateDownloadLinkResponse:
        return self._dataproxy_stub.CreateDownloadLink(create_download_link_request, metadata=self._metadata)
