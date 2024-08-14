# [RFC] Offloaded Raw Literals

**Authors:**

- @wild-endeavor
- @EngHabu
- @katrogan

## 1 Executive Summary

Flyte depends on a series of `inputs.pb` and `outputs.pb` files to do communication between nodes. This has typically served us well, except for the occasional map task that produces a large Literal collection output. This large collection typically exceeds default gRPC and configured storage limits. We sometimes also run into this issue for large dataclasses. This RFC proposes a mechanism that allows the offloading of any Literal, to get around size limitations for passing large Literal protobuf messages in the system.

## 2 Motivation
A [cursory search](https://discuss.flyte.org/?threads%5Bquery%5D=LIMIT_EXCEEDED) of Slack history shows a few times that this has come up before (and I remember some other instances, I think that search term just wasn't included). This is something that we've historically addressed by just increasing the size of the grpc message that's allowed, but this is an unsustainable solution and severely reduces the utility of large-fan-out map tasks.

## 3 Proposed Implementation
We propose configuring propeller to offload large literal collections, using the following config

```yaml
type LiteralOffloadingConfig struct {
  Enabled bool
  // Maps flytekit SDK to minimum supported version that can handle reading offloaded literals.
  SupportedSDKVersions map[string]string
  // Default, 10Mbs. Determines the size of a literal at which to trigger offloading
  MinSizeInMBForOffloading uint64
  // Fail fast threshold
  MaxSizeInMBForOffloading uint64
}

```

### 3.1 Offloaded Literal IDL
Update the `Literal` [message](https://github.com/flyteorg/flyte/blob/4a7c3c0040b1995a43939407b99ca3e87b1eb752/flyteidl/protos/flyteidl/core/literals.proto#L94-L114)
like so

```protobuf
message Literal {
    oneof value {
        // A simple value.
        Scalar scalar = 1;
        // A collection of literals to allow nesting.
        LiteralCollection collection = 2;
        // A map of strings to literals.
        LiteralMap map = 3;
    }
    ...
    // ** new below this line **
    // If this literal is offloaded, this field will contain metadata including the offload location.
    string uri = 6;
    // Includes information about the size of the literal.
    uint64 size_bytes = 7;
} 

```

### 3.2 Flyte Propeller
Once offloading is enabled in the deployment config, flytepropeller can read from the [RuntimeMetadata](https://github.com/flyteorg/flyte/blob/f448a0358d8706a09b65b96543134f629327d755/flyteidl/protos/flyteidl/core/tasks.proto#L71-L87) in the task config to determine the SDK version.

When writing outputs in the [remote_file_output_writer](https://github.com/flyteorg/flyte/blob/2ca31119d6b9258661a71f38e450f93b6692402c/flyteplugins/go/tasks/pluginmachinery/ioutils/remote_file_output_writer.go#L56-L84) the source code should detect whether the literal size exceeds the configured minimum and
- if the task is using a newer SDK version that supports reading offloaded literals, offload the literal to the configured storage backend and update the literal with the offload URI and size.
- if the task is using an older SDK version that doesn't support offloaded literals, fail the task with an error message indicating that the task output is too large and the user should update their SDK version. Downstream tasks will need to understand how to consume the offloaded literal and will need to be on newer version of the SDK as well.

### 3.3 Flytekit & Copilot
Flytekit and Copilot will both need to detect that a Literal has been offloaded and know to download it.
- in Flytekit this can be done by checking the `uri` field in the Literal message when converting a literal [to_python_value](https://github.com/flyteorg/flytekit/blob/e394af0be9f904fbf8be675eaa8b8cdc24311ced/flytekit/core/type_engine.py#L1134)
- in Copilot, the data downloader [literal handling](https://github.com/flyteorg/flyte/blob/5f4199899922ca63f7690c82dfca42a783db64c3/flytecopilot/data/download.go#L219-L248) should fetch the value

As a follow-up, we can also implement literal offloading in the SDK for conventional python tasks. Flytekit should also know how to offload the data. This should be done transparently to the user. 

We should fail fast in the SDK for too-large literals as part of the initial round of changes.

**Open Question:** How will flytekit know to fail though if propeller hasn't been updated?

### 3.4 Other Implications
#### Flytekit Remote
Flytekit Remote will need to be updated to handle offloaded literals. In order to fetch offloaded literals by URI, users must now authenticate with their cloud provider on their machines using a role which has read access to the _metadata bucket_.

#### Console
Console code should show the offloaded literal URI and gracefully handle nil Literal [values](https://github.com/flyteorg/flyte/blob/4a7c3c0040b1995a43939407b99ca3e87b1eb752/flyteidl/protos/flyteidl/core/literals.proto#L96-L105).

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

## 6 Alternatives

Alternate suggestions that were proposed include

* For map tasks, change the type of the output to a Union of the current user defined List and a new Offloaded type. We felt this would be a bit awkward since it changes the user-facing type itself (like if you were to pull up the map task definition in the API endpoint). It's also not extensible to other types of literals (maps of large dataclasses for example).

* Build off of the input wrapper construct that's still in [PR](https://github.com/flyteorg/flyte/pull/4298). The idea was to have the wrapper contain in large cases, a reference to the data, and in small cases, the data itself. We didn't fully like this idea because the entire input set or output set needs to be offloaded.
  * If the task downstream of a map task takes both the output list, along with some other input, after creating and upload the large pb file for the map task's output, Propeller would need to re-upload the entire large list or map (one time for each downstream task). If the offloading is done per literal, Propeller can just upload once and use.
* Modify the workflow CRD to include the offloading bits so that they're respected at execution time, and serialized at registration time. This is a bit heavier handed than just respecting the SDK version

## 7 Potential Impact and Dependencies

There's a couple edge cases that will just not work.

* If the map task is of an older flytekit version but for some reason the downstream task is of a newer version, Propeller will fail unnecessarily.
* If the map task is a newer version, but the downstream task is an older version, the downstream task will fail correctly.
* If workflow is using an older SDK version and launches a child workflow with a newer SDK version, the parent workflow will fail to resolve the child workflow outputs

Are there concerns about the fact that if we're offloading data once, and then sharing the pointer, we're no longer copying-by-value? Does this break any of the guarantees of Flyte and will we need to be more careful in the future around other changes to avoid issues?

## 8 Unresolved questions

Should we create a new oneof that's offloaded?

Is there anything around sampled data, or automatically computed actual metadata (like number of elements in the list) that we should do?

## 9 Conclusion

Moving to literal offloading fixes a common and frustrating pain point around map tasks. It's a relatively simple change that should have a big impact on the usability of Flyte.

```
