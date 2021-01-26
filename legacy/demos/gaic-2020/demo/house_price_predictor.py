##############################
# The example has been borrowed from Sagemaker examples
# https://github.com/awslabs/amazon-sagemaker-examples/blob/master/advanced_functionality/multi_model_xgboost_home_value/xgboost_multi_model_endpoint_home_value.ipynb
# Idea is to illustrate parallelized execution and eventually also show Sagemaker execution
import os
import typing

import joblib
import numpy as np
import pandas as pd
from flytekit.sdk.tasks import python_task, inputs, outputs
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output
from sklearn.metrics import accuracy_score
from sklearn.model_selection import train_test_split
from xgboost import XGBClassifier

NUM_HOUSES_PER_LOCATION = 1000
LOCATIONS = ['NewYork_NY', 'LosAngeles_CA', 'Chicago_IL', 'Houston_TX', 'Dallas_TX',
             'Phoenix_AZ', 'Philadelphia_PA', 'SanAntonio_TX', 'SanDiego_CA', 'SanFrancisco_CA']

PARALLEL_TRAINING_JOBS = 4  # len(LOCATIONS) if your account limits can handle it
MAX_YEAR = 2019
SPLIT_RATIOS = [0.6, 0.3, 0.1]


def gen_price(house) -> int:
    _base_price = int(house['SQUARE_FEET'] * 150)
    _price = int(_base_price + (10000 * house['NUM_BEDROOMS']) + \
                 (15000 * house['NUM_BATHROOMS']) + \
                 (15000 * house['LOT_ACRES']) + \
                 (15000 * house['GARAGE_SPACES']) - \
                 (5000 * (MAX_YEAR - house['YEAR_BUILT'])))
    return _price


def gen_random_house() -> typing.List:
    _house = {'SQUARE_FEET': int(np.random.normal(3000, 750)),
              'NUM_BEDROOMS': np.random.randint(2, 7),
              'NUM_BATHROOMS': np.random.randint(2, 7) / 2,
              'LOT_ACRES': round(np.random.normal(1.0, 0.25), 2),
              'GARAGE_SPACES': np.random.randint(0, 4),
              'YEAR_BUILT': min(MAX_YEAR, int(np.random.normal(1995, 10)))}
    _price = gen_price(_house)
    return [_price, _house['YEAR_BUILT'], _house['SQUARE_FEET'],
            _house['NUM_BEDROOMS'], _house['NUM_BATHROOMS'],
            _house['LOT_ACRES'], _house['GARAGE_SPACES']]


def gen_houses(num_houses) -> pd.DataFrame:
    _house_list = []
    for i in range(num_houses):
        _house_list.append(gen_random_house())
    _df = pd.DataFrame(_house_list,
                       columns=['PRICE', 'YEAR_BUILT', 'SQUARE_FEET', 'NUM_BEDROOMS',
                                'NUM_BATHROOMS', 'LOT_ACRES', 'GARAGE_SPACES'])
    return _df


def split_data(df: pd.DataFrame, seed: int, split: typing.List[float]) -> (np.ndarray, np.ndarray, np.ndarray):
    # split data into train and test sets
    seed = seed
    val_size = split[1]
    test_size = split[2]

    num_samples = df.shape[0]
    x1 = df.values[:num_samples, 1:]  # keep only the features, skip the target, all rows
    y1 = df.values[:num_samples, :1]  # keep only the target, all rows

    # Use split ratios to divide up into train/val/test
    x_train, x_val, y_train, y_val = train_test_split(x1, y1, test_size=(test_size + val_size), random_state=seed)
    # Of the remaining non-training samples, give proper ratio to validation and to test
    x_test, x_test, y_test, y_test = train_test_split(x_val, y_val,
                                                      test_size=(test_size / (test_size + val_size)),
                                                      random_state=seed)
    # reassemble the datasets with target in first column and features after that
    _train = np.concatenate([y_train, x_train], axis=1)
    _val = np.concatenate([y_val, x_val], axis=1)
    _test = np.concatenate([y_test, x_test], axis=1)

    return _train, _val, _test


