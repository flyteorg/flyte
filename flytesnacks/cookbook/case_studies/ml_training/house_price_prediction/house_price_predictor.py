"""
.. _single_region_house_prediction:

Predicting House Price in a Region Using XGBoost
------------------------------------------------

`XGBoost <https://xgboost.readthedocs.io/en/latest/>`__ is an optimized distributed gradient boosting library designed to be efficient, flexible, and portable.
It uses `gradient boosting <https://en.wikipedia.org/wiki/Gradient_boosting>`__ technique to implement Machine Learning algorithms.

In this tutorial, we will understand how to predict house prices using XGBoost, and Flyte.

We will split the generated dataset into train, test and validation set.

Next, we will create three Flyte tasks, that will:

1. Generate house details, and split the dataset.
2. Train the model using XGBoost.
3. Generate predictions.


Let's get started with the example!

"""

# %%
# Install the following libraries before running the model (locally):
#
# .. code-block:: python
#
#       pip install scikit-learn
#       pip install joblib
#       pip install xgboost

import os

# %%
# First, let's import the required packages into the environment.
import typing
from typing import Tuple

import flytekit
import joblib
import numpy as np
import pandas as pd
from flytekit import Resources, task, workflow
from flytekit.types.file import JoblibSerializedFile
from sklearn.model_selection import train_test_split
from xgboost import XGBRegressor

# %%
# We initialize a variable to represent columns in the dataset. The other variables help generate the dataset.
NUM_HOUSES_PER_LOCATION = 1000
COLUMNS = [
    "PRICE",
    "YEAR_BUILT",
    "SQUARE_FEET",
    "NUM_BEDROOMS",
    "NUM_BATHROOMS",
    "LOT_ACRES",
    "GARAGE_SPACES",
]
MAX_YEAR = 2021
# divide the data into train, validation, and test datasets in specific ratio.
SPLIT_RATIOS = [0.6, 0.3, 0.1]

# %%
# Data Generation
# =====================
#
# We define a function to compute the price of a house based on multiple factors (``number of bedrooms``, ``number of bathrooms``, ``area``, ``garage space``, and ``year built``).


def gen_price(house) -> int:
    _base_price = int(house["SQUARE_FEET"] * 150)
    _price = int(
        _base_price
        + (10000 * house["NUM_BEDROOMS"])
        + (15000 * house["NUM_BATHROOMS"])
        + (15000 * house["LOT_ACRES"])
        + (15000 * house["GARAGE_SPACES"])
        - (5000 * (MAX_YEAR - house["YEAR_BUILT"]))
    )
    return _price


# %%
# Next, using the above function, we generate a DataFrame object that constitutes all the house details.
def gen_houses(num_houses) -> pd.DataFrame:
    _house_list = []
    for _ in range(num_houses):
        _house = {
            "SQUARE_FEET": int(np.random.normal(3000, 750)),
            "NUM_BEDROOMS": np.random.randint(2, 7),
            "NUM_BATHROOMS": np.random.randint(2, 7) / 2,
            "LOT_ACRES": round(np.random.normal(1.0, 0.25), 2),
            "GARAGE_SPACES": np.random.randint(0, 4),
            "YEAR_BUILT": min(MAX_YEAR, int(np.random.normal(1995, 10))),
        }
        _price = gen_price(_house)
        # column names/features
        _house_list.append(
            [
                _price,
                _house["YEAR_BUILT"],
                _house["SQUARE_FEET"],
                _house["NUM_BEDROOMS"],
                _house["NUM_BATHROOMS"],
                _house["LOT_ACRES"],
                _house["GARAGE_SPACES"],
            ]
        )
    # convert the list to a DataFrame
    _df = pd.DataFrame(
        _house_list,
        columns=COLUMNS,
    )
    return _df


