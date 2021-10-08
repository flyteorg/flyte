"""
Validating and Testing Machine Learning Pipelines
-------------------------------------------------

In this example we'll show you how to use :ref:`pandera.SchemaModel <pandera:schema_models>`
to annotate dataframe inputs and outputs in an `sklearn <https://scikit-learn.org/stable/>`__
model-training pipeline.

At a high-level, the pipeline architecture involves fetching, parsing, and splitting data, then training
a model on the training set and evaluating the trained model on the test set to produce metrics:

.. mermaid::

    flowchart LR
        fetch(fetch raw data) --> raw[(raw data)]
        raw --> parse(parse data)
        parse --> parsed_data[(parsed data)]
        parsed_data --> split(split data)
        split --> train[(training set)]
        split --> test[(test set)]
        train --> train_model(train model)
        train_model --> model[(model)]
        test --> evaluate(evaluate model)
        model --> evaluate
        evaluate --> metrics

        style fetch fill:#fff2b2,stroke:#333
        style parse fill:#fff2b2,stroke:#333
        style split fill:#fff2b2,stroke:#333
        style train_model fill:#fff2b2,stroke:#333
        style evaluate fill:#fff2b2,stroke:#333


First, let's import all the necessary dependencies including ``pandas``, ``pandera``, ``sklearn``,
and ``hypothesis``.

"""

import typing

import joblib
import pandas as pd
import pandera as pa
from flytekit import task, workflow
from flytekit.types.file import JoblibSerializedFile
from pandera.typing import DataFrame, Series, Index
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score


# %%
# We also need to import the ``pandera`` flytekit plugin to enable dataframe runtime type-checking:

import flytekitplugins.pandera


# %%
# The Dataset: UCI Heart Disease
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# Before building out our pipeline it's good practice to take a closer look at our data. We'll be
# modeling the `UCI Heart Disease <https://archive.ics.uci.edu/ml/datasets/heart+Disease>`__ dataset.
#
# From the data documentation in the link above, we can see that this is a tabular dataset where each
# row contains data pertaining to a patient. There are 14 variables where 13 of them are features and one of them
# is a binary target variable indicating the presence or absence of heart disease.
#
# .. list-table:: UCI Dataset Variables
#    :widths: 25 25
#    :header-rows: 1
#
#    * - variable
#      - description
#    * - ``age``
#      - age in years
#    * - ``sex``
#      - 1 = male; 0 = female
#    * - ``cp``
#      - chest pain type (4 values)
#    * - ``trestbps``
#      - resting blood pressure
#    * - ``chol``
#      - serum cholestoral in mg/dl
#    * - ``fbs``
#      - fasting blood sugar > 120 mg/dl
#    * - ``restecg``
#      - resting electrocardiographic results (values 0,1,2)
#    * - ``thalach``
#      - maximum heart rate achieved
#    * - ``exang``
#      - exercise induced angina
#    * - ``oldpeak``
#      - ST (stress test) depression induced by exercise relative to rest
#    * - ``slope``
#      - the slope of the peak exercise ST segment
#    * - ``ca``
#      - number of major vessels (0-3) colored by flourosopy
#    * - ``thal``
#      - 3 = normal; 6 = fixed defect; 7 = reversable defect
#    * - ``target``
#      - the predicted attribute
#
# In practice, we'd want to do a little data exploration to first to get a sense of the distribution of variables.
# A useful resource for this is the `Kaggle <https://www.kaggle.com/ronitf/heart-disease-uci>`__ version of this dataset,
# which has been slightly preprocessed to be model-ready.
# 
# .. Note::
#    We'll be using the data provided by the UCI data repository since we want to work with a dataset that
#    requires some preprocessing to be model-ready.
#
# Once we've gotten a rough sense of the statistical properties of the data, we can encode that domain knowledge into
# a pandera schema:

class RawData(pa.SchemaModel):
    age: Series[int] = pa.Field(in_range={"min_value": 0, "max_value": 200})
    sex: Series[int] = pa.Field(isin=[0, 1])
    cp: Series[int] = pa.Field(
        isin=[
            1,  # typical angina
            2,  # atypical angina
            3,  # non-anginal pain
            4,  # asymptomatic
        ]
    )
    trestbps: Series[int] = pa.Field(in_range={"min_value": 0, "max_value": 200})
    chol: Series[int] = pa.Field(in_range={"min_value": 0, "max_value": 600})
    fbs: Series[int] = pa.Field(isin=[0, 1])
    restecg: Series[int] = pa.Field(
        isin=[
            0,  # normal
            1,  # having ST-T wave abnormality
            2,  # showing probable or definite left ventricular hypertrophy by Estes' criteria  
        ]
    )
    thalach: Series[int] = pa.Field(in_range={"min_value": 0, "max_value": 300})
    exang: Series[int] = pa.Field(isin=[0, 1])
    oldpeak: Series[float] = pa.Field(in_range={"min_value": 0, "max_value": 10})
    slope: Series[int] = pa.Field(
        isin=[
            1,  # upsloping
            2,  # flat
            3,  # downsloping
        ]
    )
    ca: Series[int] = pa.Field(isin=[0, 1, 2, 3])
    thal: Series[int] = pa.Field(
        isin=[
            3,  # normal
            6,  # fixed defect
            7,  # reversable defect
        ]
    )
    target: Series[int] = pa.Field(ge=0, le=4)

    class Config:
        coerce = True


