"""
This module provides the ``FlyteRemote`` object, which is the end-user's main starting point for interacting
with a Flyte backend in an interactive and programmatic way. This of this experience as kind of like the web UI
but in Python object form.
"""
from __future__ import annotations

import base64
import configparser
import functools
import hashlib
import os
import pathlib
import tempfile
import time
import typing
import uuid
from base64 import b64encode
from collections import OrderedDict
from dataclasses import asdict, dataclass
from datetime import datetime, timedelta

import click
import fsspec
import requests
from flyteidl.admin.signal_pb2 import Signal, SignalListRequest, SignalSetRequest
from flyteidl.core import literals_pb2

from flytekit import ImageSpec
from flytekit.clients.friendly import SynchronousFlyteClient
from flytekit.clients.helpers import iterate_node_executions, iterate_task_executions
from flytekit.configuration import Config, FastSerializationSettings, ImageConfig, SerializationSettings
from flytekit.core import constants, utils
from flytekit.core.artifact import Artifact
from flytekit.core.base_task import PythonTask
from flytekit.core.context_manager import FlyteContext, FlyteContextManager
from flytekit.core.data_persistence import FileAccessProvider
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.python_auto_container import PythonAutoContainerTask
from flytekit.core.reference_entity import ReferenceSpec
from flytekit.core.type_engine import LiteralsResolver, TypeEngine
from flytekit.core.workflow import WorkflowBase, WorkflowFailurePolicy
from flytekit.exceptions import user as user_exceptions
from flytekit.exceptions.user import (
    FlyteEntityAlreadyExistsException,
    FlyteEntityNotExistException,
    FlyteValueException,
)
from flytekit.loggers import logger
from flytekit.models import common as common_models
from flytekit.models import filters as filter_models
from flytekit.models import launch_plan as launch_plan_models
from flytekit.models import literals as literal_models
from flytekit.models import task as task_models
from flytekit.models import types as type_models
from flytekit.models.admin import common as admin_common_models
from flytekit.models.admin import workflow as admin_workflow_models
from flytekit.models.admin.common import Sort
from flytekit.models.core import workflow as workflow_model
from flytekit.models.core.identifier import Identifier, ResourceType, SignalIdentifier, WorkflowExecutionIdentifier
from flytekit.models.core.workflow import NodeMetadata
from flytekit.models.execution import (
    ClusterAssignment,
    ExecutionMetadata,
    ExecutionSpec,
    NodeExecutionGetDataResponse,
    NotificationList,
    WorkflowExecutionGetDataResponse,
)
from flytekit.models.launch_plan import LaunchPlanState
from flytekit.models.literals import Literal, LiteralMap
from flytekit.remote.backfill import create_backfill_workflow
from flytekit.remote.data import download_literal
from flytekit.remote.entities import FlyteLaunchPlan, FlyteNode, FlyteTask, FlyteTaskNode, FlyteWorkflow
from flytekit.remote.executions import FlyteNodeExecution, FlyteTaskExecution, FlyteWorkflowExecution
from flytekit.remote.interface import TypedInterface
from flytekit.remote.lazy_entity import LazyEntity
from flytekit.remote.remote_callable import RemoteEntity
from flytekit.remote.remote_fs import get_flyte_fs
from flytekit.tools.fast_registration import fast_package
from flytekit.tools.interactive import ipython_check
from flytekit.tools.script_mode import compress_scripts, hash_file
from flytekit.tools.translator import (
    FlyteControlPlaneEntity,
    FlyteLocalEntity,
    Options,
    get_serializable,
    get_serializable_launch_plan,
)

if typing.TYPE_CHECKING:
    try:
        from IPython.core.display import HTML
    except ImportError:
        ...

ExecutionDataResponse = typing.Union[WorkflowExecutionGetDataResponse, NodeExecutionGetDataResponse]

MOST_RECENT_FIRST = admin_common_models.Sort("created_at", admin_common_models.Sort.Direction.DESCENDING)


class RegistrationSkipped(Exception):
    """
    RegistrationSkipped error is raised when trying to register an entity that is not registrable.
    """

    pass


@dataclass
class ResolvedIdentifiers:
    project: str
    domain: str
    name: str
    version: str


def _get_latest_version(list_entities_method: typing.Callable, project: str, domain: str, name: str):
    named_entity = common_models.NamedEntityIdentifier(project, domain, name)
    entity_list, _ = list_entities_method(
        named_entity,
        limit=1,
        sort_by=Sort("created_at", Sort.Direction.DESCENDING),
    )
    admin_entity = None if not entity_list else entity_list[0]
    if not admin_entity:
        raise user_exceptions.FlyteEntityNotExistException("Named entity {} not found".format(named_entity))
    return admin_entity.id.version


def _get_entity_identifier(
    list_entities_method: typing.Callable,
    resource_type: int,  # from flytekit.models.core.identifier.ResourceType
    project: str,
    domain: str,
    name: str,
    version: typing.Optional[str] = None,
):
    return Identifier(
        resource_type,
        project,
        domain,
        name,
        version if version is not None else _get_latest_version(list_entities_method, project, domain, name),
    )


def _get_git_repo_url(source_path):
    """
    Get git repo URL from remote.origin.url
    """
    try:
        git_config = source_path / ".git" / "config"
        if not git_config.exists():
            raise ValueError(f"{source_path} is not a git repo")

        config = configparser.ConfigParser()
        config.read(git_config)
        url = config['remote "origin"']["url"]

        if url.startswith("git@"):
            # url format: git@github.com:flytekit/flytekit.git
            prefix_len, suffix_len = len("git@"), len(".git")
            return url[prefix_len:-suffix_len].replace(":", "/")
        elif url.startswith("https://"):
            # url format: https://github.com/flytekit/flytekit
            prefix_len = len("https://")
            return url[prefix_len:]
        elif url.startswith("http://"):
            # url format: http://github.com/flytekit/flytekit
            prefix_len = len("http://")
            return url[prefix_len:]
        else:
            raise ValueError("Unable to parse url")

    except Exception as e:
        logger.debug(str(e))
        return ""


