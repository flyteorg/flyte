# UnionRemote (TODO: merge FlyteRemote doc with this one)

`UnionRemote` allows you to programmatically perform certain operations on the Union control plane in a Python runtime environment.

## Creating a `UnionRemote` object

To use `UnionRemote`, install the `unionai` SDK with `pip install unionai`, then add the following import to your code:

```{code-block} python
from unionai.remote import UnionRemote
```

The `UnionRemote` class is the entrypoint for programmatically performing operations in a Python
runtime. It can be initialized by passing in the:

* `Config` object: the parent configuration object that holds all the configuration information to connect to the Flyte backend.
* `default_project`: the default project to use when fetching or executing flyte entities.
* `default_domain`: the default domain to use when fetching or executing flyte entities.
* `file_access`: the file access provider to use for offloading non-literal inputs/outputs.
* `kwargs`: additional arguments that need to be passed to create `SynchronousFlyteClient`.

A `UnionRemote` object can be created in various ways:

### Auto

The `Config` class's `auto` method can be used to automatically
construct the `Config` object.

```{code-block} python
from unionai.remote import UnionRemote
from flytekit.configuration import Config

remote = UnionRemote(config=Config.auto())
```

`auto` also accepts a `config_file` argument, which is the path to the configuration file to use.
The order of precedence that `auto` follows is:

* Finds all the environment variables that match the configuration variables.
* If no environment variables are set, it looks for a configuration file at the path specified by the `config_file` argument.
* If no configuration file is found, it uses the default values.

### Any endpoint

The `Config` class's `for_endpoint` method can be used to
construct the `Config` object to connect to a specific endpoint.

```{code-block} python
from unionai.remote import UnionRemote
from flytekit.configuration import Config

remote = UnionRemote(
    config=Config.for_endpoint(endpoint="flyte.example.net"),
    default_project="flytesnacks",
    default_domain="development",
)
```

The `for_endpoint` method also accepts:

* `insecure`: whether to use insecure connections. Defaults to `False`.
* `data_config`: can be used to configure how data is downloaded or uploaded to a specific blob storage like S3, GCS, etc.
* `config_file`: the path to the configuration file to use.

### Generalized initialization

The `Config` class can be directly used to construct the `Config` object if additional configuration is needed. You can send `configuration.PlatformConfig`, `configuration.DataConfig`,
`configuration.SecretsConfig`, and `configuration.StatsConfig` objects to the `Config` class.

| Config attribute | Description                              |
|------------------|------------------------------------------|
| `PlatformConfig` | Settings to talk to a Flyte backend.     |
| `DataConfig`     | Any data storage specific configuration. |
| `SecretsConfig`  | Configuration for secrets.               |
| `StatsConfig`    | Configuration for sending statsd.        |

For example:

```{code-block} python
from unionai.remote import UnionRemote
from flytekit.configuration import Config, PlatformConfig

remote = UnionRemote(
    config=Config(
        platform=PlatformConfig(
            endpoint="flyte.example.net",
            insecure=False,
            client_id="my-client-id",
            client_credentials_secret="my-client-secret",
            auth_mode="client_credentials",
        ),
        secrets=SecretsConfig(default_dir="/etc/secrets"),
    )
)
```

## Fetching entities

### Fetching tasks, workflows, launch plans, and executions

```{code-block} python
my_task = remote.fetch_task(name="my_task", version="v1")
my_workflow = remote.fetch_workflow(name="my_workflow", version="v1")
my_launch_plan = remote.fetch_launch_plan(name="my_launch_plan", version="v1")
my_execution = remote.fetch_execution(name="my_execution")
```

`project` and `domain` can also be specified in all the `fetch_*` calls.
If not specified, the default values given during the creation of the `UnionRemote`
object will be used.

The following is an example that fetches tasks and creates a workflow:

```{code-block} python
from flytekit import workflow

task_1 = remote.fetch_task(name="core.basic.hello_world.say_hello", version="v1")
task_2 = remote.fetch_task(
    name="core.basic.lp.greet",
    version="v13",
    project="flytesnacks",
    domain="development",
)

@workflow
def my_remote_wf(name: str) -> int:
    return task_2(task_1(name=name))
```

Another example that dynamically creates a launch plan for the `my_remote_wf` workflow:

```{code-block} python
from flytekit import LaunchPlan

my_workflow = remote.fetch_workflow(
    name="my_workflow", version="v1", project="flytesnacks", domain="development"
)
launch_plan = LaunchPlan.get_or_create(name="my_launch_plan", workflow=my_workflow)
```

### Fetching artifacts

Each artifact version has a unique URI of the form `flyte://<organization>/<project>/<domain>/<artifact_name>@<artifact_version>`.

To fetch a single artifact, pass this URI to the `get_artifact` method:

```{code-block} python
from flytekit.configuration import Config, PlatformConfig, AuthType
from unionai.remote import UnionRemote

remote = UnionRemote(Config.auto().with_params(platform=PlatformConfig(
    endpoint="example.domain.unionai.cloud", # replace with your domain
    insecure=False)))

remote.get_artifact("flyte://<organization>/<project>/<domain>/<artifact_name>@<artifact_version>") # replace with your artifact URI
```

To dynamically query for artifacts, you can pass an artifact URI with a query to the `get_artifact` method, adding as many partition key-value pairs as you wish to filter on. The query will retrieve the latest artifact version that matches the partition-based filters:

```{code-block} python
from flytekit.configuration import Config, PlatformConfig, AuthType
from unionai.remote import UnionRemote

remote = UnionRemote(Config.auto().with_params(platform=PlatformConfig(
    endpoint="example.domain.unionai.cloud", # replace with your domain
    insecure=False)))

remote.get_artifact("flyte://<organization>/<project>/<domain>/<artifact_name>?<partition_key1>=<partition_value1>&...") # replace with your artifact query
```

## Creating entities

### Creating artifacts

To create an artifact with `UnionRemote`, declare the artifact, then pass it to the `create_artifact` method:

```{code-block} python
from flytekit.configuration import Config, PlatformConfig, AuthType
from unionai.remote import UnionRemote
from flytekit.core.artifact import Artifact

remote = UnionRemote(Config.auto().with_params(platform=PlatformConfig(
    endpoint="example.domain.unionai.cloud", # replace with your domain
    insecure=False)))

BasicArtifact = Artifact(name="my_basic_artifact")
remote.create_artifact(BasicArtifact)
```

For the full list of parameters for the Artifact class, see the [Flytekit artifact documentation](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.Artifact.html)

:::{note}
If you want to create a new version of an existing artifact, be sure to set the `version` parameter. Without it, attempting to recreate the same artifact will result in an error.
:::

## Registering entities

Tasks, workflows, and launch plans can be registered using `UnionRemote`:

```{code-block} python
from flytekit.configuration import SerializationSettings

some_entity = ...
my_task = remote.register_task(
    entity=some_entity,
    serialization_settings=SerializationSettings(image_config=None),
    version="v1",
)
my_workflow = remote.register_workflow(
    entity=some_entity,
    serialization_settings=SerializationSettings(image_config=None),
    version="v1",
)
my_launch_plan = remote.register_launch_plan(entity=some_entity, version="v1")
```

* `entity`: the entity to register.
* `version`: the version that will be used to register. If not specified, the version used in serialization settings will be used.
* `serialization_settings`: the serialization settings to use. Refer to `configuration.SerializationSettings` to know all the acceptable parameters.

All the additional parameters which can be sent to the `register_*` methods can be found in the documentation for the corresponding method:
`register_task`, `register_workflow`,
and `register_launch_plan`.

The `configuration.SerializationSettings` class accepts `configuration.ImageConfig` which
holds the available images to use for the registration.