# %%
# As we can see, the ``RawData`` schema serves as both documentation and a means of enforcing some minimal quality
# checks relating to the data type of each variable as well as additional constraints about their allowable values.
#
# Fetching the Raw Data
# ^^^^^^^^^^^^^^^^^^^^^
#
# Now we're ready to write our first Flyte task:

@task
def fetch_raw_data() -> DataFrame[RawData]:
    data_url = "https://archive.ics.uci.edu/ml/machine-learning-databases/heart-disease/processed.cleveland.data"
    return (
        pd.read_csv(data_url, header=None, names=RawData.to_schema().columns.keys())
        # Demo: remove these lines to see what happens!
        .replace({"ca": {"?": None}, "thal": {"?": None}})
        .dropna(subset=["ca", "thal"])
        .astype({"ca": float, "thal": float})
    )

# %%
# This function fetches the raw data from ``data_url`` and cleaning up invalid values such as ``"?"``, dropping records
# with null values, and casting the ``ca`` and ``thal`` columns into floats since those columns serialize the number
# values in the float format, e.g. ``3.0``, so we cast them into floats before pandera coerces them into integer values.
#
# .. Note::
#   We're using the generic type ``pandera.typing.DataFrame``` and supplying the ``RawData`` schema model to
#   specify the expected fields of the dataframe. This is pandera's syntax for adding type annotations to dataframes.
#
# Parsing the Raw Data
# ^^^^^^^^^^^^^^^^^^^^
#
# For this relatively simple dataset, the only pre-processing we need to do is to convert the ``target`` column from
# having values ranging from ``0 - 4``` to a binary representation where ``0`` represents absence of heart disease and
# ``1`` represents presence of heart disease.
#
# Here we can use inheritence to define a ``ParsedData`` schema by overriding just the ``target`` attribute:

class ParsedData(RawData):
    target: Series[int] = pa.Field(isin=[0, 1])


@task
def parse_raw_data(raw_data: DataFrame[RawData]) -> DataFrame[ParsedData]:
    return raw_data.assign(target=lambda _: (_.target > 0).astype(int))

# %%
# As we can see the ``parse_raw_data`` function takes in a dataframe of type ``DataFrame[RawData]`` and outputs another
# dataframe of type ``DataFrame[ParsedData]``. The actual parsing logic is very simple, but you can imagine many cases
# in which this would involve other steps, such as one-hot-encoding. As we'll see later, since we're using a tree-based
# sklearn estimator, we can preserve the integer representations of the features since these class of estimators handles
# the discretization of these types of variables.
#
# Splitting the Data
# ^^^^^^^^^^^^^^^^^^
#
# Now it's time to split the data into a training set and a test set. Here we'll showcase the utility of
# :ref:`named outputs <_sphx_glr_auto_core_flyte_basics_named_outputs.py>` combined with pandera schemas.

import typing

DataSplits = typing.NamedTuple(
    "DataSplits",
    training_set=DataFrame[ParsedData],
    test_set=DataFrame[ParsedData]
)

@task
def split_data(parsed_data: DataFrame[ParsedData], test_size: float, random_state: int) -> DataSplits:
    training_set = parsed_data.sample(frac=test_size, random_state=random_state)
    test_set = parsed_data[~parsed_data.index.isin(training_set.index)]
    return training_set, test_set

# %%
# As we can see above, we're defining a ``DataSplits`` named tuple consisting of a ``training_set`` key and ``test_set``
# key. Since we assume that the training and test sets are drawn from the same distribution, we can specify the
# type of these data splits as ``DataFrame[ParsedData]``.
#
# Training a Model
# ^^^^^^^^^^^^^^^^
#
# Next we'll train a ``RandomForestClassifier`` to predict the absence/presence of heart disease:

def get_features_and_target(dataset):
    """Helper function for separating feature and target data."""
    X = dataset[[x for x in dataset if x != "target"]]
    y = dataset["target"]
    return X, y