class FlyteRemote(object):
    """Main entrypoint for programmatically accessing a Flyte remote backend.

    The term 'remote' is synonymous with 'backend' or 'deployment' and refers to a hosted instance of the
    Flyte platform, which comes with a Flyte Admin server on some known URI.
    """

    def __init__(
        self,
        config: Config,
        default_project: typing.Optional[str] = None,
        default_domain: typing.Optional[str] = None,
        data_upload_location: str = "flyte://my-s3-bucket/",
        **kwargs,
    ):
        """Initialize a FlyteRemote object.

        :type kwargs: All arguments that can be passed to create the SynchronousFlyteClient. These are usually grpc
            parameters, if you want to customize credentials, ssl handling etc.
        :param default_project: default project to use when fetching or executing flyte entities.
        :param default_domain: default domain to use when fetching or executing flyte entities.
        :param data_upload_location: this is where all the default data will be uploaded when providing inputs.
            The default location - `s3://my-s3-bucket/data` works for sandbox/demo environment. Please override this for non-sandbox cases.
        """
        if config is None or config.platform is None or config.platform.endpoint is None:
            raise user_exceptions.FlyteAssertion("Flyte endpoint should be provided.")

        if data_upload_location is None:
            data_upload_location = FlyteContext.current_context().file_access.raw_output_prefix
        self._kwargs = kwargs
        self._client_initialized = False
        self._config = config
        # read config files, env vars, host, ssl options for admin client
        self._default_project = default_project
        self._default_domain = default_domain

        fsspec.register_implementation("flyte", get_flyte_fs(remote=self), clobber=True)

        self._file_access = FileAccessProvider(
            local_sandbox_dir=os.path.join(config.local_sandbox_path, "control_plane_metadata"),
            raw_output_prefix=data_upload_location,
            data_config=config.data_config,
        )

        # Save the file access object locally, build a context for it and save that as well.
        self._ctx = FlyteContextManager.current_context().with_file_access(self._file_access).build()

    @property
    def context(self) -> FlyteContext:
        return self._ctx

    @property
    def client(self) -> SynchronousFlyteClient:
        """Return a SynchronousFlyteClient for additional operations."""
        if not self._client_initialized:
            self._client = SynchronousFlyteClient(self.config.platform, **self._kwargs)
            self._client_initialized = True
        return self._client

    @property
    def default_project(self) -> str:
        """Default project to use when fetching or executing flyte entities."""
        return self._default_project

    @property
    def default_domain(self) -> str:
        """Default project to use when fetching or executing flyte entities."""
        return self._default_domain

    @property
    def config(self) -> Config:
        """Image config."""
        return self._config

    @property
    def file_access(self) -> FileAccessProvider:
        """File access provider to use for offloading non-literal inputs/outputs."""
        return self._file_access

    def get(
        self, flyte_uri: typing.Optional[str] = None
    ) -> typing.Optional[typing.Union[LiteralsResolver, Literal, HTML, bytes]]:
        """
        General function that works with flyte tiny urls. This can return outputs (in the form of LiteralsResolver, or
        individual Literals for singular requests), or HTML if passed a deck link, or bytes containing HTML,
        if ipython is not available locally.
        """
        if flyte_uri is None:
            raise user_exceptions.FlyteUserException("flyte_uri cannot be empty")
        ctx = self._ctx or FlyteContextManager.current_context()
        try:
            data_response = self.client.get_data(flyte_uri)

            if data_response.HasField("literal_map"):
                lm = LiteralMap.from_flyte_idl(data_response.literal_map)
                return LiteralsResolver(lm.literals)
            elif data_response.HasField("literal"):
                return Literal.from_flyte_idl(data_response.literal)
            elif data_response.HasField("pre_signed_urls"):
                if len(data_response.pre_signed_urls.signed_url) == 0:
                    raise ValueError(f"Flyte url {flyte_uri} resolved to empty download link")
                d = data_response.pre_signed_urls.signed_url[0]
                logger.debug(f"Download link is {d}")
                fs = ctx.file_access.get_filesystem_for_path(d)

                # If the venv has IPython, then return IPython's HTML
                if ipython_check():
                    from IPython.core.display import HTML

                    logger.debug(f"IPython found, returning HTML from {flyte_uri}")
                    with fs.open(d, "rb") as r:
                        html = HTML(str(r.read()))
                        return html
                # If not return bytes
                else:
                    logger.debug(f"IPython not found, returning HTML as bytes from {flyte_uri}")
                    return fs.open(d, "rb").read()

        except user_exceptions.FlyteUserException as e:
            logger.info(f"Error from Flyte backend when trying to fetch data: {e.__cause__}")

        logger.info(f"Nothing found from {flyte_uri}")

    def remote_context(self):
        """Context manager with remote-specific configuration."""
        return FlyteContextManager.with_context(
            FlyteContextManager.current_context().with_file_access(self.file_access)
        )

    def fetch_task_lazy(
        self, project: str = None, domain: str = None, name: str = None, version: str = None
    ) -> LazyEntity:
        """
        Similar to fetch_task, just that it returns a LazyEntity, which will fetch the workflow lazily.
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")

        def _fetch():
            return self.fetch_task(project=project, domain=domain, name=name, version=version)

        return LazyEntity(name=name, getter=_fetch)

    def fetch_task(self, project: str = None, domain: str = None, name: str = None, version: str = None) -> FlyteTask:
        """Fetch a task entity from flyte admin.

        :param project: fetch entity from this project. If None, uses the default_project attribute.
        :param domain: fetch entity from this domain. If None, uses the default_domain attribute.
        :param name: fetch entity with matching name.
        :param version: fetch entity with matching version. If None, gets the latest version of the entity.
        :returns: :class:`~flytekit.remote.tasks.task.FlyteTask`

        :raises: FlyteAssertion if name is None
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")
        task_id = _get_entity_identifier(
            self.client.list_tasks_paginated,
            ResourceType.TASK,
            project or self.default_project,
            domain or self.default_domain,
            name,
            version,
        )
        admin_task = self.client.get_task(task_id)
        flyte_task = FlyteTask.promote_from_model(admin_task.closure.compiled_task.template)
        flyte_task.template._id = task_id
        return flyte_task

    def fetch_workflow_lazy(
        self, project: str = None, domain: str = None, name: str = None, version: str = None
    ) -> LazyEntity[FlyteWorkflow]:
        """
        Similar to fetch_workflow, just that it returns a LazyEntity, which will fetch the workflow lazily.
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")

        def _fetch():
            return self.fetch_workflow(project, domain, name, version)

        return LazyEntity(name=name, getter=_fetch)

    def fetch_workflow(
        self, project: str = None, domain: str = None, name: str = None, version: str = None
    ) -> FlyteWorkflow:
        """
        Fetch a workflow entity from flyte admin.
        :param project: fetch entity from this project. If None, uses the default_project attribute.
        :param domain: fetch entity from this domain. If None, uses the default_domain attribute.
        :param name: fetch entity with matching name.
        :param version: fetch entity with matching version. If None, gets the latest version of the entity.
        :raises: FlyteAssertion if name is None
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")
        workflow_id = _get_entity_identifier(
            self.client.list_workflows_paginated,
            ResourceType.WORKFLOW,
            project or self.default_project,
            domain or self.default_domain,
            name,
            version,
        )

        admin_workflow = self.client.get_workflow(workflow_id)
        compiled_wf = admin_workflow.closure.compiled_workflow

        wf_templates = [compiled_wf.primary.template]
        wf_templates.extend([swf.template for swf in compiled_wf.sub_workflows])

        node_launch_plans = {}
        # TODO: Inspect branch nodes for launch plans
        for wf_template in wf_templates:
            for node in FlyteWorkflow.get_non_system_nodes(wf_template.nodes):
                if node.workflow_node is not None and node.workflow_node.launchplan_ref is not None:
                    lp_ref = node.workflow_node.launchplan_ref
                    if node.workflow_node.launchplan_ref not in node_launch_plans:
                        admin_launch_plan = self.client.get_launch_plan(lp_ref)
                        node_launch_plans[node.workflow_node.launchplan_ref] = admin_launch_plan.spec

        return FlyteWorkflow.promote_from_closure(compiled_wf, node_launch_plans)

    def fetch_launch_plan(
        self, project: str = None, domain: str = None, name: str = None, version: str = None
    ) -> FlyteLaunchPlan:
        """Fetch a launchplan entity from flyte admin.

        :param project: fetch entity from this project. If None, uses the default_project attribute.
        :param domain: fetch entity from this domain. If None, uses the default_domain attribute.
        :param name: fetch entity with matching name.
        :param version: fetch entity with matching version. If None, gets the latest version of the entity.
        :returns: :class:`~flytekit.remote.launch_plan.FlyteLaunchPlan`

        :raises: FlyteAssertion if name is None
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")
        launch_plan_id = _get_entity_identifier(
            self.client.list_launch_plans_paginated,
            ResourceType.LAUNCH_PLAN,
            project or self.default_project,
            domain or self.default_domain,
            name,
            version,
        )
        admin_launch_plan = self.client.get_launch_plan(launch_plan_id)
        flyte_launch_plan = FlyteLaunchPlan.promote_from_model(launch_plan_id, admin_launch_plan.spec)

        wf_id = flyte_launch_plan.workflow_id
        workflow = self.fetch_workflow(wf_id.project, wf_id.domain, wf_id.name, wf_id.version)
        flyte_launch_plan._interface = workflow.interface
        flyte_launch_plan._flyte_workflow = workflow

        return flyte_launch_plan

    def fetch_execution(self, project: str = None, domain: str = None, name: str = None) -> FlyteWorkflowExecution:
        """Fetch a workflow execution entity from flyte admin.

        :param project: fetch entity from this project. If None, uses the default_project attribute.
        :param domain: fetch entity from this domain. If None, uses the default_domain attribute.
        :param name: fetch entity with matching name.
        :returns: :class:`~flytekit.remote.workflow_execution.FlyteWorkflowExecution`

        :raises: FlyteAssertion if name is None
        """
        if name is None:
            raise user_exceptions.FlyteAssertion("the 'name' argument must be specified.")
        execution = FlyteWorkflowExecution.promote_from_model(
            self.client.get_execution(
                WorkflowExecutionIdentifier(
                    project or self.default_project,
                    domain or self.default_domain,
                    name,
                )
            )
        )
        return self.sync_execution(execution)

    ######################
    #  Listing Entities  #
    ######################

    def list_signals(
        self,
        execution_name: str,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        limit: int = 100,
        filters: typing.Optional[typing.List[filter_models.Filter]] = None,
    ) -> typing.List[Signal]:
        """
        :param execution_name: The name of the execution. This is the tailend of the URL when looking at the workflow execution.
        :param project: The execution project, will default to the Remote's default project.
        :param domain: The execution domain, will default to the Remote's default domain.
        :param limit: The number of signals to fetch
        :param filters: Optional list of filters
        """
        wf_exec_id = WorkflowExecutionIdentifier(
            project=project or self.default_project, domain=domain or self.default_domain, name=execution_name
        )
        req = SignalListRequest(workflow_execution_id=wf_exec_id.to_flyte_idl(), limit=limit, filters=filters)
        resp = self.client.list_signals(req)
        s = resp.signals
        return s

    def set_signal(
        self,
        signal_id: str,
        execution_name: str,
        value: typing.Union[literal_models.Literal, typing.Any],
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        python_type: typing.Optional[typing.Type] = None,
        literal_type: typing.Optional[type_models.LiteralType] = None,
    ):
        """
        :param signal_id: The name of the signal, this is the key used in the approve() or wait_for_input() call.
        :param execution_name: The name of the execution. This is the tail-end of the URL when looking
            at the workflow execution.
        :param value: This is either a Literal or a Python value which FlyteRemote will invoke the TypeEngine to
            convert into a Literal. This argument is only value for wait_for_input type signals.
        :param project: The execution project, will default to the Remote's default project.
        :param domain: The execution domain, will default to the Remote's default domain.
        :param python_type: Provide a python type to help with conversion if the value you provided is not a Literal.
        :param literal_type: Provide a Flyte literal type to help with conversion if the value you provided
            is not a Literal
        """
        wf_exec_id = WorkflowExecutionIdentifier(
            project=project or self.default_project, domain=domain or self.default_domain, name=execution_name
        )
        if isinstance(value, Literal):
            logger.debug(f"Using provided {value} as existing Literal value")
            lit = value
        else:
            lt = literal_type or (
                TypeEngine.to_literal_type(python_type) if python_type else TypeEngine.to_literal_type(type(value))
            )
            lit = TypeEngine.to_literal(self.context, value, python_type or type(value), lt)
            logger.debug(f"Converted {value} to literal {lit} using literal type {lt}")

        req = SignalSetRequest(id=SignalIdentifier(signal_id, wf_exec_id).to_flyte_idl(), value=lit.to_flyte_idl())

        # Response is empty currently, nothing to give back to the user.
        self.client.set_signal(req)

    def recent_executions(
        self,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        limit: typing.Optional[int] = 100,
    ) -> typing.List[FlyteWorkflowExecution]:
        # Ignore token for now
        exec_models, _ = self.client.list_executions_paginated(
            project or self.default_project,
            domain or self.default_domain,
            limit,
            sort_by=MOST_RECENT_FIRST,
        )
        return [FlyteWorkflowExecution.promote_from_model(e) for e in exec_models]

    def list_tasks_by_version(
        self,
        version: str,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        limit: typing.Optional[int] = 100,
    ) -> typing.List[FlyteTask]:
        if not version:
            raise ValueError("Must specify a version")

        named_entity_id = common_models.NamedEntityIdentifier(
            project=project or self.default_project,
            domain=domain or self.default_domain,
        )
        # Ignore token for now
        t_models, _ = self.client.list_tasks_paginated(
            named_entity_id,
            filters=[filter_models.Filter.from_python_std(f"eq(version,{version})")],
            limit=limit,
        )
        return [FlyteTask.promote_from_model(t.closure.compiled_task.template) for t in t_models]

    #####################
    # Register Entities #
    #####################

    def _resolve_identifier(self, t: int, name: str, version: str, ss: SerializationSettings) -> Identifier:
        ident = Identifier(
            resource_type=t,
            project=ss.project if ss and ss.project else self.default_project,
            domain=ss.domain if ss and ss.domain else self.default_domain,
            name=name,
            version=version or ss.version,
        )
        if not ident.project or not ident.domain or not ident.name or not ident.version:
            raise ValueError(
                f"To register a new {ident.resource_type}, (project, domain, name, version) required, "
                f"received ({ident.project}, {ident.domain}, {ident.name}, {ident.version})."
            )
        return ident

    def raw_register(
        self,
        cp_entity: FlyteControlPlaneEntity,
        settings: SerializationSettings,
        version: str,
        create_default_launchplan: bool = True,
        options: Options = None,
        og_entity: FlyteLocalEntity = None,
    ) -> typing.Optional[Identifier]:
        """
        Raw register method, can be used to register control plane entities. Usually if you have a Flyte Entity like a
        WorkflowBase, Task, LaunchPlan then use other methods. This should be used only if you have already serialized entities

        :param cp_entity: The controlplane "serializable" version of a flyte entity. This is in the form that FlyteAdmin
            understands.
        :param settings: SerializationSettings to be used for registration - especially to identify the id
        :param version: Version to be registered
        :param create_default_launchplan: boolean that indicates if a default launch plan should be created
        :param options: Options to be used if registering a default launch plan
        :param og_entity: Pass in the original workflow (flytekit type) if create_default_launchplan is true
        :return: Identifier of the created entity
        """
        if isinstance(cp_entity, RemoteEntity):
            if isinstance(cp_entity, (FlyteWorkflow, FlyteTask)):
                if not cp_entity.should_register:
                    logger.debug(f"Skipping registration of remote entity: {cp_entity.name}")
                    raise RegistrationSkipped(f"Remote task/Workflow {cp_entity.name} is not registrable.")
            else:
                logger.debug(f"Skipping registration of remote entity: {cp_entity.name}")
                raise RegistrationSkipped(f"Remote task/Workflow {cp_entity.name} is not registrable.")

        if isinstance(
            cp_entity,
            (
                workflow_model.Node,
                workflow_model.WorkflowNode,
                workflow_model.BranchNode,
                workflow_model.TaskNode,
            ),
        ):
            logger.debug("Ignoring nodes for registration.")
            return None

        elif isinstance(cp_entity, ReferenceSpec):
            logger.debug(f"Skipping registration of Reference entity, name: {cp_entity.template.id.name}")
            return None

        if isinstance(cp_entity, task_models.TaskSpec):
            if isinstance(cp_entity, FlyteTask):
                version = cp_entity.id.version
            ident = self._resolve_identifier(ResourceType.TASK, cp_entity.template.id.name, version, settings)
            try:
                self.client.create_task(task_identifer=ident, task_spec=cp_entity)
            except FlyteEntityAlreadyExistsException:
                logger.info(f" {ident} Already Exists!")
            return ident

        if isinstance(cp_entity, admin_workflow_models.WorkflowSpec):
            if isinstance(cp_entity, FlyteWorkflow):
                version = cp_entity.id.version
            ident = self._resolve_identifier(ResourceType.WORKFLOW, cp_entity.template.id.name, version, settings)
            try:
                self.client.create_workflow(workflow_identifier=ident, workflow_spec=cp_entity)
            except FlyteEntityAlreadyExistsException:
                logger.info(f" {ident} Already Exists!")

            if create_default_launchplan:
                if not og_entity:
                    raise user_exceptions.FlyteValueException(
                        "To create default launch plan, please pass in the original flytekit workflow `og_entity`"
                    )

                # Let us also create a default launch-plan, ideally the default launchplan should be added
                # to the orderedDict, but we do not.
                self.file_access._get_upload_signed_url_fn = functools.partial(
                    self.client.get_upload_signed_url, project=settings.project, domain=settings.domain
                )
                default_lp = LaunchPlan.get_default_launch_plan(self.context, og_entity)
                lp_entity = get_serializable_launch_plan(
                    OrderedDict(),
                    settings,
                    default_lp,
                    recurse_downstream=False,
                    options=options,
                )
                try:
                    self.client.create_launch_plan(lp_entity.id, lp_entity.spec)
                except FlyteEntityAlreadyExistsException:
                    logger.info(f" {lp_entity.id} Already Exists!")
            return ident

        if isinstance(cp_entity, launch_plan_models.LaunchPlan):
            ident = self._resolve_identifier(ResourceType.LAUNCH_PLAN, cp_entity.id.name, version, settings)
            try:
                self.client.create_launch_plan(launch_plan_identifer=ident, launch_plan_spec=cp_entity.spec)
            except FlyteEntityAlreadyExistsException:
                logger.info(f" {ident} Already Exists!")
            return ident

        raise AssertionError(f"Unknown entity of type {type(cp_entity)}")

    def _serialize_and_register(
        self,
        entity: FlyteLocalEntity,
        settings: typing.Optional[SerializationSettings],
        version: str,
        options: typing.Optional[Options] = None,
        create_default_launchplan: bool = True,
    ) -> Identifier:
        """
        This method serializes and register the given Flyte entity
        :return: Identifier of the registered entity
        """
        m = OrderedDict()
        # Create dummy serialization settings for now.
        # TODO: Clean this up by using lazy usage of serialization settings in translator.py
        serialization_settings = settings
        is_dummy_serialization_setting = False
        if not settings:
            serialization_settings = SerializationSettings(
                ImageConfig.auto_default_image(),
                project=self.default_project,
                domain=self.default_domain,
                version=version,
            )
            is_dummy_serialization_setting = True

        if serialization_settings.version is None:
            serialization_settings.version = version

        _ = get_serializable(m, settings=serialization_settings, entity=entity, options=options)

        ident = None
        for entity, cp_entity in m.items():
            if not isinstance(cp_entity, admin_workflow_models.WorkflowSpec) and is_dummy_serialization_setting:
                # Only in the case of workflows can we use the dummy serialization settings.
                raise user_exceptions.FlyteValueException(
                    settings,
                    f"No serialization settings set, but workflow contains entities that need to be registered. {cp_entity.id.name}",
                )

            try:
                ident = self.raw_register(
                    cp_entity,
                    settings=settings,
                    version=version,
                    create_default_launchplan=create_default_launchplan,
                    options=options,
                    og_entity=entity,
                )
            except RegistrationSkipped:
                pass

        return ident

    def register_task(
        self, entity: PythonTask, serialization_settings: SerializationSettings, version: typing.Optional[str] = None
    ) -> FlyteTask:
        """
        Register a qualified task (PythonTask) with Remote
        For any conflicting parameters method arguments are regarded as overrides

        :param entity: PythonTask can be either @task or a instance of a Task class
        :param serialization_settings:  Settings that will be used to override various serialization parameters.
        :param version: version that will be used to register. If not specified will default to using the serialization settings default
        :return:
        """
        ident = self._serialize_and_register(entity=entity, settings=serialization_settings, version=version)
        ft = self.fetch_task(
            ident.project,
            ident.domain,
            ident.name,
            ident.version,
        )
        ft._python_interface = entity.python_interface
        return ft

    def register_workflow(
        self,
        entity: WorkflowBase,
        serialization_settings: typing.Optional[SerializationSettings] = None,
        version: typing.Optional[str] = None,
        default_launch_plan: typing.Optional[bool] = True,
        options: typing.Optional[Options] = None,
    ) -> FlyteWorkflow:
        """
        Use this method to register a workflow.
        :param version: version for the entity to be registered as
        :param entity: The workflow to be registered
        :param serialization_settings: The serialization settings to be used
        :param default_launch_plan: This should be true if a default launch plan should be created for the workflow
        :param options: Additional execution options that can be configured for the default launchplan
        :return:
        """
        ident = self._resolve_identifier(ResourceType.WORKFLOW, entity.name, version, serialization_settings)
        if serialization_settings:
            b = serialization_settings.new_builder()
            b.project = ident.project
            b.domain = ident.domain
            b.version = ident.version
            serialization_settings = b.build()
        ident = self._serialize_and_register(entity, serialization_settings, version, options, default_launch_plan)
        fwf = self.fetch_workflow(ident.project, ident.domain, ident.name, ident.version)
        fwf._python_interface = entity.python_interface
        return fwf

    def fast_package(self, root: os.PathLike, deref_symlinks: bool = True, output: str = None) -> (bytes, str):
        """
        Packages the given paths into an installable zip and returns the md5_bytes and the URL of the uploaded location
        :param root: path to the root of the package system that should be uploaded
        :param output: output path. Optional, will default to a tempdir
        :param deref_symlinks: if symlinks should be dereferenced. Defaults to True
        :return: md5_bytes, url
        """
        # Create a zip file containing all the entries.
        zip_file = fast_package(root, output, deref_symlinks)
        md5_bytes, _, _ = hash_file(pathlib.Path(zip_file))

        # Upload zip file to Admin using FlyteRemote.
        return self.upload_file(pathlib.Path(zip_file))

    def upload_file(
        self,
        to_upload: pathlib.Path,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
    ) -> typing.Tuple[bytes, str]:
        """
        Function will use remote's client to hash and then upload the file using Admin's data proxy service.

        :param to_upload: Must be a single file
        :param project: Project to upload under, if not supplied will use the remote's default
        :param domain: Domain to upload under, if not specified will use the remote's default
        :return: The uploaded location.
        """
        if not to_upload.is_file():
            raise ValueError(f"{to_upload} is not a single file, upload arg must be a single file.")
        md5_bytes, str_digest, _ = hash_file(to_upload)
        logger.debug(f"Text hash of file to upload is {str_digest}")

        upload_location = self.client.get_upload_signed_url(
            project=project or self.default_project,
            domain=domain or self.default_domain,
            content_md5=md5_bytes,
            filename=to_upload.name,
        )

        extra_headers = self.get_extra_headers_for_protocol(upload_location.native_url)
        encoded_md5 = b64encode(md5_bytes)
        with open(str(to_upload), "+rb") as local_file:
            content = local_file.read()
            content_length = len(content)
            headers = {"Content-Length": str(content_length), "Content-MD5": encoded_md5}
            headers.update(extra_headers)
            rsp = requests.put(
                upload_location.signed_url,
                data=content,
                headers=headers,
                verify=False
                if self._config.platform.insecure_skip_verify is True
                else self._config.platform.ca_cert_file_path,
            )

            # Check both HTTP 201 and 200, because some storage backends (e.g. Azure) return 201 instead of 200.
            if rsp.status_code not in (requests.codes["OK"], requests.codes["created"]):
                raise FlyteValueException(
                    rsp.status_code,
                    f"Request to send data {upload_location.signed_url} failed.\nResponse: {rsp.text}",
                )

        logger.debug(f"Uploading {to_upload} to {upload_location.signed_url} native url {upload_location.native_url}")

        return md5_bytes, upload_location.native_url

    @staticmethod
    def _version_from_hash(
        md5_bytes: bytes,
        serialization_settings: SerializationSettings,
        *additional_context: str,
    ) -> str:
        """
        The md5 version that we send to S3/GCS has to match the file contents exactly,
        but we don't have to use it when registering with the Flyte backend.
        To avoid changes in the For that add the hash of the compilation settings to hash of file

        :param md5_bytes:
        :param serialization_settings:
        :param additional_context: This is for additional context to factor into the version computation,
          meant for objects (like Options for instance) that don't easily consistently stringify.
        :return:
        """
        from flytekit import __version__

        additional_context = additional_context or []

        h = hashlib.md5(md5_bytes)
        h.update(bytes(serialization_settings.to_json(), "utf-8"))
        h.update(bytes(__version__, "utf-8"))

        for s in additional_context:
            h.update(bytes(s, "utf-8"))

        # Omit the character '=' from the version as that's essentially padding used by the base64 encoding
        # and does not increase entropy of the hash while making it very inconvenient to copy-and-paste.
        return base64.urlsafe_b64encode(h.digest()).decode("ascii").rstrip("=")

    def register_script(
        self,
        entity: typing.Union[WorkflowBase, PythonTask],
        image_config: typing.Optional[ImageConfig] = None,
        version: typing.Optional[str] = None,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        destination_dir: str = ".",
        copy_all: bool = False,
        default_launch_plan: bool = True,
        options: typing.Optional[Options] = None,
        source_path: typing.Optional[str] = None,
        module_name: typing.Optional[str] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
    ) -> typing.Union[FlyteWorkflow, FlyteTask]:
        """
        Use this method to register a workflow via script mode.
        :param destination_dir: The destination directory where the workflow will be copied to.
        :param copy_all: If true, the entire source directory will be copied over to the destination directory.
        :param domain: The domain to register the workflow in.
        :param project: The project to register the workflow in.
        :param image_config: The image config to use for the workflow.
        :param version: version for the entity to be registered as
        :param entity: The workflow to be registered or the task to be registered
        :param default_launch_plan: This should be true if a default launch plan should be created for the workflow
        :param options: Additional execution options that can be configured for the default launchplan
        :param source_path: The root of the project path
        :param module_name: the name of the module
        :param envs: Environment variables to be passed to the serialization
        :return:
        """
        if image_config is None:
            image_config = ImageConfig.auto_default_image()

        with tempfile.TemporaryDirectory() as tmp_dir:
            if copy_all:
                md5_bytes, upload_native_url = self.fast_package(pathlib.Path(source_path), False, tmp_dir)
            else:
                archive_fname = pathlib.Path(os.path.join(tmp_dir, "script_mode.tar.gz"))
                compress_scripts(source_path, str(archive_fname), module_name)
                md5_bytes, upload_native_url = self.upload_file(
                    archive_fname, project or self.default_project, domain or self.default_domain
                )

        serialization_settings = SerializationSettings(
            project=project,
            domain=domain,
            image_config=image_config,
            git_repo=_get_git_repo_url(source_path),
            env=envs,
            fast_serialization_settings=FastSerializationSettings(
                enabled=True,
                destination_dir=destination_dir,
                distribution_location=upload_native_url,
            ),
        )

        if version is None:

            def _get_image_names(entity: typing.Union[PythonAutoContainerTask, WorkflowBase]) -> typing.List[str]:
                if isinstance(entity, PythonAutoContainerTask) and isinstance(entity.container_image, ImageSpec):
                    return [entity.container_image.image_name()]
                if isinstance(entity, WorkflowBase):
                    image_names = []
                    for n in entity.nodes:
                        image_names.extend(_get_image_names(n.flyte_entity))
                    return image_names
                return []

            # The md5 version that we send to S3/GCS has to match the file contents exactly,
            # but we don't have to use it when registering with the Flyte backend.
            # For that add the hash of the compilation settings to hash of file
            version = self._version_from_hash(md5_bytes, serialization_settings, *_get_image_names(entity))

        if isinstance(entity, PythonTask):
            return self.register_task(entity, serialization_settings, version)
        return self.register_workflow(entity, serialization_settings, version, default_launch_plan, options)

    def register_launch_plan(
        self,
        entity: LaunchPlan,
        version: str,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        options: typing.Optional[Options] = None,
    ) -> FlyteLaunchPlan:
        """
        Register a given launchplan, possibly applying overrides from the provided options.
        :param entity: Launchplan to be registered
        :param version:
        :param project: Optionally provide a project, if not already provided in flyteremote constructor or a separate one
        :param domain: Optionally provide a domain, if not already provided in FlyteRemote constructor or a separate one
        :param options:
        :return:
        """
        ss = SerializationSettings(image_config=ImageConfig(), project=project, domain=domain, version=version)

        ident = self._resolve_identifier(ResourceType.LAUNCH_PLAN, entity.name, version, ss)
        m = OrderedDict()
        idl_lp = get_serializable_launch_plan(m, ss, entity, recurse_downstream=False, options=options)
        try:
            self.client.create_launch_plan(ident, idl_lp.spec)
        except FlyteEntityAlreadyExistsException:
            logger.debug("Launchplan already exists, ignoring")
        flp = self.fetch_launch_plan(ident.project, ident.domain, ident.name, ident.version)
        flp._python_interface = entity.python_interface
        return flp

    ####################
    # Execute Entities #
    ####################

    def _execute(
        self,
        entity: typing.Union[FlyteTask, FlyteWorkflow, FlyteLaunchPlan],
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        type_hints: typing.Optional[typing.Dict[str, typing.Type]] = None,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """Common method for execution across all entities.

        :param flyte_id: entity identifier
        :param inputs: dictionary mapping argument names to values
        :param project: project on which to execute the entity referenced by flyte_id
        :param domain: domain on which to execute the entity referenced by flyte_id
        :param execution_name: name of the execution
        :param wait: if True, waits for execution to complete
        :param type_hints: map of python types to inputs so that the TypeEngine knows how to convert the input values
          into Flyte Literals.
        :param overwrite_cache: Allows for all cached values of a workflow and its tasks to be overwritten
          for a single execution. If enabled, all calculations are performed even if cached results would
          be available, overwriting the stored data once execution finishes successfully.
        :param envs: Environment variables to set for the execution.
        :param tags: Tags to set for the execution.
        :param cluster_pool: Specify cluster pool on which newly created execution should be placed.
        :returns: :class:`~flytekit.remote.workflow_execution.FlyteWorkflowExecution`
        """
        if execution_name is not None and execution_name_prefix is not None:
            raise ValueError("Only one of execution_name and execution_name_prefix can be set, but got both set")
        execution_name_prefix = execution_name_prefix + "-" if execution_name_prefix is not None else None
        execution_name = execution_name or (execution_name_prefix or "f") + uuid.uuid4().hex[:19]
        if not options:
            options = Options()
        if options.disable_notifications is not None:
            if options.disable_notifications:
                notifications = None
            else:
                notifications = NotificationList(options.notifications)
        else:
            notifications = NotificationList([])

        type_hints = type_hints or {}
        literal_map = {}
        with self.remote_context() as ctx:
            input_flyte_type_map = entity.interface.inputs

            for k, v in inputs.items():
                if input_flyte_type_map.get(k) is None:
                    raise user_exceptions.FlyteValueException(
                        k, f"The {entity.__class__.__name__} doesn't have this input key."
                    )
                if isinstance(v, Literal):
                    lit = v
                elif isinstance(v, Artifact):
                    raise user_exceptions.FlyteValueException(v, "Running with an artifact object is not yet possible.")
                else:
                    if k not in type_hints:
                        try:
                            type_hints[k] = TypeEngine.guess_python_type(input_flyte_type_map[k].type)
                        except ValueError:
                            logger.debug(f"Could not guess type for {input_flyte_type_map[k].type}, skipping...")
                    variable = entity.interface.inputs.get(k)
                    hint = type_hints[k]
                    self.file_access._get_upload_signed_url_fn = functools.partial(
                        self.client.get_upload_signed_url,
                        project=project or self.default_project,
                        domain=domain or self.default_domain,
                    )
                    lit = TypeEngine.to_literal(ctx, v, hint, variable.type)
                literal_map[k] = lit

            literal_inputs = literal_models.LiteralMap(literals=literal_map)

        try:
            # Currently, this will only execute the flyte entity referenced by
            # flyte_id in the same project and domain. However, it is possible to execute it in a different project
            # and domain, which is specified in the first two arguments of client.create_execution. This is useful
            # in the case that I want to use a flyte entity from e.g. project "A" but actually execute the entity on a
            # different project "B". For now, this method doesn't support this use case.
            exec_id = self.client.create_execution(
                project or self.default_project,
                domain or self.default_domain,
                execution_name,
                ExecutionSpec(
                    entity.id,
                    ExecutionMetadata(
                        ExecutionMetadata.ExecutionMode.MANUAL,
                        "placeholder",  # Admin replaces this from oidc token if auth is enabled.
                        0,
                    ),
                    overwrite_cache=overwrite_cache,
                    notifications=notifications,
                    disable_all=options.disable_notifications,
                    labels=options.labels,
                    annotations=options.annotations,
                    raw_output_data_config=options.raw_output_data_config,
                    auth_role=None,
                    max_parallelism=options.max_parallelism,
                    security_context=options.security_context,
                    envs=common_models.Envs(envs) if envs else None,
                    tags=tags,
                    cluster_assignment=ClusterAssignment(cluster_pool=cluster_pool) if cluster_pool else None,
                ),
                literal_inputs,
            )
        except user_exceptions.FlyteEntityAlreadyExistsException:
            logger.warning(
                f"Execution with Execution ID {execution_name} already exists. "
                f"Assuming this is the same execution, returning!"
            )
            exec_id = WorkflowExecutionIdentifier(
                project=project or self.default_project, domain=domain or self.default_domain, name=execution_name
            )
        execution = FlyteWorkflowExecution.promote_from_model(self.client.get_execution(exec_id))

        if wait:
            return self.wait(execution)
        return execution

    def _resolve_identifier_kwargs(
        self,
        entity: typing.Any,
        project: str,
        domain: str,
        name: str,
        version: str,
    ) -> ResolvedIdentifiers:
        """
        Resolves the identifier attributes based on user input, falling back on the default project/domain and
        auto-generated version, and ultimately the entity project/domain if entity is a remote flyte entity.
        """
        ident = ResolvedIdentifiers(
            project=project or self.default_project,
            domain=domain or self.default_domain,
            name=name or entity.name,
            version=version,
        )
        if not (ident.project and ident.domain and ident.name):
            raise ValueError(
                f"Cannot launch an execution with missing project/domain/name {ident} for entity type {type(entity)}."
                f" Specify them in the execute method or when initializing FlyteRemote"
            )
        return ident

    def execute(
        self,
        entity: typing.Union[FlyteTask, FlyteLaunchPlan, FlyteWorkflow, PythonTask, WorkflowBase, LaunchPlan],
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        name: str = None,
        version: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        image_config: typing.Optional[ImageConfig] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        type_hints: typing.Optional[typing.Dict[str, typing.Type]] = None,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """
        Execute a task, workflow, or launchplan, either something that's been declared locally, or a fetched entity.

        This method supports:
        - ``Flyte{Task, Workflow, LaunchPlan}`` remote module objects.
        - ``@task``-decorated functions and ``TaskTemplate`` tasks.
        - ``@workflow``-decorated functions.
        - ``LaunchPlan`` objects.

        For local entities, this code will attempt to find the entity first, and if missing, will compile and register
        the object.

        Not all arguments are relevant in all circumstances. For example, there's no reason to use the serialization
        settings for entities that have already been registered on Admin.

        :param options:
        :param entity: entity to execute
        :param inputs: dictionary mapping argument names to values
        :param project: execute entity in this project. If entity doesn't exist in the project, register the entity
            first before executing.
        :param domain: execute entity in this domain. If entity doesn't exist in the domain, register the entity
            first before executing.
        :param name: execute entity using this name. If not None, use this value instead of ``entity.name``
        :param version: execute entity using this version. If None, uses auto-generated value.
        :param execution_name: name of the execution. If None, uses auto-generated value.
        :param image_config:
        :param wait: if True, waits for execution to complete
        :param type_hints: Python types to be passed to the TypeEngine so that it knows how to properly convert the
          input values for the execution into Flyte literals. If missing, will default to first guessing the type
          using the type engine, and then to ``type(v)``. Providing the correct Python types is particularly important
          if the inputs are containers like lists or maps, or if the Python type is one of the more complex Flyte
          provided classes (like a StructuredDataset that's annotated with columns).
        :param overwrite_cache: Allows for all cached values of a workflow and its tasks to be overwritten
          for a single execution. If enabled, all calculations are performed even if cached results would
          be available, overwriting the stored data once execution finishes successfully.
        :param envs: Environment variables to be set for the execution.
        :param tags: Tags to be set for the execution.
        :param cluster_pool: Specify cluster pool on which newly created execution should be placed.

        .. note:

            The ``name`` and ``version`` arguments do not apply to ``FlyteTask``, ``FlyteLaunchPlan``, and
            ``FlyteWorkflow`` entity inputs. These values are determined by referencing the entity identifier values.
        """
        if entity.python_interface:
            type_hints = type_hints or entity.python_interface.inputs
        if isinstance(entity, FlyteTask) or isinstance(entity, FlyteLaunchPlan):
            return self.execute_remote_task_lp(
                entity=entity,
                inputs=inputs,
                project=project,
                domain=domain,
                execution_name=execution_name,
                execution_name_prefix=execution_name_prefix,
                options=options,
                wait=wait,
                type_hints=type_hints,
                overwrite_cache=overwrite_cache,
                envs=envs,
                tags=tags,
                cluster_pool=cluster_pool,
            )
        if isinstance(entity, FlyteWorkflow):
            return self.execute_remote_wf(
                entity=entity,
                inputs=inputs,
                project=project,
                domain=domain,
                execution_name=execution_name,
                execution_name_prefix=execution_name_prefix,
                options=options,
                wait=wait,
                type_hints=type_hints,
                overwrite_cache=overwrite_cache,
                envs=envs,
                tags=tags,
                cluster_pool=cluster_pool,
            )
        if isinstance(entity, PythonTask):
            return self.execute_local_task(
                entity=entity,
                inputs=inputs,
                project=project,
                domain=domain,
                name=name,
                version=version,
                execution_name=execution_name,
                execution_name_prefix=execution_name_prefix,
                image_config=image_config,
                wait=wait,
                overwrite_cache=overwrite_cache,
                envs=envs,
                tags=tags,
                cluster_pool=cluster_pool,
            )
        if isinstance(entity, WorkflowBase):
            return self.execute_local_workflow(
                entity=entity,
                inputs=inputs,
                project=project,
                domain=domain,
                name=name,
                version=version,
                execution_name=execution_name,
                execution_name_prefix=execution_name_prefix,
                image_config=image_config,
                options=options,
                wait=wait,
                overwrite_cache=overwrite_cache,
                envs=envs,
                tags=tags,
                cluster_pool=cluster_pool,
            )
        if isinstance(entity, LaunchPlan):
            return self.execute_local_launch_plan(
                entity=entity,
                inputs=inputs,
                version=version,
                project=project,
                domain=domain,
                name=name,
                execution_name=execution_name,
                execution_name_prefix=execution_name_prefix,
                options=options,
                wait=wait,
                overwrite_cache=overwrite_cache,
                envs=envs,
                tags=tags,
                cluster_pool=cluster_pool,
            )
        raise NotImplementedError(f"entity type {type(entity)} not recognized for execution")

    # Flyte Remote Entities
    # ---------------------

    def execute_remote_task_lp(
        self,
        entity: typing.Union[FlyteTask, FlyteLaunchPlan],
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        type_hints: typing.Optional[typing.Dict[str, typing.Type]] = None,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """Execute a FlyteTask, or FlyteLaunchplan.

        NOTE: the name and version arguments are currently not used and only there consistency in the function signature
        """
        return self._execute(
            entity,
            inputs,
            project=project,
            domain=domain,
            execution_name=execution_name,
            execution_name_prefix=execution_name_prefix,
            wait=wait,
            options=options,
            type_hints=type_hints,
            overwrite_cache=overwrite_cache,
            envs=envs,
            tags=tags,
            cluster_pool=cluster_pool,
        )

    def execute_remote_wf(
        self,
        entity: FlyteWorkflow,
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        type_hints: typing.Optional[typing.Dict[str, typing.Type]] = None,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """Execute a FlyteWorkflow.

        NOTE: the name and version arguments are currently not used and only there consistency in the function signature
        """
        launch_plan = self.fetch_launch_plan(entity.id.project, entity.id.domain, entity.id.name, entity.id.version)
        return self.execute_remote_task_lp(
            launch_plan,
            inputs,
            project=project,
            domain=domain,
            execution_name=execution_name,
            execution_name_prefix=execution_name_prefix,
            options=options,
            wait=wait,
            type_hints=type_hints,
            overwrite_cache=overwrite_cache,
            envs=envs,
            tags=tags,
            cluster_pool=cluster_pool,
        )

    # Flytekit Entities
    # -----------------

    def execute_local_task(
        self,
        entity: PythonTask,
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        name: str = None,
        version: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        image_config: typing.Optional[ImageConfig] = None,
        wait: bool = False,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """
        Execute a @task-decorated function or TaskTemplate task.

        :param entity: local task entity.
        :param inputs: register the task, which requires compiling the task, before running it.
        :param project: The execution project, will default to the Remote's default project.
        :param domain: The execution domain, will default to the Remote's default domain.
        :param name: specific name of the task to run.
        :param version: specific version of the task to run.
        :param execution_name: If provided, will use this name for the execution.
        :param image_config: If provided, will use this image config in the pod.
        :param wait: If True, will wait for the execution to complete before returning.
        :param overwrite_cache: If True, will overwrite the cache.
        :param envs: Environment variables to set for the execution.
        :param tags: Tags to set for the execution.
        :param cluster_pool: Specify cluster pool on which newly created execution should be placed.
        :return: FlyteWorkflowExecution object.
        """
        resolved_identifiers = self._resolve_identifier_kwargs(entity, project, domain, name, version)
        resolved_identifiers_dict = asdict(resolved_identifiers)
        try:
            flyte_task: FlyteTask = self.fetch_task(**resolved_identifiers_dict)
        except FlyteEntityNotExistException:
            if isinstance(entity, PythonAutoContainerTask):
                if not image_config:
                    raise ValueError(f"PythonTask {entity.name} not already registered, but image_config missing")
            ss = SerializationSettings(
                image_config=image_config,
                project=project or self.default_project,
                domain=domain or self._default_domain,
                version=version,
            )
            flyte_task: FlyteTask = self.register_task(entity, ss)

        return self.execute(
            flyte_task,
            inputs,
            project=resolved_identifiers.project,
            domain=resolved_identifiers.domain,
            execution_name=execution_name,
            execution_name_prefix=execution_name_prefix,
            wait=wait,
            type_hints=entity.python_interface.inputs,
            overwrite_cache=overwrite_cache,
            envs=envs,
            tags=tags,
            cluster_pool=cluster_pool,
        )

    def execute_local_workflow(
        self,
        entity: WorkflowBase,
        inputs: typing.Dict[str, typing.Any],
        project: str = None,
        domain: str = None,
        name: str = None,
        version: str = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        image_config: typing.Optional[ImageConfig] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """
        Execute an @workflow decorated function.
        :param entity:
        :param inputs:
        :param project:
        :param domain:
        :param name:
        :param version:
        :param execution_name:
        :param image_config:
        :param options:
        :param wait:
        :param overwrite_cache:
        :param envs:
        :param tags:
        :param cluster_pool:
        :return:
        """
        resolved_identifiers = self._resolve_identifier_kwargs(entity, project, domain, name, version)
        resolved_identifiers_dict = asdict(resolved_identifiers)
        ss = SerializationSettings(
            image_config=image_config,
            project=resolved_identifiers.project,
            domain=resolved_identifiers.domain,
            version=resolved_identifiers.version,
        )
        try:
            # Just fetch to see if it already exists
            # todo: Add logic to check that the fetched workflow is functionally equivalent.
            self.fetch_workflow(**resolved_identifiers_dict)
        except FlyteEntityNotExistException:
            logger.info("Registering workflow because it wasn't found in Flyte Admin.")
            if not image_config:
                raise ValueError("Need image config since we are registering")
            self.register_workflow(entity, ss, version=version, options=options)

        try:
            flyte_lp = self.fetch_launch_plan(**resolved_identifiers_dict)
        except FlyteEntityNotExistException:
            logger.info("Try to register default launch plan because it wasn't found in Flyte Admin!")
            default_lp = LaunchPlan.get_default_launch_plan(self.context, entity)
            self.register_launch_plan(
                default_lp,
                project=resolved_identifiers.project,
                domain=resolved_identifiers.domain,
                version=version,
                options=options,
            )
            flyte_lp = self.fetch_launch_plan(**resolved_identifiers_dict)

        return self.execute(
            flyte_lp,
            inputs,
            project=project,
            domain=domain,
            execution_name=execution_name,
            execution_name_prefix=execution_name_prefix,
            wait=wait,
            options=options,
            type_hints=entity.python_interface.inputs,
            overwrite_cache=overwrite_cache,
            envs=envs,
            tags=tags,
            cluster_pool=cluster_pool,
        )

    def execute_local_launch_plan(
        self,
        entity: LaunchPlan,
        inputs: typing.Dict[str, typing.Any],
        version: str,
        project: typing.Optional[str] = None,
        domain: typing.Optional[str] = None,
        name: typing.Optional[str] = None,
        execution_name: typing.Optional[str] = None,
        execution_name_prefix: typing.Optional[str] = None,
        options: typing.Optional[Options] = None,
        wait: bool = False,
        overwrite_cache: typing.Optional[bool] = None,
        envs: typing.Optional[typing.Dict[str, str]] = None,
        tags: typing.Optional[typing.List[str]] = None,
        cluster_pool: typing.Optional[str] = None,
    ) -> FlyteWorkflowExecution:
        """

        :param entity: The locally defined launch plan object
        :param inputs: Inputs to be passed into the execution as a dict with Python native values.
        :param version: The version to look up/register the launch plan (if not already exists)
        :param project: The same as version, but will default to the Remote object's project
        :param domain: The same as version, but will default to the Remote object's domain
        :param name: The same as version, but will default to the entity's name
        :param execution_name: If specified, will be used as the execution name instead of randomly generating.
        :param options: Options to be passed into the execution.
        :param wait: If True, will wait for the execution to complete before returning.
        :param overwrite_cache: If True, will overwrite the cache.
        :param envs: Environment variables to be passed into the execution.
        :param tags: Tags to be passed into the execution.
        :param cluster_pool: Specify cluster pool on which newly created execution should be placed.
        :return: FlyteWorkflowExecution object
        """
        resolved_identifiers = self._resolve_identifier_kwargs(entity, project, domain, name, version)
        resolved_identifiers_dict = asdict(resolved_identifiers)
        project = resolved_identifiers.project
        domain = resolved_identifiers.domain
        try:
            flyte_launchplan: FlyteLaunchPlan = self.fetch_launch_plan(**resolved_identifiers_dict)
        except FlyteEntityNotExistException:
            flyte_launchplan: FlyteLaunchPlan = self.register_launch_plan(
                entity,
                version=version,
                project=project,
                domain=domain,
            )
        return self.execute_remote_task_lp(
            flyte_launchplan,
            inputs,
            project=project,
            domain=domain,
            execution_name=execution_name,
            execution_name_prefix=execution_name_prefix,
            options=options,
            wait=wait,
            type_hints=entity.python_interface.inputs,
            overwrite_cache=overwrite_cache,
            envs=envs,
            tags=tags,
            cluster_pool=cluster_pool,
        )

    ###################################
    # Wait for Executions to Complete #
    ###################################

    def wait(
        self,
        execution: FlyteWorkflowExecution,
        timeout: typing.Optional[timedelta] = None,
        poll_interval: typing.Optional[timedelta] = None,
        sync_nodes: bool = True,
    ) -> FlyteWorkflowExecution:
        """Wait for an execution to finish.

        :param execution: execution object to wait on
        :param timeout: maximum amount of time to wait
        :param poll_interval: sync workflow execution at this interval
        :param sync_nodes: passed along to the sync call for the workflow execution
        """
        poll_interval = poll_interval or timedelta(seconds=30)
        time_to_give_up = datetime.max if timeout is None else datetime.utcnow() + timeout

        while datetime.utcnow() < time_to_give_up:
            execution = self.sync_execution(execution, sync_nodes=sync_nodes)
            if execution.is_done:
                return execution
            time.sleep(poll_interval.total_seconds())

        raise user_exceptions.FlyteTimeout(f"Execution {self} did not complete before timeout.")

    ########################
    # Sync Execution State #
    ########################

    def sync(
        self,
        execution: FlyteWorkflowExecution,
        entity_definition: typing.Union[FlyteWorkflow, FlyteTask] = None,
        sync_nodes: bool = False,
    ) -> FlyteWorkflowExecution:
        """
        This function was previously a singledispatchmethod. We've removed that but this function remains
        so that we don't break people.

        :param execution:
        :param entity_definition:
        :param sync_nodes: By default sync will fetch data on all underlying node executions (recursively,
          so subworkflows will also get picked up). Set this to False in order to prevent that (which
          will make this call faster).
        :return: Returns the same execution object, but with additional information pulled in.
        """
        if not isinstance(execution, FlyteWorkflowExecution):
            raise ValueError(f"remote.sync should only be called on workflow executions, got {type(execution)}")
        return self.sync_execution(execution, entity_definition, sync_nodes)

    def sync_execution(
        self,
        execution: FlyteWorkflowExecution,
        entity_definition: typing.Union[FlyteWorkflow, FlyteTask] = None,
        sync_nodes: bool = False,
    ) -> FlyteWorkflowExecution:
        """
        Sync a FlyteWorkflowExecution object with its corresponding remote state.
        """
        if entity_definition is not None:
            raise ValueError("Entity definition arguments aren't supported when syncing workflow executions")

        # Update closure, and then data, because we don't want the execution to finish between when we get the data,
        # and then for the closure to have is_done to be true.
        execution._closure = self.client.get_execution(execution.id).closure
        execution_data = self.client.get_execution_data(execution.id)
        lp_id = execution.spec.launch_plan
        underlying_node_executions = []
        if sync_nodes:
            underlying_node_executions = [
                FlyteNodeExecution.promote_from_model(n) for n in iterate_node_executions(self.client, execution.id)
            ]

        # This condition is only true for single-task executions
        if execution.spec.launch_plan.resource_type == ResourceType.TASK:
            flyte_entity = self.fetch_task(lp_id.project, lp_id.domain, lp_id.name, lp_id.version)
            node_interface = flyte_entity.interface
            if sync_nodes:
                # Need to construct the mapping. There should've been returned exactly three nodes, a start,
                # an end, and a task node.
                task_node_exec = [
                    x
                    for x in filter(
                        lambda x: x.id.node_id != constants.START_NODE_ID and x.id.node_id != constants.END_NODE_ID,
                        underlying_node_executions,
                    )
                ]
                # We need to manually make a map of the nodes since there is none for single task executions
                # Assume the first one is the only one.
                node_mapping = (
                    {
                        task_node_exec[0].id.node_id: FlyteNode(
                            id=flyte_entity.id,
                            upstream_nodes=[],
                            bindings=[],
                            metadata=NodeMetadata(name=""),
                            task_node=FlyteTaskNode(flyte_entity),
                        )
                    }
                    if len(task_node_exec) >= 1
                    else {}  # This is for the case where node executions haven't appeared yet
                )
        # This is the default case, an execution of a normal workflow through a launch plan
        else:
            fetched_lp = self.fetch_launch_plan(lp_id.project, lp_id.domain, lp_id.name, lp_id.version)
            node_interface = fetched_lp.flyte_workflow.interface
            execution._flyte_workflow = fetched_lp.flyte_workflow
            node_mapping = fetched_lp.flyte_workflow._node_map

        # update node executions (if requested), and inputs/outputs
        if sync_nodes:
            node_execs = {}
            for n in underlying_node_executions:
                node_execs[n.id.node_id] = self.sync_node_execution(n, node_mapping)  # noqa
            execution._node_executions = node_execs
        return self._assign_inputs_and_outputs(execution, execution_data, node_interface)

    def sync_node_execution(
        self,
        execution: FlyteNodeExecution,
        node_mapping: typing.Dict[str, FlyteNode],
    ) -> FlyteNodeExecution:
        """
        Get data backing a node execution. These FlyteNodeExecution objects should've come from Admin with the model
        fields already populated correctly. For purposes of the remote experience, we'd like to supplement the object
        with some additional fields:
        - inputs/outputs
        - task/workflow executions, and/or underlying node executions in the case of parent nodes
        - TypedInterface (remote wrapper type)

        A node can have several different types of executions behind it. That is, the node could've run (perhaps
        multiple times because of retries):
        - A task
        - A static subworkflow
        - A dynamic subworkflow (which in turn may have run additional tasks, subwfs, and/or launch plans)
        - A launch plan

        The data model is complicated, so ascertaining which of these happened is a bit tricky. That logic is
        encapsulated in this function.
        """
        # For single task execution - the metadata spec node id is missing. In these cases, revert to regular node id
        node_id = execution.metadata.spec_node_id
        # This case supports single-task execution compiled workflows.
        if node_id and node_id not in node_mapping and execution.id.node_id in node_mapping:
            node_id = execution.id.node_id
            logger.debug(
                f"Using node execution ID {node_id} instead of spec node id "
                f"{execution.metadata.spec_node_id}, single-task execution likely."
            )
        # This case supports single-task execution compiled workflows with older versions of admin/propeller
        if not node_id:
            node_id = execution.id.node_id
            logger.debug(f"No metadata spec_node_id found, using {node_id}")

        # First see if it's a dummy node, if it is, we just skip it.
        if constants.START_NODE_ID in node_id or constants.END_NODE_ID in node_id:
            return execution

        # Look for the Node object in the mapping supplied
        if node_id in node_mapping:
            execution._node = node_mapping[node_id]
        else:
            raise Exception(f"Missing node from mapping: {node_id}")

        # Get the node execution data
        node_execution_get_data_response = self.client.get_node_execution_data(execution.id)

        # Calling a launch plan directly case
        # If a node ran a launch plan directly (i.e. not through a dynamic task or anything) then
        # the closure should have a workflow_node_metadata populated with the launched execution id.
        # The parent node flag should not be populated here
        # This is the simplest case
        if not execution.metadata.is_parent_node and execution.closure.workflow_node_metadata:
            launched_exec_id = execution.closure.workflow_node_metadata.execution_id
            # This is a recursive call, basically going through the same process that brought us here in the first
            # place, but on the launched execution.
            launched_exec = self.fetch_execution(
                project=launched_exec_id.project, domain=launched_exec_id.domain, name=launched_exec_id.name
            )
            self.sync_execution(launched_exec)
            if launched_exec.is_done:
                # The synced underlying execution should've had these populated.
                execution._inputs = launched_exec.inputs
                execution._outputs = launched_exec.outputs
            execution._workflow_executions.append(launched_exec)
            execution._interface = launched_exec._flyte_workflow.interface
            return execution

        # If a node ran a static subworkflow or a dynamic subworkflow then the parent flag will be set.
        if execution.metadata.is_parent_node:
            # We'll need to query child node executions regardless since this is a parent node
            child_node_executions = iterate_node_executions(
                self.client,
                workflow_execution_identifier=execution.id.execution_id,
                unique_parent_id=execution.id.node_id,
            )
            child_node_executions = [x for x in child_node_executions]

            # If this was a dynamic task, then there should be a CompiledWorkflowClosure inside the
            # NodeExecutionGetDataResponse
            if node_execution_get_data_response.dynamic_workflow is not None:
                compiled_wf = node_execution_get_data_response.dynamic_workflow.compiled_workflow
                node_launch_plans = {}
                # TODO: Inspect branch nodes for launch plans
                for node in FlyteWorkflow.get_non_system_nodes(compiled_wf.primary.template.nodes):
                    if (
                        node.workflow_node is not None
                        and node.workflow_node.launchplan_ref is not None
                        and node.workflow_node.launchplan_ref not in node_launch_plans
                    ):
                        node_launch_plans[node.workflow_node.launchplan_ref] = self.client.get_launch_plan(
                            node.workflow_node.launchplan_ref
                        ).spec

                dynamic_flyte_wf = FlyteWorkflow.promote_from_closure(compiled_wf, node_launch_plans)
                execution._underlying_node_executions = [
                    self.sync_node_execution(FlyteNodeExecution.promote_from_model(cne), dynamic_flyte_wf._node_map)
                    for cne in child_node_executions
                ]
                execution._task_executions = [
                    node_exes.task_executions for node_exes in execution.subworkflow_node_executions.values()
                ]

                execution._interface = dynamic_flyte_wf.interface

            # Handle the case where it's a static subworkflow
            elif isinstance(execution._node.flyte_entity, FlyteWorkflow):
                sub_flyte_workflow = execution._node.flyte_entity
                sub_node_mapping = {n.id: n for n in sub_flyte_workflow.flyte_nodes}
                execution._underlying_node_executions = [
                    self.sync_node_execution(FlyteNodeExecution.promote_from_model(cne), sub_node_mapping)
                    for cne in child_node_executions
                ]
                execution._interface = sub_flyte_workflow.interface

            # Handle the case where it's a branch node
            elif execution._node.branch_node is not None:
                logger.info(
                    "Skipping branch node execution for now - branch nodes will "
                    "not have inputs and outputs filled in"
                )
                return execution
            else:
                logger.error(f"NE {execution} undeterminable, {type(execution._node)}, {execution._node}")
                raise Exception(f"Node execution undeterminable, entity has type {type(execution._node)}")

        # Handle the case for gate nodes
        elif execution._node.gate_node is not None:
            logger.info("Skipping gate node execution for now - gate nodes don't have inputs and outputs filled in")
            return execution

        # This is the plain ol' task execution case
        else:
            execution._task_executions = [
                self.sync_task_execution(
                    FlyteTaskExecution.promote_from_model(t), node_mapping[node_id].task_node.flyte_task
                )
                for t in iterate_task_executions(self.client, execution.id)
            ]
            execution._interface = execution._node.flyte_entity.interface

        self._assign_inputs_and_outputs(
            execution,
            node_execution_get_data_response,
            execution.interface,
        )

        return execution

    def sync_task_execution(
        self, execution: FlyteTaskExecution, entity_definition: typing.Optional[FlyteTask] = None
    ) -> FlyteTaskExecution:
        """Sync a FlyteTaskExecution object with its corresponding remote state."""
        execution._closure = self.client.get_task_execution(execution.id).closure
        execution_data = self.client.get_task_execution_data(execution.id)
        task_id = execution.id.task_id
        if entity_definition is None:
            entity_definition = self.fetch_task(task_id.project, task_id.domain, task_id.name, task_id.version)
        return self._assign_inputs_and_outputs(execution, execution_data, entity_definition.interface)

    #############################
    # Terminate Execution State #
    #############################

    def terminate(self, execution: FlyteWorkflowExecution, cause: str):
        """Terminate a workflow execution.

        :param execution: workflow execution to terminate
        :param cause: reason for termination
        """
        self.client.terminate_execution(execution.id, cause)

    ##################
    # Helper Methods #
    ##################

    def _assign_inputs_and_outputs(
        self,
        execution: typing.Union[FlyteWorkflowExecution, FlyteNodeExecution, FlyteTaskExecution],
        execution_data,
        interface: TypedInterface,
    ):
        """Helper for assigning synced inputs and outputs to an execution object."""
        input_literal_map = self._get_input_literal_map(execution_data)
        execution._inputs = LiteralsResolver(input_literal_map.literals, interface.inputs, self.context)

        if execution.is_done and not execution.error:
            output_literal_map = self._get_output_literal_map(execution_data)
            execution._outputs = LiteralsResolver(output_literal_map.literals, interface.outputs, self.context)
        return execution

    def _get_input_literal_map(self, execution_data: ExecutionDataResponse) -> literal_models.LiteralMap:
        # Inputs are returned inline unless they are too big, in which case a url blob pointing to them is returned.
        if bool(execution_data.full_inputs.literals):
            return execution_data.full_inputs
        elif execution_data.inputs.bytes > 0:
            with self.remote_context() as ctx:
                tmp_name = os.path.join(ctx.file_access.local_sandbox_dir, "inputs.pb")
                ctx.file_access.get_data(execution_data.inputs.url, tmp_name)
                return literal_models.LiteralMap.from_flyte_idl(
                    utils.load_proto_from_file(literals_pb2.LiteralMap, tmp_name)
                )
        return literal_models.LiteralMap({})

    def _get_output_literal_map(self, execution_data: ExecutionDataResponse) -> literal_models.LiteralMap:
        # Outputs are returned inline unless they are too big, in which case a url blob pointing to them is returned.
        if bool(execution_data.full_outputs.literals):
            return execution_data.full_outputs
        elif execution_data.outputs.bytes > 0:
            with self.remote_context() as ctx:
                tmp_name = os.path.join(ctx.file_access.local_sandbox_dir, "outputs.pb")
                ctx.file_access.get_data(execution_data.outputs.url, tmp_name)
                return literal_models.LiteralMap.from_flyte_idl(
                    utils.load_proto_from_file(literals_pb2.LiteralMap, tmp_name)
                )
        return literal_models.LiteralMap({})

    def generate_console_http_domain(self) -> str:
        """
        This should generate the domain where console is hosted.

        :return:
        """
        # If the console endpoint is explicitly set, return it, else derive it from the admin config
        if self.config.platform.console_endpoint:
            return self.config.platform.console_endpoint
        protocol = "http" if self.config.platform.insecure else "https"
        endpoint = self.config.platform.endpoint
        # N.B.: this assumes that in case we have an identical configuration as the sandbox default config we are running single binary. The intent here is
        # to ensure that the urls produced in the getting started guide point to the correct place.
        if self.config.platform == Config.for_sandbox().platform:
            endpoint = "localhost:30080"
        return protocol + f"://{endpoint}"

    def generate_console_url(
        self,
        entity: typing.Union[
            FlyteWorkflowExecution, FlyteNodeExecution, FlyteTaskExecution, FlyteWorkflow, FlyteTask, FlyteLaunchPlan
        ],
    ):
        """
        Generate a Flyteconsole URL for the given Flyte remote endpoint.
        This will automatically determine if this is an execution or an entity and change the type automatically
        """
        if isinstance(entity, (FlyteWorkflowExecution, FlyteNodeExecution, FlyteTaskExecution)):
            return f"{self.generate_console_http_domain()}/console/projects/{entity.id.project}/domains/{entity.id.domain}/executions/{entity.id.name}"  # noqa

        if not isinstance(entity, (FlyteWorkflow, FlyteTask, FlyteLaunchPlan)):
            raise ValueError(f"Only remote entities can be looked at in the console, got type {type(entity)}")
        rt = "workflow"
        if entity.id.resource_type == ResourceType.TASK:
            rt = "task"
        elif entity.id.resource_type == ResourceType.LAUNCH_PLAN:
            rt = "launch_plan"
        return f"{self.generate_console_http_domain()}/console/projects/{entity.id.project}/domains/{entity.id.domain}/{rt}/{entity.name}/version/{entity.id.version}"  # noqa

    def launch_backfill(
        self,
        project: str,
        domain: str,
        from_date: datetime,
        to_date: datetime,
        launchplan: str,
        launchplan_version: str = None,
        execution_name: str = None,
        version: str = None,
        dry_run: bool = False,
        execute: bool = True,
        parallel: bool = False,
        failure_policy: typing.Optional[WorkflowFailurePolicy] = None,
    ) -> typing.Optional[FlyteWorkflowExecution, FlyteWorkflow, WorkflowBase]:
        """
        Creates and launches a backfill workflow for the given launchplan. If launchplan version is not specified,
        then the latest launchplan is retrieved.
        The from_date is exclusive and end_date is inclusive and backfill run for all instances in between. ::
            -> (start_date - exclusive, end_date inclusive)

        If dry_run is specified, the workflow is created and returned.
        If execute==False is specified then the workflow is created and registered.
        In the last case, the workflow is created, registered and executed.

        The `parallel` flag can be used to generate a workflow where all launchplans can be run in parallel. Default
        is that execute backfill is run sequentially

        :param project: str project name
        :param domain: str domain name
        :param from_date: datetime generate a backfill starting at this datetime (exclusive)
        :param to_date:  datetime generate a backfill ending at this datetime (inclusive)
        :param launchplan: str launchplan name in the flyte backend
        :param launchplan_version: str (optional) version for the launchplan. If not specified the most recent will be retrieved
        :param execution_name: str (optional) the generated execution will be named so. this can help in ensuring idempotency
        :param version: str (optional) version to be used for the newly created workflow.
        :param dry_run: bool do not register or execute the workflow
        :param execute: bool Register and execute the wwkflow.
        :param parallel: if the backfill should be run in parallel. False (default) will run each bacfill sequentially.
        :param failure_policy: WorkflowFailurePolicy (optional) to be used for the newly created workflow. This can
                control failure behavior - whether to continue on failure or stop immediately on failure
        :return: In case of dry-run, return WorkflowBase, else if no_execute return FlyteWorkflow else in the default
            case return a FlyteWorkflowExecution
        """
        lp = self.fetch_launch_plan(project=project, domain=domain, name=launchplan, version=launchplan_version)
        wf, start, end = create_backfill_workflow(
            start_date=from_date, end_date=to_date, for_lp=lp, parallel=parallel, failure_policy=failure_policy
        )
        if dry_run:
            logger.warning("Dry Run enabled. Workflow will not be registered and or executed.")
            return wf

        unique_fingerprint = f"{start}-{end}-{launchplan}-{launchplan_version}"
        h = hashlib.md5()
        h.update(unique_fingerprint.encode("utf-8"))
        unique_fingerprint_encoded = base64.urlsafe_b64encode(h.digest()).decode("ascii")
        if not version:
            version = unique_fingerprint_encoded
        ss = SerializationSettings(
            image_config=ImageConfig.auto(),
            project=project,
            domain=domain,
            version=version,
        )
        remote_wf = self.register_workflow(wf, serialization_settings=ss)

        if not execute:
            return remote_wf

        return self.execute(remote_wf, inputs={}, project=project, domain=domain, execution_name=execution_name)

    @staticmethod
    def get_extra_headers_for_protocol(native_url):
        if native_url.startswith("abfs://"):
            return {"x-ms-blob-type": "BlockBlob"}
        return {}

    def activate_launchplan(self, ident: Identifier):
        """
        Given a launchplan, activate it, all previous versions are deactivated.
        """
        self.client.update_launch_plan(id=ident, state=LaunchPlanState.ACTIVE)

    def download(
        self, data: typing.Union[LiteralsResolver, Literal, LiteralMap], download_to: str, recursive: bool = True
    ):
        """
        Download the data to the specified location. If the data is a LiteralsResolver, LiteralMap and if recursive is
        specified, then all file like objects will be recursively downloaded (e.g. FlyteFile/Dir (blob),
         StructuredDataset etc).

        Note: That it will use your sessions credentials to access the remote location. For sandbox, this should be
        automatically configured, assuming you are running sandbox locally. For other environments, you will need to
        configure your credentials appropriately.

        :param data: data to be downloaded
        :param download_to: location to download to (str) that should be a valid path
        :param recursive: if the data is a LiteralsResolver or LiteralMap, then this flag will recursively download
        """
        download_to = pathlib.Path(download_to)
        if isinstance(data, Literal):
            download_literal(self.file_access, "data", data, download_to)
        else:
            if not recursive:
                raise click.UsageError("Please specify --recursive to download all variables in a literal map.")
            if isinstance(data, LiteralsResolver):
                lm = data.literals
            else:
                lm = data
            for var, literal in lm.items():
                download_literal(self.file_access, var, literal, download_to)
