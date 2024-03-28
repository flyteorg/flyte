# [RFC] Offloaded Raw Literals

**Authors:**

- @wild-endeavor
- @EngHabu

## 1 Executive Summary

Flyte depends on a series of `inputs.pb` and `outputs.pb` files to do communication between nodes. This has typically served us well, except for the occasional map task that produces a large Literal output. We sometimes also run into this issue for large dataclasses. This RFC proposes a mechanism that allows the offloading of any Literal, which would be done only of course for now for size reasons.

## 2 Motivation
A [cursory search](https://discuss.flyte.org/?threads%5Bquery%5D=LIMIT_EXCEEDED) of Slack history shows a few times that this has come up before (and I remember some other instances, I think that search term just wasn't included). This is something that we've historically addressed by just increasing the size of the grpc message that's allowed, but this is an unsustainable solution.

## 3 Proposed Implementation

### 3.1 Offloaded Literal IDL
To the `Literal` [message](https://github.com/flyteorg/flyte/blob/cb6384ac6ea60f8b9421a71cfda4279f3579d3cb/flyteidl/protos/flyteidl/core/literals.proto#L95), add a new field called `starp` that will point to a location in the "metadata" bucket of the Flyte backend. The offloaded bytes should be deserialzable into a `Literal` object.

Questions: How will things like metadata be handled? Should they be merged? What should be in the `value` field of the main parent Literal?

### 3.2 Flyte Propeller
* When writing map task outputs, depending on the size, Propeller will need to offload the LiteralCollection after constructing it, and create a new Literal for downstream tasks to use, with the 
* Also Propeller will need to check the flytekit version of the map task. If it's an older version (i.e. before the change proposed in this RFC), and it's large enough to need to be offloaded, it should fail the task. The assumption here is that if the map task is of the older version then downstream tasks will probably also be of those older versions which won't know how to resolved these offloaded literals. 

### 3.3 Flytekit & Copilot
Flytekit and Copilot will both need to detect that a Literal has been offloaded and know to download it.

For large outputs (like large maps of large dataclasses), Flytekit should also know how to offload the data. This should be done transparently to the user. How will propeller know to fail though if propeller hasn't been updated?

### 3.4 Other Implications
Does console need to change at all?

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

## 6 Alternatives

Alternate suggestions that were proposed include

* For map tasks, change the type of the output to a Union of the current user defined List and a new Offloaded type. We felt this would be a bit awkward since it changes the user-facing type itself (like if you were to pull up the map task definition in the API endpoint). It's also not extensible to other types of literals (maps of large dataclasses for example).

* Build off of the input wrapper construct that's still in PR. The idea was to have the wrapper contain in large cases, a reference to the data, and in small cases, the data itself. We didn't fully like this idea because the entire input set or output set needs to be offloaded.
  * If the task downstream of a map task takes both the output list, along with some other input, after creating and upload the large pb file for the map task's output, Propeller would need to re-upload the entire large list or map (one time for each downstream task). If the offloading is done per literal, Propeller can just upload once and use.

## 7 Potential Impact and Dependencies

There's a couple edge cases that will just not work.

* If the map task is of an older flytekit version but for some reason the downstream task is of a newer version, Propeller will fail unnecessarily.
* If the map task is a newer version, but the downstream task is an older version, the downstream task will fail correctly.

Are there concerns about the fact that if we're offloading data once, and then sharing the pointer, we're no longer copying-by-value? Does this break any of the guarantees of Flyte and will we need to be more careful in the future around other changes to avoid issues?

## 8 Unresolved questions

Should we create a new oneof that's offloaded?

Is there anything around sampled data, or automatically computed actual metadata (like number of elements in the list) that we should do?

## 9 Conclusion

*Here, we briefly outline why this is the right decision to make at this time and move forward!*

