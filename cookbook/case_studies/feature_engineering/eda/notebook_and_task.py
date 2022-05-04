"""
EDA and Feature Engineering in Jupyter Notebook and Modeling in a Flyte Task
============================================================================

In this example, we will implement a simple pipeline that takes hyperparameters, does EDA, feature engineering
(step 1: EDA and feature engineering in notebook), and measures the Gradient Boosting model's performance using mean absolute error (MAE)
(step 2: Modeling in a Flyte Task).
"""

# %%
# First, let's import the libraries we will use in this example.
import os
import pathlib
from dataclasses import dataclass

import numpy as np
import pandas as pd
from dataclasses_json import dataclass_json
from flytekit import Resources, kwtypes, task, workflow
from flytekitplugins.papermill import NotebookTask
from sklearn.ensemble import GradientBoostingRegressor
from sklearn.model_selection import cross_val_score, train_test_split
from sklearn.preprocessing import RobustScaler


# %%
# We define a ``dataclass`` to store the hyperparameters of the Gradient Boosting Regressor.
@dataclass_json
@dataclass
class Hyperparameters(object):
    n_estimators: int = 150
    max_depth: int = 3
    max_features: str = "sqrt"
    min_samples_split: int = 4
    random_state: int = 2
    nfolds: int = 10


# %%
# We define a ``NotebookTask`` to run the `Jupyter notebook
# <https://github.com/flyteorg/flytesnacks/blob/master/cookbook/case_studies/feature_engineering/eda/supermarket_regression_1.ipynb>`__.
# This notebook returns ``dummified_data`` and ``dataset`` as the outputs.
#
# .. note::
#   ``dummified_data`` is used in this example, and ``dataset`` is used in the upcoming example.
nb = NotebookTask(
    name="eda-feature-eng-nb",
    notebook_path=os.path.join(
        pathlib.Path(__file__).parent.absolute(), "supermarket_regression_1.ipynb"
    ),
    outputs=kwtypes(dummified_data=pd.DataFrame, dataset=str),
    requests=Resources(mem="500Mi"),
)

# %%
# Next, we define a ``cross_validate`` function and a ``modeling`` task to compute the MAE score of the data against
# the Gradient Boosting Regressor.


def cross_validate(model, nfolds, feats, targets):
    score = -1 * (
        cross_val_score(
            model, feats, targets, cv=nfolds, scoring="neg_mean_absolute_error"
        )
    )
    return np.mean(score)


@task
def modeling(
    dataset: pd.DataFrame,
    hyperparams: Hyperparameters,
) -> float:
    y_target = dataset["Product_Supermarket_Sales"].tolist()
    dataset.drop(["Product_Supermarket_Sales"], axis=1, inplace=True)

    X_train, X_test, y_train, _ = train_test_split(dataset, y_target, test_size=0.3)

    scaler = RobustScaler()

    scaler.fit(X_train)

    X_train = scaler.transform(X_train)
    X_test = scaler.transform(X_test)

    gb_model = GradientBoostingRegressor(
        n_estimators=hyperparams.n_estimators,
        max_depth=hyperparams.max_depth,
        max_features=hyperparams.max_features,
        min_samples_split=hyperparams.min_samples_split,
        random_state=hyperparams.random_state,
    )

    return cross_validate(gb_model, hyperparams.nfolds, X_train, y_train)


# %%
# We define a ``workflow`` to run the notebook and the ``modeling`` task.
@workflow
def notebook_wf(hyperparams: Hyperparameters = Hyperparameters()) -> float:
    output = nb()
    mae_score = modeling(dataset=output.dummified_data, hyperparams=hyperparams)
    return mae_score


# %%
# We can now run the notebook and the modeling task locally.
if __name__ == "__main__":
    print(notebook_wf())
