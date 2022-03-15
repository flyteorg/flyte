# Introduction

This document proposes, **Flytectl** as a single CLI that interacts with FlyteAdmin service. It is proposed to write the CLI in **Golang** and would support both gRPC and REST endpoints of
FlyteAdmin. We will start with gRPC endpoint since the client can be easily generated. In the future, we will work on generating a Swagger-based REST client from the gRPC specification. As we build
more SDKs in different languages we will support a common way of interacting with the API. This doesn't mean that some SDKs may provide native ways of interacting with the Admin API (for e.g.
Flytekit), but the intention is to eventually replace **Flytekit/Flytecli** with Flytectl exclusively.

Flytectl has been designed to deliver user features without having to rely on the UI. Flytectl will follow the standard oauth2 for
authentication, supported by FlyteAdmin. Moreover, Flytectl should be readily available on almost any platform - OSX, Linux and Windows. We will strive to keep it relatively lean and fast.

# Why One CLI?

As we build multiple SDKs, we need a native way of interacting with the API. Having multiple CLIs makes it hard to keep all of them in sync as we rapidly evolve the API and add more features.

*Diagram here*


# Why Golang?
- Most of Flyte backend is written in Golang.
- Golang offers great CLI tooling support with viper and cobra.
- Golang toolchain helps create cross-compiled small, light weight binary, which is efficient and easy to use.
- We already generate Golang proto and clients for all our IDL.
- We have multiple common libraries available to ease the development of this tool.
- Kubectl is a stellar example of a CLI done well.

