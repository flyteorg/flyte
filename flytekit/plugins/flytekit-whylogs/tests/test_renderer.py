from typing import Optional

import numpy as np
import pandas as pd
import whylogs as why
from flytekitplugins.whylogs.renderer import WhylogsConstraintsRenderer, WhylogsSummaryDriftRenderer
from whylogs.core.constraints import ConstraintsBuilder, MetricConstraint, MetricsSelector

import flytekit
from flytekit import task, workflow


@task
def make_data(n_rows: int) -> pd.DataFrame:
    data = {
        "sepal_length": np.random.random_sample(n_rows),
        "sepal_width": np.random.random_sample(n_rows),
        "petal_length": np.random.random_sample(n_rows),
        "petal_width": np.random.random_sample(n_rows),
        "species": np.random.choice(["virginica", "setosa", "versicolor"], n_rows),
        "species_id": np.random.choice([1, 2, 3], n_rows),
    }
    return pd.DataFrame(data)


@task
def run_constraints(df: pd.DataFrame, min_value: Optional[float] = 0.0, max_value: Optional[float] = 4.0) -> bool:
    # This API constraints workflow is very flexible but a bit cumbersome.
    # It will be simplified in the future, so for now we'll stick with injecting
    # a Constraints object to the renderer.
    profile_view = why.log(df).view()
    builder = ConstraintsBuilder(profile_view)
    num_constraint = MetricConstraint(
        name=f"numbers between {min_value} and {max_value} only",
        condition=lambda x: x.min > min_value and x.max < max_value,
        metric_selector=MetricsSelector(metric_name="distribution", column_name="sepal_length"),
    )

    builder.add_constraint(num_constraint)
    constraints = builder.build()

    renderer = WhylogsConstraintsRenderer()
    flytekit.Deck("constraints", renderer.to_html(constraints=constraints))

    return constraints.validate()


@workflow
def whylogs_renderers_workflow(min_value: float, max_value: float) -> bool:
    df = make_data(n_rows=10)
    validated = run_constraints(df=df, min_value=min_value, max_value=max_value)
    return validated


def test_constraints_passing():
    validated = whylogs_renderers_workflow(min_value=0.0, max_value=1.0)
    assert validated is True


def test_constraints_failing():
    validated = whylogs_renderers_workflow(min_value=-1.0, max_value=0.0)
    assert validated is False


def test_summary_drift_report_is_written():
    renderer = WhylogsSummaryDriftRenderer()
    new_data = make_data(n_rows=10)
    reference_data = make_data(n_rows=100)

    report = renderer.to_html(target_data=new_data, reference_data=reference_data)
    assert report is not None
    assert isinstance(report, str)
    assert "Profile Summary" in report
