import asyncio
import inspect
import signal
from contextlib import asynccontextmanager
from datetime import datetime, timedelta
from functools import partial, wraps
from typing import List, Optional

from flytekit import Deck, Secret, current_context
from flytekit.configuration import DataConfig, PlatformConfig, S3Config
from flytekit.core.base_task import PythonTask
from flytekit.core.context_manager import ExecutionState, FlyteContext, FlyteContextManager
from flytekit.core.python_function_task import PythonFunctionTask
from flytekit.core.task import task
from flytekit.core.workflow import WorkflowBase
from flytekit.loggers import logger
from flytekit.models.core.execution import WorkflowExecutionPhase
from flytekit.remote import FlyteRemote

FLYTE_SANDBOX_INTERNAL_ENDPOINT = "flyte-sandbox-grpc.flyte:8089"
FLYTE_SANDBOX_MINIO_ENDPOINT = "http://flyte-sandbox-minio.flyte:9000"

NODE_HTML_TEMPLATE = """
<style>
    #flyte-frame-container > div.active {{font-family: Open sans;}}
</style>

<style>
    #flyte-frame-container div.input-output {{
        font-family: monospace;
        background: #f0f0f0;
        padding: 10px 15px;
        margin: 15px 0;
    }}
</style>

<h3>{entity_type}: {entity_name}</h3>

<p>
    <strong>Execution:</strong>
    <a target="_blank" href="{url}">{execution_name}</a>
</p>

<details>
<summary>Inputs</summary>
<div class="input-output">{inputs}</div>
</details>

<details>
<summary>Outputs</summary>
<div class="input-output">{outputs}</div>
</details>

<hr>
"""


class EagerException(Exception):
    """Raised when a node in an eager workflow encounters an error.

    This exception should be used in an :py:func:`@eager <flytekit.experimental.eager>` workflow function to
    catch exceptions that are raised by tasks or subworkflows.

    .. code-block:: python

        from flytekit import task
        from flytekit.experimental import eager, EagerException

        @task
        def add_one(x: int) -> int:
            if x < 0:
                raise ValueError("x must be positive")
            return x + 1

        @task
        def double(x: int) -> int:
            return x * 2

        @eager
        async def eager_workflow(x: int) -> int:
            try:
                out = await add_one(x=x)
            except EagerException:
                # The ValueError error is caught
                # and raised as an EagerException
                raise
            return await double(x=out)
    """