# %%
# Data Preprocessing and Splitting
# ===================================
#
# We split the data into train, test, and validation subsets.
def split_data(
    df: pd.DataFrame, seed: int, split: typing.List[float]
) -> Tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:

    seed = seed
    val_size = split[1]  # 0.3
    test_size = split[2]  # 0.1

    num_samples = df.shape[0]
    # retain the features, skip the target column
    x1 = df.values[:num_samples, 1:]
    # retain the target column
    y1 = df.values[:num_samples, :1]

    # divide the features and target column into random train and test subsets, based on `test_size`
    x_train, x_test, y_train, y_test = train_test_split(
        x1, y1, test_size=test_size, random_state=seed
    )
    # divide the train data into train and validation subsets, based on `test_size`
    x_train, x_val, y_train, y_val = train_test_split(
        x_train,
        y_train,
        test_size=(val_size / (1 - test_size)),  # here, `test_size` computes to 0.3
        random_state=seed,
    )

    # reassemble the datasets by placing `target` as first column and `features` in the subsequent columns
    _train = np.concatenate([y_train, x_train], axis=1)
    _val = np.concatenate([y_val, x_val], axis=1)
    _test = np.concatenate([y_test, x_test], axis=1)

    # return three DataFrames with train, test, and validation data
    return (
        pd.DataFrame(
            _train,
            columns=COLUMNS,
        ),
        pd.DataFrame(
            _val,
            columns=COLUMNS,
        ),
        pd.DataFrame(
            _test,
            columns=COLUMNS,
        ),
    )


# %%
# Next, we create a ``NamedTuple`` to map a variable name to its respective data type.
dataset = typing.NamedTuple(
    "GenerateSplitDataOutputs",
    train_data=pd.DataFrame,
    val_data=pd.DataFrame,
    test_data=pd.DataFrame,
)

# %%
# We define a task to call the aforementioned functions.


@task(cache=True, cache_version="0.1", limits=Resources(mem="600Mi"))
def generate_and_split_data(number_of_houses: int, seed: int) -> dataset:
    _houses = gen_houses(number_of_houses)
    return split_data(_houses, seed, split=SPLIT_RATIOS)


# %%
# Training
# ==========
#
# We fit an ``XGBRegressor`` model on our data, serialize the model using `joblib`, and return a :py:obj:`~flytekit:flytekit.types.file.JoblibSerializedFile`.
@task(cache_version="1.0", cache=True, limits=Resources(mem="600Mi"))
def fit(loc: str, train: pd.DataFrame, val: pd.DataFrame) -> JoblibSerializedFile:

    # fetch the features and target columns from the train dataset
    x = train[train.columns[1:]]
    y = train[train.columns[0]]

    # fetch the features and target columns from the validation dataset
    eval_x = val[val.columns[1:]]
    eval_y = val[val.columns[0]]

    m = XGBRegressor()
    # fit the model to the train data
    m.fit(x, y, eval_set=[(eval_x, eval_y)])

    working_dir = flytekit.current_context().working_directory
    fname = os.path.join(working_dir, f"model-{loc}.joblib.dat")
    joblib.dump(m, fname)

    # return the serialized model
    return JoblibSerializedFile(path=fname)


# %%
# Generating Predictions
# ========================
#
# Next, we unserialize the XGBoost model using `joblib` to generate the predictions.
@task(cache_version="1.0", cache=True, limits=Resources(mem="600Mi"))
def predict(
    test: pd.DataFrame,
    model_ser: JoblibSerializedFile,
) -> typing.List[float]:

    # load the model
    model = joblib.load(model_ser)

    # load the test data
    x_df = test[test.columns[1:]]

    # generate predictions
    y_pred = model.predict(x_df).tolist()

    # return the predictions
    return y_pred


# %%
# Lastly, we define a workflow to run the pipeline.
@workflow
def house_price_predictor_trainer(
    seed: int = 7, number_of_houses: int = NUM_HOUSES_PER_LOCATION
) -> typing.List[float]:

    # generate the data and split it into train test, and validation data
    split_data_vals = generate_and_split_data(
        number_of_houses=number_of_houses, seed=seed
    )

    # fit the XGBoost model
    model = fit(
        loc="NewYork_NY", train=split_data_vals.train_data, val=split_data_vals.val_data
    )

    # generate predictions
    predictions = predict(model_ser=model, test=split_data_vals.test_data)

    return predictions


# %%
# Running the Model Locally
# ==========================
#
# We can run the workflow locally provided the required libraries are installed. The output would be a list of house prices, generated using the XGBoost model.
if __name__ == "__main__":
    print(house_price_predictor_trainer())
