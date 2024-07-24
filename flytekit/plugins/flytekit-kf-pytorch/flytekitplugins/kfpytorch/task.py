"""
This Plugin adds the capability of running distributed pytorch training to Flyte using backend plugins, natively on
Kubernetes. It leverages `Pytorch Job <https://github.com/kubeflow/pytorch-operator>`_ Plugin from kubeflow.
"""
import os
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, List, NamedTuple, Optional, Union

from flyteidl.plugins.kubeflow import common_pb2 as kubeflow_common
from flyteidl.plugins.kubeflow import pytorch_pb2 as pytorch_task
from google.protobuf.json_format import MessageToDict

import flytekit
from flytekit import PythonFunctionTask, Resources, lazy_module
from flytekit.configuration import SerializationSettings
from flytekit.core.resources import convert_resources_to_resource_model
from flytekit.exceptions.user import FlyteRecoverableException
from flytekit.extend import IgnoreOutputs, TaskPlugins
from flytekit.loggers import logger

from .error_handling import create_recoverable_error_file, is_recoverable_worker_error

cloudpickle = lazy_module("cloudpickle")

TORCH_IMPORT_ERROR_MESSAGE = "PyTorch is not installed. Please install `flytekitplugins-kfpytorch['elastic']`."


@dataclass
class RestartPolicy(Enum):
    """
    RestartPolicy describes how the replicas should be restarted
    """

    ALWAYS = kubeflow_common.RESTART_POLICY_ALWAYS
    FAILURE = kubeflow_common.RESTART_POLICY_ON_FAILURE
    NEVER = kubeflow_common.RESTART_POLICY_NEVER


@dataclass
class CleanPodPolicy(Enum):
    """
    CleanPodPolicy describes how to deal with pods when the job is finished.
    """

    NONE = kubeflow_common.CLEANPOD_POLICY_NONE
    ALL = kubeflow_common.CLEANPOD_POLICY_ALL
    RUNNING = kubeflow_common.CLEANPOD_POLICY_RUNNING


@dataclass
class RunPolicy:
    """
    RunPolicy describes some policy to apply to the execution of a kubeflow job.
    Args:
        clean_pod_policy (int): Defines the policy for cleaning up pods after the PyTorchJob completes. Default to None.
        ttl_seconds_after_finished (int): Defines the TTL for cleaning up finished PyTorchJobs.
        active_deadline_seconds (int): Specifies the duration (in seconds) since startTime during which the job.
        can remain active before it is terminated. Must be a positive integer. This setting applies only to pods.
        where restartPolicy is OnFailure or Always.
        backoff_limit (int): Number of retries before marking this job as failed.
    """

    clean_pod_policy: CleanPodPolicy = None
    ttl_seconds_after_finished: Optional[int] = None
    active_deadline_seconds: Optional[int] = None
    backoff_limit: Optional[int] = None


@dataclass
class Worker:
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    replicas: Optional[int] = None
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class Master:
    """
    Configuration for master replica group. Master should always have 1 replica, so we don't need a `replicas` field
    """

    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class PyTorch(object):
    """
    Configuration for an executable `PyTorch Job <https://github.com/kubeflow/pytorch-operator>`_. Use this
    to run distributed PyTorch training on Kubernetes. Please notice, in most cases, you should not worry
    about the configuration of the master and worker groups. The default configuration should work. The only
    field you should change is the number of workers. Both replicas will use the same image, and the same
    resources inherited from task function decoration.

    Args:
        master: Configuration for the master replica group.
        worker: Configuration for the worker replica group.
        run_policy: Configuration for the run policy.
        num_workers: [DEPRECATED] This argument is deprecated. Use `worker.replicas` instead.
    """

    master: Master = field(default_factory=lambda: Master())
    worker: Worker = field(default_factory=lambda: Worker())
    run_policy: Optional[RunPolicy] = field(default_factory=lambda: None)
    # Support v0 config for backwards compatibility
    num_workers: Optional[int] = None


