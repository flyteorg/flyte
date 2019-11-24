from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import joblib
import pandas as pd
from flytekit.sdk.tasks import python_task, outputs, inputs
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Output, Input
from sklearn.metrics import accuracy_score
from sklearn.model_selection import train_test_split
from xgboost import XGBClassifier

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
TYPED_COLUMNS = [
    ('#preg', Types.Integer),
    ('pgc_2h', Types.Integer),
    ('diastolic_bp', Types.Integer),
    ('tricep_skin_fold_mm', Types.Integer),
    ('serum_insulin_2h', Types.Integer),
    ('bmi', Types.Float),
    ('diabetes_pedigree', Types.Float),
    ('age', Types.Integer),
    ('class', Types.Integer),
]
# the input dataset schema
DATASET_SCHEMA = Types.Schema(TYPED_COLUMNS)
# the first 8 columns are features
FEATURES_SCHEMA = Types.Schema(TYPED_COLUMNS[:8])
# the last column is the class
CLASSES_SCHEMA = Types.Schema([TYPED_COLUMNS[-1]])


class XGBoostModelHyperparams(object):
    """
    These are the xgboost hyper parameters available in scikit-learn library.
    """

    def __init__(self, max_depth=3, learning_rate=0.1, n_estimators=100,
                 objective="binary:logistic", booster='gbtree',
                 n_jobs=1, **kwargs):
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
@inputs(dataset=Types.CSV, seed=Types.Integer, test_split_ratio=Types.Float)
@outputs(x_train=FEATURES_SCHEMA, x_test=FEATURES_SCHEMA, y_train=CLASSES_SCHEMA, y_test=CLASSES_SCHEMA)
@python_task(cache_version='1.0', cache=True, memory_limit="200Mi")
def get_traintest_splitdatabase(ctx, dataset, seed, test_split_ratio, x_train, x_test, y_train, y_test):
    """
    Retrieves the training dataset from the given blob location and then splits it using the split ratio and returns the result
    This splitter is only for the dataset that has the format as specified in the example csv. The last column is assumed to be
    the class and all other columns 0-8 the features.

    The data is returned as a schema, which gets converted to a parquet file in the back.
    """
    dataset.download()
    column_names = [k for k in DATASET_SCHEMA.columns.keys()]
    df = pd.read_csv(dataset.local_path, names=column_names)

    # Select all features
    x = df[column_names[:8]]
    # Select only the classes
    y = df[[column_names[-1]]]

    # split data into train and test sets
    _x_train, _x_test, _y_train, _y_test = train_test_split(
        x, y, test_size=test_split_ratio, random_state=seed)

    # TODO also add support for Spark dataframe, but make the pyspark dependency optional
    x_train.set(_x_train)
    x_test.set(_x_test)
    y_train.set(_y_train)
    y_test.set(_y_test)


@inputs(x=FEATURES_SCHEMA, y=CLASSES_SCHEMA, hyperparams=Types.Generic)  # TODO support arbitrary jsonifiable classes
@outputs(model=Types.Blob)  # TODO: Support for subtype format=".joblib.dat"))
@python_task(cache_version='1.0', cache=True, memory_limit="200Mi")
def fit(ctx, x, y, hyperparams, model):
    """
    This function takes the given input features and their corresponding classes to train a XGBClassifier.
    NOTE: We have simplified the number of hyper parameters we take for demo purposes
    """
    with x as r:
        x_df = r.read()
    with y as r:
        y_df = r.read()

    hp = XGBoostModelHyperparams.from_dict(hyperparams)
    # fit model no training data
    m = XGBClassifier(n_jobs=hp.n_jobs, max_depth=hp.max_depth, n_estimators=hp.n_estimators, booster=hp.booster,
                      objective=hp.objective, learning_rate=hp.learning_rate)
    m.fit(x_df, y_df)

    # TODO model Blob should be a file like object
    fname = "model.joblib.dat"
    joblib.dump(m, fname)
    model.set(fname)


@inputs(x=FEATURES_SCHEMA, model_ser=Types.Blob)  # TODO: format=".joblib.dat"))
@outputs(predictions=CLASSES_SCHEMA)
@python_task(cache_version='1.0', cache=True, memory_limit="200Mi")
def predict(ctx, x, model_ser, predictions):
    """
    Given a any trained model, serialized using joblib (this method can be shared!) and features, this method returns
    predictions.
    """
    model_ser.download()
    model = joblib.load(model_ser.local_path)
    # make predictions for test data
    with x as r:
        x_df = r.read()
    y_pred = model.predict(x_df)

    col = [k for k in CLASSES_SCHEMA.columns.keys()]
    y_pred_df = pd.DataFrame(y_pred, columns=col, dtype="int64")
    y_pred_df.round(0)
    predictions.set(y_pred_df)


@inputs(predictions=CLASSES_SCHEMA, y=CLASSES_SCHEMA)
@outputs(accuracy=Types.Float)
@python_task(cache_version='1.0', cache=True, memory_limit="200Mi")
def metrics(ctx, predictions, y, accuracy):
    """
    Compares the predictions with the actuals and returns the accuracy score.
    """
    with predictions as r:
        pred_df = r.read()

    with y as r:
        y_df = r.read()

    # evaluate predictions
    acc = accuracy_score(y_df, pred_df)

    print("Accuracy: %.2f%%" % (acc * 100.0))
    accuracy.set(float(acc))


@workflow_class
class DiabetesXGBoostModelTrainer(object):
    """
    This pipeline trains an XGBoost mode for any given dataset that matches the schema as specified in
    https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names.
    """

    # Inputs dataset, fraction of the dataset to be split out for validations and seed to use to perform the split
    dataset = Input(Types.CSV, default=Types.CSV.create_at_known_location(
        "https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv"),
                    help="A CSV File that matches the format https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names")

    test_split_ratio = Input(Types.Float, default=0.33, help="Ratio of how much should be test to Train")
    seed = Input(Types.Integer, default=7, help="Seed to use for splitting.")

    # the actual algorithm
    split = get_traintest_splitdatabase(dataset=dataset, seed=seed, test_split_ratio=test_split_ratio)
    fit_task = fit(x=split.outputs.x_train, y=split.outputs.y_train, hyperparams=XGBoostModelHyperparams(
        max_depth=4,
    ).to_dict())
    predicted = predict(model_ser=fit_task.outputs.model, x=split.outputs.x_test)
    score_task = metrics(predictions=predicted.outputs.predictions, y=split.outputs.y_test)

    # Outputs: joblib seralized model and accuracy of the model
    model = Output(fit_task.outputs.model, sdk_type=Types.Blob)
    accuracy = Output(score_task.outputs.accuracy, sdk_type=Types.Float)
