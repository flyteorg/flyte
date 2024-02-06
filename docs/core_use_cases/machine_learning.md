---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_machine_learning)=

# Machine Learning

Flyte can handle a full spectrum of machine learning workloads, from
training small models to gpu-accelerated deep learning and hyperparameter
optimization.

## Getting the Data

In this simple example, we train a binary classification model on the
[wine dataset](https://scikit-learn.org/stable/datasets/toy_dataset.html#wine-dataset)
that is available through the `scikit-learn` package:

```{code-cell} ipython3
import pandas as pd
from flytekit import Resources, task, workflow
from sklearn.datasets import load_wine
from sklearn.linear_model import LogisticRegression

import flytekit.extras.sklearn


@task(requests=Resources(mem="500Mi"))
def get_data() -> pd.DataFrame:
    """Get the wine dataset."""
    return load_wine(as_frame=True).frame
```

## Define a Training Workflow

Then, we define `process_data` and `train_model` tasks along with a
`training_workflow` to put all the pieces together for a model-training
pipeline.

```{code-cell} ipython3
@task
def process_data(data: pd.DataFrame) -> pd.DataFrame:
    """Simplify the task from a 3-class to a binary classification problem."""
    return data.assign(target=lambda x: x["target"].where(x["target"] == 0, 1))


@task
def train_model(data: pd.DataFrame, hyperparameters: dict) -> LogisticRegression:
    """Train a model on the wine dataset."""
    features = data.drop("target", axis="columns")
    target = data["target"]
    return LogisticRegression(max_iter=5000, **hyperparameters).fit(features, target)


@workflow
def training_workflow(hyperparameters: dict) -> LogisticRegression:
    """Put all of the steps together into a single workflow."""
    data = get_data()
    processed_data = process_data(data=data)
    return train_model(
        data=processed_data,
        hyperparameters=hyperparameters,
    )

```

```{important}
Even though you can use a `dict` type to represent the model's hyperparameters,
we recommend using {ref}`dataclasses <dataclass>` to define a custom
`Hyperparameter` Python object that provides more type information to the Flyte
compiler. For example, Flyte uses this type information to auto-generate
type-safe launch forms on the Flyte UI. Learn more in the
{ref}`Extending Flyte <customizing_flyte_types>` guide.
```

## Computing Predictions

Executing this workflow locally, we can call the `model.predict` method to make
sure we can use our newly trained model to make predictions based on some
feature matrix.

```{code-cell} ipython3
model = training_workflow(hyperparameters={"C": 0.01})
X, _ = load_wine(as_frame=True, return_X_y=True)
model.predict(X.sample(10, random_state=41))
```

## Extending your ML Workloads

There are many ways to extend your workloads:

```{list-table}
:header-rows: 0
:widths: 20 30

* - **🏔 Vertical Scaling**
  - Use the {py:class}`~flytekit.Resources` task keyword argument to request
    additional CPUs, GPUs, and/or memory.
* - **🗺 Horizontal Scaling**
  - With constructs like {py:func}`~flytekit.dynamic` workflows and
    {py:func}`~flytekit.map_task`s, implement gridsearch, random search,
    and even [bayesian optimization](https://github.com/flyteorg/flytekit-python-template/tree/main/bayesian-optimization/%7B%7Bcookiecutter.project_name%7D%7D).
* - **🔧 Specialized Tuning Libraries**
  - Use the {ref}`Ray Integration <kube-ray-op>` and leverage tools like
    [Ray Tune](https://docs.ray.io/en/latest/tune/index.html) for hyperparameter
    optimization, all orchestrated by Flyte as ephemerally-provisioned Ray clusters.
* - **📦 Ephemeral Cluster Resources**
  - Use the {ref}`MPI Operator <kf-mpi-op>`, {ref}`Sagemaker <aws-sagemaker>`,
    {ref}`Kubeflow Tensorflow <kftensorflow-plugin>`, {ref}`Kubeflow Pytorch<kf-pytorch-op>`
    and {doc}`more <_tags/DistributedComputing>` to do distributed training.
* - **🔎 Experiment Tracking**
  - Auto-capture training logs with the {py:func}`~flytekitplugins.mlflow.mlflow_autolog`
    decorator, which can be viewed as Flyte Decks with `@task(disable_decks=False)`.
* - **⏩ Inference Acceleration**
  - Serialize your models in ONNX format using the {ref}`ONNX plugin <onnx>`, which
    supports ScikitLearn, TensorFlow, and PyTorch.
```

```{admonition} Learn More
:class: important

See the {ref}`Tutorials <tutorials>` for more machine learning examples.
```
