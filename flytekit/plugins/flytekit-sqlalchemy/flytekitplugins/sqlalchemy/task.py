import typing
from dataclasses import dataclass

from flytekit import current_context, kwtypes, lazy_module
from flytekit.configuration import SerializationSettings
from flytekit.configuration.default_images import DefaultImages, PythonVersion
from flytekit.core.base_sql_task import SQLTask
from flytekit.core.python_customized_container_task import PythonCustomizedContainerTask
from flytekit.core.shim_task import ShimTaskExecutor
from flytekit.loggers import logger
from flytekit.models import task as task_models
from flytekit.models.security import Secret
from flytekit.types.schema import FlyteSchema

pd = lazy_module("pandas")
pandas_io_sql = lazy_module("pandas.io.sql")
sqlalchemy = lazy_module("sqlalchemy")


class SQLAlchemyDefaultImages(DefaultImages):
    """Default images for the sqlalchemy flytekit plugin."""

    _DEFAULT_IMAGE_PREFIXES = {
        PythonVersion.PYTHON_3_8: "cr.flyte.org/flyteorg/flytekit:py3.8-sqlalchemy-",
        PythonVersion.PYTHON_3_9: "cr.flyte.org/flyteorg/flytekit:py3.9-sqlalchemy-",
        PythonVersion.PYTHON_3_10: "cr.flyte.org/flyteorg/flytekit:py3.10-sqlalchemy-",
        PythonVersion.PYTHON_3_11: "cr.flyte.org/flyteorg/flytekit:py3.11-sqlalchemy-",
    }


@dataclass
class SQLAlchemyConfig(object):
    """
    Use this configuration to configure task. String should be standard
    sqlalchemy connector format
    (https://docs.sqlalchemy.org/en/14/core/engines.html#database-urls).
    Database can be found:
    - within the container
    - or from a publicly accessible source

    Args:
        uri: default sqlalchemy connector
        connect_args: sqlalchemy kwarg overrides -- ex: host
        secret_connect_args: flyte secrets loaded into sqlalchemy connect args
            -- ex: {"password": flytekit.models.security.Secret(name=SECRET_NAME, group=SECRET_GROUP)}
    """

    uri: str
    connect_args: typing.Optional[typing.Dict[str, typing.Any]] = None
    secret_connect_args: typing.Optional[typing.Dict[str, Secret]] = None

    @staticmethod
    def _secret_to_dict(secret: Secret) -> typing.Dict[str, typing.Optional[str]]:
        return {
            "group": secret.group,
            "key": secret.key,
            "group_version": secret.group_version,
            "mount_requirement": secret.mount_requirement.value,
        }

    def secret_connect_args_to_dicts(self) -> typing.Optional[typing.Dict[str, typing.Dict[str, typing.Optional[str]]]]:
        if self.secret_connect_args is None:
            return None

        secret_connect_args_dicts = {}
        for key, secret in self.secret_connect_args.items():
            secret_connect_args_dicts[key] = self._secret_to_dict(secret)

        return secret_connect_args_dicts


class SQLAlchemyTask(PythonCustomizedContainerTask[SQLAlchemyConfig], SQLTask[SQLAlchemyConfig]):
    """
    Makes it possible to run client side SQLAlchemy queries that optionally return a FlyteSchema object
    """

    # TODO: How should we use pre-built containers for running portable tasks like this? Should this always be a referenced task type?

    _SQLALCHEMY_TASK_TYPE = "sqlalchemy"

    def __init__(
        self,
        name: str,
        query_template: str,
        task_config: SQLAlchemyConfig,
        inputs: typing.Optional[typing.Dict[str, typing.Type]] = None,
        output_schema_type: typing.Optional[typing.Type[FlyteSchema]] = FlyteSchema,
        container_image: str = SQLAlchemyDefaultImages.default_image(),
        **kwargs,
    ):
        if output_schema_type:
            outputs = kwtypes(results=output_schema_type)
        else:
            outputs = None

        super().__init__(
            name=name,
            task_config=task_config,
            executor_type=SQLAlchemyTaskExecutor,
            task_type=self._SQLALCHEMY_TASK_TYPE,
            query_template=query_template,
            container_image=container_image,
            inputs=inputs,
            outputs=outputs,
            **kwargs,
        )

    @property
    def output_columns(self) -> typing.Optional[typing.List[str]]:
        c = self.python_interface.outputs["results"].column_names()
        return c if c else None

    def get_custom(self, settings: SerializationSettings) -> typing.Dict[str, typing.Any]:
        return {
            "query_template": self.query_template,
            "uri": self.task_config.uri,
            "connect_args": self.task_config.connect_args or {},
            "secret_connect_args": self.task_config.secret_connect_args_to_dicts(),
        }


class SQLAlchemyTaskExecutor(ShimTaskExecutor[SQLAlchemyTask]):
    def execute_from_model(self, tt: task_models.TaskTemplate, **kwargs) -> typing.Any:
        if tt.custom["secret_connect_args"] is not None:
            for key, secret_dict in tt.custom["secret_connect_args"].items():
                value = current_context().secrets.get(group=secret_dict["group"], key=secret_dict["key"])
                tt.custom["connect_args"][key] = value

        engine = sqlalchemy.create_engine(tt.custom["uri"], connect_args=tt.custom["connect_args"], echo=False)
        logger.info(f"Connecting to db {tt.custom['uri']}")

        interpolated_query = SQLAlchemyTask.interpolate_query(tt.custom["query_template"], **kwargs)
        logger.info(f"Interpolated query {interpolated_query}")
        with engine.begin() as connection:
            df = None
            if tt.interface.outputs:
                df = pd.read_sql_query(sqlalchemy.text(interpolated_query), connection)
            else:
                pandas_io_sql.pandasSQL_builder(connection).execute(sqlalchemy.text(interpolated_query))
        return df
