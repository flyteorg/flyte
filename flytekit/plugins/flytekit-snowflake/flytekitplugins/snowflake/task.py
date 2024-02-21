from dataclasses import dataclass
from typing import Dict, Optional, Type

from flytekit.configuration import SerializationSettings
from flytekit.extend import SQLTask
from flytekit.extend.backend.base_agent import AsyncAgentExecutorMixin
from flytekit.models import task as _task_model
from flytekit.types.structured import StructuredDataset

_USER_FIELD = "user"
_ACCOUNT_FIELD = "account"
_DATABASE_FIELD = "database"
_SCHEMA_FIELD = "schema"
_WAREHOUSE_FIELD = "warehouse"
_TABLE_FIELD = "table"


@dataclass
class SnowflakeConfig(object):
    """
    SnowflakeConfig should be used to configure a Snowflake Task.
    """

    # The user to query against
    user: Optional[str] = None
    # The account to query against
    account: Optional[str] = None
    # The database to query against
    database: Optional[str] = None
    # The optional schema to separate query execution.
    schema: Optional[str] = None
    # The optional warehouse to set for the given Snowflake query
    warehouse: Optional[str] = None
    # The optional table to set for the given Snowflake query
    table: Optional[str] = None


class SnowflakeTask(AsyncAgentExecutorMixin, SQLTask[SnowflakeConfig]):
    """
    This is the simplest form of a Snowflake Task, that can be used even for tasks that do not produce any output.
    """

    # This task is executed using the snowflake handler in the backend.
    _TASK_TYPE = "snowflake"

    def __init__(
        self,
        name: str,
        query_template: str,
        task_config: Optional[SnowflakeConfig] = None,
        inputs: Optional[Dict[str, Type]] = None,
        output_schema_type: Optional[Type[StructuredDataset]] = None,
        **kwargs,
    ):
        """
        To be used to query Snowflake databases.

        :param name: Name of this task, should be unique in the project
        :param query_template: The actual query to run. We use Flyte's Golang templating format for Query templating.
          Refer to the templating documentation
        :param task_config: SnowflakeConfig object
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
            task_config = SnowflakeConfig()
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

    def get_config(self, settings: SerializationSettings) -> Dict[str, str]:
        return {
            _USER_FIELD: self.task_config.user,
            _ACCOUNT_FIELD: self.task_config.account,
            _DATABASE_FIELD: self.task_config.database,
            _SCHEMA_FIELD: self.task_config.schema,
            _WAREHOUSE_FIELD: self.task_config.warehouse,
            _TABLE_FIELD: self.task_config.table,
        }

    def get_sql(self, settings: SerializationSettings) -> Optional[_task_model.Sql]:
        sql = _task_model.Sql(statement=self.query_template, dialect=_task_model.Sql.Dialect.ANSI)
        return sql
