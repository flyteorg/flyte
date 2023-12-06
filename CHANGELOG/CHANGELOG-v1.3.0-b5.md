
# Flyte v1.3.0-b5 Changelog

This pulls in Databricks support. Please see the [GH issue](https://github.com/flyteorg/flyte/issues/3173) for a listing of the relevant PRs.
There are other changes included in this beta release as well.

## Try it out locally
You can try out these changes "locally", as they've been included in the `flytectl demo` image for this beta release, but since the demo cluster is meant
to be an isolated, local-only cluster, you'll have to make some changes to get it to talk to a live databricks account. You'll also need to configure
access to a real S3 bucket (as opposed to Minio, which is what the demo local cluster typically relies on).

### S3 Setup
Follow the [AWS instructions](https://docs.aws.amazon.com/powershell/latest/userguide/pstools-appendix-sign-up.html) for generating access and secret
keys that can be used to hit your S3 bucket of choice.

### Flyte Demo Cluster
#### Starting the cluster
Run flytectl demo start with the image argument

```bash
flytectl demo start --image ghcr.io/flyteorg/flyte-sandbox-bundled:sha-e240038bea1f3bdfe2092823688d35dc78fb6e6b
```

#### Configure the Demo Cluster
1. Update the Flyte configmap
    ```bash
    kubectl -n flyte edit cm sandbox-flyte-binary-config
    ```
  1. Update the `003-storage.yaml` section
     Make the storage section look like the following. You should update the propeller `rawoutput-prefix` setting as well.
      ```
      storage:
        type: s3
        container: "your-bucket"
        stow:
          kind: s3
          config:
            access_key_id: AKIAYOURKEY
            auth_type: accesskey
            secret_key: YOUR+SECRET
            disable_ssl: true
            region: us-east-2
      ```
  1. Update the `010-inline-config.yaml` section
     1. Under the existing `plugins` section, as a sibling to `k8s`, add
        ```
        databricks:
          databricksInstance: dbc-abc-123.cloud.databricks.com
          entrypointFile: dbfs:///FileStore/tables/entrypoint.py
        ```
     2. In the `k8s` section, update the `default-env-vars` section
        ```
        - FLYTE_AWS_ACCESS_KEY_ID: AKIAYOURKEY
        - AWS_DEFAULT_REGION: us-east-2
        - FLYTE_AWS_SECRET_ACCESS_KEY: YOUR+SECRET
        ```
         These are the same values as in the storage section above.

     3. Add in an section for data proxy
        ```
        remoteData:
           region: us-east-2
           scheme: aws
           signedUrls:
             durationMinutes: 3
        ```
     4. Enable databricks plugin
        ```shell
        task-plugins:
          default-for-task-types:
            container: container
            container_array: k8s-array
            sidecar: sidecar
            ray: ray
            spark: databricks
        enabled-plugins:
          - container
          - databricks
          - ray
          - sidecar
          - k8s-array

        ```

1. Update the Flyte deployment
   ```
   kubectl -n flyte edit deploy sandbox-flyte-binary
   ```
   
   Add an environment variable for your databricks token to the flyte pod
   ```
      - name: FLYTE_SECRET_FLYTE_DATABRICKS_API_TOKEN
        value: dapixyzxyzxyz
    ```
    
1. Restart the deployment
   ```
   kubectl -n flyte rollout restart deploy sandbox-flyte-binary
   ```

### Databricks Code
You'll need to upload an [entrypoint](https://gist.github.com/pingsutw/482e7f0134414dac437500344bac5134) file to your dbfs (or S3). This is the referenced gist from the primary [Databricks plugin documentation](https://github.com/flyteorg/flyte/blob/master/docs/deployment/plugin_setup/webapi/databricks.rst) as well, which currently only covers the `flyte-core` Helm chart installation.


### User Code
1. a sample py file that has a simple spark task.

```python
import datetime
import random
from operator import add

import flytekit
from flytekit import Resources, task, workflow

from flytekitplugins.spark.task import Databricks

# %%
# You can create a Spark task by adding a ``@task(task_config=Spark(...)...)`` decorator.
# ``spark_conf`` can have configuration options that are typically used when configuring a Spark cluster.
# To run a Spark job on Databricks platform, just add Databricks config to the task config. Databricks Config is the same as the databricks job request.
# Refer to `Databricks job request <https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure>`__
@task(
    task_config=Databricks(
        # this configuration is applied to the spark cluster
        spark_conf={
            "spark.driver.memory": "1000M",
            "spark.executor.memory": "1000M",
            "spark.executor.cores": "1",
            "spark.executor.instances": "2",
            "spark.driver.cores": "1",
        },
        databricks_conf={
           "run_name": "flytekit databricks plugin example",
           "new_cluster": {
               "spark_version": "11.0.x-scala2.12",
               "node_type_id": "r3.xlarge",
               "aws_attributes": {
                   "availability": "ON_DEMAND",
                   "instance_profile_arn": "arn:aws:iam::123123123:instance-profile/databricks-s3-role",
               },
               "num_workers": 4,
           },
           "timeout_seconds": 3600,
           "max_retries": 1,
       }
    ),
    limits=Resources(mem="2000M"),
    cache_version="1",
)
def hello_spark(partitions: int) -> float:
    print("Starting Spark with Partitions: {}".format(partitions))

    n = 100000 * partitions
    sess = flytekit.current_context().spark_session
    count = (
        sess.sparkContext.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
    )
    pi_val = 4.0 * count / n
    print("Pi val is :{}".format(pi_val))
    return pi_val

# Let's define a function on which the map-reduce operation is called within the Spark cluster.
def f(_):
    x = random.random() * 2 - 1
    y = random.random() * 2 - 1
    return 1 if x**2 + y**2 <= 1 else 0


# Next, we define a regular Flyte task which will not execute on the Spark cluster.
@task(cache_version="1")
def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))
    return 1


# This workflow shows that a spark task and any python function (or a Flyte task) can be chained together as long as they match the parameter specifications.
@workflow
def my_databricks_job(triggered_date: datetime.datetime) -> float:
    """
    Using the workflow is still as any other workflow. As image is a property of the task, the workflow does not care
    about how the image is configured.
    """
    pi = hello_spark(partitions=50)
    print_every_time(value_to_print=pi, date_triggered=triggered_date)
    return pi


# Workflows with spark tasks can be executed locally. Some aspects of spark, like links to :ref:`Hive <Hive>` meta stores may not work, but these are limitations of using Spark and are not introduced by Flyte.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running To run a Spark job on Databricks platform(triggered_date=datetime.datetime.now()){my_databricks_job(triggered_date=datetime.datetime.now())}"
    )

```
2. Build a custom image for spark clusters
```dockerfile
FROM databricksruntime/standard:11.3-LTS
ENV PATH $PATH:/databricks/python3/bin
ENV PYTHONPATH /databricks/driver
# Install custom package
RUN /databricks/python3/bin/pip install awscli flytekitplugins-spark==v1.3.0b5

# Copy the actual code
COPY ./ /databricks/driver
```
3. image building command if necessary.
```shell
docker build -t pingsutw/databricks:test -f Dockerfile .
```
4. pyflyte command to register the flyte workflow and task.

```shell
pyflyte --config ~/.flyte/config-sandbox.yaml register --destination-dir . --image pingsutw/databricks:test databricks.py
```

