"""
Feature Engineering Tasks
-------------------------

Let's define some feature engineering tasks to be used in conjunction with the Flyte workflow.
"""

# %%
# Import the necessary libraries.
import numpy as np
import pandas as pd
from flytekit import task
from numpy.core.fromnumeric import sort
from sklearn.feature_selection import SelectKBest, f_classif
from sklearn.impute import SimpleImputer

# %%
# There are a specific set of columns for which imputation isn't required. Ignore them.
NO_IMPUTATION_COLS = [
    "Hospital Number",
    "surgery",
    "Age",
    "outcome",
    "surgical lesion",
    "timestamp",
]

# %%
# Use the `SimpleImputer <https://scikit-learn.org/stable/modules/generated/sklearn.impute.SimpleImputer.html>`__ class from the ``scikit-learn`` library
# to fill in the missing values of the dataset.
@task
def mean_median_imputer(
    dataframe: pd.DataFrame,
    imputation_method: str,
) -> pd.DataFrame:
    dataframe = dataframe.replace("?", np.nan)
    if imputation_method not in ["median", "mean"]:
        raise ValueError("imputation_method takes only values 'median' or 'mean'")

    imputer = SimpleImputer(missing_values=np.nan, strategy=imputation_method)

    imputer = imputer.fit(
        dataframe[dataframe.columns[~dataframe.columns.isin(NO_IMPUTATION_COLS)]]
    )
    dataframe[
        dataframe.columns[~dataframe.columns.isin(NO_IMPUTATION_COLS)]
    ] = imputer.transform(
        dataframe[dataframe.columns[~dataframe.columns.isin(NO_IMPUTATION_COLS)]]
    )
    return dataframe


# %%
# The `SelectKBest <https://scikit-learn.org/stable/modules/generated/sklearn.feature_selection.SelectKBest.html#sklearn.feature_selection.SelectKBest>`__ method
# removes all but the highest scoring features.
@task
def univariate_selection(
    dataframe: pd.DataFrame, num_features: int, data_class: str
) -> pd.DataFrame:
    # remove ``timestamp`` and ``Hospital Number`` columns as they ought to be present in the dataset
    dataframe = dataframe.drop(["event_timestamp", "Hospital Number"], axis=1)

    if num_features > 9:
        raise ValueError(
            f"Number of features must be <= 9; you've given {num_features}"
        )

    X = dataframe.iloc[:, dataframe.columns != data_class]
    y = dataframe.loc[:, data_class]
    test = SelectKBest(score_func=f_classif, k=num_features)
    fit = test.fit(X, y)
    indices = sort((-fit.scores_).argsort()[:num_features])
    column_names = list(map(X.columns.__getitem__, indices))
    column_names.extend([data_class])
    features = fit.transform(X)
    return pd.DataFrame(np.c_[features, y.to_numpy()], columns=column_names)
