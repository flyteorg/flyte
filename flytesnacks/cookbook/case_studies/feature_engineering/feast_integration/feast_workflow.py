"""
Flyte Pipeline with Feast
-------------------------

In this example, let's build a Flyte pipeline with Feast as the feature store.

Here's the step-by-step process:

* Fetch SQLite3 data as a Pandas DataFrame
* Perform mean-median-imputation
* Build a feature store
* Store the features in an offline store
* Retrieve the features from an offline store
* Perform univariate-feature-selection
* Train a Naive Bayes model
* Load features into an online store
* Fetch a random feature vector for inference
* Generate a prediction

.. note::

    Run ``flytectl demo start`` before running the workflow locally.
"""

# %%
# Import the necessary dependencies.
#
# .. note::
#
#   If running the workflow locally, do an absolute import of the feature engineering tasks.
#
#   .. code-block::
#
#       from feature_eng_tasks import mean_median_imputer, univariate_selection
import logging
import os
from datetime import datetime, timedelta

import boto3
import flytekit
import joblib
import numpy as np
import pandas as pd
from feast import Entity, FeatureStore, FeatureView, Field, FileSource
from feast.infra.offline_stores.file import FileOfflineStoreConfig
from feast.infra.online_stores.sqlite import SqliteOnlineStoreConfig
from feast.repo_config import RepoConfig
from feast.types import Float64, String
from flytekit import FlyteContext, Resources, task, workflow
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.types.file import FlyteFile, JoblibSerializedFile
from flytekit.types.structured import StructuredDataset
from sklearn.model_selection import train_test_split
from sklearn.naive_bayes import GaussianNB

from .feature_eng_tasks import mean_median_imputer, univariate_selection

logger = logging.getLogger(__file__)


# %%
# Set the endpoint, import the feature engineering tasks, and initialize the AWS environment variables.
if os.getenv("DEMO") is None:
    # local execution
    os.environ["FEAST_S3_ENDPOINT_URL"] = ENDPOINT = "http://localhost:30084"
else:
    # execution on demo cluster
    os.environ["FEAST_S3_ENDPOINT_URL"] = ENDPOINT = "http://minio.flyte:9000"

os.environ["AWS_ACCESS_KEY_ID"] = "minio"
os.environ["AWS_SECRET_ACCESS_KEY"] = "miniostorage"

# %%
# Define the necessary data holders.
#
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


# %%
# This task exists just for the demo cluster case as Feast needs an explicit S3 bucket and path.
# This unfortunately makes the workflow less portable.
@task
def create_bucket(
    bucket_name: str, registry_path: str, online_store_path: str
) -> RepoConfig:
    client = boto3.client(
        "s3",
        aws_access_key_id=os.getenv("AWS_ACCESS_KEY_ID"),
        aws_secret_access_key=os.getenv("AWS_SECRET_ACCESS_KEY"),
        use_ssl=False,
        endpoint_url=ENDPOINT,
    )

    try:
        client.create_bucket(Bucket=bucket_name)
    except client.exceptions.BucketAlreadyOwnedByYou:
        logger.info(f"Bucket {bucket_name} has already been created by you.")
        pass

    return RepoConfig(
        registry=f"s3://{bucket_name}/{registry_path}",
        project="horsecolic",
        provider="local",
        offline_store=FileOfflineStoreConfig(),
        online_store=SqliteOnlineStoreConfig(path=online_store_path),
        entity_key_serialization_version=2,
    )


# %%
# Define a ``SQLite3Task`` that fetches data from a data source for feature ingestion.
load_horse_colic_sql = SQLite3Task(
    name="sqlite3.load_horse_colic",
    query_template="select * from data",
    output_schema_type=StructuredDataset,
    task_config=SQLite3Config(
        uri=DATABASE_URI,
        compressed=True,
    ),
)


# %%
# Set the datatype of the timestamp column in the underlying DataFrane to ``datetime``, which would otherwise be a string.
@task
def convert_timestamp_column(
    dataframe: pd.DataFrame, timestamp_column: str
) -> pd.DataFrame:
    dataframe[timestamp_column] = pd.to_datetime(dataframe[timestamp_column])
    return dataframe


