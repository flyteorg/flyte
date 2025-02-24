# Flyte 1.15.0 release notes

## Added

### Streaming Decks ([#2779](https://github.com/flyteorg/flytekit/pull/2779))
[Decks](https://docs.flyte.org/en/latest/user_guide/development_lifecycle/decks.html) is a Flyte feature that provides customizable visualization tools for tasks, offering both default and user-defined insights
For use cases like Hyperparameter Optimization, seeing the metrics, data, logs, and links on the Flyte Deck while the task is running is extremely useful. However, the Flyte Deck HTML is uploaded to the remote bucket only after the task succeeds.
With this change, flytepropeller now sends the Deck URI to flyteadmin once the task starts running. This feature is also helpful for debugging, identifying your task’s dependencies and even looking at the source code that is being executed.

![Screenshot 2025-02-14 at 2.10.27 PM](https://hackmd.io/_uploads/S14BHMTKke.png)

### Define shared memory in the task definition

Several ML libraries make use of shared memory, so this is now easier to do in Flyte. This new argument can be either a bool or a string value, the former means "memory backed volumes are sized to node allocatable memory", whereas the latter means that a volume of that size will be defined in the podspec used to run the task. For example, the simple case:

```python=
import os
from flytekit import task

@task(shared_memory=True)
def t1() -> bool:
    # /dev/shm is available in the task
    ...
```

or a more complex case:

```python=
import os
from flytekit import task

@task(shared_memory="10Gi")
def t1() -> bool:
    # /dev/shm of size 10Gi is available in the task
    ...
```

### Progress bars to help visualize package creation and upload

Flytekit includes now progress tracking, including package creation/compression progress visualization.

Here's how the progress bar looks like:

![image](https://hackmd.io/_uploads/BysASVTt1e.png)

And after files are compressed and uploaded:

![image](https://hackmd.io/_uploads/H1FnUEpYkl.png)



### Debug @dynamic tasks

Advancing the debug story we've been chipping away, we now support debugging `@dynamic` tasks.

## Changed
### Better Secrets handling ([#3048](https://github.com/flyteorg/flytekit/pull/3048))
Flyte enables tasks to requests secrets managed by Kubernetes or by external systems. 
If you want to map and mount a Kubernetes secret with an environment variable, flytekit required you to export an environment variable as `_FSEC_<GROUP>_<GROUP_VERSION>_<KEY>` with `GROUP` corresponding to the Secret name, and `KEY` being the actual key to retrieve from the secret. This was inconvenient and too verbose. 
Starting with this release, you can easily configure secrets from environment variables::

```python
@task(secret_requests=[Secret(..., env_var="MY_ENVVAR", mount_requirement=Secret.MountType.ENV_VAR)
def hello():
    ...
```
Or from files:

```python
@task(secret_requests=[Secret(..., env_var="MY_TOKEN_PATH", mount_requirement=Secret.MountType.FILE)
def hello():
    ...
```
### Override Pod template using .with_overrides ([#2981](https://github.com/flyteorg/flytekit/pull/2981))

Flyte lets you specify default resources for task executions to enable deterministic scheduling behavior.
The `.with_overrides` method allows you to override the global resource defaults dynamically:

```python
@workflow
def my_pipeline(x: typing.List[int]) -> int:
    return square_1(x=count_unique_numbers_1(x=x)).with_overrides(limits=Resources(cpu="6", mem="500Mi"))
```

This method has been available for Resources only. Starting with this release, you can now override the Pod template too.

Similar to Resources, you can instruct `flytepropeller` to use a custom Pod template for the task executions, setting the template at different levels: from namespace-wide to the task configuration. 
Starting with this release, you can override dynamically the Pod template when the task is invoked in the workflow:

```python
@task
def say_hello() -> str:
    return "Hello, World!"

@workflow
def hello_world_wf() -> str:
    res = say_hello().with_overrides(limits=Resources(cpu="2", mem="600Mi"),pod_template=PodTemplate(
        primary_container_name="primary-nelson",
        labels={"lKeyA": "lValA", "lKeyB": "lValB"},
        annotations={"aKeyA": "aValA", "aKeyB": "aValB"},
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="primary-nelson",
                    image="arbaobao/flyte-test-images:pythonpath5",
                    env=[V1EnvVar(name="eKeyC", value="eValC"), V1EnvVar(name="eKeyD", value="eValD")],
                ),
                V1Container(
                    name="primary-nelson2",
                    image="arbaobao/flyte-test-images:pythonpath5",
                    env=[V1EnvVar(name="eKeyC", value="eValC"), V1EnvVar(name="eKeyD", value="eValD")],
                ),
            ],
        )
    ))
    return res
```

This change provides more flexible pod configuration management while maintaining compatibility with existing resource override patterns.
### Introducing Caching Policies ([#3129](https://github.com/flyteorg/flytekit/pull/3129))
Flyte provides robust caching mechanisms to optimize workflow performance. Two key features users can leverage today include:
#### a. Task Output Caching with Deterministic Inputs

Users can enable task-level caching by setting `cache=True` in the `@task` decorator. This ensures that tasks with identical inputs reuse previously computed outputs, saving time and resources.

Example:
```python
@task(cache=True, cache_version="1.0")  
def square(n: int) -> int:  
    return n * n 
```
#### b. Custom Hashing for Offloaded Objects (e.g., DataFrames)

Flyte allows users to define custom hash functions for non-deterministic objects (like pandas DataFrames) using Annotated types and HashMethod. This ensures caching works even for complex data types.

Example:

```python
def hash_pandas_dataframe(df: pd.DataFrame) -> str:  
    return str(pd.util.hash_pandas_object(df))  

@task  
def data_loader(cache=True, cache_version="1.0",
) -> Annotated[pd.DataFrame, HashMethod(hash_pandas_dataframe)]:  
    return pd.DataFrame(...) 
```
This release brings a refactoring of the Flyte’s caching system into a unified `Cache` object, deprecating legacy parameters and adding advanced features:

##### 1. Unified Cache Configuration Object

All caching parameters (existing and new) are now grouped under a single Cache class:
```python
class Cache:  
    version: Optional[str] = None # Replaces `cache_version`  
    serialize: bool = False      # Replaces `cache_serialize`  
    ignored_inputs: Union[Tuple[str, ...], str] = ()  # New: Exclude inputs from cache key  
    salt: str = ""           New: Add unique salt to cache key  
    policies: Optional[Union[List[CachePolicy], CachePolicy]] = None  # New: Dynamic versioning 
```
> **Deprecation Notice**: Legacy parameters (e.g., `cache_version`, `cache_serialize`) are deprecated but remain supported.

##### 2. Advanced Cache Key Customization

- Cache Policies: Define dynamic versioning logic via CachePolicy protocols.

- Policies generate version strings through the `get_version` method. Those are then concatenated and hashed (`SHA-256`) to form the final cache key. 

> In case a Cache object is defined, we require the definition of policies 

- Salt: Add a unique string to the cache key to avoid collisions (e.g., differentiate between teams sharing the same task).

- Ignored Inputs: Exclude specific inputs (e.g., `request_id`) from cache key calculation using `ignored_inputs`.


#### c. Migration Path

Existing code using `cache=True` or `cache_version` will continue to work, but users are encouraged to migrate to the `Cache` object:

```python
@task(cache=Cache(version="1.0", serialize=True, ignored_inputs=("debug_mode",)))  
def my_task(...) -> ...: 
```

### Faster GitIgnore checks

Algorithmic speed-ups are always welcome and this is exactly what happened in https://github.com/flyteorg/flytekit/pull/3007. Enjoy the massive speedup!


## Full changelog
* docs: Update datatype mapping for polars Dataframe and LazyFrame by @HansBambel in https://github.com/flyteorg/flyte/pull/6109
* Add custom info field to ExternalResourceInfo by @hamersaw in https://github.com/flyteorg/flyte/pull/6115
* improve error message in the agent plugin by @pingsutw in https://github.com/flyteorg/flyte/pull/6114
* upstream ArrayNode DataMode + BindingData LiteralOffloadedMetadata idl changes by @pvditt in https://github.com/flyteorg/flyte/pull/6074
* Pass CustomInfo metadata through ArrayNode ExternalResourceInfo (#591) by @hamersaw in https://github.com/flyteorg/flyte/pull/6116
* [Docs] Add new page for jupyter notebook interactive mode by @Mecoli1219 in https://github.com/flyteorg/flyte/pull/6036
* Use child dir for branch taken by @andrewwdye in https://github.com/flyteorg/flyte/pull/6120
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6047
* Improve resource manager test coverage by @luckyarthur in https://github.com/flyteorg/flyte/pull/5973
* Add nodeName template var and update doc by @ketian-indeed in https://github.com/flyteorg/flyte/pull/6088
* Allow flytepropeller manager to have different resource request from flyte propeller by @cpaulik in https://github.com/flyteorg/flyte/pull/5974
* Unpin flytectl from single-binary tests by @eapolinario in https://github.com/flyteorg/flyte/pull/6113
* Add new `sandbox-bundled-offloaded-functional-tests` for offloading literals unit test by @Mecoli1219 in https://github.com/flyteorg/flyte/pull/6111
* chore: fix some function names in comment by @loselarry in https://github.com/flyteorg/flyte/pull/6122
* Flyteadmin digest comparison should rely on database semantics by @popojk in https://github.com/flyteorg/flyte/pull/6058
* Upgrade golang.org/x/crypto lib to address vulnerability by @katrogan in https://github.com/flyteorg/flyte/pull/6133
* chore: using domain-qualified finalizers by @trutx in https://github.com/flyteorg/flyte/pull/6023
* Update contribution guides  by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6150
* [Docs] Align fd code lines with Flytesnacks examples by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/6149
* docs: update URL in local workflow execution guide to reflect project… by @vincent0426 in https://github.com/flyteorg/flyte/pull/6137
* Print K8s Resolver Event for Agent Grpc request debugging by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6146
* Fix: Upgrade stow to fix listing error files by @fg91 in https://github.com/flyteorg/flyte/pull/6159
* Add retries to sendgrid emailer (#618) by @katrogan in https://github.com/flyteorg/flyte/pull/6164
* [Bug] correctly set task execution phase for terminal array node by @pvditt in https://github.com/flyteorg/flyte/pull/5136
* Add name to ContainerPort by @pingsutw in https://github.com/flyteorg/flyte/pull/6127
* Add securityContext for individual containers in flyte-binary Deployment by @marrrcin in https://github.com/flyteorg/flyte/pull/6168
* Add env_name to Secrets  by @thomasjpfan in https://github.com/flyteorg/flyte/pull/6160
* Add missing execution-keyed entities for custom filters by @katrogan in https://github.com/flyteorg/flyte/pull/6172
* Minor auto_refresh cleanup by @Sovietaced in https://github.com/flyteorg/flyte/pull/5813
* Add a property bag to node metadata by @EngHabu in https://github.com/flyteorg/flyte/pull/6169
* reserve field in execution metadata by @troychiu in https://github.com/flyteorg/flyte/pull/6174
* New Deck IDL by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6175
* Rename Secret.env_name to Secret.env_var by @thomasjpfan in https://github.com/flyteorg/flyte/pull/6176
* stripping metadata from dataclass types (#626) by @hamersaw in https://github.com/flyteorg/flyte/pull/6177
* [flyteagent] Default Service Config Using Round Robin Mechanism by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6179
* Add instructions on how to yank a flytectl release by @eapolinario in https://github.com/flyteorg/flyte/pull/6105
* chore: fix function name and comment by @dashangcun in https://github.com/flyteorg/flyte/pull/6163
* [Docs] Update flytectl_update_workflow-execution-config.rst by @pranshustuff in https://github.com/flyteorg/flyte/pull/5865
* Add option to disable synchronization on creation for auto refresh cache by @Sovietaced in https://github.com/flyteorg/flyte/pull/5940
* Make it possible to configure flyte with an externally managed Secret by @mhotan in https://github.com/flyteorg/flyte/pull/5700
* [Docs] Update auth_setup.rst by @pranshustuff in https://github.com/flyteorg/flyte/pull/5863
* [flytepropeller][flyteadmin] Streaming Decks V2 by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6053
* [Docs] Add the K8s agent set up document by @shuyingliang in https://github.com/flyteorg/flyte/pull/6184
* Configure ray pod runtime class based on custom pod specs by @Sovietaced in https://github.com/flyteorg/flyte/pull/6199
* Use generated mocks in flyteadmin by @popojk in https://github.com/flyteorg/flyte/pull/6197
* Define env var for docsearch API key in dev-docs build script by @eapolinario in https://github.com/flyteorg/flyte/pull/6187
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6130
* Support overriding task pod_template via with_overrides by @arbaobao in https://github.com/flyteorg/flyte/pull/6118
* Pass an error message to the failure node by @popojk in https://github.com/flyteorg/flyte/pull/6181
* Bump golang.org/x/net from 0.28.0 to 0.33.0 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/6213
* Bump golang.org/x/net from 0.27.0 to 0.33.0 in /flytestdlib by @dependabot in https://github.com/flyteorg/flyte/pull/6214
* Update Serverless link by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6216
* Enable 0 workers replicaSpec for pytorchjob by @punkerpunker in https://github.com/flyteorg/flyte/pull/6201
* update mocks in async package to mockery v2 generated mocks by @popojk in https://github.com/flyteorg/flyte/pull/6220
* Update type mapping table to reflect 1.14 changes by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6191
* Remove Deck URI when `OutputReader` is nil by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6223
* fix pod template override error  by @arbaobao in https://github.com/flyteorg/flyte/pull/6224
* Prevent nil pointer in pod helper by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6225
* utilize sync maps to map project mappings of prom watches by @pvditt in https://github.com/flyteorg/flyte/pull/6222
* Adds shared memory into extended resources by @thomasjpfan in https://github.com/flyteorg/flyte/pull/6193
* Return NotFound error when decks uri is empty (#642) by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6228
* fix: typos in documentation files by @maximevtush in https://github.com/flyteorg/flyte/pull/6233
* Add on failure node passing err message doc by @popojk in https://github.com/flyteorg/flyte/pull/6209
* [FEAT] add driverPod/executorPod in Spark by @machichima in https://github.com/flyteorg/flyte/pull/6085
* Add memray docs by @fiedlerNr9 in https://github.com/flyteorg/flyte/pull/5971
* Disable offloading of literals by default by @eapolinario in https://github.com/flyteorg/flyte/pull/6246
* use mockery to generate mocks in data pkg by @popojk in https://github.com/flyteorg/flyte/pull/6244
* use mockery generate mock in clusterresource pkg by @popojk in https://github.com/flyteorg/flyte/pull/6243
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6210

## New Contributors
* @HansBambel made their first contribution in https://github.com/flyteorg/flyte/pull/6109
* @Mecoli1219 made their first contribution in https://github.com/flyteorg/flyte/pull/6036
* @luckyarthur made their first contribution in https://github.com/flyteorg/flyte/pull/5973
* @ketian-indeed made their first contribution in https://github.com/flyteorg/flyte/pull/6088
* @loselarry made their first contribution in https://github.com/flyteorg/flyte/pull/6122
* @dashangcun made their first contribution in https://github.com/flyteorg/flyte/pull/6163
* @pranshustuff made their first contribution in https://github.com/flyteorg/flyte/pull/5865
* @shuyingliang made their first contribution in https://github.com/flyteorg/flyte/pull/6184
* @arbaobao made their first contribution in https://github.com/flyteorg/flyte/pull/6118
* @punkerpunker made their first contribution in https://github.com/flyteorg/flyte/pull/6201
* @maximevtush made their first contribution in https://github.com/flyteorg/flyte/pull/6233
* @fiedlerNr9 made their first contribution in https://github.com/flyteorg/flyte/pull/5971

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.14.1...v1.15.0