@dataclass
class Elastic(object):
    """
    Configuration for `torch elastic training <https://pytorch.org/docs/stable/elastic/run.html>`_.

    Use this to run single- or multi-node distributed pytorch elastic training on k8s.

    Single-node elastic training is executed in a k8s pod when `nnodes` is set to 1.
    Multi-node training is executed otherwise using a `Pytorch Job <https://github.com/kubeflow/training-operator>`_.

    Args:
        nnodes (Union[int, str]): Number of nodes, or the range of nodes in form <minimum_nodes>:<maximum_nodes>.
        nproc_per_node (str): Number of workers per node.
        start_method (str): Multiprocessing start method to use when creating workers.
        monitor_interval (int): Interval, in seconds, to monitor the state of workers.
        max_restarts (int): Maximum number of worker group restarts before failing.
        rdzv_configs (Dict[str, Any]): Additional rendezvous configs to pass to torch elastic, e.g. `{"timeout": 1200, "join_timeout": 900}`.
            See `torch.distributed.launcher.api.LaunchConfig` and `torch.distributed.elastic.rendezvous.dynamic_rendezvous.create_handler`.
    """

    nnodes: Union[int, str] = 1
    nproc_per_node: int = 1
    start_method: str = "spawn"
    monitor_interval: int = 5
    max_restarts: int = 0
    rdzv_configs: Dict[str, Any] = field(default_factory=dict)


class PyTorchFunctionTask(PythonFunctionTask[PyTorch]):
    """
    Plugin that submits a PyTorchJob (see https://github.com/kubeflow/pytorch-operator)
        defined by the code within the _task_function to k8s cluster.
    """

    _PYTORCH_TASK_TYPE = "pytorch"

    def __init__(self, task_config: PyTorch, task_function: Callable, **kwargs):
        if task_config.num_workers and task_config.worker.replicas:
            raise ValueError(
                "Cannot specify both `num_workers` and `worker.replicas`. Please use `worker.replicas` as `num_workers` is depreacated."
            )
        if task_config.num_workers is None and task_config.worker.replicas is None:
            raise ValueError(
                "Must specify either `num_workers` or `worker.replicas`. Please use `worker.replicas` as `num_workers` is depreacated."
            )
        super().__init__(
            task_config,
            task_function,
            task_type=self._PYTORCH_TASK_TYPE,
            # task_type_version controls the version of the task template, do not change
            task_type_version=1,
            **kwargs,
        )

    def _convert_replica_spec(
        self, replica_config: Union[Master, Worker]
    ) -> pytorch_task.DistributedPyTorchTrainingReplicaSpec:
        resources = convert_resources_to_resource_model(requests=replica_config.requests, limits=replica_config.limits)
        replicas = 1
        # Master should always have 1 replica
        if not isinstance(replica_config, Master):
            replicas = replica_config.replicas
        return pytorch_task.DistributedPyTorchTrainingReplicaSpec(
            replicas=replicas,
            image=replica_config.image,
            resources=resources.to_flyte_idl() if resources else None,
            restart_policy=replica_config.restart_policy.value if replica_config.restart_policy else None,
        )

    def _convert_run_policy(self, run_policy: RunPolicy) -> kubeflow_common.RunPolicy:
        return kubeflow_common.RunPolicy(
            clean_pod_policy=run_policy.clean_pod_policy.value if run_policy.clean_pod_policy else None,
            ttl_seconds_after_finished=run_policy.ttl_seconds_after_finished,
            active_deadline_seconds=run_policy.active_deadline_seconds,
            backoff_limit=run_policy.backoff_limit,
        )

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        worker = self._convert_replica_spec(self.task_config.worker)
        # support v0 config for backwards compatibility
        if self.task_config.num_workers:
            worker.replicas = self.task_config.num_workers

        run_policy = self._convert_run_policy(self.task_config.run_policy) if self.task_config.run_policy else None
        pytorch_job = pytorch_task.DistributedPyTorchTrainingTask(
            worker_replicas=worker,
            master_replicas=self._convert_replica_spec(self.task_config.master),
            run_policy=run_policy,
        )
        return MessageToDict(pytorch_job)


# Register the Pytorch Plugin into the flytekit core plugin system
TaskPlugins.register_pythontask_plugin(PyTorch, PyTorchFunctionTask)


class ElasticWorkerResult(NamedTuple):
    """
    A named tuple representing the result of a torch elastic worker process.

    Attributes:
        return_value (Any): The value returned by the task function in the worker process.
        decks (list[flytekit.Deck]): A list of flytekit Deck objects created in the worker process.
    """

    return_value: Any
    decks: List[flytekit.Deck]