# %%
# Define ``store_offline`` and ``load_historical_features`` tasks to store and retrieve the historial features, respectively.
#
# .. list-table:: Decoding the ``Feast`` Jargon
#    :widths: 25 25
#
#    * - ``FeatureStore``
#      - A FeatureStore object is used to define, create, and retrieve features.
#    * - ``Entity``
#      - Represents a collection of entities and associated metadata. It's usually the primary key of your data.
#    * - ``FeatureView``
#      - A FeatureView defines a logical grouping of serve-able features.
#    * - ``FileSource``
#      - File data sources allow for the retrieval of historical feature values from files on disk for building training datasets, as well as for materializing features into an online store.
#    * - ``apply()``
#      - Registers objects to metadata store and updates the related infrastructure.
#    * - ``get_historical_features()``
#      - Enriches an entity dataframe with historical feature values for either training or batch scoring.
#
# .. note::
#
#   The Feast feature store is mutable, so be careful, as Flyte workflows can be highly concurrent!
#   TODO: use postgres db as the registry to support concurrent writes.
@task(limits=Resources(mem="400Mi"))
def store_offline(repo_config: RepoConfig, dataframe: StructuredDataset) -> FlyteFile:
    horse_colic_entity = Entity(name="Hospital Number")

    ctx = flytekit.current_context()
    data_dir = os.path.join(ctx.working_directory, "parquet-data")
    os.makedirs(data_dir, exist_ok=True)

    FlyteContext.current_context().file_access.get_data(
        dataframe._literal_sd.uri + "/00000",
        dataframe._literal_sd.uri + ".parquet",
    )

    horse_colic_feature_view = FeatureView(
        name="horse_colic_stats",
        entities=[horse_colic_entity],
        schema=[
            Field(name="rectal temperature", dtype=Float64),
            Field(name="total protein", dtype=Float64),
            Field(name="peripheral pulse", dtype=Float64),
            Field(name="surgical lesion", dtype=String),
            Field(name="abdominal distension", dtype=Float64),
            Field(name="nasogastric tube", dtype=String),
            Field(name="outcome", dtype=String),
            Field(name="packed cell volume", dtype=Float64),
            Field(name="nasogastric reflux PH", dtype=Float64),
        ],
        source=FileSource(
            path=dataframe._literal_sd.uri + ".parquet",
            timestamp_field="timestamp",
            s3_endpoint_override=ENDPOINT,
        ),
        ttl=timedelta(days=1),
    )

    # ingest the data into Feast
    FeatureStore(config=repo_config).apply(
        [horse_colic_entity, horse_colic_feature_view]
    )

    return FlyteFile(path=repo_config.online_store.path)


@task(limits=Resources(mem="400Mi"))
def load_historical_features(repo_config: RepoConfig) -> pd.DataFrame:
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

    historical_features = (
        FeatureStore(config=repo_config)
        .get_historical_features(entity_df=entity_df, features=FEAST_FEATURES)
        .to_df()
    )  # noqa

    return historical_features


# %%
# Train a Naive Bayes model by fetching features from the offline store and the corresponding data from the parquet file.
@task
def train_model(dataset: pd.DataFrame, data_class: str) -> JoblibSerializedFile:
    x_train, _, y_train, _ = train_test_split(
        dataset[dataset.columns[~dataset.columns.isin([data_class])]],
        dataset[data_class],
        test_size=0.33,
        random_state=42,
    )
    model = GaussianNB()
    model.fit(x_train.values, y_train.values)
    model.feature_names = list(x_train.columns.values)
    fname = "/tmp/model.joblib.dat"
    joblib.dump(model, fname)
    return fname


# %%
# To generate predictions, define ``store_online`` and ``retrieve_online`` tasks.
#
# .. list-table:: Decoding the ``Feast`` Jargon
#    :widths: 25 25
#
#    * - ``materialize()``
#      - Materializes data from offline to an online store.
#    * - ``get_online_features()``
#      - Retrieves the latest online feature data.
#
# .. note::
#
#   One key difference between an online and offline store is that only the latest feature values are stored per entity
#   key in an online store, unlike an offline store where all feature values are stored.
#   Our dataset has two such entries with the same ``Hospital Number`` but different time stamps.
#   Only data point with the latest timestamp will be stored in the online store.
@task(limits=Resources(mem="400Mi"))
def store_online(repo_config: RepoConfig, online_store: FlyteFile) -> FlyteFile:
    # download the online store file and copy the content to the actual online store path
    FlyteContext.current_context().file_access.get_data(
        online_store.download(), repo_config.online_store.path
    )

    FeatureStore(config=repo_config).materialize(
        start_date=datetime.utcnow() - timedelta(days=2000),
        end_date=datetime.utcnow() - timedelta(minutes=10),
    )

    return FlyteFile(path=repo_config.online_store.path)


