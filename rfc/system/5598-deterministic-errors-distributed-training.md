# Deterministic error propagation for distributed (training) tasks

**Authors:**

- @bgedik
- @fg91

## 1 Executive Summary

Flyte can schedule distributed training jobs leverging e.g. the [kubeflow training operator](https://github.com/kubeflow/training-operator/tree/f55a91d03f23498cdb465ac26c78566228077c51) and its `PyTorchJob`, `TFJob`, `MPIJob`, ...

For these distributed jobs, multiple Kubernetes pods are launched. Any of these worker pods can crash, causing all other worker pods in the distributed job to fail subsequently because one worker disappeared.

Error propagation, in Flyte, happens by the pod entrypoint uploading a file called `error.pb` to blob storage which contains (among other things) the error message and the information whether the error is retriable.

In a failed distributed training job, all worker pods currently try to create the same `error.pb` file in blob storage - leading to a race condition. It is not guaranteed that the root-cause error is the one being reported to the user and used to determine whether the task can be retried. In fact, the current behavior typically results in the worst outcome, as the latter errors override the former ones, which is the exact opposite of the desired behavior of identifying the first error as the root cause.

## 2 Motivation

* As a Flyte user trying to understand why a distributed training task failed, I currently cannot rely on the error reported in the Flyte Console (UI) being the root cause error.
    * Instead, I have to search the logs of each worker pod. For distributed training jobs with dozens or even hundreds of worker pods, this can be tedious.
    * (Current remedies include combining all worker pods in stackdriver logs using a wildcard in the pod name and then filtering by severity.)
* As a Flyte user marking specific errors as retriable (using a `FlyteRecoverableException`), I want Flyte to deterministically determine the root cause error that killed the distributed job so that the retry behaviour does not suffer from a race condition.

## 3 Proposed Implementation

When a distributed training job dies, one of the worker pods often dies due to a certain root-cause error. The other worker pods subsequently crash because one of the workers disappeared. We are interested in the root-cause error, not the error that one worker disappeared.

As done in torch distributed elastic, we propose to use the timestamp of the exception as a proxy to determine the root-cause. Pytorch distributed (which is used by the `flytekitplugins.kfpytorch.Elastic` task type), for instance, raises a [ChildFailedError](https://github.com/pytorch/pytorch/blob/36d24925c66661037349cad3759dc33850ed0291/torch/distributed/elastic/multiprocessing/errors/__init__.py#L199C16-L199C17) exception which contains a so-called [ProcessFailure](https://github.com/pytorch/pytorch/blob/36d24925c66661037349cad3759dc33850ed0291/torch/distributed/elastic/multiprocessing/errors/__init__.py#L90) which contains the exception timestamp.

We acknowledge that other frameworks might choose to determine the root cause error in a different way which is why we propose to introduce the concept of an *error aggregation strategy* employed by flytepropeller to identity the root-cause error in a distributed job. The authors of this RFC aim to implement the strategy `"earliest"` for the two kubeflow pytorch task types (`task_config=PyTorch` and `task_config=Elastic` provided by `flytekitplugins.kfpytorch`) but propose to structure the introduced changes in a way that allows potential other strategies.

### Flyteplugins - Creation of Kubernetes resources for tasks

The [pod entrypoint `pyflyte-execute`](https://github.com/flyteorg/flytekit/blob/master/flytekit/bin/entrypoint.py) needs to be configured to handle multiple error files for a distributed task.

For this purpose, we propose that distributed plugins in `flyteplugins` like `pytorch` inject two new (optional) `FLYTE_INTERNAL_` environment variables:

* `FLYTE_INTERNAL_WORKER_NAME`: One of our goals is that the UI eventually tells the user in which worker the root-cause error occurred. For this purpose, the pod entrypoint reads the value from this environment variable. Plugin's can choose for themselves what value this environment variable should have. For the pytorch plugin we aim to set it to the pod name via the [Kubernetes downward api](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/).
* `FLYTE_INTERNAL_DIST_ERROR_STRATEGY` tells the pod entrypoint which strategy `flytepropeller` will use to aggregate the error from the worker pods. The pod entrypoint needs this information to determine which information it needs to provide as part of the error file. More on this below.

(We propose to define the keys for these environment variables [here](https://github.com/flyteorg/flyte/blob/815f85d0ce90a3ace61cce17c0bfb441ac2dbcc3/flyteplugins/go/tasks/pluginmachinery/flytek8s/k8s_resource_adds.go#L20) where the existing `FLYTE_INTERNAL_` environment variables are defined while the respective plugins are responsible for actually setting them to the desired value in the respective pod specs.)

### Flytekit

#### Preventing the race condition

For distributed training tasks, the [pod entrypoint `pyflyte-execute`](https://github.com/flyteorg/flytekit/blob/master/flytekit/bin/entrypoint.py) must not upload a single file called [`error.pb`](https://github.com/flyteorg/flytekit/blob/77d056ab9fda40ec6b2312a4d197b9107cdb70dc/flytekit/core/constants.py#L4) (which is the same for all worker pods) but instead choose a file name which differs for each worker pod. We propose to simply include a random uuid in the filename `error-<uuid>.pb` to prevent the race condition. Furthermore, we propose that these error files get grouped in an `errors/` folder under the raw output prefix.

#### Providing relevant error information to the backend an UI

The pod entrypoint needs to provide the information in which worker the error occurred in order to display the name in the UI. For the strategy `"earliest"`, it needs to also provide the timestamp when the error occurred.

We therefore propose to add optional attributes `worker` and `timestamp` (unix epoch time with micro- or nanoseconds granularity) to flyteidl's [`message ContainerError`](https://github.com/flyteorg/flyte/blob/30d33149159c90d0de44f6351b8d5d7309242e59/flyteidl/protos/flyteidl/core/errors.proto#L11).


Furthermore, we propose to add an optional `timestamp` attributes to all [flytekit exceptions](https://github.com/flyteorg/flytekit/tree/master/flytekit/exceptions).

The flytekit pytorch elastic plugin, for instance, catches `ChildFailedError`s [here](https://github.com/flyteorg/flytekit/blob/77d056ab9fda40ec6b2312a4d197b9107cdb70dc/plugins/flytekit-kf-pytorch/flytekitplugins/kfpytorch/task.py#L449), would extract the timestamp, and re-raise it as a Flyte exception which contains a timestamp. (Other plugins, e.g. non-elastic pytorch, which don't come with built-in exception types that include error timestamps, can themselves record the timestamp when the `task_function` raises an exception.)

The entrypoint `pyflyte-execute` will transfer the timestamp from the flytekit exception into the protobuf `ContainerError`. It will also set the `worker` attribute of the `ContainerError` according to the `FLYTE_INTERNAL_WORKER_NAME` environment variable introduced above.

### Flytepropeller/Flyteplugins - Aggregate the errors in the backend

In the [kubernetes plugin machinery](https://github.com/flyteorg/flyte/blob/815f85d0ce90a3ace61cce17c0bfb441ac2dbcc3/flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go) we propose to define the error aggregation strategy and allow plugins to configure it via their `PluginProperties`:

```go
type ErrorAggregationStrategy int

const (
	// Single error file from a single container
	Default ErrorAggregationStrategy = iota

	// Earliest error from potentially multiple error files
	Earliest
)

// System level properties that this Plugin supports
type PluginProperties struct {
	...
	ErrorAggregationStrategy ErrorAggregationStrategy
}
```

Currently, [here](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flytepropeller/pkg/controller/nodes/task/k8s/plugin_manager.go#L290) in the plugin manager, upon completion of a node execution, a new [`RemoteFileOutputReader`](https://github.com/flyteorg/flyte/blob/d6da838627d57cd27d60beea004e974ce1fb3ca5/flyteplugins/go/tasks/pluginmachinery/ioutils/remote_file_output_reader.go#L14) is constructed which is responsible for reading the error file uploaded to blob storage. This `RemoteFileOutputReader` implements the [`OutputReader` interface](https://github.com/flyteorg/flyte/blob/1e54d21c4d4ee74245f799a57b4bb8a5534e8368/flyteplugins/go/tasks/pluginmachinery/io/iface.go#L32).

We propose to implement a new `MultiErrorFileRemoteFileOutputReader` which (for future flexibility) can be configured with the different strategies we define. Initially, the only available strategy will be `"earliest"` which the RFC authors aim to use for the kubeflow pytorch plugin. This output reader will search for all error files in the `/errors` folder under the raw output prefix and aggregate the error as specified by the strategy.

If in [the plugin manager](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flytepropeller/pkg/controller/nodes/task/k8s/plugin_manager.go#L290) the respective plugin is found to configure an error aggregation strategy other than `Default`, we instantiate such a `MultiErrorFileRemoteFileOutputReader` reader (instead of the existing `RemoteFileOutputReader`) and configure it with the respective strategy.

For the strategy `Earliest`, it will determine the `ContainerError` with the earliest timestamp, will use this one to determine retriability, and will communicate this specific error message to flyteadmin (and finally the UI).

#### Backwards compatibility
We propose that the new `MultiErrorFileRemoteFileOutputReader` falls back to reading the `error.pb` (behaviour of the default `RemoteFileOutputReader`) if no `error-<pod-name>.pb` files are found in order to solve the problem of backwards compatibility:

* If flytekit uses a version that supports multiple error files but the backend does not yet, `pyflyte-execute` will not upload multiple error files for distributed tasks since the `FLYTE_INTERNAL_DIST_ERROR_STRATEGY` environment variable will not be set.
* If flytekit uses an older version that does not support multiple error files while the backend does, a single error file will be uploaded despite `FLYTE_INTERNAL_DIST_ERROR_STRATEGY` being set. The output reader will, however, fall back to reading the single `error.pb`.


### Displaying the name of the worker which experienced the root cause error in the UI

We propose that in the UI, in addition to the root-cause error message, for distributed tasks we display the name of the worker pod which experienced the root-cause error. As a user trying to debug a failure, this allows to quickly identify the logs of the relevant pod out of potentially hundreds of pods.

To communicate the name of the worker which experienced the root-cause error from flytepropeller to flyteadmin and eventually the UI, we propose to add the (optional) attribute `worker` also in the [`core.ExecutionError` protobuf message](https://github.com/flyteorg/flyte/blob/815f85d0ce90a3ace61cce17c0bfb441ac2dbcc3/flyteidl/protos/flyteidl/core/execution.proto#L61).

In `ReadError` of the new `MultiErrorFileRemoteFileOutputReader`, we will then transfer the name of the respective worker pod which experienced the root-cause error from the `ContainerError` in the `ErrorDocument` to the `core.ExecutionError` (as is already done today in the [`RemoteFileOutputReader` for the error message](https://github.com/flyteorg/flyte/blob/815f85d0ce90a3ace61cce17c0bfb441ac2dbcc3/flyteplugins/go/tasks/pluginmachinery/ioutils/remote_file_output_reader.go#L65)).

With these changes, flyteadmin's `/api/v1/executions/<project>/<domain>/<execution id>` endpoint, which today provides the error message to the UI, then also provides the information which worker experienced the root cause error. `flyteconsole` needs to be modified to show this information.

## 4 Metrics & Dashboards

-

## 5 Drawbacks

We don't see any drawbacks to making the error handling of distributed training tasks deterministic and making it easier for users to identify which pod in a distributed job failed first.

## 6 Alternatives

A poor man's version would be to not override the error file if it already exists. While this is a worse solution than proposed above as there still is a race condition, this would still be better than the current behavior because at least we would *favor* earlier errors instead of later ones.

## 7 Potential Impact and Dependencies

The authors of this RFC have experience with pytorch (elastic and non-elastic) distributed training jobs and will implement the proposed changes for the pytorch plugin. The improvement proposed in this RFC might be relevant for community members using e.g. the distributed tensorflow or mpi plugins. If possible, they should be included in the RFC and implementation process so that all distributed task plugins can benefit from the improved error handling.

## 8 Unresolved questions

-

## 9 Conclusion

With ML models getting bigger and bigger, distributed training jobs become increasingly important to the Flyte community. Removing the race condition outlined above from Flyte's error handling for such jobs will significantly improve the UX because we will be able to determine recoverability and report the root-cause error in the Flyte UI in a deterministic way.