def spawn_helper(
    fn: bytes, raw_output_prefix: str, checkpoint_dest: str, checkpoint_src: str, kwargs
) -> ElasticWorkerResult:
    """Help to spawn worker processes.

    The purpose of this function is to 1) be pickleable so that it can be used with
    the multiprocessing start method `spawn` and 2) to call a cloudpickle-serialized
    function passed to it. This function itself doesn't have to be pickleable. Without
    such a helper task functions, which are not pickleable, couldn't be used with the
    start method `spawn`.

    Args:
        fn (bytes): Cloudpickle-serialized target function to be executed in the worker process.
        raw_output_prefix (str): Where to write offloaded data (files, directories, dataframes).
        checkpoint_dest (str): If a previous checkpoint exists, this path should is set to the folder
            that contains the checkpoint information.
        checkpoint_src (str): Location where the new checkpoint should be copied to.

    Returns:
        ElasticWorkerResult: A named tuple containing the return value of the task function and a list of
            flytekit Deck objects created in the worker process.
    """
    from flytekit.bin.entrypoint import setup_execution

    with setup_execution(
        raw_output_data_prefix=raw_output_prefix,
        checkpoint_path=checkpoint_dest,
        prev_checkpoint=checkpoint_src,
    ):
        fn = cloudpickle.loads(fn)

        try:
            return_val = fn(**kwargs)
        except Exception as e:
            # See explanation in `create_recoverable_error_file` why we check
            # for recoverable errors here in the worker processes.
            if isinstance(e, FlyteRecoverableException):
                create_recoverable_error_file()
            raise
        return ElasticWorkerResult(return_value=return_val, decks=flytekit.current_context().decks)


