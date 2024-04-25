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
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Flyte Pipeline in One Jupyter Notebook

In this example, we will implement a simple pipeline that takes hyperparameters, does EDA, feature engineering, and measures the Gradient
Boosting model's performance using mean absolute error (MAE), all in one notebook.

+++ {"lines_to_next_cell": 0}

First, let's import the libraries we will use in this example.

```{code-cell}
import os
import pathlib

from flytekit import Resources, kwtypes, workflow
from flytekitplugins.papermill import NotebookTask
```

+++ {"lines_to_next_cell": 0}

We define a `NotebookTask` to run the [Jupyter notebook](https://github.com/flyteorg/flytesnacks/blob/master/examples/exploratory_data_analysis/exploratory_data_analysis/supermarket_regression.ipynb).

```{eval-rst}
.. list-table:: ``NotebookTask`` Parameters
   :widths: 25 25

   * - ``notebook_path``
     - Path to the Jupyter notebook file
   * - ``inputs``
     - Inputs to be sent to the notebook
   * - ``outputs``
     - Outputs to be returned from the notebook
   * - ``requests``
     - Specify compute resource requests for your task.
```

This notebook returns `mae_score` as the output.

```{code-cell}
nb = NotebookTask(
    name="pipeline-nb",
    notebook_path=os.path.join(pathlib.Path(__file__).parent.absolute(), "supermarket_regression.ipynb"),
    inputs=kwtypes(
        n_estimators=int,
        max_depth=int,
        max_features=str,
        min_samples_split=int,
        random_state=int,
    ),
    outputs=kwtypes(mae_score=float),
    requests=Resources(mem="500Mi"),
)
```

Since a task need not be defined, we create a `workflow` and return the MAE score.

```{code-cell}
@workflow
def notebook_wf(
    n_estimators: int = 150,
    max_depth: int = 3,
    max_features: str = "sqrt",
    min_samples_split: int = 4,
    random_state: int = 2,
) -> float:
    output = nb(
        n_estimators=n_estimators,
        max_depth=max_depth,
        max_features=max_features,
        min_samples_split=min_samples_split,
        random_state=random_state,
    )
    return output.mae_score
```

+++ {"lines_to_next_cell": 0}

We can now run the notebook locally.

```{code-cell}
if __name__ == "__main__":
    print(notebook_wf())
```