class AsyncEntity:
    """A wrapper around a Flyte entity (task, workflow, launch plan) that allows it to be executed asynchronously."""

    def __init__(
        self,
        entity,
        remote: Optional[FlyteRemote],
        ctx: FlyteContext,
        async_stack: "AsyncStack",
        timeout: Optional[timedelta] = None,
        poll_interval: Optional[timedelta] = None,
        local_entrypoint: bool = False,
    ):
        self.entity = entity
        self.ctx = ctx
        self.async_stack = async_stack
        self.execution_state = self.ctx.execution_state.mode
        self.remote = remote
        self.local_entrypoint = local_entrypoint
        if self.remote is not None:
            logger.debug(f"Using remote config: {self.remote.config}")
        else:
            logger.debug("Not using remote, executing locally")
        self._timeout = timeout
        self._poll_interval = poll_interval
        self._execution = None

    async def __call__(self, **kwargs):
        logger.debug(f"Calling {self.entity}: {self.entity.name}")

        # ensure async context is provided
        if "async_ctx" in kwargs:
            kwargs.pop("async_ctx")

        if getattr(self.entity, "execution_mode", None) == PythonFunctionTask.ExecutionBehavior.DYNAMIC:
            raise EagerException(
                "Eager workflows currently do not work with dynamic workflows. "
                "If you need to use a subworkflow, use a static @workflow or nested @eager workflow."
            )

        if not self.local_entrypoint and self.ctx.execution_state.is_local_execution():
            # If running as a local workflow execution, just execute the python function
            try:
                if isinstance(self.entity, WorkflowBase):
                    out = self.entity._workflow_function(**kwargs)
                    if inspect.iscoroutine(out):
                        # need to handle invocation of AsyncEntity tasks within the workflow
                        out = await out
                    return out
                elif isinstance(self.entity, PythonTask):
                    # invoke the task-decorated entity
                    out = self.entity(**kwargs)
                    if inspect.iscoroutine(out):
                        out = await out
                    return out
                else:
                    raise ValueError(f"Entity type {type(self.entity)} not supported for local execution")
            except Exception as exc:
                raise EagerException(
                    f"Error executing {type(self.entity)} {self.entity.name} with {type(exc)}: {exc}"
                ) from exc

        # this is a hack to handle the case when the task.name doesn't contain the fully
        # qualified module name
        entity_name = (
            f"{self.entity._instantiated_in}.{self.entity.name}"
            if self.entity._instantiated_in not in self.entity.name
            else self.entity.name
        )

        if isinstance(self.entity, WorkflowBase):
            remote_entity = self.remote.fetch_workflow(name=entity_name)
        elif isinstance(self.entity, PythonTask):
            remote_entity = self.remote.fetch_task(name=entity_name)
        else:
            raise ValueError(f"Entity type {type(self.entity)} not supported for local execution")

        execution = self.remote.execute(remote_entity, inputs=kwargs, type_hints=self.entity.python_interface.inputs)
        self._execution = execution

        url = self.remote.generate_console_url(execution)
        msg = f"Running flyte {type(self.entity)} {entity_name} on remote cluster: {url}"
        if self.local_entrypoint:
            logger.info(msg)
        else:
            logger.debug(msg)

        node = AsyncNode(self, entity_name, execution, url)
        self.async_stack.set_node(node)

        poll_interval = self._poll_interval or timedelta(seconds=30)
        time_to_give_up = datetime.max if self._timeout is None else datetime.utcnow() + self._timeout

        while datetime.utcnow() < time_to_give_up:
            execution = self.remote.sync(execution)
            if execution.closure.phase in {WorkflowExecutionPhase.FAILED}:
                raise EagerException(f"Error executing {self.entity.name} with error: {execution.closure.error}")
            elif execution.is_done:
                break
            await asyncio.sleep(poll_interval.total_seconds())

        outputs = {}
        for key, type_ in self.entity.python_interface.outputs.items():
            outputs[key] = execution.outputs.get(key, as_type=type_)

        if len(outputs) == 1:
            out, *_ = outputs.values()
            return out
        return outputs

    async def terminate(self):
        execution = self.remote.sync(self._execution)
        logger.debug(f"Cleaning up execution: {execution}")
        if not execution.is_done:
            self.remote.terminate(
                execution,
                f"Execution terminated by eager workflow execution {self.async_stack.parent_execution_id}.",
            )

            poll_interval = self._poll_interval or timedelta(seconds=6)
            time_to_give_up = datetime.max if self._timeout is None else datetime.utcnow() + self._timeout

            while datetime.utcnow() < time_to_give_up:
                execution = self.remote.sync(execution)
                if execution.is_done:
                    break
                await asyncio.sleep(poll_interval.total_seconds())

        return True


class AsyncNode:
    """A node in the async callstack."""

    def __init__(self, async_entity, entity_name, execution=None, url=None):
        self.entity_name = entity_name
        self.async_entity = async_entity
        self.execution = execution
        self._url = url

    @property
    def url(self) -> str:
        # make sure that internal flyte sandbox endpoint is replaced with localhost endpoint when rendering the urls
        # for flyte decks
        endpoint_root = FLYTE_SANDBOX_INTERNAL_ENDPOINT.replace("http://", "")
        if endpoint_root in self._url:
            return self._url.replace(endpoint_root, "localhost:30080")
        return self._url

    @property
    def entity_type(self) -> str:
        if (
            isinstance(self.async_entity.entity, PythonTask)
            and getattr(self.async_entity.entity, "execution_mode", None) == PythonFunctionTask.ExecutionBehavior.EAGER
        ):
            return "Eager Workflow"
        elif isinstance(self.async_entity.entity, PythonTask):
            return "Task"
        elif isinstance(self.async_entity.entity, WorkflowBase):
            return "Workflow"
        return str(type(self.async_entity.entity))

    def __repr__(self):
        ex_id = self.execution.id
        execution_id = None if self.execution is None else f"{ex_id.project}:{ex_id.domain}:{ex_id.name}"
        return (
            "<async_node | "
            f"entity_type: {self.entity_type} | "
            f"entity: {self.entity_name} | "
            f"execution: {execution_id}"
        )


