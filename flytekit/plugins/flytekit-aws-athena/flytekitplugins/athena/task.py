from dataclasses import dataclass
from typing import Any, Dict, Optional, Type

from google.protobuf.json_format import MessageToDict

from flytekit.configuration import SerializationSettings
from flytekit.extend import SQLTask
from flytekit.models.presto import PrestoQuery
from flytekit.types.schema import FlyteSchema


@dataclass
class AthenaConfig(object):
    """
    AthenaConfig should be used to configure a Athena Task.
    """

    # The database to query against
    database: Optional[str] = None
    # The optional workgroup to separate query execution.
    workgroup: Optional[str] = None
    # The catalog to set for the given Presto query
    catalog: Optional[str] = None


class AthenaTask(SQLTask[AthenaConfig]):
    """
    This is the simplest form of a Athena Task, that can be used even for tasks that do not produce any output.
    """

    # This task is executed using the presto handler in the backend.
    _TASK_TYPE = "presto"

    def __init__(
        self,
        name: str,
        query_template: str,
        task_config: Optional[AthenaConfig] = None,
        inputs: Optional[Dict[str, Type]] = None,
        output_schema_type: Optional[Type[FlyteSchema]] = None,
        **kwargs,
    ):
        """
        To be used to query Athena databases.

        :param name: Name of this task, should be unique in the project
        :param query_template: The actual query to run. We use Flyte's Golang templating format for Query templating.
          Refer to the templating documentation
        :param task_config: AthenaConfig object
        :param inputs: Name and type of inputs specified as an ordered dictionary
        :param output_schema_type: If some data is produced by this query, then you can specify the output schema type
        :param kwargs: All other args required by Parent type - SQLTask
        """
        outputs = None
        if output_schema_type is not None:
            outputs = {
                "results": output_schema_type,
            }
        if task_config is None:
            task_config = AthenaConfig()
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

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        # This task is executed using the presto handler in the backend.
        job = PrestoQuery(
            statement=self.query_template,
            schema=self.task_config.database,
            routing_group=self.task_config.workgroup,
            catalog=self.task_config.catalog,
        )
        return MessageToDict(job.to_flyte_idl())
