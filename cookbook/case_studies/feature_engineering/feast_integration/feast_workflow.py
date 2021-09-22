"""
Flyte Pipeline with Feast
-------------------------

This workflow makes use of the feature engineering tasks defined in the other file. We'll build an end-to-end Flyte pipeline utilizing "Feast". 
Here is the step-by-step process:

* Fetch the SQLite3 data as a Pandas DataFrame
* Perform mean-median-imputation
* Build a feature store
* Store the updated features in an offline store
* Retrieve the features from an offline store
* Perform univariate-feature-selection
* Train a Naive Bayes model
* Load features into an online store
* Fetch one feature vector for inference
* Generate prediction
"""

import logging
import random
import typing

# %%
# Let's import the libraries.
from datetime import datetime, timedelta

import joblib
import pandas as pd
from feast import Entity, Feature, FeatureStore, FeatureView, FileSource, ValueType
from feast_dataobjects import FeatureStore, FeatureStoreConfig
from feature_eng_tasks import mean_median_imputer, univariate_selection
from flytekit import task, workflow
from flytekit.core.node_creation import create_node
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.types.file import JoblibSerializedFile
from flytekit.types.schema import FlyteSchema
from sklearn.model_selection import train_test_split
from sklearn.naive_bayes import GaussianNB

logger = logging.getLogger(__file__)


# %%
# We define the necessary data holders.

# TODO: find a better way to define these features.
FEAST_FEATURES = [
    "horse_colic_stats:rectal temperature",
    "horse_colic_stats:total protein",
    "horse_colic_stats:peripheral pulse",
    "horse_colic_stats:surgical lesion",
    "horse_colic_stats:abdominal distension",
    "horse_colic_stats:nasogastric tube",
    "horse_colic_stats:outcome",
    "horse_colic_stats:packed cell volume",
    "horse_colic_stats:nasogastric reflux PH",
]
DATABASE_URI = "https://cdn.discordapp.com/attachments/545481172399030272/861575373783040030/horse_colic.db.zip"
DATA_CLASS = "surgical lesion"


sql_task = SQLite3Task(
    name="sqlite3.horse_colic",
    query_template="select * from data",
    output_schema_type=FlyteSchema,
    task_config=SQLite3Config(
        uri=DATABASE_URI,
        compressed=True,
    ),
)

# %%
# We define two tasks, namely ``store_offline`` and ``load_historical_features`` to store and retrieve the historial features.
#
# .. list-table:: Decoding the ``Feast`` Jargon
#    :widths: 25 25
#
#    * - ``FeatureStore``
#      - A FeatureStore object is used to define, create, and retrieve features.
#    * - ``Entity``
#      - Represents a collection of entities and associated metadata. It's usually the primary key of your data.
#    * - ``FeatureView``
#      - A FeatureView defines a logical grouping of serveable features.
#    * - ``FileSource``
#      - File data sources allow for the retrieval of historical feature values from files on disk for building training datasets, as well as for materializing features into an online store.
#    * - ``apply()``
#      - Register objects to metadata store and update related infrastructure.
#    * - ``get_historical_features()``
#      - Enrich an entity dataframe with historical feature values for either training or batch scoring.
@task
def store_offline(feature_store: FeatureStore, dataframe: FlyteSchema):
    horse_colic_entity = Entity(name="Hospital Number", value_type=ValueType.STRING)

    horse_colic_feature_view = FeatureView(
        name="horse_colic_stats",
        entities=["Hospital Number"],
        features=[
            Feature(name="rectal temperature", dtype=ValueType.FLOAT),
            Feature(name="total protein", dtype=ValueType.FLOAT),
            Feature(name="peripheral pulse", dtype=ValueType.FLOAT),
            Feature(name="surgical lesion", dtype=ValueType.STRING),
            Feature(name="abdominal distension", dtype=ValueType.FLOAT),
            Feature(name="nasogastric tube", dtype=ValueType.STRING),
            Feature(name="outcome", dtype=ValueType.STRING),
            Feature(name="packed cell volume", dtype=ValueType.FLOAT),
            Feature(name="nasogastric reflux PH", dtype=ValueType.FLOAT),
        ],
        batch_source=FileSource(
            path=str(dataframe.remote_path),
            event_timestamp_column="timestamp",
        ),
        ttl=timedelta(days=1),
    )

    # Ingest the data into feast
    feature_store.apply([horse_colic_entity, horse_colic_feature_view])


@task
def load_historical_features(feature_store: FeatureStore) -> FlyteSchema:
    entity_df = pd.DataFrame.from_dict(
        {
            "Hospital Number": [
                "530101",
                "5290409",
                "5291329",
                "530051",
                "529518",
                "530101",
                "529340",
                "5290409",
                "530034",
            ],
            "event_timestamp": [
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 7, 5, 11, 36, 1),
                datetime(2021, 6, 25, 16, 36, 27),
                datetime(2021, 7, 5, 11, 50, 40),
                datetime(2021, 6, 25, 16, 36, 27),
            ],
        }
    )

    return feature_store.get_historical_features(
        entity_df=entity_df,
        features=FEAST_FEATURES,
    )


