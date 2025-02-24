import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler

from flytekit import task, workflow


@task
def get_preprocessor() -> StandardScaler:
    return StandardScaler()


@task
def get_model() -> LinearRegression:
    return LinearRegression()


@task
def make_pipeline(preprocessor: StandardScaler, model: LinearRegression) -> Pipeline:
    return Pipeline([("scaler", preprocessor), ("model", model)])


@task
def fit_pipeline(pipeline: Pipeline) -> Pipeline:
    x = np.random.normal(size=(10, 2))
    y = np.random.randint(2, size=(10,))
    pipeline.fit(x, y)
    return pipeline


@task
def num_features(pipeline: Pipeline) -> int:
    return pipeline.n_features_in_


@workflow
def wf():
    preprocessor = get_preprocessor()
    model = get_model()
    pipeline = make_pipeline(preprocessor=preprocessor, model=model)
    pipeline = fit_pipeline(pipeline=pipeline)
    num_features(pipeline=pipeline)


@workflow
def test_wf():
    wf()
