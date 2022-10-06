"""
whylogs Example
---------------

This examples shows users how to profile pandas DataFrames with whylogs,
pass them within tasks and also use our renderers to create a SummaryDriftReport
and a ConstraintsReport with failed and passed constraints.
"""

# %%
# First, let's make all the necessary imports for our example to run properly
import os

import flytekit
import numpy as np
import pandas as pd
import whylogs as why
from flytekit import conditional, task, workflow
from flytekitplugins.whylogs.renderer import WhylogsConstraintsRenderer, WhylogsSummaryDriftRenderer
from flytekitplugins.whylogs.schema import WhylogsDatasetProfileTransformer
from sklearn.datasets import load_diabetes
from whylogs.core import DatasetProfileView
from whylogs.core.constraints import ConstraintsBuilder
from whylogs.core.constraints.factories import (
    greater_than_number,
    mean_between_range,
    null_percentage_below_number,
    smaller_than_number,
)


# %%
# Next thing is defining a task to read our reference dataset.
# For this, we will take scikit-learn's entire example Diabetes dataset
@task
def get_reference_data() -> pd.DataFrame:
    diabetes = load_diabetes()
    df = pd.DataFrame(diabetes.data, columns=diabetes.feature_names)
    df["target"] = pd.DataFrame(diabetes.target)
    return df


# %%
# To some extent, we wanted to show kinds of drift in our example,
# so in order to reproduce some of what real-life data behaves
# we will take an arbitrary subset of the reference dataset
@task
def get_target_data() -> pd.DataFrame:
    diabetes = load_diabetes()
    df = pd.DataFrame(diabetes.data, columns=diabetes.feature_names)
    df["target"] = pd.DataFrame(diabetes.target)
    return df.mask(df["age"] < 0.0).dropna(axis=0)


# %%
# Now we will define a task that can take in any pandas DataFrame
# and return a ``DatasetProfileView``, which is our data profile.
# With it, users can either visualize and check overall statistics
# or even run a constraint suite on top of it.
@task
def create_profile_view(df: pd.DataFrame) -> DatasetProfileView:
    result = why.log(df)
    return result.view()


# %%
# And we will also define a constraints report task
# that will run some checks in our existing profile.
@task
def constraints_report(profile_view: DatasetProfileView) -> bool:
    builder = ConstraintsBuilder(dataset_profile_view=profile_view)
    builder.add_constraint(greater_than_number(column_name="age", number=-11.0))
    builder.add_constraint(smaller_than_number(column_name="bp", number=20.0))
    builder.add_constraint(mean_between_range(column_name="s3", lower=-1.5, upper=1.5))
    builder.add_constraint(null_percentage_below_number(column_name="sex", number=0.0))

    constraints = builder.build()

    renderer = WhylogsConstraintsRenderer()
    flytekit.Deck("constraints", renderer.to_html(constraints=constraints))

    return constraints.validate()


# %%
# This is a representation of a prediction task. Since we are looking
# to take some of the complexity away from our demonstrations,
# our model prediction here will be represented by generating a bunch of
# random numbers with numpy. This task will take place if we pass our
# constraints suite.
@task
def make_predictions(input_data: pd.DataFrame, output_path: str) -> str:
    input_data["predictions"] = np.random.random(size=len(input_data))
    if not os.path.exists(output_path):
        os.makedirs(output_path)
    input_data.to_csv(os.path.join(output_path, "predictions.csv"))
    return f"wrote predictions successfully to {output_path}"


# %%
# Lastly, if the constraint checks fail, we will create a FlyteDeck
# with the Summary Drift Report, which can provide further intuition into
# whether there was a data drift to the failed constraint checks.
@task
def summary_drift_report(new_data: pd.DataFrame, reference_data: pd.DataFrame) -> str:
    renderer = WhylogsSummaryDriftRenderer()
    report = renderer.to_html(target_data=new_data, reference_data=reference_data)
    flytekit.Deck("summary drift", report)
    return f"reported summary drift for target dataset with n={len(new_data)}"


# %%
# Finally, we can then create a Flyte workflow that will
# chain together our example data pipeline
@workflow
def wf() -> str:
    # 1. Read data
    target_df = get_target_data()

    # 2. Profile data and validate it
    profile_view = create_profile_view(df=target_df)
    validated = constraints_report(profile_view=profile_view)

    # 3. Conditional actions if data is valid or not
    return (
        conditional("stop_if_fails")
        .if_(validated.is_false())
        .then(
            summary_drift_report(
                new_data=target_df,
                reference_data=get_reference_data(),
            )
        )
        .else_()
        .then(make_predictions(input_data=target_df, output_path="./data"))
    )


# %%
if __name__ == "__main__":
    wf()
