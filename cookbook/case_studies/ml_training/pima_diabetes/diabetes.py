"""
Train and Validate a Diabetes Classification XGBoost Model
----------------------------------------------------------

Watch a demo of sandbox creation and a sample execution of the pima diabetes pipeline below.

..  youtube:: YEvs0MHXZnY

"""
import typing
from collections import OrderedDict
from dataclasses import dataclass
from typing import Tuple

import joblib
import pandas as pd
from dataclasses_json import dataclass_json
from flytekit import Resources, task, workflow
from flytekit.types.file import FlyteFile
from flytekit.types.schema import FlyteSchema
from sklearn.metrics import accuracy_score
from sklearn.model_selection import train_test_split
from xgboost import XGBClassifier

# %%
# Since we are working with a specific dataset, we will create a strictly typed schema for the dataset.
# If we wanted a generic data splitter we could use a Generic schema without any column type and name information
# `Example file <https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv>`_
# CSV Columns
#
# #. Number of times pregnant
# #. Plasma glucose concentration a 2 hours in an oral glucose tolerance test
# #. Diastolic blood pressure (mm Hg)
# #. Triceps skin fold thickness (mm)
# #. 2-Hour serum insulin (mu U/ml)
# #. Body mass index (weight in kg/(height in m)^2)
# #. Diabetes pedigree function
# #. Age (years)
# #. Class variable (0 or 1)
#
# Example Row: 6,148,72,35,0,33.6,0.627,50,1
# the input dataset schema
DATASET_COLUMNS = OrderedDict(
    {
        "#preg": int,
        "pgc_2h": int,
        "diastolic_bp": int,
        "tricep_skin_fold_mm": int,
        "serum_insulin_2h": int,
        "bmi": float,
        "diabetes_pedigree": float,
        "age": int,
        "class": int,
    }
)
# %%
# The first 8 columns are features
FEATURE_COLUMNS = OrderedDict(
    {k: v for k, v in DATASET_COLUMNS.items() if k != "class"}
)
# %%
# The last column is the class
CLASSES_COLUMNS = OrderedDict({"class": int})


# %%
# Let us declare a task that accepts a CSV file with the previously defined
# columns and converts it to a typed schema.
# An example CSV file is available at
# `here <https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv>`__
@task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
def split_traintest_dataset(
    dataset: FlyteFile[typing.TypeVar("csv")], seed: int, test_split_ratio: float
) -> Tuple[
    FlyteSchema[FEATURE_COLUMNS],
    FlyteSchema[FEATURE_COLUMNS],
    FlyteSchema[CLASSES_COLUMNS],
    FlyteSchema[CLASSES_COLUMNS],
]:
    """
    Retrieves the training dataset from the given blob location and then splits it using the split ratio and returns the result
    This splitter is only for the dataset that has the format as specified in the example csv. The last column is assumed to be
    the class and all other columns 0-8 the features.

    The data is returned as a schema, which gets converted to a parquet file in the back.
    """
    column_names = [k for k in DATASET_COLUMNS.keys()]
    df = pd.read_csv(dataset, names=column_names)

    # Select all features
    x = df[column_names[:8]]
    # Select only the classes
    y = df[[column_names[-1]]]

    # split data into train and test sets
    return train_test_split(x, y, test_size=test_split_ratio, random_state=seed)


# %%
# It is also possible to defined the output file type. This is useful in
# combining tasks, where one task may only accept models serialized in ``.joblib.dat``
MODELSER_JOBLIB = typing.TypeVar("joblib.dat")


# %%
# It is also possible in Flyte to pass custom objects, as long as they are
# declared as ``dataclass``es and also decorated with ``@dataclass_json``.
@dataclass_json
@dataclass
class XGBoostModelHyperparams(object):
    """
    These are the xgboost hyper parameters available in scikit-learn library.
    """

    max_depth: int = 3
    learning_rate: float = 0.1
    n_estimators: int = 100
    objective: str = "binary:logistic"
    booster: str = "gbtree"
    n_jobs: int = 1


model_file = typing.NamedTuple("Model", model=FlyteFile[MODELSER_JOBLIB])
workflow_outputs = typing.NamedTuple(
    "WorkflowOutputs", model=FlyteFile[MODELSER_JOBLIB], accuracy=float
)


@task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
def fit(
    x: FlyteSchema[FEATURE_COLUMNS],
    y: FlyteSchema[CLASSES_COLUMNS],
    hyperparams: XGBoostModelHyperparams,
) -> model_file:
    """
    This function takes the given input features and their corresponding classes to train a XGBClassifier.
    NOTE: We have simplified the number of hyper parameters we take for demo purposes
    """
    x_df = x.open().all()
    y_df = y.open().all()

    # fit model no training data
    m = XGBClassifier(
        n_jobs=hyperparams.n_jobs,
        max_depth=hyperparams.max_depth,
        n_estimators=hyperparams.n_estimators,
        booster=hyperparams.booster,
        objective=hyperparams.objective,
        learning_rate=hyperparams.learning_rate,
    )
    m.fit(x_df, y_df)

    # TODO model Blob should be a file like object
    fname = "model.joblib.dat"
    joblib.dump(m, fname)
    return (fname,)


@task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
def predict(
    x: FlyteSchema[FEATURE_COLUMNS],
    model_ser: FlyteFile[MODELSER_JOBLIB],
) -> FlyteSchema[CLASSES_COLUMNS]:
    """
    Given a any trained model, serialized using joblib (this method can be shared!) and features, this method returns
    predictions.
    """
    model = joblib.load(model_ser)
    # make predictions for test data
    x_df = x.open().all()
    y_pred = model.predict(x_df)

    col = [k for k in CLASSES_COLUMNS.keys()]
    y_pred_df = pd.DataFrame(y_pred, columns=col, dtype="int64")
    y_pred_df.round(0)
    return y_pred_df


@task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
def score(
    predictions: FlyteSchema[CLASSES_COLUMNS], y: FlyteSchema[CLASSES_COLUMNS]
) -> float:
    """
    Compares the predictions with the actuals and returns the accuracy score.
    """
    pred_df = predictions.open().all()
    y_df = y.open().all()
    # evaluate predictions
    acc = accuracy_score(y_df, pred_df)
    print("Accuracy: %.2f%%" % (acc * 100.0))
    return float(acc)


# %%
# Workflow sample here
@workflow
def diabetes_xgboost_model(
    dataset: FlyteFile[
        typing.TypeVar("csv")
    ] = "https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv",
    test_split_ratio: float = 0.33,
    seed: int = 7,
) -> workflow_outputs:
    """
    This pipeline trains an XGBoost mode for any given dataset that matches the schema as specified in
    https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names.
    """
    x_train, x_test, y_train, y_test = split_traintest_dataset(
        dataset=dataset, seed=seed, test_split_ratio=test_split_ratio
    )
    model = fit(
        x=x_train,
        y=y_train,
        hyperparams=XGBoostModelHyperparams(max_depth=4),
    )
    predictions = predict(x=x_test, model_ser=model.model)
    return model.model, score(predictions=predictions, y=y_test)


# %%
# The entire workflow can be executed locally as follows.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(diabetes_xgboost_model())