def generate_data(loc: str, number_of_houses: int, seed: int) -> (np.ndarray, np.ndarray, np.ndarray):
    _houses = gen_houses(number_of_houses)
    _train, _val, _test = split_data(_houses, seed, split=SPLIT_RATIOS)
    return _train, _val, _test


def save_to_dir(path: str, n: str, arr: np.ndarray) -> str:
    d = os.path.join(path, n)
    os.makedirs(d, exist_ok=True)
    save_to_file(d, n, arr)
    return d


def save_to_file(path: str, n: str, arr: np.ndarray) -> str:
    f = f'{n}.csv'
    f = os.path.join(path, f)
    np.savetxt(f, arr, delimiter=',', fmt='%.2f')
    return f


# NOTE we are generated multipart csv (or a directory) because Sagemaker wants a directory and cannot work with a
# single file
@inputs(loc=Types.String, number_of_houses=Types.Integer, seed=Types.Integer)
@outputs(train=Types.MultiPartCSV, val=Types.MultiPartCSV, test=Types.CSV)
@python_task(cache=True, cache_version="0.1", memory_request="200Mi")
def generate_and_split_data(wf_params, loc, number_of_houses, seed, train, val, test):
    _train, _val, _test = generate_data(loc, number_of_houses, seed)
    os.makedirs("data", exist_ok=True)
    train.set(save_to_dir("data", "train", _train))
    val.set(save_to_dir("data", "val", _val))
    test.set(save_to_file("data", "test", _test))


@inputs(train=Types.MultiPartCSV)
@outputs(model=Types.Blob)
@python_task(cache_version='1.0', cache=True, memory_request="200Mi")
def fit(ctx, train, model):
    """
    This function takes the given input features and their corresponding classes to train a XGBClassifier.
    NOTE: We have simplified the algorithm for demo. Default hyper params & no validation dataset :P.
    """
    train.download()
    files = os.listdir(train.local_path)
    # We know we are writing just one file, so we will just read the one file
    df = pd.read_csv(os.path.join(train.local_path, files[0]), header=None)
    y = df[df.columns[0]]
    x = df[df.columns[1:]]
    # fit model no training data
    m = XGBClassifier()
    m.fit(x, y)

    # TODO model Blob should be a file like object
    fname = "model.joblib.dat"
    joblib.dump(m, fname)
    model.set(fname)


@inputs(test=Types.CSV, model_ser=Types.Blob)  # TODO: format=".joblib.dat"))
@outputs(predictions=Types.List(Types.Float), accuracy=Types.Float)
@python_task(cache_version='1.0', cache=True, memory_request="200Mi")
def predict(ctx, test, model_ser, predictions, accuracy):
    """
    Given a any trained model, serialized using joblib (this method can be shared!) and features, this method returns
    predictions.
    """
    # Load model
    model_ser.download()
    model = joblib.load(model_ser.local_path)
    # Load test data
    test.download()
    test_df = pd.read_csv(test.local_path, header=None)
    x_df = test_df[test_df.columns[1:]]
    y_df = test_df[test_df.columns[0]]
    y_pred = model.predict(x_df)
    predictions.set(y_pred.tolist())

    acc = accuracy_score(y_df, y_pred)

    print("Accuracy: %.2f%%" % (acc * 100.0))
    accuracy.set(float(acc))


@workflow_class
class HousePricePredictionModelTrainer(object):
    """
    This pipeline trains an XGBoost model, also generated synthetic data and runs predictions against test dataset
    """

    loc = Input(Types.String, help="Location for where to train the model.")
    seed = Input(Types.Integer, default=7, help="Seed to use for splitting.")
    num_houses = Input(Types.Integer, default=1000, help="Number of houses to generate data for")

    # the actual algorithm
    split = generate_and_split_data(loc=loc, number_of_houses=num_houses, seed=seed)
    fit_task = fit(train=split.outputs.train)
    predicted = predict(model_ser=fit_task.outputs.model, test=split.outputs.test)

    # Outputs: joblib seralized model and accuracy of the model
    model = Output(fit_task.outputs.model, sdk_type=Types.Blob)
    accuracy = Output(predicted.outputs.accuracy, sdk_type=Types.Float)