class PytorchElasticFunctionTask(PythonFunctionTask[Elastic]):
    """
    Plugin for distributed training with torch elastic/torchrun (see
    https://pytorch.org/docs/stable/elastic/run.html).
    """

    _ELASTIC_TASK_TYPE = "pytorch"
    _ELASTIC_TASK_TYPE_STANDALONE = "python-task"

    def __init__(self, task_config: Elastic, task_function: Callable, **kwargs):
        task_type = self._ELASTIC_TASK_TYPE_STANDALONE if task_config.nnodes == 1 else self._ELASTIC_TASK_TYPE

        super(PytorchElasticFunctionTask, self).__init__(
            task_config=task_config,
            task_type=task_type,
            task_function=task_function,
            # task_type_version controls the version of the task template, do not change
            task_type_version=1,
            **kwargs,
        )
        try:
            from torch.distributed import run
        except ImportError:
            raise ImportError(TORCH_IMPORT_ERROR_MESSAGE)
        self.min_nodes, self.max_nodes = run.parse_min_max_nnodes(str(self.task_config.nnodes))

        """
        c10d is the backend recommended by torch elastic.
        https://pytorch.org/docs/stable/elastic/run.html#note-on-rendezvous-backend

        For c10d, no backend server has to be deployed.
        https://pytorch.org/docs/stable/elastic/run.html#deployment
        Instead, the workers will use the master's address as the rendezvous point.
        """
        self.rdzv_backend = "c10d"

    def _execute(self, **kwargs) -> Any:
        """
        Execute the task function using torch distributed's `elastic_launch`.

        Returns:
            The result of (global) rank zero.

        Raises:
            FlyteRecoverableException: If the first exception raised in the local worker group is or
                inherits from `FlyteRecoverableException`.
            RuntimeError: If the first exception raised in the local worker group is not and does not
                inherit from `FlyteRecoverableException`.
            IgnoreOutputs: Raised when the task is successful in any worker group with index > 0.
        """
        try:
            from torch.distributed.launcher.api import LaunchConfig, elastic_launch
        except ImportError:
            raise ImportError(TORCH_IMPORT_ERROR_MESSAGE)

        dist_env_vars_set = os.environ.get("PET_NNODES") is not None
        if not dist_env_vars_set and self.min_nodes > 1:
            logger.warning(
                (
                    f"`nnodes` is set to {self.task_config.nnodes} in elastic task but execution appears "
                    "to not run in a `PyTorchJob`. Rendezvous might timeout."
                )
            )

        config = LaunchConfig(
            run_id=flytekit.current_context().execution_id.name,
            min_nodes=self.min_nodes,
            max_nodes=self.max_nodes,
            nproc_per_node=self.task_config.nproc_per_node,
            rdzv_backend=self.rdzv_backend,  # rdzv settings
            rdzv_configs=self.task_config.rdzv_configs,
            rdzv_endpoint=os.environ.get("PET_RDZV_ENDPOINT", "localhost:0"),
            max_restarts=self.task_config.max_restarts,
            monitor_interval=self.task_config.monitor_interval,
            start_method=self.task_config.start_method,
        )

        if self.task_config.start_method == "spawn":
            """
            We use cloudpickle to serialize the non-pickleable task function.
            The torch elastic launcher then launches the spawn_helper function (which is pickleable)
            instead of the task function. This helper function, in the child-process, then deserializes
            the task function, again with cloudpickle, and executes it.
            """
            launcher_target_func = spawn_helper

            dumped_target_function = cloudpickle.dumps(self._task_function)

            ctx = flytekit.current_context()
            try:
                checkpoint_dest = ctx.checkpoint._checkpoint_dest
                checkpoint_src = ctx.checkpoint._checkpoint_src
            except NotImplementedError:
                # Not using checkpointing in parent process
                checkpoint_dest = None
                checkpoint_src = None

            launcher_args = (dumped_target_function, ctx.raw_output_prefix, checkpoint_dest, checkpoint_src, kwargs)
        elif self.task_config.start_method == "fork":
            """
            The torch elastic launcher doesn't support passing kwargs to the target function,
            only args. Flyte only works with kwargs. Thus, we create a closure which already has
            the task kwargs bound. We tell the torch elastic launcher to start this function in
            the child processes.
            """

            def fn_partial():
                """Closure of the task function with kwargs already bound."""
                try:
                    return_val = self._task_function(**kwargs)
                except Exception as e:
                    # See explanation in `create_recoverable_error_file` why we check
                    # for recoverable errors here in the worker processes.
                    if isinstance(e, FlyteRecoverableException):
                        create_recoverable_error_file()
                    raise
                return ElasticWorkerResult(return_value=return_val, decks=flytekit.current_context().decks)

            launcher_target_func = fn_partial
            launcher_args = ()

        else:
            raise Exception("Bad start method")

        from torch.distributed.elastic.multiprocessing.api import SignalException
        from torch.distributed.elastic.multiprocessing.errors import ChildFailedError

        try:
            out = elastic_launch(
                config=config,
                entrypoint=launcher_target_func,
            )(*launcher_args)
        except ChildFailedError as e:
            _, first_failure = e.get_first_failure()
            if is_recoverable_worker_error(first_failure):
                raise FlyteRecoverableException(e.format_msg())
            else:
                raise RuntimeError(e.format_msg())
        except SignalException as e:
            logger.exception(f"Elastic launch agent process terminating: {e}")
            raise IgnoreOutputs()

        # `out` is a dictionary of rank (not local rank) -> result
        # Rank 0 returns the result of the task function
        if 0 in out:
            # For rank 0, we transfer the decks created in the worker process to the parent process
            ctx = flytekit.current_context()
            for deck in out[0].decks:
                if not isinstance(deck, flytekit.deck.deck.TimeLineDeck):
                    ctx.decks.append(deck)

            return out[0].return_value
        else:
            raise IgnoreOutputs()

    def execute(self, **kwargs) -> Any:
        """
        This method will be invoked to execute the task.

        Handles the exception scope for the `_execute` method.
        """
        from flytekit.exceptions import scopes as exception_scopes

        return exception_scopes.user_entry_point(self._execute)(**kwargs)

    def get_custom(self, settings: SerializationSettings) -> Optional[Dict[str, Any]]:
        if self.task_config.nnodes == 1:
            """
            Torch elastic distributed training is executed in a normal k8s pod so that this
            works without the kubeflow train operator.
            """
            return super().get_custom(settings)
        else:
            from flyteidl.plugins.kubeflow.pytorch_pb2 import ElasticConfig

            elastic_config = ElasticConfig(
                rdzv_backend=self.rdzv_backend,
                min_replicas=self.min_nodes,
                max_replicas=self.max_nodes,
                nproc_per_node=self.task_config.nproc_per_node,
                max_restarts=self.task_config.max_restarts,
            )
            job = pytorch_task.DistributedPyTorchTrainingTask(
                worker_replicas=pytorch_task.DistributedPyTorchTrainingReplicaSpec(
                    replicas=self.max_nodes,
                ),
                elastic_config=elastic_config,
            )
            return MessageToDict(job)


# Register the PytorchElastic Plugin into the flytekit core plugin system
TaskPlugins.register_pythontask_plugin(Elastic, PytorchElasticFunctionTask)