The following example showcases how to register a workflow using an existing image if the workflow is created locally:

```{code-block} python
from flytekit.configuration import ImageConfig

img = ImageConfig.from_images(
    "docker.io/xyz:latest", {"spark": "docker.io/spark:latest"}
)
wf2 = remote.register_workflow(
    my_remote_wf,
    serialization_settings=SerializationSettings(image_config=img),
    version="v1",
)
```

## Executing entities

You can execute a task, workflow, or launch plan using the `execute` method
which returns a `FlyteWorkflowExecution` object:

```{code-block} python
some_entity = ...  # one of FlyteTask, FlyteWorkflow, or FlyteLaunchPlan
execution = remote.execute(
    some_entity, inputs={...}, execution_name="my_execution", wait=True
)
```

* `inputs`: the inputs to the entity.
* `execution_name`: the name of the execution. This is useful to avoid de-duplication of executions.
* `wait`: synchronously wait for the execution to complete.

Additional arguments include:

* `project`: the project on which to execute the entity.
* `domain`: the domain on which to execute the entity.
* `type_hints`: a dictionary mapping Python types to their corresponding Flyte types.
* `options`: options can be configured for a launch plan during registration or overridden during execution. Refer to `Options` to know all the acceptable parameters.

The following is an example demonstrating how to use the `Options` class to configure a Flyte entity:

```{code-block} python
from flytekit.models.common import AuthRole, Labels
from flytekit.tools.translator import Options

some_entity = ...  # one of FlyteTask, FlyteWorkflow, or FlyteLaunchPlan
execution = remote.execute(
    some_entity,
    inputs={...},
    execution_name="my_execution",
    wait=True,
    options=Options(
        raw_data_prefix="s3://my-bucket/my-prefix",
        auth_role=AuthRole(assumable_iam_role="my-role"),
        labels=Labels({"my-label": "my-value"}),
    ),
)
```

## Retrieving & inspecting executions

After an execution is completed, you can retrieve the execution using the `fetch_execution` method. The fetched execution can be used to retrieve the inputs and outputs of an execution:

```{code-block} python
execution = remote.fetch_execution(
    name="fb22e306a0d91e1c6000", project="flytesnacks", domain="development"
)
input_keys = execution.inputs.keys()
output_keys = execution.outputs.keys()
```

The `inputs` and `outputs` correspond to the top-level execution or the workflow itself.

To fetch a specific output, say, a model file:

```{code-block} python
model_file = execution.outputs["model_file"]
with open(model_file) as f:
    # use mode
    ...
```

You can use `sync` to sync the entity object's state with the remote state during the execution run:

```{code-block} python
synced_execution = remote.sync(execution, sync_nodes=True)
node_keys = synced_execution.node_executions.keys()
```

:::{note}
During the sync, you may encounter an error of `Received message larger than max (xxx vs. 4194304)` if the message size is too large. In that case, reach out to the Union team.
:::

`node_executions` will fetch all the underlying node executions recursively.

To fetch output of a specific node execution:

```{code-block} python
node_execution_output = synced_execution.node_executions["n1"].outputs["model_file"]
```

Node here can correspond to a task, workflow, or branch node.

## Listing entities

To list recent executions, use the `recent_executions` method:

```{code-block} python
recent_executions = remote.recent_executions(project="flytesnacks", domain="development", limit=10)
```

The `limit` parameter is optional and defaults to 100.

To list tasks by version, use the `UnionRemote.list_tasks_by_version` method.

```{code-block} python
tasks = remote.list_tasks_by_version(project="flytesnacks", domain="development", version="v1")
```

## Terminating an execution

To terminate an execution, use the `terminate` method:

```{code-block} python
execution = remote.fetch_execution(name="fb22e306a0d91e1c6000", project="flytesnacks", domain="development")
remote.terminate(execution, cause="Code needs to be updated")
```
