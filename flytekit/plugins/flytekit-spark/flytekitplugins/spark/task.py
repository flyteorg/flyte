import os
from dataclasses import dataclass
from typing import Any, Callable, Dict, Optional, Union, cast

from google.protobuf.json_format import MessageToDict

from flytekit import FlyteContextManager, PythonFunctionTask, lazy_module, logger
from flytekit.configuration import DefaultImages, SerializationSettings
from flytekit.core.context_manager import ExecutionParameters
from flytekit.extend import ExecutionState, TaskPlugins
from flytekit.extend.backend.base_agent import AsyncAgentExecutorMixin
from flytekit.image_spec import ImageSpec

from .models import SparkJob, SparkType

pyspark_sql = lazy_module("pyspark.sql")
SparkSession = pyspark_sql.SparkSession


@dataclass
class Spark(object):
    """
    Use this to configure a SparkContext for a your task. Task's marked with this will automatically execute
    natively onto K8s as a distributed execution of spark

    Args:
        spark_conf: Dictionary of spark config. The variables should match what spark expects
        hadoop_conf: Dictionary of hadoop conf. The variables should match a typical hadoop configuration for spark
        executor_path: Python binary executable to use for PySpark in driver and executor.
        applications_path: MainFile is the path to a bundled JAR, Python, or R file of the application to execute.
    """

    spark_conf: Optional[Dict[str, str]] = None
    hadoop_conf: Optional[Dict[str, str]] = None
    executor_path: Optional[str] = None
    applications_path: Optional[str] = None

    def __post_init__(self):
        if self.spark_conf is None:
            self.spark_conf = {}

        if self.hadoop_conf is None:
            self.hadoop_conf = {}


@dataclass
class Databricks(Spark):
    """
    Use this to configure a Databricks task. Task's marked with this will automatically execute
    natively onto databricks platform as a distributed execution of spark

    Args:
        databricks_conf: Databricks job configuration compliant with API version 2.1, supporting 2.0 use cases.
        For the configuration structure, visit here.https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure
        For updates in API 2.1, refer to: https://docs.databricks.com/en/workflows/jobs/jobs-api-updates.html
        databricks_token: Databricks access token. https://docs.databricks.com/dev-tools/api/latest/authentication.html.
        databricks_instance: Domain name of your deployment. Use the form <account>.cloud.databricks.com.
    """

    databricks_conf: Optional[Dict[str, Union[str, dict]]] = None
    databricks_token: Optional[str] = None
    databricks_instance: Optional[str] = None


# This method does not reset the SparkSession since it's a bit hard to handle multiple
# Spark sessions in a single application as it's described in:
# https://stackoverflow.com/questions/41491972/how-can-i-tear-down-a-sparksession-and-create-a-new-one-within-one-application.
def new_spark_session(name: str, conf: Dict[str, str] = None):
    """
    Optionally creates a new spark session and returns it.
    In cluster mode (running in hosted flyte, this will disregard the spark conf passed in)

    This method is safe to be used from any other method. That is one reason why, we have duplicated this code
    fragment with the pre-execute. For example in the notebook scenario we might want to call it from a separate kernel
    """
    import pyspark as _pyspark

    # We run in cluster-mode in Flyte.
    # Ref https://github.com/lyft/flyteplugins/blob/master/go/tasks/v1/flytek8s/k8s_resource_adds.go#L46
    sess_builder = _pyspark.sql.SparkSession.builder.appName(f"FlyteSpark: {name}")
    if "FLYTE_INTERNAL_EXECUTION_ID" not in os.environ and conf is not None:
        # If either of above cases is not true, then we are in local execution of this task
        # Add system spark-conf for local/notebook based execution.
        sess_builder = sess_builder.master("local[*]")
        spark_conf = _pyspark.SparkConf()
        for k, v in conf.items():
            spark_conf.set(k, v)
        spark_conf.set("spark.driver.bindAddress", "127.0.0.1")
        # In local execution, propagate PYTHONPATH to executors too. This makes the spark
        # execution hermetic to the execution environment. For example, it allows running
        # Spark applications using Bazel, without major changes.
        if "PYTHONPATH" in os.environ:
            spark_conf.setExecutorEnv("PYTHONPATH", os.environ["PYTHONPATH"])
        sess_builder = sess_builder.config(conf=spark_conf)

    # If there is a global SparkSession available, get it and try to stop it.
    _pyspark.sql.SparkSession.builder.getOrCreate().stop()

    return sess_builder.getOrCreate()
    # SparkSession.Stop does not work correctly, as it stops the session before all the data is written
    # sess.stop()


