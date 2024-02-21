import typing
from functools import partial, wraps

import flytekit
from flytekit import FlyteContextManager, lazy_module
from flytekit.bin.entrypoint import get_one_of
from flytekit.deck.renderer import TopFrameRenderer

go = lazy_module("plotly.graph_objects")
plotly_subplots = lazy_module("plotly.subplots")
pd = lazy_module("pandas")
mlflow = lazy_module("mlflow")


def metric_to_df(metrics: typing.List[mlflow.entities.metric.Metric]) -> pd.DataFrame:
    """
    Converts mlflow Metric object to a dataframe of 2 columns ['timestamp', 'value']
    """
    t = []
    v = []
    for m in metrics:
        t.append(m.timestamp)
        v.append(m.value)
    return pd.DataFrame(list(zip(t, v)), columns=["timestamp", "value"])


def get_run_metrics(c: mlflow.MlflowClient, run_id: str) -> typing.Dict[str, pd.DataFrame]:
    """
    Extracts all metrics and returns a dictionary of metric name to the list of metric for the given run_id
    """
    r = c.get_run(run_id)
    metrics = {}
    for k in r.data.metrics.keys():
        metrics[k] = metric_to_df(metrics=c.get_metric_history(run_id=run_id, key=k))
    return metrics


def get_run_params(c: mlflow.MlflowClient, run_id: str) -> typing.Optional[pd.DataFrame]:
    """
    Extracts all parameters and returns a dictionary of metric name to the list of metric for the given run_id
    """
    r = c.get_run(run_id)
    name = []
    value = []
    if r.data.params == {}:
        return None
    for k, v in r.data.params.items():
        name.append(k)
        value.append(v)
    return pd.DataFrame(list(zip(name, value)), columns=["name", "value"])


def plot_metrics(metrics: typing.Dict[str, pd.DataFrame]) -> typing.Optional[go.Figure]:
    v = len(metrics)
    if v == 0:
        return None

    # Initialize figure with subplots
    fig = plotly_subplots.make_subplots(rows=v, cols=1, subplot_titles=list(metrics.keys()))

    # Add traces
    row = 1
    for k, v in metrics.items():
        v["timestamp"] = (v["timestamp"] - v["timestamp"][0]) / 1000
        fig.add_trace(go.Scatter(x=v["timestamp"], y=v["value"], name=k), row=row, col=1)
        row = row + 1

    fig.update_xaxes(title_text="Time (s)")
    fig.update_layout(height=700, width=900)
    return fig


def mlflow_autolog(fn=None, *, framework=mlflow.sklearn, experiment_name: typing.Optional[str] = None):
    """MLFlow decorator to enable autologging of training metrics.

    This decorator can be used as a nested decorator for a ``@task`` and it will automatically enable mlflow autologging,
    for the given ``framework``. By default autologging is enabled for ``sklearn``.

    .. code-block:: python

        @task
        @mlflow_autolog(framework=mlflow.tensorflow)
        def my_tensorflow_trainer():
            ...

    One benefit of doing so is that the mlflow metrics are then rendered inline using FlyteDecks and can be viewed
    in jupyter notebook, as well as in hosted Flyte environment:

    .. code-block:: python

        # jupyter notebook cell
        with flytekit.new_context() as ctx:
            my_tensorflow_trainer()
            ctx.get_deck()  # IPython.display

    When the task is called in a Flyte backend, the decorator starts a new MLFlow run using the Flyte execution name
    by default, or a user-provided ``experiment_name`` in the decorator.

    :param fn: Function to generate autologs for.
    :param framework: The mlflow module to use for autologging
    :param experiment_name: The MLFlow experiment name. If not provided, uses the Flyte execution name.
    """

    @wraps(fn)
    def wrapper(*args, **kwargs):
        framework.autolog()
        params = FlyteContextManager.current_context().user_space_params
        ctx = FlyteContextManager.current_context()

        experiment = experiment_name or "local workflow"
        run_name = None  # MLflow will generate random name if value is None

        if not ctx.execution_state.is_local_execution():
            experiment = f"{get_one_of('FLYTE_INTERNAL_EXECUTION_WORKFLOW', '_F_WF')}" or experiment_name
            run_name = f"{params.execution_id.name}.{params.task_id.name.split('.')[-1]}"

        mlflow.set_experiment(experiment)
        with mlflow.start_run(run_name=run_name):
            out = fn(*args, **kwargs)
            run = mlflow.active_run()
            if run is not None:
                client = mlflow.MlflowClient()
                run_id = run.info.run_id
                metrics = get_run_metrics(client, run_id)
                figure = plot_metrics(metrics)
                if figure:
                    flytekit.Deck("mlflow metrics", figure.to_html())
                params = get_run_params(client, run_id)
                if params is not None:
                    flytekit.Deck("mlflow params", TopFrameRenderer(max_rows=10).to_html(params))
        return out

    if fn is None:
        return partial(mlflow_autolog, framework=framework, experiment_name=experiment_name)

    return wrapper
