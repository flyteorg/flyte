# Flytekit Great Expectations Plugin

Great Expectations helps enforce data quality. The plugin supports the usage of Great Expectations as task and type.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-great-expectations
```

## Task Example
```python
import os

import pandas as pd
from flytekit import Resources, kwtypes, task, workflow
from flytekitplugins.great_expectations import BatchRequestConfig, GreatExpectationsTask

simple_task_object = GreatExpectationsTask(
    name="great_expectations_task_simple",
    datasource_name="data",
    inputs=kwtypes(dataset=str),
    expectation_suite_name="test.demo",
    data_connector_name="data_example_data_connector",
    context_root_dir="great_expectations",
)

@task(limits=Resources(mem="500Mi"))
def simple_task(csv_file: str) -> int:
    result = simple_task_object(dataset=csv_file)
    df = pd.read_csv(os.path.join("greatexpectations", "data", csv_file))
    return df.shape[0]

@workflow
def simple_wf(dataset: str = "yellow_tripdata_sample_2019-01.csv") -> int:
    return simple_task(csv_file=dataset)
```

## Type Example
```python
from flytekit import workflow
from flytekitplugins.great_expectations import (
    BatchRequestConfig,
    GreatExpectationsFlyteConfig,
    GreatExpectationsType,
)

def simple_task(
    directory: GreatExpectationsType[
        str,
        GreatExpectationsFlyteConfig(
            datasource_name="data",
            expectation_suite_name="test.demo",
            data_connector_name="my_data_connector",
            batch_request_config=BatchRequestConfig(
                data_connector_query={
                    "batch_filter_parameters": {
                        "year": "2019",
                        "month": "01",
                    },
                    "limit": 10,
                },
            ),
            context_root_dir="great_expectations",
        ),
    ]
) -> str:
    return f"Validation works for {directory}!"


@workflow
def simple_wf(directory: str = "my_assets") -> str:
    return simple_task(directory=directory)
```

[More examples](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/greatexpectations/index.html) can be found in the documentation.
