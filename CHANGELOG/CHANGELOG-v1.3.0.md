# Flyte v1.3.0

The main features of this 1.3 release are

* Databricks support as part of the Spark plugin
* New Helm chart that offers a simpler deployment using just one Flyte service
* Signaling/gate node support (human in the loop tasks)
* User documentation support (backend and flytekit only, limited types)

The latter two are pending some work in Flyte console, they will be piped through fully by the end of Q1. Support for setting and approving gate nodes is supported in `FlyteRemote` however, only a limited set of types can be passed in.

## Notes
There are a couple things to point out with this release.

### Caching on Structured Dataset
Please take a look at the flytekit [PR notes](https://github.com/flyteorg/flytekit/pull/1159) for more information but if you haven't bumped Propeller to version v1.1.36 (aka Flyte v1.2) or later, tasks that take as input a dataframe or a structured dataset type, that are cached, will trigger a cache miss. If you've upgraded Propeller, it will not.

### Flytekit Remote Types
In the `FlyteRemote` experience, fetched [tasks](https://github.com/flyteorg/flytekit/blob/0585b1394c6a6a90a35e8a337bc079bead6f7eb2/flytekit/remote/entities.py#L35) and [workflows](https://github.com/flyteorg/flytekit/blob/0585b1394c6a6a90a35e8a337bc079bead6f7eb2/flytekit/remote/entities.py#L478) will now be based on their respective "spec" classes in the IDL ([task](https://github.com/flyteorg/flyteidl/blob/fd05a1329597230c352372c5948fc1bd5af48b44/protos/flyteidl/admin/task.proto#L55)/[wf](https://github.com/flyteorg/flyteidl/blob/fd05a1329597230c352372c5948fc1bd5af48b44/protos/flyteidl/admin/workflow.proto#L54)) rather than the template. The spec messages are a superset of the template messages so no information is lost.  If you have code that was accessing elements of the templates directly however, these will need to be updated.

## Usage Overview
### Databricks
Please refer to the [documentation](https://docs.flyte.org/en/latest/deployment/plugin_setup/webapi/databricks.html#deployment-plugin-setup-webapi-databricks) for setting up Databricks.
Databricks is a subclass of the Spark task configuration so you'll be able to use the new class in place of the more general `Spark` configuration.

```python
from flytekitplugins.spark import Databricks
@task(
    task_config=Databricks(
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
                    "instance_profile_arn": "arn:aws:iam::1237657460:instance-profile/databricks-s3-role",
                },
                "num_workers": 4,
            },
            "timeout_seconds": 3600,
            "max_retries": 1,
        }
    ))
```

### New Deployment Type
A couple releases ago, we introduced a new Flyte [executable](https://github.com/flyteorg/flyte/blob/master/cmd/main.go) that combined all the functionality of Flyte's backend into one command. This simplifies the deployment in that only one image needs to run now.  This approach is now our recommended way for new comers to the project to install and administer Flyte and there is a [new Helm chart](https://github.com/flyteorg/flyte/tree/master/charts/flyte-binary) also. Documentation has been updated to take this into account. For new installations of Flyte, clusters that do not already have the `flyte-core` or `flyte` charts installed, users can
```bash
helm install flyte-server flyteorg/flyte-binary --namespace flyte --values your_values.yaml
```

#### New local demo environment
Users may have noticed that the environment provided by `flytectl demo start` has also been updated to use this new style of deployment, and internally now installs this new Helm chart. The demo cluster now also exposes an internal docker registry on port `30000`. That is, with the new demo cluster up, you can tag and push to `localhost:30000/yourimage:tag123` and the image will be accessible to the internal Docker daemon. The web interface is still at `localhost:30080`, Postgres has been moved to `30001` and the Minio API (not web server) has been moved to `30002`.

### Human-in-the-loop Workflows
Users can now insert sleeps, approval, and input requests, in the form of gate nodes. Check out one of our earlier [issues](https://github.com/flyteorg/flyte/issues/208) for background information.

```python
from flytekit import wait_for_input, approve, sleep

@workflow
def mainwf(a: int):
    x = t1(a=a)
    s1 = wait_for_input("signal-name", timeout=timedelta(hours=1), expected_type=bool)
    s2 = wait_for_input("signal name 2", timeout=timedelta(hours=2), expected_type=int)
    z = t1(a=5)
    zzz = sleep(timedelta(seconds=10))
    y = t2(a=s2)
    q = t2(a=approve(y, "approvalfory", timeout=timedelta(hours=2)))
    x >> s1
    s1 >> z
    z >> zzz
    ...
```

These also work inside `@dynamic` tasks. Interacting with signals from flytekit's remote experience looks like
```python
from flytekit.remote.remote import FlyteRemote
from flytekit.configuration import Config
r = FlyteRemote(
    Config.auto(config_file="/Users/ytong/.flyte/dev.yaml"),
   default_project="flytesnacks",
   default_domain="development",
)
r.list_signals("atc526g94gmlg4w65dth")
r.set_signal("signal-name", "execidabc123", True)
```

### Overwritten Cached Values on Execution
Users can now configure workflow execution to overwrite the cache. Each task in the workflow execution, regardless of previous cache status, will execute and write cached values - overwriting previous values if necessary. This allows previously corrupted cache values to be corrected without the tedious process of incrementing the `cache_version` and re-registering Flyte workflows / tasks.


### Support for Dask
Users will be able to spawn [Dask](https://www.dask.org/) ephemeral clusters as part of their workflows, similar to the support for [Ray](https://docs.flyte.org/en/latest/flytesnacks/examples/ray_plugin/index.html) and [Spark](https://docs.flyte.org/en/latest/flytesnacks/examples/k8s_spark_plugin/index.html).


## Looking Ahead
In the coming release, we are focusing on...
1. [Out of core plugin](https://hackmd.io/k_hMtUsGTbKl2IksC3IjkA): Make backend plugin scalable and easy to author. No need of code generation, using tools that MLEs and Data Scientists are not accustomed to using.
2. [Performance Observability](https://github.com/flyteorg/flyte/blob/60ba3a603ed1e4fcd47da3ed89dde422faa4d188/rfc/system/2995-performance-benchmarking.md): We have made great progress on exposing both finer-grained runtime metrics and Flytes orchestration metrics. This is important to better understand workflow evaluation performance and mitigate inefficiencies thereof.
