import sys
import typing
from collections import OrderedDict

import pytest
from typing_extensions import Annotated

from flytekit import Resources
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.types.file import FileExt, FlyteFile
from flytekit.types.schema import FlyteSchema


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_diabetes():
    import pandas as pd

    # Since we are working with a specific dataset, we will create a strictly typed schema for the dataset.
    # If we wanted a generic data splitter we could use a Generic schema without any column type and name information
    # Example file: https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv
    # CSV Columns
    #  1. Number of times pregnant
    #  2. Plasma glucose concentration a 2 hours in an oral glucose tolerance test
    #  3. Diastolic blood pressure (mm Hg)
    #  4. Triceps skin fold thickness (mm)
    #  5. 2-Hour serum insulin (mu U/ml)
    #  6. Body mass index (weight in kg/(height in m)^2)
    #  7. Diabetes pedigree function
    #  8. Age (years)
    #  9. Class variable (0 or 1)
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
    # the first 8 columns are features
    FEATURE_COLUMNS = OrderedDict({k: v for k, v in DATASET_COLUMNS.items() if k != "class"})
    # the last column is the class
    CLASSES_COLUMNS = OrderedDict({"class": int})

    MODELSER_JOBLIB = Annotated[str, FileExt("joblib.dat")]

    class XGBoostModelHyperparams(object):
        """
        These are the xgboost hyper parameters available in scikit-learn library.
        """

        def __init__(
            self,
            max_depth=3,
            learning_rate=0.1,
            n_estimators=100,
            objective="binary:logistic",
            booster="gbtree",
            n_jobs=1,
            **kwargs,
        ):
            self.n_jobs = int(n_jobs)
            self.booster = booster
            self.objective = objective
            self.n_estimators = int(n_estimators)
            self.learning_rate = learning_rate
            self.max_depth = int(max_depth)

        def to_dict(self):
            return self.__dict__

        @classmethod
        def from_dict(cls, d):
            return cls(**d)

    # load data
    # Example file: https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv
    @task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
    def split_traintest_dataset(
        dataset: FlyteFile[typing.TypeVar("csv")], seed: int, test_split_ratio: float
    ) -> typing.Tuple[
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

        # We will fake train test split. Just return the same dataset multiple times
        return x, x, y, y

    nt = typing.NamedTuple("Outputs", model=FlyteFile[MODELSER_JOBLIB])

    @task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
    def fit(x: FlyteSchema[FEATURE_COLUMNS], y: FlyteSchema[CLASSES_COLUMNS], hyperparams: dict) -> nt:
        """
        This function takes the given input features and their corresponding classes to train a XGBClassifier.
        NOTE: We have simplified the number of hyper parameters we take for demo purposes
        """
        x_df = x.open().all()
        print(x_df)
        y_df = y.open().all()
        print(y_df)

        hp = XGBoostModelHyperparams.from_dict(hyperparams)
        print(hp)
        # fit model no training data
        # Faking fit

        fname = "model.joblib.dat"
        with open(fname, "w") as f:
            f.write("Some binary data")
        return nt(model=fname)  # type: ignore

    @task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
    def predict(x: FlyteSchema[FEATURE_COLUMNS], model_ser: FlyteFile[MODELSER_JOBLIB]) -> FlyteSchema[CLASSES_COLUMNS]:
        """
        Given a any trained model, serialized using joblib (this method can be shared!) and features, this method returns
        predictions.
        """
        # make predictions for test data
        x_df = x.open().all()
        print(x_df)
        col = [k for k in CLASSES_COLUMNS.keys()]
        y_pred_df = pd.DataFrame(data=[{col[0]: [0, 1]}], columns=col, dtype="int64")
        y_pred_df.round(0)
        return y_pred_df

    @task(cache_version="1.0", cache=True, limits=Resources(mem="200Mi"))
    def score(predictions: FlyteSchema[CLASSES_COLUMNS], y: FlyteSchema[CLASSES_COLUMNS]) -> float:
        """
        Compares the predictions with the actuals and returns the accuracy score.
        """
        pred_df = predictions.open().all()
        print(pred_df)
        y_df = y.open().all()
        print(y_df)
        # evaluate predictions
        return 0.2

    @workflow
    def diabetes_xgboost_model(
        dataset: FlyteFile[typing.TypeVar("csv")],
        # = "https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv",
        test_split_ratio: float = 0.33,
        seed: int = 7,
    ) -> typing.NamedTuple("Outputs", model=FlyteFile[MODELSER_JOBLIB], accuracy=float):
        """
        This pipeline trains an XGBoost mode for any given dataset that matches the schema as specified in
        https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names.
        """
        x_train, x_test, y_train, y_test = split_traintest_dataset(
            dataset=dataset, seed=seed, test_split_ratio=test_split_ratio
        )
        model = fit(x=x_train, y=y_train, hyperparams=XGBoostModelHyperparams(max_depth=4).to_dict())
        predictions = predict(x=x_test, model_ser=model.model)
        return model.model, score(predictions=predictions, y=y_test)

    # TODO enable this after the defaults are working
    # diabetes_xgboost_model(dataset="/Users/kumare/Downloads/pima-indians-diabetes.data.csv")