class AsyncStack:
    """A stack of async nodes that are executed in chronological order."""

    def __init__(self, parent_task_id: Optional[str], parent_execution_id: Optional[str]):
        self.parent_task_id = parent_task_id
        self.parent_execution_id = parent_execution_id
        self._call_stack: List[AsyncNode] = []

    def __repr__(self):
        return f"<parent_task_id: '{self.parent_task_id}' call_stack: {self._call_stack}>"

    @property
    def call_stack(self) -> List[AsyncNode]:
        return self._call_stack

    def set_node(self, node: AsyncNode):
        self._call_stack.append(node)


async def render_deck(async_stack):
    """Render the callstack as a deck presentation to be shown after eager workflow execution."""

    def get_io(dict_like):
        try:
            return {k: dict_like.get(k) for k in dict_like}
        except Exception:
            return dict_like

    output = "<h2>Nodes</h2><hr>"
    for node in async_stack.call_stack:
        node_inputs = get_io(node.execution.inputs)
        if node.execution.closure.phase in {WorkflowExecutionPhase.FAILED}:
            node_outputs = None
        else:
            node_outputs = get_io(node.execution.outputs)

        output = f"{output}\n" + NODE_HTML_TEMPLATE.format(
            entity_type=node.entity_type,
            entity_name=node.entity_name,
            execution_name=node.execution.id.name,
            url=node.url,
            inputs=node_inputs,
            outputs=node_outputs,
        )

    Deck("eager workflow", output)


@asynccontextmanager
async def eager_context(
    fn,
    remote: Optional[FlyteRemote],
    ctx: FlyteContext,
    async_stack: AsyncStack,
    timeout: Optional[timedelta] = None,
    poll_interval: Optional[timedelta] = None,
    local_entrypoint: bool = False,
):
    """This context manager overrides all tasks in the global namespace with async versions."""

    _original_cache = {}

    # override tasks with async version
    for k, v in fn.__globals__.items():
        if isinstance(v, (PythonTask, WorkflowBase)):
            _original_cache[k] = v
            fn.__globals__[k] = AsyncEntity(v, remote, ctx, async_stack, timeout, poll_interval, local_entrypoint)

    try:
        yield
    finally:
        # restore old tasks
        for k, v in _original_cache.items():
            fn.__globals__[k] = v


async def node_cleanup_async(sig, loop, async_stack: AsyncStack):
    """Clean up subtasks when eager workflow parent is done.

    This applies either if the eager workflow completes successfully, fails, or is cancelled by the user.
    """
    logger.debug(f"Cleaning up async nodes on signal: {sig}")
    terminations = []
    for node in async_stack.call_stack:
        terminations.append(node.async_entity.terminate())
    results = await asyncio.gather(*terminations)
    logger.debug(f"Successfully terminated subtasks {results}")


def node_cleanup(sig, frame, loop, async_stack: AsyncStack):
    """Clean up subtasks when eager workflow parent is done.

    This applies either if the eager workflow completes successfully, fails, or is cancelled by the user.
    """
    logger.debug(f"Cleaning up async nodes on signal: {sig}")
    terminations = []
    for node in async_stack.call_stack:
        terminations.append(node.async_entity.terminate())
    results = asyncio.gather(*terminations)
    results = asyncio.run(results)
    logger.debug(f"Successfully terminated subtasks {results}")
    loop.close()


