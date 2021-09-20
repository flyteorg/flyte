import os
from datetime import datetime, timedelta

from flytekit.core.context_manager import FlyteContext

import random
import joblib
import logging
import typing
import pandas as pd
from feast import (
    Entity,
    Feature,
    FeatureStore,
    FeatureView,
    FileSource,
    RepoConfig,
    ValueType,
    online_response,
    registry,
)
from flytekit.core.node_creation import create_node
from feast.infra.offline_stores.file import FileOfflineStoreConfig
from feast.infra.online_stores.sqlite import SqliteOnlineStoreConfig
from flytekit import reference_task, task, workflow, Workflow
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.types.file import JoblibSerializedFile
from flytekit.types.file.file import FlyteFile
from flytekit.types.schema import FlyteSchema
from sklearn.model_selection import train_test_split
from sklearn.naive_bayes import GaussianNB
from flytekit.configuration import aws
from feature_eng_tasks import mean_median_imputer, univariate_selection
from feast_dataobjects import FeatureStore, FeatureStoreConfig


logger = logging.getLogger(__file__)
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
# Next, we train the Naive Bayes model using the data that's been fetched from the feature store.
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
# We define a task to test the model using the inference point fetched earlier.
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


@task
def convert_timestamp_column(
    dataframe: FlyteSchema, timestamp_column: str
) -> FlyteSchema:
    df = dataframe.open().all()
    df[timestamp_column] = pd.to_datetime(df[timestamp_column])
    return df

@task
def build_feature_store(s3_bucket: str, registry_path: str, online_store_path: str) -> FeatureStore:
    feature_store_config = FeatureStoreConfig(project="horsecolic", s3_bucket=s3_bucket, registry_path=registry_path, online_store_path=online_store_path)
    return FeatureStore(config=feature_store_config)


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
    dataframe = mean_median_imputer(dataframe=df, imputation_method=imputation_method)
    # Need to convert timestamp column in the underlying dataframe, otherwise its type is written as
    # string. There is probably a better way of doing this conversion.
    converted_df = convert_timestamp_column(
        dataframe=dataframe, timestamp_column="timestamp"
    )

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

    # Use a feature retrieved from the online store for inference on a trained model
    selected_features = univariate_selection(
        dataframe=load_historical_features_node.o0,
        num_features=num_features_univariate,
        data_class=DATA_CLASS,
    )
    trained_model = train_model(
        dataset=selected_features,
        data_class=DATA_CLASS,
    )
    prediction = test_model(
        model_ser=trained_model,
        inference_point=retrieve_online_node.o0,
    )

    return prediction


if __name__ == "__main__":
    print(f"{feast_workflow()}")
