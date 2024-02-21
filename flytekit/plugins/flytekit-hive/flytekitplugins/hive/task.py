from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Type

from google.protobuf.json_format import MessageToDict

from flytekit.configuration import SerializationSettings
from flytekit.extend import SQLTask
from flytekit.models.qubole import HiveQuery, QuboleHiveJob
from flytekit.types.schema import FlyteSchema


@dataclass
class HiveConfig(object):
    """
    HiveConfig should be used to configure a Hive Task. Config values are statically defined during the construction
    of a task, and are not parameterized.
    Note: A separate story is in progress to dynamically alter configuration for an execution launch.

    Args:
        cluster_label: A string value that helps in identifying the cluster label.
        tags: Any tags that should be associated with the remote execution request.
    """

    cluster_label: str = ""
    tags: Optional[List[str]] = None

    def __post_init__(self):
        if self.tags is None:
            self.tags = []


class HiveTask(SQLTask[HiveConfig]):
    """
    This is the simplest form of a Hive Task, that can be used even for tasks that do not produce any output.
    """

    _TASK_TYPE = "hive"

    def __init__(
        self,
        name: str,
        query_template: str,
        task_config: Optional[HiveConfig] = None,
        inputs: Optional[Dict[str, Type]] = None,
        output_schema_type: Optional[Type[FlyteSchema]] = None,
        **kwargs,
    ):
        """
        Args:
            name: Name of this task, should be unique in the project
            config: Type HiveConfig object
            inputs: Name and type of inputs specified as an ordered dictionary
            query_template: The actual query to run. We use Flyte's Golang templating format for Query templating.
                            Refer to the templating documentation
            output_schema_type: If some data is produced by this query, then you can specify the output schema type
            **kwargs: All other args required by Parent type - SQLTask
        """
        outputs = None
        if output_schema_type is not None:
            outputs = {
                "results": output_schema_type,
            }
        if task_config is None:
            task_config = HiveConfig()
        super().__init__(
            name=name,
            task_config=task_config,
            query_template=query_template,
            inputs=inputs,
            outputs=outputs,
            task_type=self._TASK_TYPE,
            **kwargs,
        )
        self._output_schema_type = output_schema_type

    @property
    def cluster_label(self) -> str:
        return self.task_config.cluster_label

    @property
    def output_schema_type(self) -> Optional[Type[FlyteSchema]]:
        return self._output_schema_type

    @property
    def tags(self) -> List[str]:
        return self.task_config.tags

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        # timeout_sec and retry_count will become deprecated, please use timeout and retry settings on the Task
        query = HiveQuery(query=self.query_template, timeout_sec=0, retry_count=0)
        job = QuboleHiveJob(
            query=query,
            cluster_label=self.cluster_label,
            tags=self.tags,
        )
        return MessageToDict(job.to_flyte_idl())


class HiveSelectTask(HiveTask):
    _HIVE_QUERY_FORMATTER = """
        {stage_query_str}

        CREATE TEMPORARY TABLE {{{{ .PerRetryUniqueKey }}}}_tmp AS {select_query_str};
        CREATE EXTERNAL TABLE {{{{ .PerRetryUniqueKey }}}} LIKE {{{{ .PerRetryUniqueKey }}}}_tmp STORED AS PARQUET;
        ALTER TABLE {{{{ .PerRetryUniqueKey }}}} SET LOCATION '{{{{ .RawOutputDataPrefix }}}}';

        INSERT OVERWRITE TABLE {{{{ .PerRetryUniqueKey }}}}
            SELECT *
            FROM {{{{ .PerRetryUniqueKey }}}}_tmp;
        DROP TABLE {{{{ .PerRetryUniqueKey }}}};
        """

    def __init__(
        self,
        name: str,
        select_query: str,
        inputs: Optional[Dict[str, Type]],
        output_schema_type: Optional[Type[FlyteSchema]] = None,  # Should default to a generic schema object?
        config: Optional[HiveConfig] = None,
        stage_query: Optional[str] = None,
        **kwargs,
    ):
        """
        Args:
            select_query: Singular query that returns a Tabular dataset
            stage_query: optional query that should be executed before the actual ``select_query``. This can usually
                be used for setting memory or the an alternate execution engine like `tez <https://tez.apache.org/>`__
        """
        query_template = HiveSelectTask._HIVE_QUERY_FORMATTER.format(
            stage_query_str=stage_query or "", select_query_str=select_query.strip().strip(";")
        )
        super().__init__(
            name=name,
            config=config,
            inputs=inputs,
            query_template=query_template,
            output_schema_type=output_schema_type,
        )