# %%
# Next, we train a naive bayes model using the data from the feature store.
@task
def train_model(dataset: pd.DataFrame, data_class: str) -> JoblibSerializedFile:
    X_train, _, y_train, _ = train_test_split(
        dataset[dataset.columns[~dataset.columns.isin([data_class])]],
        dataset[data_class],
        test_size=0.33,
        random_state=42,
    )
    model = GaussianNB()
    model.fit(X_train, y_train)
    model.feature_names = list(X_train.columns.values)
    fname = "/tmp/model.joblib.dat"
    joblib.dump(model, fname)
    return fname

# %%
# To perform inferencing, we define two tasks: ``store_online`` and ``retrieve_online``.
#
# .. list-table:: Decoding the ``Feast`` Jargon
#    :widths: 25 25
#
#    * - ``materialize()``
#      - Materialize data from the offline store into the online store.
#    * - ``get_online_features()``
#      - Retrieves the latest online feature data.
#
# .. note::
#   One key difference between an online and offline store is that only the latest feature values are stored per entity key in an 
#   online store, unlike an offline store where all feature values are stored.
#   Our dataset has two such entries with the same ``Hospital Number`` but different time stamps. 
#   Only data point with the latest timestamp will be stored in the online store.
@task
def store_online(feature_store: FeatureStore):
    feature_store.materialize(
        start_date=datetime.utcnow() - timedelta(days=250),
        end_date=datetime.utcnow() - timedelta(minutes=10),
    )

@task
def retrieve_online(
    feature_store: FeatureStore, dataset: pd.DataFrame
) -> dict:
    inference_data = random.choice(dataset["Hospital Number"])
    logger.info(f"Hospital Number chosen for inference is: {inference_data}")
    entity_rows = [{"Hospital Number": inference_data}]

    return feature_store.get_online_features(FEAST_FEATURES, entity_rows)


# %%
# We define a task to test our model using the inference point fetched earlier.
@task
def test_model(
    model_ser: JoblibSerializedFile,
    inference_point: dict,
) -> typing.List[str]:

    # Load model
    model = joblib.load(model_ser)
    f_names = model.feature_names

    test_list = []
    for each_name in f_names:
        test_list.append(inference_point[each_name][0])
    prediction = model.predict([test_list])
    return prediction

# %%
# Next, we need to convert timestamp column in the underlying dataframe, otherwise its type is written as string.
@task
def convert_timestamp_column(
    dataframe: FlyteSchema, timestamp_column: str
) -> FlyteSchema:
    df = dataframe.open().all()
    df[timestamp_column] = pd.to_datetime(df[timestamp_column])
    return df

# %%
# The ``build_feature_store`` task is a medium to access Feast methods by building a feature store.
@task
def build_feature_store(s3_bucket: str, registry_path: str, online_store_path: str) -> FeatureStore:
    feature_store_config = FeatureStoreConfig(project="horsecolic", s3_bucket=s3_bucket, registry_path=registry_path, online_store_path=online_store_path)
    return FeatureStore(config=feature_store_config)

# %%
# Finally, we define a workflow that streamlines the whole pipeline building and feature serving process.
@workflow
def feast_workflow(
    imputation_method: str = "mean",
    num_features_univariate: int = 7,
    s3_bucket: str = "feast-integration",
    registry_path: str = "registry.db",
    online_store_path: str = "online.db",
) -> typing.List[str]:

    # Load parquet file from sqlite task
    df = sql_task()

    # Perfrom mean median imputation
    dataframe = mean_median_imputer(dataframe=df, imputation_method=imputation_method)

    # Convert timestamp column from string to datetime.
    converted_df = convert_timestamp_column(
        dataframe=dataframe, timestamp_column="timestamp"
    )

    # Build feature store
    feature_store = build_feature_store(s3_bucket=s3_bucket, registry_path=registry_path, online_store_path=online_store_path)

    # Ingest data into offline store
    store_offline_node = create_node(store_offline, feature_store=feature_store, dataframe=converted_df)

    # Demonstrate how to load features from offline store
    load_historical_features_node = create_node(load_historical_features, feature_store=feature_store)

    # Ingest data into online store
    store_online_node = create_node(store_online, feature_store=feature_store)

    # Retrieve feature data from online store
    retrieve_online_node = create_node(retrieve_online, feature_store=feature_store, dataset=converted_df)

    # Enforce order in which tasks that interact with the feast SDK have to run
    store_offline_node >> load_historical_features_node
    load_historical_features_node >> store_online_node
    store_online_node >> retrieve_online_node

    # Perform univariate feature selection
    selected_features = univariate_selection(
        dataframe=load_historical_features_node.o0,
        num_features=num_features_univariate,
        data_class=DATA_CLASS,
    )

    # Train the Naive Bayes model
    trained_model = train_model(
        dataset=selected_features,
        data_class=DATA_CLASS,
    )

    # Use a feature retrieved from the online store for inference
    prediction = test_model(
        model_ser=trained_model,
        inference_point=retrieve_online_node.o0,
    )

    return prediction

if __name__ == "__main__":
    print(f"{feast_workflow()}")

# %%
# You should see prediction against the test input as the workflow output.