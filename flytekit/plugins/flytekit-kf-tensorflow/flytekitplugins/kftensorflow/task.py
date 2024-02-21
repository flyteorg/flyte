"""
This Plugin adds the capability of running distributed tensorflow training to Flyte using backend plugins, natively on
Kubernetes. It leverages `TF Job <https://github.com/kubeflow/tf-operator>`_ Plugin from kubeflow.
"""
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, Optional, Union

from flyteidl.plugins.kubeflow import common_pb2 as kubeflow_common
from flyteidl.plugins.kubeflow import tensorflow_pb2 as tensorflow_task
from google.protobuf.json_format import MessageToDict

from flytekit import PythonFunctionTask, Resources
from flytekit.configuration import SerializationSettings
from flytekit.core.resources import convert_resources_to_resource_model
from flytekit.extend import TaskPlugins


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
    RunPolicy describes a set of policies to apply to the execution of a Kubeflow job.

    Args:
        clean_pod_policy: The policy for cleaning up pods after the job completes. Defaults to None.
        ttl_seconds_after_finished: The time-to-live (TTL) in seconds for cleaning up finished jobs.
        active_deadline_seconds: The duration (in seconds) since startTime during which the job can remain
            active before it is terminated. Must be a positive integer. This setting applies only to pods
            where restartPolicy is OnFailure or Always.
        backoff_limit: The number of retries before marking this job as failed.
    """

    clean_pod_policy: CleanPodPolicy = None
    ttl_seconds_after_finished: Optional[int] = None
    active_deadline_seconds: Optional[int] = None
    backoff_limit: Optional[int] = None


@dataclass
class Chief:
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    replicas: Optional[int] = None
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class PS:
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    replicas: Optional[int] = None
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class Worker:
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    replicas: Optional[int] = None
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class Evaluator:
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None
    replicas: int = 0
    restart_policy: Optional[RestartPolicy] = None


@dataclass
class TfJob:
    """
    Configuration for an executable `TensorFlow Job <https://github.com/kubeflow/tf-operator>`_. Use this
    to run distributed TensorFlow training on Kubernetes.

    Args:
        chief: Configuration for the chief replica group.
        ps: Configuration for the parameter server (PS) replica group.
        worker: Configuration for the worker replica group.
        evaluator: Configuration for the evaluator replica group.
        run_policy: Configuration for the run policy.
        num_workers: [DEPRECATED] This argument is deprecated. Use `worker.replicas` instead.
        num_ps_replicas: [DEPRECATED] This argument is deprecated. Use `ps.replicas` instead.
        num_chief_replicas: [DEPRECATED] This argument is deprecated. Use `chief.replicas` instead.
    """

    chief: Chief = field(default_factory=lambda: Chief())
    ps: PS = field(default_factory=lambda: PS())
    worker: Worker = field(default_factory=lambda: Worker())
    evaluator: Evaluator = field(default_factory=lambda: Evaluator())
    run_policy: Optional[RunPolicy] = field(default_factory=lambda: None)
    # Support v0 config for backwards compatibility
    num_workers: Optional[int] = None
    num_ps_replicas: Optional[int] = None
    num_chief_replicas: Optional[int] = None
    num_evaluator_replicas: Optional[int] = None


class TensorflowFunctionTask(PythonFunctionTask[TfJob]):
    """
    Plugin that submits a TFJob (see https://github.com/kubeflow/tf-operator)
        defined by the code within the _task_function to k8s cluster.
    """

    _TF_JOB_TASK_TYPE = "tensorflow"

    def __init__(self, task_config: TfJob, task_function: Callable, **kwargs):
        if task_config.num_workers and task_config.worker.replicas:
            raise ValueError(
                "Cannot specify both `num_workers` and `worker.replicas`. Please use `worker.replicas` as `num_workers` is depreacated."
            )
        if task_config.num_workers is None and task_config.worker.replicas is None:
            raise ValueError(
                "Must specify either `num_workers` or `worker.replicas`. Please use `worker.replicas` as `num_workers` is depreacated."
            )
        if task_config.num_chief_replicas and task_config.chief.replicas:
            raise ValueError(
                "Cannot specify both `num_chief_replicas` and `chief.replicas`. Please use `chief.replicas` as `num_chief_replicas` is depreacated."
            )
        if task_config.num_chief_replicas is None and task_config.chief.replicas is None:
            raise ValueError(
                "Must specify either `num_chief_replicas` or `chief.replicas`. Please use `chief.replicas` as `num_chief_replicas` is depreacated."
            )
        if task_config.num_ps_replicas and task_config.ps.replicas:
            raise ValueError(
                "Cannot specify both `num_ps_replicas` and `ps.replicas`. Please use `ps.replicas` as `num_ps_replicas` is depreacated."
            )
        if task_config.num_ps_replicas is None and task_config.ps.replicas is None:
            raise ValueError(
                "Must specify either `num_ps_replicas` or `ps.replicas`. Please use `ps.replicas` as `num_ps_replicas` is depreacated."
            )
        if task_config.num_evaluator_replicas and task_config.evaluator.replicas > 0:
            raise ValueError(
                "Cannot specify both `num_evaluator_replicas` and `evaluator.replicas`. Please use `evaluator.replicas` as `num_evaluator_replicas` is depreacated."
            )
        super().__init__(
            task_type=self._TF_JOB_TASK_TYPE,
            task_config=task_config,
            task_function=task_function,
            task_type_version=1,
            **kwargs,
        )

    def _convert_replica_spec(
        self, replica_config: Union[Chief, PS, Worker, Evaluator]
    ) -> tensorflow_task.DistributedTensorflowTrainingReplicaSpec:
        resources = convert_resources_to_resource_model(requests=replica_config.requests, limits=replica_config.limits)
        return tensorflow_task.DistributedTensorflowTrainingReplicaSpec(
            replicas=replica_config.replicas,
            image=replica_config.image,
            resources=resources.to_flyte_idl() if resources else None,
            restart_policy=replica_config.restart_policy.value if replica_config.restart_policy else None,
        )

    def _convert_run_policy(self, run_policy: RunPolicy) -> kubeflow_common.RunPolicy:
        return kubeflow_common.RunPolicy(
            clean_pod_policy=run_policy.clean_pod_policy.value if run_policy.clean_pod_policy.value else None,
            ttl_seconds_after_finished=run_policy.ttl_seconds_after_finished,
            active_deadline_seconds=run_policy.active_deadline_seconds,
            backoff_limit=run_policy.backoff_limit,
        )

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        chief = self._convert_replica_spec(self.task_config.chief)
        if self.task_config.num_chief_replicas:
            chief.replicas = self.task_config.num_chief_replicas

        worker = self._convert_replica_spec(self.task_config.worker)
        if self.task_config.num_workers:
            worker.replicas = self.task_config.num_workers

        ps = self._convert_replica_spec(self.task_config.ps)
        if self.task_config.num_ps_replicas:
            ps.replicas = self.task_config.num_ps_replicas

        evaluator = self._convert_replica_spec(self.task_config.evaluator)
        if self.task_config.num_evaluator_replicas:
            evaluator.replicas = self.task_config.num_evaluator_replicas

        run_policy = self._convert_run_policy(self.task_config.run_policy) if self.task_config.run_policy else None
        training_task = tensorflow_task.DistributedTensorflowTrainingTask(
            chief_replicas=chief,
            worker_replicas=worker,
            ps_replicas=ps,
            evaluator_replicas=evaluator,
            run_policy=run_policy,
        )

        return MessageToDict(training_task)


# Register the Tensorflow Plugin into the flytekit core plugin system
TaskPlugins.register_pythontask_plugin(TfJob, TensorflowFunctionTask)
