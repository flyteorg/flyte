"""
Data Cleaning Tasks
-------------------------

We'll define the relevant feature engineering tasks to clean up the SQLite3 data.
"""

# %%
# Firstly, let's import the required libraries.
import numpy as np
import pandas as pd
from flytekit import task
from numpy.core.fromnumeric import sort
from sklearn.feature_selection import SelectKBest, f_classif
from sklearn.impute import SimpleImputer

# %%
# Next, we define a ``mean_median_imputer`` task to fill in the missing values of the dataset, for which we use `SimpleImputer <https://scikit-learn.org/stable/modules/generated/sklearn.impute.SimpleImputer.html>`__ class picked from the ``scikit-learn`` library.
@task
def mean_median_imputer(
    dataframe: pd.DataFrame,
    imputation_method: str,
) -> pd.DataFrame:

    dataframe = dataframe.replace("?", np.nan)
    if imputation_method not in ["median", "mean"]:
        raise ValueError("imputation_method takes only values 'median' or 'mean'")

    imputer = SimpleImputer(missing_values=np.nan, strategy=imputation_method)

    imputer = imputer.fit(dataframe)
    dataframe[:] = imputer.transform(dataframe)

    return dataframe

# %%
# This task returns the filled-in dataframe.

# %%
# Let's define one other task called ``univariate_selection`` that does feature selection. 
# The `SelectKBest <https://scikit-learn.org/stable/modules/generated/sklearn.feature_selection.SelectKBest.html#sklearn.feature_selection.SelectKBest>`__ method removes all but the highest scoring features.

@task
def univariate_selection(
    dataframe: pd.DataFrame, split_mask: int, num_features: int
) -> pd.DataFrame:

    X = dataframe.iloc[:, 0:split_mask]
    y = dataframe.iloc[:, split_mask]
    test = SelectKBest(score_func=f_classif, k=num_features)
    fit = test.fit(X, y)
    indices = sort((-fit.scores_).argsort()[:num_features])
    column_names = map(dataframe.columns.__getitem__, indices)
    features = fit.transform(X)
    return pd.DataFrame(features, columns=column_names)

# %%
# This task returns a dataframe with the specified number of columns.
