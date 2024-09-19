import json
from typing import Dict, List, NamedTuple, Optional, Union

from flytekit import PythonInstanceTask, lazy_module
from flytekit.extend import Interface
from flytekit.types.structured.structured_dataset import StructuredDataset

duckdb = lazy_module("duckdb")
pd = lazy_module("pandas")
pa = lazy_module("pyarrow")


class QueryOutput(NamedTuple):
    counter: int = -1
    output: Optional[str] = None


class DuckDBQuery(PythonInstanceTask):
    _TASK_TYPE = "duckdb"

    def __init__(
        self,
        name: str,
        query: Union[str, List[str]],
        inputs: Optional[Dict[str, Union[StructuredDataset, list]]] = None,
        **kwargs,
    ):
        """
        This method initializes the DuckDBQuery.

        Args:
            name: Name of the task
            query: DuckDB query to execute
            inputs: The query parameters to be used while executing the query
        """
        self._query = query
        # create an in-memory database that's non-persistent
        self._con = duckdb.connect(":memory:")

        outputs = {"result": StructuredDataset}

        super(DuckDBQuery, self).__init__(
            name=name,
            task_type=self._TASK_TYPE,
            task_config=None,
            interface=Interface(inputs=inputs, outputs=outputs),
            **kwargs,
        )

    def _execute_query(self, params: list, query: str, counter: int, multiple_params: bool):
        """
        This method runs the DuckDBQuery.

        Args:
            params: Query parameters to use while executing the query
            query: DuckDB query to execute
            counter: Use counter to map user-given arguments to the query parameters
            multiple_params: Set flag to indicate the presence of params for multiple queries
        """
        if any(x in query for x in ("$", "?")):
            if multiple_params:
                counter += 1
                if not counter < len(params):
                    raise ValueError("Parameter doesn't exist.")
                if "insert" in query.lower():
                    # run executemany disregarding the number of entries to store for an insert query
                    yield QueryOutput(output=self._con.executemany(query, params[counter]), counter=counter)
                else:
                    yield QueryOutput(output=self._con.execute(query, params[counter]), counter=counter)
            else:
                if params:
                    yield QueryOutput(output=self._con.execute(query, params), counter=counter)
                else:
                    raise ValueError("Parameter not specified.")
        else:
            yield QueryOutput(output=self._con.execute(query), counter=counter)

    def execute(self, **kwargs) -> StructuredDataset:
        # TODO: Enable iterative download after adding the functionality to structured dataset code.
        params = None
        for key in self.python_interface.inputs.keys():
            val = kwargs.get(key)
            if isinstance(val, StructuredDataset):
                # register structured dataset
                self._con.register(key, val.open(pa.Table).all())
            elif isinstance(val, (pd.DataFrame, pa.Table)):
                # register pandas dataframe/arrow table
                self._con.register(key, val)
            elif isinstance(val, list):
                # copy val into params
                params = val
            elif isinstance(val, str):
                # load into a list
                params = json.loads(val)
            else:
                raise ValueError(f"Expected inputs of type StructuredDataset, str or list, received {type(val)}")

        final_query = self._query
        query_output = QueryOutput()
        # set flag to indicate the presence of params for multiple queries
        multiple_params = isinstance(params[0], list) if params else False

        if isinstance(self._query, list) and len(self._query) > 1:
            # loop until the penultimate query
            for query in self._query[:-1]:
                query_output = next(
                    self._execute_query(
                        params=params, query=query, counter=query_output.counter, multiple_params=multiple_params
                    )
                )
            final_query = self._query[-1]

        # fetch query output from the last query
        # expecting a SELECT query
        dataframe = next(
            self._execute_query(
                params=params, query=final_query, counter=query_output.counter, multiple_params=multiple_params
            )
        ).output.arrow()

        return StructuredDataset(dataframe=dataframe)
