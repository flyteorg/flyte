# [RFC] Caching of offloaded types

**Authors:**

- @eapolinario

## 1 Executive Summary

We propose a way to override the default behavior of [caching task executions](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/flyte_basics/task_cache.html), enabling cache-by-value semantics for certain categories of objects.

## 2 Motivation

The behavior displayed by the cache, in some cases, does not match the users intuitions. For example, this code makes use of pandas dataframes:

```python
@task
def foo(a: int, b: str) -> pd.DataFrame:
    df = pd.Dataframe(...)
    ...
    return df

@task(cached=True, version="1.0")
def bar(df: pd.Dataframe) -> int:
    ...
    
@workflow
def wf(a: int, b: str):
    df = foo(a=a, b=b) 
    v = bar(df=df)
```

As the system currently stands, calls to `bar` will never be cached. We cannot honor the promise that cache keys will take into the account the value of the dataframes, because of how we represent dataframes in the backend.

This RFC discusses how to extend this model to allow users to override the representation of objects used for caching purposes.

## 3 Proposed Implementation

First we're going to enumerate the problems tackled in this proposal and discuss an implementation for each one. Namely:
  1. how to model cache-by-value semantics for non-Flyte objects?
  2. how to signal to users that a cached tasks contain hashable objects?

### Problem 1: Cache-by-value semantics for non-Flyte objects

For a certain category of objects, for example pandas dataframes, in order to pass data around in a performant way, Flyte writes the actual data to an object in the configured blob store and uses the path to this random object as part of the representation of the object. In other words, from the perspective of the backend, for all intents and purposes we offer cache-by-reference semantics for these objects. 

So how to offer a cache-by-value semantics for the case of these offloaded objects? We're going to expose a way for users to override the hash of objects and use that hash as part of the caching computation.

#### Exposing a hash in Literals

We will expose a new field in [Literal](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L68-L79) objects called `hash`, which will be used to represent that literal in cache key calculations. 

Each client will then define the mechanics of setting that field when it's appropriate. Specifically for flytekit, we are going to use the extensions to the typing module proposed in https://www.python.org/dev/peps/pep-0593/, more specifically [`typing.Annotated`](https://docs.python.org/3/library/typing.html#typing.Annotated). 

The following example illustrates how these annotated return objects are going to look like:


```python
@task
def foo(a: int, b: str) -> Annotated[pd.DataFrame, HashMethod(hash_pandas_dataframe_function) :
    df = pd.Dataframe(...)
    ...
    return df

@task(cached=True, version="1.0")
def bar(df: pd.Dataframe) -> int:
    ...
    
@workflow
def wf(a: int, b: str):
    df = foo(a=a, b=b)
    # Note that:
    #   1. the return type of `foo` wraps around a pandas datataframe and adds some metadata to it.
    #   2. the task `bar` is marked as cached and since the dataframe returned by `foo` overrides its hash
    #      we will check the cache using the dataframe's hash as opposed to the literal representation.
    v = bar(df=df) 
```

It's worth noting that this is a strictly opt-in feature, controlled at the level of Type Transformers. In other words, annotating types for which Type Transformers are not marked as opted in will be a no-op.

In the backend, during the construction of the cache key, prior to calling data catalog, in case the `hash` field is set, we will use it to represent the literal, otherwise we will keep [the current behavior](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L68-L79). A similar reasoning applies to the local executions case.

#### Adding the hash to FlyteIDL

The ``Literal`` will contain a new field of type string:

```protobuf
// A simple value. This supports any level of nesting (e.g. array of array of array of Blobs) as well as simple primitives.
message Literal {
    oneof value {
        // A simple value.
        Scalar scalar = 1;

        // A collection of literals to allow nesting.
        LiteralCollection collection = 2;

        // A map of strings to literals.
        LiteralMap map = 3;
    }

    // Hash to be used during cache key calculation.
    string hash = 4;
}
```

#### Flytekit 

The crux of the mechanics in flytekit revolves around how to expose enough flexibility to allow for arbitrary hash functions to be used, while at the same time providing enough information to flytekit. We propose the use of a `HashMethod` metadata object used in annotated return types. The idea being that during the process of converting from a python value to a literal we apply that hash method and set it in the literal.

We'll lean on [`typing.Annotated`](https://docs.python.org/3/library/typing.html#typing.Annotated) to annotate types with `HashMethod` objects.

