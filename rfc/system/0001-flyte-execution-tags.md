# Flyte Execution Tags and Metadata

**Authors:**

- @kumare3

## 1 Executive Summary

Flyte currently provides no visual ways of grouping executions and other
entities with a “project” and “domain”. Usually a project is used by a team of
engineers and/or researchers and sometimes they may be  experimenting in
isolation or running multiple co-horts of experiments. Also there are cases in
which grouping executions by a group might make them easier to find and
associate them better. This document provides a solution of how we could
improve the experience of discovering executions within Flyte. It also provides
motivation of how this feature could further improve other ecosystem projects.

## 2 Motivation

As a User I want to
 - Group a certain number of executions into an experiment group - for ease of debugging and discovery.
 - I want to mark certain executions as “blessed” or “released”. This could be done through providing a semantic version after the execution is successful 
 - I want to group all “sub launchplan” executions with the parent execution.  
 - External systems could group executions based on some identifiers.  
 - Users could name their executions without having to worry about the character limits, uniqueness constraints and limited characterset.
 - Simplify filtering of certain executions

## 3 Proposed Implementation

### Support for tags

We propose to solve the problem of discovery by supporting arbitrary metadata association with an entity. This is similar to conceppt of “tags” as in AWS. The tags are represented as `“key”: “value”` pairs. In kuberenetes this can be represented using “labels” and “annotations”. Labels and annotations are already supported per execution as documented in - [ExecutionCreateRequest](https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#executioncreaterequest) -> [ExecutionSpec](https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#executionspec). Moreover,  every project supports default labels [Project](https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#project). Thus the final execution will have a the union of the project default labels + the user specified labels as it exists today.

Currently the resultant labels are not persisted and are only applied to the
execution in Kubernetes. As a first step, we recommend that these labels are
persisted associated with an execution and then ListExecutions API is updated
to return all 
 - filtered executions by labels with supported `and` `or` queries
 - all associated executions with every execution
 - limit total number of labels per exection to 10-15

Once this is implemented the UI and CLI can be updated to support these
queries.

### CLI Interface 

A workflow or task can be executed using

```bash
pyflyte run --remote --labels k:v --labels k1:v1 test.py wf --input1=10
```
 (or equivalently in flytectl)

flytectl and flyteremote can support filtering of executions by labels. Example
in flytectl,
```bash
flytectl get execution -p flytesnacks -d development --filter.labels="k:v"
```

### UI Interface

Two approaches

#### Approach 1: Treat all labels the same way and allow search/filter and click based grouping
In this approach users will get the regular executions view with all the labels
available on each execution. The users are allowed to filter an exection simply
by clicking on a label and then all executions are filtered by that label. 

#### Approach 2: Certain label keys are treated special
 - “group” will group everything
 - “experiment” will also group everything with higher priority. 
 - “name” will override the execution id with the name?


### Support for descriptions

Users want to describe their executions and especially once an experiment
succeeds they may want to add a lot more description and data about the
experiment. Thus, we propose to add descriptions to the execution as well.

Allow add description when you start an execution
```bash
pyflyte run --remote --labels k:v --labels k1:v1 --description "........" test.py wf --input1=10
```
 
It should be possible to add the description as a Markdown
```bash
pyflyte run --remote --labels k:v --labels k1:v1 --description README.md test.py wf --input1=10
```
 
It should be possible to add a description for an execution in the UI, after
the execution has been created.


## 4 Metrics & Dashboards
NA

## 5 Drawbacks
It is important to understand that this may add a little more stress to the
metadata database. Back of the envelope calculation

 -> 1 million executions * 20 labels each. 
 -> each label has "key" and "value". Key is 10 characters, value is 64
characters
 -> 1.5 * 10^9 bytes (assuming one byte per character" -> 1.5GB

This is not significant and will increase as executions increase

## 6 Alternatives
NA


## 7 Potential Impact and Dependencies
We this this is one of the most requested features in Flyte and will solve
a lot of problems.


## 8 Unresolved questions
NA

## 9 Conclusion
WIP