## Generating Swagger code
We started exploring this (Flytetools)[https://github.com/lyft/flytetools#tools] has some work. The Swagger code-gen maintainer also approached us to see if they could help.

# API

## Top level commands

```bash
$ flytectl [options]
  version
  configure
  get
  create
  update
  delete 
```

### base options
- *endpoint* endpoint where Flyteadmin is available
- *insecure* use if Oauth is not available
- optional *project* project for which we need to retrieve details
- optional *domain* domain for which we need to retrieve details
- TBD

### version
returns the version of the CLI, version of Admin service and version of the Platform that is deployed

### configure
Allows configuring Flytectl for your own usage (low pri). Needed for especially storing Auth tokens.

### get/delete
Get retrieves a list of resources that is qualified by a further sub-command. for example
```bash
$ flytectl --endpoint "example.flyte.net" get projects
$ flytectl --endpoint "example.flyte.net" --project "p" --domain "d" delete workflows
```
This returns a list of projects

To retrieve just one project
```bash
$ flytectl --endpoint "example.flyte.net" get projects <project-name>
$ flytectl --endpoint "example.flyte.net" --project "p" --domain "d" delete workflows "W1"
```

### Create is special
Create may need more information than can be easily passed in command line and we recommend using files to create an entity. The file could be in protobuf, jsonpb (json) or jsonpb (yaml) form.
Eventually we may want to simplify the json and yaml representations but that is not required in first pass. We may also want to create just a separate option for that.

The create for Task and Workflow is essential what is encompassed in the pyflyte as the registration process. We will decouple the registration process such that pyflyte, jflyte (other native cli's or
code methods) can dump a serialized representations of the workflows and tasks that are directly consumed by **Flytectl**. Thus Flytectl is essential in every flow for the user.


#### Create Templatization
User-facing SDKs can serialize workflow code to protobuf representations but these will be incomplete. Specifically, the _project_, _domain_, and _version_ parameters must be supplied at create time since these are attributes of the registerable, rather than serialized object. Placeholder template variables including:

* `{{ .project }}`
* `{{ .domain }}`
* `{{ .version }}`
* [auth](https://github.com/flyteorg/flyteidl/blob/c3baba8983019680ef57b6244cea36ba951233ed/protos/flyteidl/admin/common.proto#L241): including the assumable_iam_role and/or kubernetes_service_account
* the [output_location_prefix](https://github.com/flyteorg/flyteidl/blob/c3baba8983019680ef57b6244cea36ba951233ed/protos/flyteidl/admin/common.proto#L250)

will be included in the serialized protobuf that must be substituted at **create** time.  Eventually the hope is that substitution will be done server-side.

Furthermore, to reproduce the equivalent **fast-register** code path for the flyte-cli defined in flytekit an equivalent _fast-create_ command must fill in additional template variables in the [task container args](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L142). These serialized, templatized args will appear like so:

```
"pyflyte-fast-execute",
"--additional-distribution",
"{{ .remote_package_path }}",
"--dest-dir",
"{{ .dest_dir }}",
"--",
"pyflyte-execute",
...
```

The `remote package path` is determined by uploading the compressed user code (produced in the serialize step) to a user-specified remote directory (called `additional-distribution-dir` in flytekit). In the case of fast-create the code _version_ arg can be deterministcally assigned when serializing the code. Compressed code archives uploaded as individual files to the remote directory can assume the version name to guarantee uniqueness.

The `dest dir` is an optional argument specified by the user to designate where code is downloaded at execution time.

![Registration process](flytectl_interaction.png)

### update
This is a lower priority option as most entities in flyte are immutable and do not support updates. For the ones where update is supported, we should look into retrieving the existing and allow editing in an editor, like kubectl edit does.

**To be specified**


# Details of each resource

## Projects
Projects are top level entity in Flyte. You can fetch multiple projects or one project using the CLI. Think about projects like namespaces.

 - create
```bash
$ flytectl create projects --name "Human readable Name of project" --id project-id --labels key=value --labels key=value --description "long string"
Alternatively
$ flytectl create project -f project.yaml
```

```yaml
project.yaml
name: Human readable project name
id: project-x
labels:
  - k: v
  - k1: v1
description: |
  Long description
```
 - get
```bash
$ flytectl get projects [project-name] [-o yaml | -o json | default -o table]
```
 - update
```bash
$ flytectl update projects --id project-x ...
# You can only update one project at a time
```

## Tasks
 - get
```bash
$ flytectl get tasks [task-name] [-o yaml | -o json | default -o table] [--filters...] [--sort-by...] [--selectors...]
```
 - get specific version and get a template to launch
   Create an execution is complicated as the user needs to know all the input types and  way to simplify this could be to create a YAML template locally from the launchplan (the interface, etc)
```bash
$ flytectl get task task-name --execution-template -o YAML
yaml.template (TBD)
This is a special version of get launch-plan which can be executed by passing it to create execution.

```
 - create
 - create
 - update

## Workflows
Support
 - get
```bash
$ flytectl get workflows [workflow-name] [-o yaml | -o json | default -o table] [--filters...] [--sort-by...] [--selectors...]
```
 - create
 - update

## Launch Plans
Support
 - get
```bash
$ flytectl get launch-plans [launchplan-name] [-o yaml | -o json | default -o table] [--filters...] [--sort-by...] [--selectors...]
```
 - get specific version and get a template to launch
   Create an execution is complicated as the user needs to know all the input types and  way to simplify this could be to create a YAML template locally from the launchplan (the interface, etc)
```bash
$ flytectl get launch-plans launch-plan-name --execution-template -o YAML
yaml.template (TBD)
This is a special version of get launch-plan which can be executed by passing it to create execution.

```
 - create
 - update

## Execution
Create or retrieve an execution. 
 - get
Get all executions or get a single execution.
```bash
$ flytectl get execution [exec-name] [-o yaml | -o json | default -o table] [--filters...] [--sort-by...] [--selectors...]
```
An interesting feature in get-execution might be to filter within the execution only the execution of a node, or quickly find the ones that have failed. 
Visualizing the execution is also challenging. We may want to visualize
We could use https://graphviz.org/ to visualize the DAG.
Within the DAG, NodeExecutions and corresponding task executions need to be fetched.
 - create
   Create an execution for a LaunchPlan or a Task. This is very interesting as it should accept inputs for the execution.
```bash
$ flytectl create execution -f template.yaml (see get-template command)
OR
$ flytectl create execution --launch-plan "name" --inputs "key=value"
```
 - delete - here refers to terminate

## MatchableEntity
Ability to retrieve matchable entity and edit its details
 - get
 - create
 - update

## Outputs
Support
 - get
 - create
 - update

# No resource interactions

## Install all examples
Today Flytesnacks houses a few examples for Flyte usage in python. When a user wants to get started with Flyte quickly it would be preferable that all Flytesnacks examples are serialized and stored as artifacts in flytesnacks for every checkin. This can be done for python flytekit using `pyflyte serialize` command. Once they are posted as serialized blobs, flytectl could easily retrieve them and register them in a specific project as desired by the user.

```bash
$ flytectl examples register-all [cookbook|plugins|--custom-path=remote-path] [--semver semantic-version-of-flytesnacks-examples] --target-project --target-domain
```
The remote has to follow a protocol. It should be an archive - `tar.gz` with two folders `example-set/ -tasks/*.pb -workflows/*.pb` All the workflows in this path will be installed to the target project / domain

## Setup a repository with dockerfile for writing code for Flyte
Maybe we should look at `boilr` or some other existing framework to do this
```bash
$ flytectl init project --archetype tensorflow-2.0
$ flytectl init project --archetype spark-3.0
$ flytectl init project --archetype xgboost
...
```
For this to work, all these archetypes should be available in a separate repository. An archetype is essentially a template with dockerfile and folder setup with flytekit.config
