# Flytekit whylogs Plugin

whylogs is an open source library for logging any kind of data. With whylogs,
you are able to generate summaries of datasets (called whylogs profiles) which
can be used to:

- Create data constraints to know whether your data looks the way it should
- Quickly visualize key summary statistics about a dataset
- Track changes in a dataset over time

```bash
pip install flytekitplugins-whylogs
```

To generate profiles, you can add a task like the following:

```python
import whylogs as why
from whylogs.core import DatasetProfileView

import pandas as pd

from flytekit import task

@task
def profile(df: pd.DataFrame) -> DatasetProfileView:
    result = why.log(df) # Various overloads for different common data types exist
    profile_view = result.view()
    return profile
```

>**NOTE:** You'll be passing around `DatasetProfileView` from tasks, not `DatasetProfile`.

## Validating Data

A common step in data pipelines is data validation. This can be done in
`whylogs` through the constraint feature. You'll be able to create failure tasks
if the data in the workflow doesn't conform to some configured constraints, like
min/max values on features, data types on features, etc.

```python
from whylogs.core.constraints.factories import greater_than_number, mean_between_range

@task
def validate_data(profile_view: DatasetProfileView):
    builder = ConstraintsBuilder(dataset_profile_view=profile_view)
    builder.add_constraint(greater_than_number(column_name="my_column", number=0.14))
    builder.add_constraint(mean_between_range(column_name="my_other_column", lower=2, upper=3))
    constraint = builder.build()
    valid = constraint.validate()

    if valid is False:
        print(constraint.report())
        raise Exception("Invalid data found")
```

If you want to learn more about whylogs, check out our [example notebooks](https://github.com/whylabs/whylogs/tree/mainline/python/examples).
