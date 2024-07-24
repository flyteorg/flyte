import re
from typing import Any, Dict, Optional, Tuple, Type, TypeVar

from flytekit.core.base_task import PythonTask, TaskMetadata
from flytekit.core.interface import Interface

T = TypeVar("T")


class SQLTask(PythonTask[T]):
    """
    Base task types for all SQL tasks. See :py:class:`flytekit.extras.sqlite3.task.SQLite3Task`
    and :py:class:`flytekitplugins.athena.task.AthenaTask` for examples of how to use it as a base class.

    .. autoclass:: flytekit.extras.sqlite3.task.SQLite3Task
       :noindex:
    """

    _INPUT_REGEX = re.compile(r"({{\s*.inputs.(\w+)\s*}})", re.IGNORECASE)

    def __init__(
        self,
        name: str,
        query_template: str,
        task_config: Optional[T] = None,
        task_type="sql_task",
        inputs: Optional[Dict[str, Tuple[Type, Any]]] = None,
        metadata: Optional[TaskMetadata] = None,
        outputs: Optional[Dict[str, Type]] = None,
        **kwargs,
    ):
        """
        This SQLTask should mostly just be used as a base class for other SQL task types and should not be used
        directly. See :py:class:`flytekit.extras.sqlite3.task.SQLite3Task`
        """
        super().__init__(
            task_type=task_type,
            name=name,
            interface=Interface(inputs=inputs or {}, outputs=outputs or {}),
            metadata=metadata,
            task_config=task_config,
            **kwargs,
        )
        self._query_template = re.sub(r"\s+", " ", query_template.replace("\n", " ").replace("\t", " ")).strip()

    @property
    def query_template(self) -> str:
        return self._query_template

    def execute(self, **kwargs) -> Any:
        raise Exception("Cannot run a SQL Task natively, please mock.")

    def get_query(self, **kwargs) -> str:
        return self.interpolate_query(self.query_template, **kwargs)

    @classmethod
    def interpolate_query(cls, query_template, **kwargs) -> Any:
        """
        This function will fill in the query template with the provided kwargs and return the interpolated query.
        Please note that when SQL tasks run in Flyte, this step is done by the task executor.
        """
        modified_query = query_template
        matched = set()
        for match in cls._INPUT_REGEX.finditer(query_template):
            expr = match.groups()[0]
            var = match.groups()[1]
            if var not in kwargs:
                raise ValueError(f"Variable {var} in Query (part of {expr}) not found in inputs {kwargs.keys()}")
            matched.add(var)
            val = kwargs[var]
            # str conversion should be deliberate, with right conversion for each type
            modified_query = modified_query.replace(expr, str(val))

        if len(matched) < len(kwargs.keys()):
            diff = set(kwargs.keys()).difference(matched)
            raise ValueError(f"Extra Inputs have no matches in query template - missing {diff}")
        return modified_query
