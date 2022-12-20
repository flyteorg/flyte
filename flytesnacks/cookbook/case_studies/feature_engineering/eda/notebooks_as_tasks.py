"""
EDA and Feature Engineering in One Jupyter Notebook and Modeling in the Other
=============================================================================

In this example, we will implement a simple pipeline that takes hyperparameters, does EDA, feature engineering
(step 1: EDA and feature engineering in notebook), and measures the Gradient Boosting model's performance using mean absolute error
(MAE) (step 2: Modeling in notebook).
"""

# %%
# First, let's import the libraries we will use in this example.
import os
import pathlib

import pandas as pd
from flytekit import Resources, kwtypes, workflow
from flytekitplugins.papermill import NotebookTask

# %%
# We define a ``NotebookTask`` to run the `Jupyter notebook
# <https://github.com/flyteorg/flytesnacks/blob/master/cookbook/case_studies/feature_engineering/eda/supermarket_regression_1.ipynb>`__ (EDA).
# This notebook returns ``dummified_data`` and ``dataset`` as the outputs.
#
# .. note::
#   ``dataset`` is used in this example, and ``dummified_data`` is used in the previous example.
#   ``dataset`` lets us send the DataFrame as a JSON string to the subsequent notebook because DataFrame input cannot be sent
#   directly to the notebook as per Papermill.
nb_1 = NotebookTask(
    name="eda-featureeng-nb",
    notebook_path=os.path.join(
        pathlib.Path(__file__).parent.absolute(), "supermarket_regression_1.ipynb"
    ),
    outputs=kwtypes(dummified_data=pd.DataFrame, dataset=str),
    requests=Resources(mem="500Mi"),
)

# %%
# We define a ``NotebookTask`` to run the `Jupyter notebook
# <https://github.com/flyteorg/flytesnacks/blob/master/cookbook/case_studies/feature_engineering/eda/supermarket_regression_2.ipynb>`__
# (Modeling).
#
# This notebook returns ``mae_score`` as the output.
nb_2 = NotebookTask(
    name="regression-nb",
    notebook_path=os.path.join(
        pathlib.Path(__file__).parent.absolute(),
        "supermarket_regression_2.ipynb",
    ),
    inputs=kwtypes(
        dataset=str,
        n_estimators=int,
        max_depth=int,
        max_features=str,
        min_samples_split=int,
        random_state=int,
    ),
    outputs=kwtypes(mae_score=float),
    requests=Resources(mem="500Mi"),
)

# %%
# We define a ``Workflow`` to run the notebook tasks.


@workflow
def notebook_wf(
    n_estimators: int = 150,
    max_depth: int = 3,
    max_features: str = "sqrt",
    min_samples_split: int = 4,
    random_state: int = 2,
) -> float:
    eda_output = nb_1()
    regression_output = nb_2(
        dataset=eda_output.dataset,
        n_estimators=n_estimators,
        max_depth=max_depth,
        max_features=max_features,
        min_samples_split=min_samples_split,
        random_state=random_state,
    )
    return regression_output.mae_score


# %%
# We can now run the two notebooks locally.
if __name__ == "__main__":
    print(notebook_wf())
