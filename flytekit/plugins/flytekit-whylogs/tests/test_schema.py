from datetime import datetime
from typing import Type

import pandas as pd
import whylogs as why
from flytekitplugins.whylogs.schema import WhylogsDatasetProfileTransformer
from whylogs.core import DatasetProfileView

from flytekit import task, workflow
from flytekit.core.context_manager import FlyteContextManager


@task
def whylogs_profiling() -> DatasetProfileView:
    df = pd.DataFrame({"a": [1, 2, 3, 4]})
    result = why.log(pandas=df)
    return result.view()


@task
def fetch_whylogs_datetime(profile_view: DatasetProfileView) -> datetime:
    return profile_view.dataset_timestamp


@workflow
def whylogs_wf() -> datetime:
    profile_view = whylogs_profiling()
    return fetch_whylogs_datetime(profile_view=profile_view)


def test_task_returns_whylogs_profile_view() -> None:
    actual_profile = whylogs_profiling()
    assert actual_profile is not None
    assert isinstance(actual_profile, DatasetProfileView)


def test_profile_view_gets_passed_on_tasks() -> None:
    result = whylogs_wf()
    assert result is not None
    assert isinstance(result, datetime)


def test_to_html_method() -> None:
    tf = WhylogsDatasetProfileTransformer()
    profile_view = whylogs_profiling()
    report = tf.to_html(FlyteContextManager.current_context(), profile_view, Type[DatasetProfileView])

    assert isinstance(report, str)
    assert "Profile Visualizer" in report
