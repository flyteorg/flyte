# [RFC Template] Simplify Resources

**Authors:**

- @granthamtaylor

## 1 Executive Summary

Defining resources for tasks is error prone and counter-intuitive due to the nuances around `requests` and `limits`.

This RFC discusses an alternative solution to simplify the definition of resources, remove some footguns, and smoothen flytekit altogether.

## 2 Motivation

Some of the footguns around `requests` and `limits` are non-deterministic and not intuitive, thus can escape detection during development but have significant and detrimental impact in production.

For example, the possibility of setting `requests < limits` for `mem` can lead to a pod becoming overloaded. For `cpu`, doing so can lead to CPU availability, for some multi-processing applications, to be non-deterministic, thus resulting in unpredictable performance degradation.

Additionally, setting `requests < limits` for `gpu` and `ephemeral_storage` should not be allowed.

Having `limits` and `requests` be configured under two distinct instances of `flytekit.Resources` makes validation challenging.

Lastly, the implementation of `limits` and `requests` as two distinct instances of `flytekit.Resources` makes for an awkward experience of defining new resource configurations (`accelerator`, `shared_memory`, `processor`, etc)

## 3 Proposed Implementation

- Deprecate `limits` and `requests` for every entity that requires them (`task`, `dynamic`, `eager`, etc)
    - For `cpu`, the value in `Resources` should define the `cpu` `request`, because CPU allocation can be non-deterministic if `limits > requests`
    - For `mem`, the value in `Resources` should define the `mem` `request`, because memory allocation can be non-deterministic if `limits > requests`, potentially resulting in OOM errors.
    - For `gpu`, the value in `Resources` should define the `gpu` `limit`, as having a different value (`requests != limits`) is not even allowed anyways.
    - For `ephemeral_storage`, the value in `Resources` should define the `ephemeral_storage` `limit`, because memory allocation can be non-deterministic if `limits > requests`, potentially resulting in pod evictions.
- Add the argument with just `resources: flytekit.Resources`
- Add the `accelerator` configuration to `flytekit.Resources` and deprecate it from `task`, `dynamic`, `eager`, etc.

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

## 5 Drawbacks

We are losing some flexibility here for power users. However, generally speaking, being able to set both `requests` and `limits` for CPU / memory is a well known footgun.

## 6 Alternatives

Should any resource type (`CPU`) need both `limits` and `requests`, we should allow definition of them via a `tuple` (IE `(400m, 600m)`).

## 7 Potential Impact and Dependencies

This requires deprecating some long-standing and widely used functionality. We should be mindful of the impact of deprecating `requests` and `limits` here.

Ideally, we would have `requests`, `limits`, and `resources` all defined simultaneously for at least one full minor version, during which time `requests` and `limits` will raise a deprecation warning.

Should users attempt to define (`requests` or `limits`) and `resources` we should raise an error. Should users attempt to define any new resource type (IE `accelerator`) in `Resources` and pass that `Resources` instance to `requests` or `limits`, we should raise an error.

## 8 Unresolved questions

- Are there any circumstances in which users _need_ to define `requests` AND `limits` such that they are different?

## 9 Conclusion

This simplifies a messy footgun from Flyte, simplifies the SDK, and paves the way for `Resources` to be used to define multiple the resources for other implementations, such as `agents`.