@task
def train_model(training_set: DataFrame[ParsedData], random_state: int) -> JoblibSerializedFile:
    model = RandomForestClassifier(n_estimators=100, random_state=random_state)
    X, y = get_features_and_target(training_set)
    model.fit(X, y)
    model_fp = "/tmp/model.joblib"
    joblib.dump(model, model_fp)
    return JoblibSerializedFile(path=model_fp)

# %%
# This task serializes the model with joblib and returns a ``JoblibSerializedFile`` type, which is understood and
# automatically handled by Flyte so that the pointer to the actual serialized file can be passed onto the last
# step of the pipeline.
#
# Model Evaluation
# ^^^^^^^^^^^^^^^^
#
# Next we assess the accuracy score of the model on the test set:

@task
def evaluate_model(model: JoblibSerializedFile, test_set: DataFrame[ParsedData]) -> float:
    with open(model, "rb") as f:
        model = joblib.load(f)
    X, y = get_features_and_target(test_set)
    preds = model.predict(X)
    return accuracy_score(y, preds)

# %%
# Finally, we put all of the pieces together in a Flyte workflow:

@workflow
def pipeline(data_random_state: int, model_random_state: int) -> float:
    raw_data = fetch_raw_data()
    parsed_data = parse_raw_data(raw_data=raw_data)
    training_set, test_set = split_data(parsed_data=parsed_data, test_size=0.2, random_state=data_random_state)
    model = train_model(training_set=training_set, random_state=model_random_state)
    return evaluate_model(model=model, test_set=test_set)


# %%
# Unit Testing Pipeline Code
# ^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# A powerful feature offered by pandera is `data synthesis strategies <https://pandera.readthedocs.io/en/stable/data_synthesis_strategies.html>`__.
# It harnesses the power of `hypothesis <https://hypothesis.readthedocs.io/en/latest/>`__, which is a property-based
# testing library that makes it easy to test your code not based on hand-crafted test cases, but by defining a
# contract of types and assertions that your code needs to fulfill. ``pandera`` strategies enables us to sample valid
# data under the constraints of our ``SchemaModel`` so that we can make sure the code implementation is correct.
#
# Typically we'd want to define these tests under the conventions of a testing framework like `pytest <https://docs.pytest.org/en/6.2.x/>`__,
# but in this example we'll just define it under a ``if __name__ == "__main__"`` block:

if __name__ == "__main__":
    import hypothesis
    import hypothesis.strategies as st
    from hypothesis import given

    class PositiveExamples(ParsedData):
        target: Series[int] = pa.Field(eq=1)

    class NegativeExamples(ParsedData):
        target: Series[int] = pa.Field(eq=0)

    @given(
        PositiveExamples.strategy(size=5),
        NegativeExamples.strategy(size=5),
        st.integers(min_value=0, max_value=2**32)
    )
    @hypothesis.settings(
        deadline=1000,
        suppress_health_check=[
            hypothesis.HealthCheck.too_slow,
            hypothesis.HealthCheck.filter_too_much,
        ],
    )
    def test_train_model(positive_examples, negative_examples, model_random_state):
        training_set = pd.concat([positive_examples, negative_examples])
        model = train_model(training_set=training_set, random_state=model_random_state)
        model = joblib.load(model)
        X, _ = get_features_and_target(training_set)
        preds = model.predict(X)
        for pred in preds:
            assert pred in {0, 1}


    def run_test_suite():
        test_train_model()
        print("tests completed!")

    run_test_suite()


# %%
# In the test code above, we can see that we're defining two schema models for the purposes of testing: ``PositiveExamples``
# and ``NegativeExamples``. We're doing this so that we guarantee that we can control the class balance of the sampled
# data.
#
# Then, we define a test called ``test_train_model``, which is decorated with ``@given`` so that ``hypothesis`` can
# handle the sampling and shrinkage of test cases, trying to find a scenario that falsifies the assumptions in the
# test function body. In this case, all we're doing is checking the following properties of the ``train_model``
# task:
#
# - With some sample of data, it can output a trained ``model``
# - That model can be used to generate predictions in the set ``{0, 1}``
#
# Conclusion
# ^^^^^^^^^^
#
# In this example we've learned how to:
#
# 1. Define pandera ``SchemaModel`` s to document and enforce the properties of dataframes as they pass through our
#    model-training pipeline.
# 2. Type-annotate ``flytekit`` tasks to automatically type-check dataframes at runtime using the ``flytekitplugins.pandera`` plugin.
# 3. Use pandera ``SchemaModel`` s in unit tests to make sure that our code is working as expected under certain operating assumptions.
