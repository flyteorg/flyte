"""
Chain Flyte Entities
--------------------

Data passing between tasks or workflows need not necessarily happen through parameters.
In such a case, if you want to explicitly construct the dependency, flytekit provides a mechanism to chain Flyte entities using the ``>>`` operator.

Tasks
^^^^^

Let's enforce an order for ``read()`` to happen after ``write()``, and for ``write()`` to happen after ``create_bucket()``.

.. note::
    To run the example locally, spin up the demo cluster using ``flytectl demo start``.
"""

# %%
# Import the necessary dependencies.
import logging
from io import StringIO

import boto3
import pandas as pd
from flytekit import task, workflow

logger = logging.getLogger(__file__)

CSV_FILE = "iris.csv"
BUCKET_NAME = "chain-flyte-entities"


BOTO3_CLIENT = boto3.client(
    "s3",
    aws_access_key_id="minio",
    aws_secret_access_key="miniostorage",
    use_ssl=False,
    endpoint_url="http://minio.flyte:9000",
)

# %%
# Create an s3 bucket.
# This task exists just for the sandbox case.
@task(cache=True, cache_version="1.0")
def create_bucket():
    try:
        BOTO3_CLIENT.create_bucket(Bucket=BUCKET_NAME)
    except BOTO3_CLIENT.exceptions.BucketAlreadyOwnedByYou:
        logger.info(f"Bucket {BUCKET_NAME} has already been created by you.")
        pass


# %%
# Define a ``read()`` task to read from the s3 bucket.
@task
def read() -> pd.DataFrame:
    data = pd.read_csv(
        BOTO3_CLIENT.get_object(Bucket=BUCKET_NAME, Key=CSV_FILE)["Body"]
    )
    return data


# %%
# Define a ``write()`` task to write the dataframe to a CSV file in the s3 bucket.
@task(cache=True, cache_version="1.0")
def write():
    df = pd.DataFrame(  # noqa : F841
        data={
            "sepal_length": [5.3],
            "sepal_width": [3.8],
            "petal_length": [0.1],
            "petal_width": [0.3],
            "species": ["setosa"],
        }
    )
    csv_buffer = StringIO()
    df.to_csv(csv_buffer)
    BOTO3_CLIENT.put_object(
        Body=csv_buffer.getvalue(), Bucket=BUCKET_NAME, Key=CSV_FILE
    )


# %%
# We want to enforce an order here: ``create_bucket()`` followed by ``write()`` followed by ``read()``.
# Since no data-passing happens between the tasks, use ``>>`` operator on the tasks.
@workflow
def chain_tasks_wf() -> pd.DataFrame:
    create_bucket_promise = create_bucket()
    write_promise = write()
    read_promise = read()

    create_bucket_promise >> write_promise
    write_promise >> read_promise

    return read_promise


# %%
# Chain SubWorkflows
# ^^^^^^^^^^^^^^^^^^
#
# Similar to tasks, you can chain :ref:`subworkflows <subworkflows>`.
@workflow
def write_sub_workflow():
    write()


@workflow
def read_sub_workflow() -> pd.DataFrame:
    return read()


# %%
# Use ``>>`` to chain the subworkflows.
@workflow
def chain_workflows_wf() -> pd.DataFrame:
    create_bucket_promise = create_bucket()
    write_sub_wf = write_sub_workflow()
    read_sub_wf = read_sub_workflow()

    create_bucket_promise >> write_sub_wf
    write_sub_wf >> read_sub_wf

    return read_sub_wf


#%%
# Run the workflows locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running chain_tasks_wf()... {chain_tasks_wf()}")
    print(f"Running chain_workflows_wf()... {chain_workflows_wf()}")