Since we cannot assume any hashing characteristics of the hash functions (since we are not talking about perfect hashing in the general case), the `HashMethod` will expose a simple interface, composed of a callable of type `Callable[[T], str]`. That said, we will provide plenty of examples to showcase the flexibility of this approach.

#### Cache key calculation

As mentioned above, during the regular cache key calculation, flyte takes the input literal map as part of the key to generate a tag, which is then associated with the artifact (i.e. the task output). Going forward we will associate multiple tags, with and without the hashes, so that lookups still work transparently for both references and hash-annotated values.

#### Other clients

Although nothing prevents the adoption of this feature in other clients, flytekit-java and other clients are outside of the scope of this RFC.

### Problem 2: Bubbling up caching information

A natural question to ask is in what ways we can help users in the process of authoring and visualizing workflows that use these hash-annotated objects? 

First of all, we're going to augment the [`LiteralType` metadata](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/types.proto#L91) with a `hashable` field. Setting this field will indicate that for that type we expect the hash to be overridden in case that type is used as an input in a cacheable task.

#### At registration time

During the registration phase we will be able to detect cases where cached tasks[^1] inputs do not set the `hashable` field, even though they should in most cases. For example, by doing this we're able to catch the following case:

[^1]: For the purposes of this discussion, every time we mention cached tasks in this section we are assuming that we're talking about tasks where the cache bit is set to True and that at least one of the input types is hashable.

```python
@task
def foo(a: int, b: str) -> pd.DataFrame:
    df = pd.Dataframe(...)
    ...
    return df

@task(cached=True, version="1.0")
def bar(df: pd.Dataframe) -> int:
    ...
    
@workflow
def wf(a: int, b: str):
    df = foo(a=a, b=b)
    # At registration time we will be able to signal to the user that calls to the task `bar` will never be cached, since that even though the hashable bit is set,
    # the upstream task is not returning a hash-annotated object.
    v = bar(df=df) 
    ...
```

Whenever we detect such cases, we'll raise a warning at registration time, telling the user that it's very likely that they forgot to annotate the return type of `foo`. It's worth noting that we are able to tach this case, but in the general case, users will have to decide if such warnings are valid.

#### During cache hits

We're going to augment the [metadata](https://github.com/flyteorg/flyteidl/blob/989fe6e4f51eed3a5a68b8f563244bd38b765f4c/protos/flyteidl/admin/node_execution.proto#L175) returned by flytepropeller to indicate that the hash was used. We'll investigate ways to display this information in flyteconsole in the case of cached tasks.


## 4 Metrics & Dashboards

N/A?

## 5 Drawbacks

Although this feature does *not* impact the other aspects of how the cache works, for example,  changes to the version or the signature of the task invalidate the cache entries, it still might cause confusion since from the perspective of a task execution we will not be able to tell if an object will have its hash overridden. 

## 6 Alternatives

A few options were discussed in https://github.com/flyteorg/flyte/issues/1581:

**Expose a `hash` method in Type Transformers and annotate the task in case the type transformer does not contain a hash function:** The major drawbacks of this alternative are two-fold: 
    (1) There's no way to opt-out, i.e. all objects produced by that Type Transformer will be cached.
    (2) The UX is a bit clunky in the case of user-defined hash functions for two reasons:
        - the return type does not contain any indication that the object is being cached, i.e. we'd have to look in the parameters set in the @task decorator

**Offload the hash calculation to an external system like s3:** The idea would be to rely on etags produced by the act of storing an object in s3. This turns out to be a chicken-and-egg problem as we need to produce etags in order to compare if that tag was already produced. Also, even if we were able to overcome that hurdle, handing this computation to a blackbox system like s3 risks making cache entries unrecoverable, since there is no guarantee that the algorithm s3 uses to compute etags will not change.

**Define the new hash field in the Scalar instead of the Literal**: One could argue that the hash could be "tucked away" in the [Scalar](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L55-L65) instead. At this time I couldn't find a great argument for going that route, although most of the reasoning around the use of the hash during cache key calculation still applies (adjusting the access to the hash field).

## 7 Potential Impact and Dependencies

It might be useful to cap the size of hashes produced by client code so as to limit the size of the cache keys. Performing this capping should be done in an uniform manner, which implies that it should happen in the backend.

## 8 Unresolved questions

How does this affect @dynamic tasks, subworkflows, and launchplans?

What's the observability provided by data catalog?

## 9 Conclusion

Users have been requesting a cache-by-value semantics for non-Flyte objects for a long time and the proposal described in this document achieves that goal while not imposing a huge cost to components of the Flyte system.

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [x]  Copy template
- [x]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