def eager(
    _fn=None,
    *,
    remote: Optional[FlyteRemote] = None,
    client_secret_group: Optional[str] = None,
    client_secret_key: Optional[str] = None,
    timeout: Optional[timedelta] = None,
    poll_interval: Optional[timedelta] = None,
    local_entrypoint: bool = False,
    **kwargs,
):
    """Eager workflow decorator.

    :param remote: A :py:class:`~flytekit.remote.FlyteRemote` object to use for executing Flyte entities.
    :param client_secret_group: The client secret group to use for this workflow.
    :param client_secret_key: The client secret key to use for this workflow.
    :param timeout: The timeout duration specifying how long to wait for a task/workflow execution within the eager
        workflow to complete or terminate. By default, the eager workflow will wait indefinitely until complete.
    :param poll_interval: The poll interval for checking if a task/workflow execution within the eager workflow has
        finished. If not specified, the default poll interval is 6 seconds.
    :param local_entrypoint: If True, the eager workflow will can be executed locally but use the provided
        :py:func:`~flytekit.remote.FlyteRemote` object to create task/workflow executions. This is useful for local
        testing against a remote Flyte cluster.
    :param kwargs: keyword-arguments forwarded to :py:func:`~flytekit.task`.

    This type of workflow will execute all flyte entities within it eagerly, meaning that all python constructs can be
    used inside of an ``@eager``-decorated function. This is because eager workflows use a
    :py:class:`~flytekit.remote.remote.FlyteRemote` object to kick off executions when a flyte entity needs to produce a
    value.

    For example:

    .. code-block:: python

        from flytekit import task
        from flytekit.experimental import eager

        @task
        def add_one(x: int) -> int:
            return x + 1

        @task
        def double(x: int) -> int:
            return x * 2

        @eager
        async def eager_workflow(x: int) -> int:
            out = await add_one(x=x)
            return await double(x=out)

        # run locally with asyncio
        if __name__ == "__main__":
            import asyncio

            result = asyncio.run(eager_workflow(x=1))
            print(f"Result: {result}")  # "Result: 4"

    Unlike :py:func:`dynamic workflows <flytekit.dynamic>`, eager workflows are not compiled into a workflow spec, but
    uses python's `async <https://docs.python.org/3/library/asyncio.html>`__ capabilities to execute flyte entities.

    .. note::

       Eager workflows only support `@task`, `@workflow`, and `@eager` entities. Dynamic workflows and launchplans are
       currently not supported.

    Note that for the ``@eager`` function is an ``async`` function. Under the hood, tasks and workflows called inside
    an ``@eager`` workflow are executed asynchronously. This means that task and workflow calls will return an awaitable,
    which need to be awaited.

    .. important::

       A ``client_secret_group`` and ``client_secret_key`` is needed for authenticating via
       :py:class:`~flytekit.remote.remote.FlyteRemote` using the ``client_credentials`` authentication, which is
       configured via :py:class:`~flytekit.configuration.PlatformConfig`.

       .. code-block:: python

            from flytekit.remote import FlyteRemote
            from flytekit.configuration import Config

            @eager(
                remote=FlyteRemote(config=Config.auto(config_file="config.yaml")),
                client_secret_group="my_client_secret_group",
                client_secret_key="my_client_secret_key",
            )
            async def eager_workflow(x: int) -> int:
                out = await add_one(x)
                return await double(one)

       Where ``config.yaml`` contains is a flytectl-compatible config file.
       For more details, see `here <https://docs.flyte.org/projects/flytectl/en/latest/#configuration>`__.

       When using a sandbox cluster started with ``flytectl demo start``, however, the ``client_secret_group``
       and ``client_secret_key`` are not needed, :

       .. code-block:: python

            @eager(remote=FlyteRemote(config=Config.for_sandbox()))
            async def eager_workflow(x: int) -> int:
                ...

    .. important::

       When using ``local_entrypoint=True`` you also need to specify the ``remote`` argument. In this case, the eager
       workflow runtime will be local, but all task/subworkflow invocations will occur on the specified Flyte cluster.
       This argument is primarily used for testing and debugging eager workflow logic locally.

    """

    if _fn is None:
        return partial(
            eager,
            remote=remote,
            client_secret_group=client_secret_group,
            client_secret_key=client_secret_key,
            local_entrypoint=local_entrypoint,
            **kwargs,
        )

    if local_entrypoint and remote is None:
        raise ValueError("Must specify remote argument if local_entrypoint is True")

    @wraps(_fn)
    async def wrapper(*args, **kws):
        # grab the "async_ctx" argument injected by PythonFunctionTask.execute
        logger.debug("Starting")
        _remote = remote

        # locally executed nested eager workflows won't have async_ctx injected into the **kws input
        ctx = kws.pop("async_ctx", None)
        task_id, execution_id = None, None
        if ctx:
            exec_params = ctx.user_space_params
            task_id = exec_params.task_id
            execution_id = exec_params.execution_id

        async_stack = AsyncStack(task_id, execution_id)
        _remote = _prepare_remote(_remote, ctx, client_secret_group, client_secret_key, local_entrypoint)

        # make sure sub-nodes as cleaned up on termination signal
        loop = asyncio.get_event_loop()
        node_cleanup_partial = partial(node_cleanup_async, async_stack=async_stack)
        cleanup_fn = partial(asyncio.ensure_future, node_cleanup_partial(signal.SIGTERM, loop))
        signal.signal(signal.SIGTERM, partial(node_cleanup, loop=loop, async_stack=async_stack))

        async with eager_context(_fn, _remote, ctx, async_stack, timeout, poll_interval, local_entrypoint):
            try:
                if _remote is not None:
                    with _remote.remote_context():
                        out = await _fn(*args, **kws)
                else:
                    out = await _fn(*args, **kws)
                # need to await for _fn to complete, then invoke the deck
                await render_deck(async_stack)
                return out
            finally:
                # in case the cleanup function hasn't been called yet, call it at the end of the eager workflow
                await cleanup_fn()

    secret_requests = kwargs.pop("secret_requests", None) or []
    if client_secret_group is not None and client_secret_key is not None:
        secret_requests.append(Secret(group=client_secret_group, key=client_secret_key))

    return task(
        wrapper,
        secret_requests=secret_requests,
        enable_deck=True,
        execution_mode=PythonFunctionTask.ExecutionBehavior.EAGER,
        **kwargs,
    )