# %%
# Retrieve a feature vector from the online store.
@task
def retrieve_online(
    repo_config: RepoConfig,
    online_store: FlyteFile,
    data_point: int,
) -> dict:
    # retrieve a data point
    logger.info(f"Hospital Number chosen for inference is: {data_point}")
    entity_rows = [{"Hospital Number": data_point}]

    # download the online store file and copy the content to the actual online store path
    FlyteContext.current_context().file_access.get_data(
        online_store.download(), repo_config.online_store.path
    )

    # get the feature vector
    feature_vector = (
        FeatureStore(config=repo_config)
        .get_online_features(FEAST_FEATURES, entity_rows)
        .to_dict()
    )
    return feature_vector


# %%
# Use the inference data point fetched earlier to generate the prediction.
@task
def predict(model_ser: JoblibSerializedFile, features: dict) -> np.ndarray:
    model = joblib.load(model_ser)
    f_names = model.feature_names

    test_list = []
    for each_name in f_names:
        test_list.append(features[each_name][0])

    if all(test_list):
        prediction = model.predict([[float(x) for x in test_list]])
    else:
        prediction = ["The requested data is not found in the online feature store"]
    return prediction


# %%
# Define a workflow that loads the data from SQLite3 database, does feature engineering, and stores the offline features in a feature store.
@workflow
def featurize(
    repo_config: RepoConfig, imputation_method: str = "mean"
) -> (StructuredDataset, FlyteFile):
    # load parquet file from sqlite task
    df = load_horse_colic_sql()

    # perform mean median imputation
    df = mean_median_imputer(dataframe=df, imputation_method=imputation_method)

    # convert timestamp column from string to datetime
    converted_df = convert_timestamp_column(dataframe=df, timestamp_column="timestamp")

    online_store = store_offline(
        repo_config=repo_config,
        dataframe=converted_df,
    )

    return df, online_store


# %%
# Define a workflow that trains a Naive Bayes model.
@workflow
def trainer(
    df: StructuredDataset, num_features_univariate: int = 7
) -> JoblibSerializedFile:
    # perform univariate feature selection
    selected_features_dataset = univariate_selection(
        dataframe=df,  # noqa
        num_features=num_features_univariate,
        data_class=DATA_CLASS,
    )

    # train the Naive Bayes model
    trained_model = train_model(
        dataset=selected_features_dataset,
        data_class=DATA_CLASS,
    )

    return trained_model


# %%
# Lastly, define a workflow to encapsulate and run the previously defined tasks and workflows.
@workflow
def feast_workflow(
    s3_bucket: str = "feast-integration",
    registry_path: str = "registry.db",
    online_store_path: str = "online.db",
    imputation_method: str = "mean",
    num_features_univariate: int = 7,
) -> (JoblibSerializedFile, np.ndarray):
    repo_config = create_bucket(
        bucket_name=s3_bucket,
        registry_path=registry_path,
        online_store_path=online_store_path,
    )

    # feature engineering and store features in an offline store
    df, online_store = featurize(
        repo_config=repo_config,
        imputation_method=imputation_method,
    )

    # load features from the offline store
    historical_features = load_historical_features(repo_config=repo_config)

    # load features after performing feature engineering
    df >> historical_features

    # train a Naive Bayes model
    model = trainer(
        df=historical_features, num_features_univariate=num_features_univariate
    )

    # materialize features into an online store
    loaded_online_store = store_online(
        repo_config=repo_config,
        online_store=online_store,
    )

    # retrieve features from the online store
    feature_vector = retrieve_online(
        repo_config=repo_config,
        online_store=loaded_online_store,
        data_point=533738,
    )

    # generate a prediction
    prediction = predict(model_ser=model, features=feature_vector)  # noqa

    return model, prediction


if __name__ == "__main__":
    print(f"{feast_workflow()}")

# %%
# You should see prediction against the test input as the workflow output.