class PysparkFunctionTask(AsyncAgentExecutorMixin, PythonFunctionTask[Spark]):
    """
    Actual Plugin that transforms the local python code for execution within a spark context
    """

    _SPARK_TASK_TYPE = "spark"

    def __init__(
        self,
        task_config: Spark,
        task_function: Callable,
        container_image: Optional[Union[str, ImageSpec]] = None,
        **kwargs,
    ):
        self.sess: Optional[SparkSession] = None
        self._default_executor_path: str = task_config.executor_path
        self._default_applications_path: str = task_config.applications_path

        if isinstance(container_image, ImageSpec):
            if container_image.base_image is None:
                img = f"cr.flyte.org/flyteorg/flytekit:spark-{DefaultImages.get_version_suffix()}"
                container_image.base_image = img
                # default executor path and applications path in apache/spark-py:3.3.1
                self._default_executor_path = self._default_executor_path or "/usr/bin/python3"
                self._default_applications_path = (
                    self._default_applications_path or "local:///usr/local/bin/entrypoint.py"
                )
        super(PysparkFunctionTask, self).__init__(
            task_config=task_config,
            task_type=self._SPARK_TASK_TYPE,
            task_function=task_function,
            container_image=container_image,
            **kwargs,
        )

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        job = SparkJob(
            spark_conf=self.task_config.spark_conf,
            hadoop_conf=self.task_config.hadoop_conf,
            application_file=self._default_applications_path or "local://" + settings.entrypoint_settings.path,
            executor_path=self._default_executor_path or settings.python_interpreter,
            main_class="",
            spark_type=SparkType.PYTHON,
        )
        if isinstance(self.task_config, Databricks):
            cfg = cast(Databricks, self.task_config)
            job._databricks_conf = cfg.databricks_conf
            job._databricks_token = cfg.databricks_token
            job._databricks_instance = cfg.databricks_instance

        return MessageToDict(job.to_flyte_idl())

    def pre_execute(self, user_params: ExecutionParameters) -> ExecutionParameters:
        import pyspark as _pyspark

        ctx = FlyteContextManager.current_context()
        sess_builder = _pyspark.sql.SparkSession.builder.appName(f"FlyteSpark: {user_params.execution_id}")
        if not (ctx.execution_state and ctx.execution_state.mode == ExecutionState.Mode.TASK_EXECUTION):
            # If either of above cases is not true, then we are in local execution of this task
            # Add system spark-conf for local/notebook based execution.
            spark_conf = _pyspark.SparkConf()
            spark_conf.set("spark.driver.bindAddress", "127.0.0.1")
            for k, v in self.task_config.spark_conf.items():
                spark_conf.set(k, v)
            # In local execution, propagate PYTHONPATH to executors too. This makes the spark
            # execution hermetic to the execution environment. For example, it allows running
            # Spark applications using Bazel, without major changes.
            if "PYTHONPATH" in os.environ:
                spark_conf.setExecutorEnv("PYTHONPATH", os.environ["PYTHONPATH"])
            sess_builder = sess_builder.config(conf=spark_conf)

        self.sess = sess_builder.getOrCreate()
        return user_params.builder().add_attr("SPARK_SESSION", self.sess).build()

    def execute(self, **kwargs) -> Any:
        if isinstance(self.task_config, Databricks):
            # Use the Databricks agent to run it by default.
            try:
                ctx = FlyteContextManager.current_context()
                if not ctx.file_access.is_remote(ctx.file_access.raw_output_prefix):
                    raise ValueError(
                        "To submit a Databricks job locally,"
                        " please set --raw-output-data-prefix to a remote path. e.g. s3://, gcs//, etc."
                    )
                if ctx.execution_state and ctx.execution_state.is_local_execution():
                    return AsyncAgentExecutorMixin.execute(self, **kwargs)
            except Exception as e:
                logger.error(f"Agent failed to run the task with error: {e}")
                logger.info("Falling back to local execution")
        return PythonFunctionTask.execute(self, **kwargs)


# Inject the Spark plugin into flytekits dynamic plugin loading system
TaskPlugins.register_pythontask_plugin(Spark, PysparkFunctionTask)
TaskPlugins.register_pythontask_plugin(Databricks, PysparkFunctionTask)
