# Deterministic error propagation for distributed (training) tasks

**Authors:**

- @bgedik
- @eapolinario
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

### Flytekit

#### Preventing the race condition

For distributed training tasks, the [pod entrypoint `pyflyte-execute`](https://github.com/flyteorg/flytekit/blob/master/flytekit/bin/entrypoint.py) must not upload a file called [`error.pb`](https://github.com/flyteorg/flytekit/blob/77d056ab9fda40ec6b2312a4d197b9107cdb70dc/flytekit/core/constants.py#L4) (which is the same for all worker pods) but instead choose a file name which differs for each worker pod. We propose to simply include the pod name in the `error-<pod name>.pb` as this prevents the race condition.

For this purpose, we propose that `flyteplugins` injects the environment variable `FLYTE_INTERNAL_POD_NAME` using the Kubernetes [downward api](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef).

Furthermore, we propose that distributed task plugins in `flyteplugins` inject the environment variable `FLYTE_INTERNAL_ERROR_PROPAGATION=earliest` (where `earliest` is the first of potentially multiple strategies to determine the root cause error, see below).

If the `FLYTE_INTERNAL_ERROR_PROPAGATION` environment variable is set, `pyflyte-execute` includes the pod name in the error file.

#### Including the error time stamp

When a distributed training job dies, one of the worker pods often dies due to a certain root-cause error. The other worker pods subsequently crash because one of the workers disappeared. We are interested in the root-cause error, not the error that one worker disappeared.

We propose to use the timestamp of the exception as a proxy to determine the root-cause. Pytorch distributed, for instance, raises a [ChildFailedError](https://github.com/pytorch/pytorch/blob/36d24925c66661037349cad3759dc33850ed0291/torch/distributed/elastic/multiprocessing/errors/__init__.py#L199C16-L199C17) exception which contains a so-called [ProcessFailure](https://github.com/pytorch/pytorch/blob/36d24925c66661037349cad3759dc33850ed0291/torch/distributed/elastic/multiprocessing/errors/__init__.py#L90) which contains the exception timestamp.

We propose to add an optional `timestamp` attributes to all [flytekit exceptions](https://github.com/flyteorg/flytekit/tree/master/flytekit/exceptions) (system and user scope, recoverable, ...).

The flytekit pytorch elastic plugin, for instance, catches `ChildFailedError`s [here](https://github.com/flyteorg/flytekit/blob/77d056ab9fda40ec6b2312a4d197b9107cdb70dc/plugins/flytekit-kf-pytorch/flytekitplugins/kfpytorch/task.py#L449), would extract the timestamp, and re-raise it as a Flyte exception which contains a timestamp. (Other plugins, e.g. non-elastic pytorch, which don't come with built-in exception types that include error timestamps, can themselves record the timestamp when the `task_function` raises an exception.)

We furthermore propose to add a timestamp to flyteidl's [`message ContainerError`](https://github.com/flyteorg/flyte/blob/30d33149159c90d0de44f6351b8d5d7309242e59/flyteidl/protos/flyteidl/core/errors.proto#L11).

The entrypoint `pyflyte-execute` will transfer the timestamp from the flytekit exception into the protobuf `ContainerError`.

### Flytepropeller (Backend)

Currently, [here](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flytepropeller/pkg/controller/nodes/task/k8s/plugin_manager.go#L290) in the plugin manager, upon completion of a node execution, a new [`RemoteFileOutputReader`](https://github.com/flyteorg/flyte/blob/d6da838627d57cd27d60beea004e974ce1fb3ca5/flyteplugins/go/tasks/pluginmachinery/ioutils/remote_file_output_reader.go#L14) is constructed which is responsible for reading the error file uploaded to blob storage. This `RemoteFileOutputReader` implements the [`OutputReader` interface](https://github.com/flyteorg/flyte/blob/1e54d21c4d4ee74245f799a57b4bb8a5534e8368/flyteplugins/go/tasks/pluginmachinery/io/iface.go#L32).

We propose to implement a new `MultiErrorFileRemoteFileOutputReader` which (for future flexibility) can be configured with different policies that determine which of multiple errors to report downstream. Initially, the only available policy is "earliest".

Open questions:

* How do we configure for distributed plugins to use this new `MultiErrorFileRemoteFileOutputReader` reader instead of the default one?
    * We could add a `MultipleErrorFiles` property to `PluginProperties` (see https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go#L34). The PyTorch plugin, for instance, would then pass `true` for `MultipleErrorFiles` [here](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flyteplugins/go/tasks/plugins/k8s/kfoperators/pytorch/pytorch.go#L31).
    
    Currently, [here](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flytepropeller/pkg/controller/nodes/task/k8s/plugin_manager.go#L290) in the plugin manager, where we call `NewRemoteFileOutputReader`, we do have access to `e.plugin`, and thus to `PluginProperties` and could make use of that information to instantiate another output reader.
    * Could we alternatively add an `OutputReader` to the [`PluginContext`](https://github.com/flyteorg/flyte/blob/4514860cf56ba62717f6c207f269410a8c1a5461/flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go#L51)? Where would we customize this plugin context for e.g. the kubeflow plugins?

#### Backwards compatibility
We propose that the new `MultiErrorFileRemoteFileOutputReader` falls back to reading the `error.pb` if no `error-<pod-name>.pb` files are found in order to solve the problem of backwards compatibility:

* If flytekit uses a version that supports multiple error files but the backend does not yet, `pyflyte-execute` will not upload multiple error files for distributed tasks since the `FLYTE_INTERNAL_ERROR_PROPAGATION` environment variable will not be set.
* If flytekit uses an older version that does not support multiple error files while the backend does, a single error file will be uploaded despite `FLYTE_INTERNAL_ERROR_PROPAGATION` being set. The output reader will, however, fall back to reading the single `error.pb`.

### Flyteconsole (UI)

We propose that in the UI, in addition to the error message, we display the name of the pod which experienced the root-cause error. As a user trying to debug the failure, this allows me to quickly identify the logs of the relevant pod in a distributed job with potentially hundreds of workers.

To transport this information from `pyflyte-execute` to flytepropeller, we propose to add an additional field `pod_name` (or `container_name`) to `message ContainerError`.

Open question:
* Where does flytepropeller add this info for it to be displayed as part of the error in the UI?

## 4 Metrics & Dashboards

-

## 5 Drawbacks

We don't see any drawbacks to making the error handling of distributed training tasks deterministic and making it easier for users to identify which pod in a distributed job failed first.

## 6 Alternatives

A poor man's version would be to not override the error file if it already exists. While this is a worse solution than proposed above as there still is a race condition, this would still be better than the current behavior because at least we would *favor* earlier errors instead of later ones.

## 7 Potential Impact and Dependencies

The authors of this RFC have experience with pytorch (elastic and non-elastic) distributed training jobs and will implement the proposed changes for the pytorch plugin. The improvement proposed in this RFC might be relevant for community members using e.g. the distributed tensorflow or mpi plugins. If possible, they should be included in the RFC and implementation process so that all distributed task plugins can benefit from the improved error handling.

## 8 Unresolved questions

See open questions in section 3.

## 9 Conclusion

With ML models getting bigger and bigger, distributed training jobs become increasingly important to the Flyte community. Removing the race condition outlined above from Flyte's error handling for such jobs will significantly improve the UX because we will be able to determine recoverability and report the root-cause error in the Flyte UI in a deterministic way.
