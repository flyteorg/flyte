# Flyte Execution Tags and Metadata

**Authors:**

- @kumare3

## 1 Executive Summary

Flyte currently provides no visual ways of grouping executions and other
entities apart from with a “project” and “domain”. Usually a project is used by a team of
engineers and/or researchers, and sometimes they may be  experimenting in
isolation or running multiple co-horts of experiments. Also, there are cases in
which grouping executions by a group might make them easier to find and
associate them better. This document provides a solution of how we could
improve the experience of discovering executions within Flyte. It also provides
motivation of how this feature could further improve other ecosystem projects.

## 2 Motivation

As a User I want to
 - Group a certain number of executions into an experiment group - for ease of debugging and discovery - when launching the workflows, while they are running or after they have already finished.
 - I want to mark certain executions as “blessed” or “released”. This could be done through providing a semantic version after the execution is successful 
 - I want to group all “sub launchplan” executions with the parent execution.  
 - External systems could group executions based on some identifiers.  
 - Users could name their executions without having to worry about the character limits, uniqueness constraints and limited characterset.
 - Simplify filtering of certain executions
 - Be able to remove tags from an execution after it has been started, or after it has finished.

## 3 Proposed Implementation

### Support for tags

We propose to solve the problem of discovery by supporting arbitrary metadata association with an entity. This is similar to concept of “tags” as in AWS.
The tags are represented as plain string.
We'll add tags to [ExecutionCreateRequest](https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#executioncreaterequest)  -> [ExecutionSpec](https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/admin/admin.html#executionspec).

The resultant tags will be persisted in the database, instead of being applied to the
execution in Kubernetes. We'll create two new tables in the flyteadmin database.
- ``execution_admin_tags`` is a join table, and it contains admin_tags ID and execution ID.
- ``admin_tags`` saves all the tag names

As a first step, we recommend that these tags are
persisted associated with an execution and limit total number of tags per execution to 10-15
The ListExecutions API is updated to return all 
 - Filtered executions by tags with supported `and` `or` queries
 - All associated executions with every execution

For the second step, we will support attaching tags to the project, task, and workflow. In addition,
we will enable users to update the tags of an execution after it has been created. This will be done
through a flyteadmin API (need to create a new endpoint for updating the tags).

Once this is implemented the UI and CLI can be updated to support these
queries.

### CLI Interface 

A workflow or task can be executed using

```bash
pyflyte run --remote --tags '["hello", "world"]' test.py wf --input1=10
```
 (or equivalently in flytectl)

flytectl and flyte remote can support filtering of executions by tags. Example
in flytectl,
```bash
flytectl get execution -p flytesnacks -d development --filter.tags="hello,world"
```

### UI Interface
All tags are treated the same way and allow search/filter and click based grouping.
Users will get the regular executions view with all the tags available on each execution.
The users are allowed to filter an execution simply by clicking on a tag and then all
executions are filtered by that tag. 

![Grouping / Filtering UX](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/rfc/tags/labels-filter.png)


### Support for descriptions

Users want to describe their executions and especially once an experiment
succeeds they may want to add a lot more description and data about the
experiment. Thus, we propose to add descriptions to the execution as well.

Allow add description when you start an execution
```bash
pyflyte run --remote --tags '["key1", "key2"]' --description "........" test.py wf --input1=10
```
 
It should be possible to add the description as a Markdown
```bash
pyflyte run --remote --tags '["key1", "key2"]' --description README.md test.py wf --input1=10
```
 
It should be possible to add a description for an execution in the UI, after
the execution has been created.
![UI Descriptions](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/rfc/tags/description-edit.png)

## 4 Metrics & Dashboards
NA

## 5 Drawbacks
It is important to understand that this may add a little more stress to the
metadata database. Back of the envelope calculation

 -> 1 million executions * 20 tags each. 
 -> Each tag is a plain string, which has 64 characters.
 -> 1.28 * 10^9 bytes (assuming one byte per character" -> 1.28GB

This is not significant even though it will increase as executions increase

## 6 Alternatives
NA


## 7 Potential Impact and Dependencies
This is one of the most requested features in Flyte and will solve
a lot of problems.


## 8 Unresolved questions
NA

## 9 Conclusion
With tags, users can discover their executions and other flyte entities easily.
By storing tags in the database, it allows flyte easily search and filter executions, and also allows us to easily update/remove the tags of an execution after it has been created. 
