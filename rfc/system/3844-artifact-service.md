# Artifact Service

**Authors:**

- @wild-endeavor

## 1 Executive Summary
This RFC introduces the idea of an Artifact service. We've noticed for a long time that data produced by Flyte are not easily searchable nor reusable. The data catalog service in Flyte never lived up to its name. It is used merely for memoization. 

There are also implications in this doc for a future lineage type service.

## 2 Motivation
### Background
The notion of a tiny URL for flyte artifacts was already [added](https://github.com/flyteorg/flyteadmin/pull/546) (and [amended](https://github.com/flyteorg/flyteadmin/pull/565)) as exploratory code. The goal was to unlock the following use-cases 

* Be able to quickly pull an output from console to a Jupyter notebook.
* Be able to quickly pass the output of one workflow/task execution as the start of another workflow.

These two are doable now with the exploratory changes to some capacity. There's definitely more to do to make the experience more usable of course (like this doesn't work in console yet and the Admin change to directly handle these URLs in an execution create request is missing).

So let’s say you have a workflow that produced a dataframe at `flyte://v1/flytesnacks/dev/abc/n0/o/my_df` if you want to run a different workflow like

```python
@workflow
def wf(df: pd.DataFrame): ...
```

Instead of getting this:
![image-20230504-234827](https://github.com/flyteorg/flyte/assets/2896568/0ade1cf2-24d0-4982-ba06-0a1a5c7ebde4)

You should get a form in which you can just put in that string, and run with it.

Type checking will happen in Admin as soon as the CreateExecutionRequest comes in, it can retrieve the underlying workflow/task output types and perform the type check.

### Introduction of Artifact
If you squint at the above from a one-thousand meter view, the behaviour of the URL starts to look like a cataloging service. There's a unique identifier for everything produced by Flyte, it should be usable across the platform and it should be persisted. Since the catalog name was already taken, we are calling this Artifact. An Artifact is nothing more than a tracked piece of metadata associated with a Literal and LiteralType.

* Artifacts have a unique identifier per Flyte installation 
* All outputs in Flyte automatically come with an Artifact (every output of every task/workflow execution).
* Artifacts can have additional metadata data
  * Can keep type information (dataset, model, maybe a "value" which is used for Scalars that are important)
  * Can have aliases, so you should be able to alias an artifact with `latest` and then reference it.

Data modeling side note: currently the Flyte URL produced by the PRs have two forms - the node level output, or the task execution level output. For the purpose of writing to the Artifact store, we're going to take the task output.

### Artifact User Experience
The Artifact story should support the following UX. All code examples subject to change/debate.

#### Use an existing artifact locally
Use case: I want to inspect the data locally, download it and use it.

```python
a = remote.get("flyte://blah")
a = Artifact.get("flyte://blah", remote)
a.download()  # errors if it's not an off-loaded type, downloads otherwise
a.as_python_val(type_hint=...)  # call to the TypeEngine, guessing the Python type if there's no type hint
```

#### Use an existing artifact as input to a new workflow execution
Use case: A task in another workflow produced some data. I now want to use that data as the input to a new and different workflow run.

```python
df_artifact = Artifact("flyte://a1")
remote.execute(wf, inputs={"a": df_artifact})
```

Since the Artifact service can take a link and return a Literal, this part is pretty simple since remote.execute already can take Literals. Will just have to fetch and add it to the LiteralMap.  When remote runs, it’ll have to fetch the Literal.

Note: with this construct, two different versions/aliases/etc of the same dataset/artifact can be used in one workflow - think about a workflow that compares the performance of two models.

#### Specify metadata at Artifact creation time
Use case: When this task or workflow runs, I want to tag it with certain tags and give it a version alias.

```python
@task
def t1() -> Annotated[nn.Module, Artifact(name="my.artifact.name", alias={"version": "latest"}, type=Model)]
    ...
```
There needs to be some unique-ness aspect here. The ability to uniquely identify an Artifact given some label like 'latest' or something is predicated on the fact that it's the latest given some additional fields. Like, given an alias key `alias_key` which has a value `alias_value`, when combined with Project, Domain and the Name (here `my.artifact.name`), this should correspond to exactly one Artifact at any given time.

#### Create a new one locally
Use case: I have some data locally, I want to upload this to Flyte and make the Artifact service aware of this, as if it were the output of a workflow execution.

```python
a = Artifact.new(python_val, pytype, schema, metadata, ...)  # remote detected from current context maybe.
a = Artifact.new(python_val, pytype, schema, metadata, remote, type=Dataset ...)
new_a = Artifact.new(large_important_str)  # .new(my_df), .new(Path("/root/data/tokened/"))
remote.create(new_a)
```

Given an object in Python memory, whether it's a float, a pd.DataFrame, or a io.StringIO buffer, a user should be able to turn this into an Artifact. How this might work is:
* FlyteRemote adds itself into the data persistence layer as an fsspec compatible filesystem by adding a new context that overrides the output data prefix with `flyte://dv1/proj/domain/`. See this [ongoing work](https://github.com/flyteorg/flytekit/pull/1674/files).
* Run the Python object through the flytekit Type Engine.
* Return the Artifact object containing a possibly different flyte URL to the user. Attributes of the created Artifact might contain
  * an address that's universal per Artifact service (aka Flyte installaion).
  * a name that was autogenerated using the item stored.
  * aliases that make sense that are autogenerated
  * it will be associated with a Literal and a Literal type and also store any schema or metadata information.
  * bookkeeping information, created_at, user who created it, etc.

#### Queryable and bindable queries
Use case: Identify the artifact with such and such name and the "latest" alias. I have a workflow whose default input I want to be whatever is returned by that search when the workflow is kicked off.

Basic functionality here on the search side might look like
* Given an Artifact URL, list all executions it was used in.
* Given an Artifact URL, find the execution that produced it, or let the user know if it was an upload.
* Given a set of aliases or tags, return Artifact objects.

On the authoring side, we would like users to be able to do the following
```python
@task
def t1() -> Annotated[nn.Module, Artifact(name="models.nn.lidar", alias="latest", overwrite_alias=True)]: ...

# Possibly in a completely different repo
@workflow
def wf(model: nn.Module = Artifact.get_query(name="models.nn.lidar", alias="latest")): ...
```
The binding is made at compilation time but it is not resolved until run time. That is, when `wf` is run, Admin will have to call out to the Artifact service, get back precisely one Artifact, and use the Literal therein contained as the input. If the type is not a match, then it can result it in a user run time error, which is okay.

## 3 Proposed Implementation
We are working through an implementation in the following PRs:
IDL: https://github.com/flyteorg/flyteidl/pull/408/files
Admin: https://github.com/flyteorg/flyteadmin/pull/572/files
Flytekit: https://github.com/flyteorg/flytekit/pull/1655/files

This Flytekit [PR](https://github.com/flyteorg/flytekit/pull/1674/files) was also referenced earlier and will be done as a separate initiative to fix long-standing issues with `pyflyte run`

There is also considerable work left to be done on the UI side. If the tiny URL construct holds it would be nice to see a click-copying button from the UI for example that would copy a Python snippet that would allow users to immediately hydrate a Literal.

## 4 Metrics & Dashboards
Will address last.


## 5 Drawbacks
None.

## 6 Alternatives
One alternative not to the idea presented here but to the implementation is to use the existing data catalog service instead, either in whole or in part. Everything from using the grpc service itself with well-crafted calls, to implementing the new service but with the data catalog code.

## 7 Potential Impact and Dependencies
Everything will be implemented before merging anything major or anything that can't be pulled back out.

## 8 Unresolved questions
If you look in the IDL change, you will see that the Artifact message comes with a `source` field. This is nice but doesn't really make sense if you think about it. There's a better name for the act of identifying where an Artifact came from, and also for identifying things in the search capability like what other executions it was used in. That term is lineage. We're not sure if lineage should be a different service or is an integral part of this Artifact service yet. I can see arguments both ways. It is clear that they are closely intertwined though.

## 9 Conclusion