def _prepare_remote(
    remote: Optional[FlyteRemote],
    ctx: FlyteContext,
    client_secret_group: Optional[str] = None,
    client_secret_key: Optional[str] = None,
    local_entrypoint: bool = False,
) -> Optional[FlyteRemote]:
    """Prepare FlyteRemote object for accessing Flyte cluster in a task running on the same cluster."""

    is_local_execution_mode = ctx.execution_state.mode in {
        ExecutionState.Mode.LOCAL_TASK_EXECUTION,
        ExecutionState.Mode.LOCAL_WORKFLOW_EXECUTION,
    }

    if remote is not None and local_entrypoint and is_local_execution_mode:
        # when running eager workflows as a local entrypoint, we don't have to modify the remote object
        # because we can assume that the user is running this from their local machine and can do browser-based
        # authentication.
        logger.info("Running eager workflow as local entrypoint")
        return remote

    if remote is None or is_local_execution_mode:
        # if running the "eager workflow" (which is actually task) locally, run the task as a function,
        # which doesn't need a remote object
        return None

    # Handle the case where this the task is running in a Flyte cluster and needs to access the cluster itself
    # via FlyteRemote.
    if remote.config.platform.endpoint.startswith("localhost"):
        # replace sandbox endpoints with internal dns, since localhost won't exist within the Flyte cluster
        return _internal_demo_remote(remote)
    return _internal_remote(remote, client_secret_group, client_secret_key)


def _internal_demo_remote(remote: FlyteRemote) -> FlyteRemote:
    """Derives a FlyteRemote object from a sandbox yaml configuration, modifying parts to make it work internally."""
    # replace sandbox endpoints with internal dns, since localhost won't exist within the Flyte cluster
    return FlyteRemote(
        config=remote.config.with_params(
            platform=PlatformConfig(
                endpoint=FLYTE_SANDBOX_INTERNAL_ENDPOINT,
                insecure=True,
                auth_mode="Pkce",
                client_id=remote.config.platform.client_id,
            ),
            data_config=DataConfig(
                s3=S3Config(
                    endpoint=FLYTE_SANDBOX_MINIO_ENDPOINT,
                    access_key_id=remote.config.data_config.s3.access_key_id,
                    secret_access_key=remote.config.data_config.s3.secret_access_key,
                ),
            ),
        ),
        default_domain=remote.default_domain,
        default_project=remote.default_project,
    )


def _internal_remote(
    remote: FlyteRemote,
    client_secret_group: str,
    client_secret_key: str,
) -> FlyteRemote:
    """Derives a FlyteRemote object from a yaml configuration file, modifying parts to make it work internally."""
    assert client_secret_group is not None, "secret_group must be defined when using a remote cluster"
    assert client_secret_key is not None, "secret_key must be defined a remote cluster"
    secrets_manager = current_context().secrets
    client_secret = secrets_manager.get(client_secret_group, client_secret_key)
    # get the raw output prefix from the context that's set from the pyflyte-execute entrypoint
    # (see flytekit/bin/entrypoint.py)
    ctx = FlyteContextManager.current_context()
    return FlyteRemote(
        config=remote.config.with_params(
            platform=PlatformConfig(
                endpoint=remote.config.platform.endpoint,
                insecure=remote.config.platform.insecure,
                auth_mode="client_credentials",
                client_id=remote.config.platform.client_id,
                client_credentials_secret=remote.config.platform.client_credentials_secret or client_secret,
            ),
        ),
        default_domain=remote.default_domain,
        default_project=remote.default_project,
        data_upload_location=ctx.file_access.raw_output_prefix,
    )
