---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

# KNN Classifier

In this example, let's understand how effortlessly the Modin DataFrames can be used with tasks and workflows in a simple classification pipeline.
Modin uses [Ray](https://github.com/ray-project/ray/) or [Dask](https://dask.org/) as the compute engine. We will use Ray in this example.

To install Modin with Ray as the backend,

```bash
pip install modin[ray]
```

:::{note}
To install Modin with Dask as the backend,

```bash
pip install modin[dask]
```
:::

Let's dive right in!

+++ {"lines_to_next_cell": 0}

Let's import the necessary dependencies.

```{code-cell}
from typing import List, NamedTuple

import flytekitplugins.modin  # noqa: F401
import modin.pandas
import ray
from flytekit import task, workflow
from sklearn.datasets import load_wine
from sklearn.metrics import accuracy_score
from sklearn.model_selection import train_test_split
from sklearn.neighbors import KNeighborsClassifier

ray.shutdown()  # close previous instance of ray (if any)
ray.init(num_cpus=2)  # open a new instance of ray


split_data = NamedTuple(
    "split_data",
    train_features=modin.pandas.DataFrame,
    test_features=modin.pandas.DataFrame,
    train_labels=modin.pandas.DataFrame,
    test_labels=modin.pandas.DataFrame,
)
```

+++ {"lines_to_next_cell": 0}

We define a task that processes the wine dataset after loading it into the environment.

```{code-cell}
@task
def preprocess_data() -> split_data:
    wine = load_wine(as_frame=True)

    # convert features and target (numpy arrays) into Modin DataFrames
    wine_features = modin.pandas.DataFrame(data=wine.data, columns=wine.feature_names)
    wine_target = modin.pandas.DataFrame(data=wine.target, columns=["target"])

    # split the dataset
    X_train, X_test, y_train, y_test = train_test_split(wine_features, wine_target, test_size=0.33, random_state=101)

    return split_data(
        train_features=X_train,
        test_features=X_test,
        train_labels=y_train,
        test_labels=y_test,
    )
```

+++ {"lines_to_next_cell": 0}

Next, we define a task that:

1. trains a KNeighborsClassifier model,
2. fits the model to the data, and
3. predicts the output for the test dataset.

```{code-cell}
@task
def fit_and_predict(
    X_train: modin.pandas.DataFrame,
    X_test: modin.pandas.DataFrame,
    y_train: modin.pandas.DataFrame,
) -> List[int]:
    lr = KNeighborsClassifier()  # create a KNeighborsClassifier model
    lr.fit(X_train, y_train)  # fit the model to the data
    predicted_vals = lr.predict(X_test)  # predict values for test data
    return predicted_vals.tolist()
```

+++ {"lines_to_next_cell": 0}

We compute accuracy of the model.

```{code-cell}
@task
def calc_accuracy(y_test: modin.pandas.DataFrame, predicted_vals_list: List[int]) -> float:
    return accuracy_score(y_test, predicted_vals_list)
```

+++ {"lines_to_next_cell": 0}

Lastly, we define a workflow.

```{code-cell}
@workflow
def pipeline() -> float:
    split_data_vals = preprocess_data()
    predicted_vals_output = fit_and_predict(
        X_train=split_data_vals.train_features,
        X_test=split_data_vals.test_features,
        y_train=split_data_vals.train_labels,
    )
    return calc_accuracy(y_test=split_data_vals.test_labels, predicted_vals_list=predicted_vals_output)


if __name__ == "__main__":
    print(f"Accuracy of the model is {pipeline()}%")
```
