# Contributing to Flyte

Thank you for taking the time to contribute to Flyte!
Please read our [Code of Conduct](https://lfprojects.org/policies/code-of-conduct/) before contributing to Flyte.

Here are some guidelines for you to follow, which will make your first and follow-up contributions easier.

TL;DR: Find the repo-specific contribution guidelines in the [Component Reference](#component-reference) section.

## Becoming a Contributor

An issue tagged with [`good first issue`](https://github.com/flyteorg/flyte/labels/good%20first%20issue) is the best place to start for first-time contributors.

**Appetizer for every repo: Fork and clone the concerned repository. Create a new branch on your fork and make the required changes. Create a pull request once your work is ready for review.** 

> To open a pull request, follow [the guide](https://guides.github.com/activities/forking/) by GitHub. 

Example PR for your reference: [GitHub PR](https://github.com/flyteorg/flytepropeller/pull/242). 
A couple of checks are introduced to help maintain the robustness of the project. 

- To get through DCO, [sign off](https://github.com/src-d/guide/blob/master/developer-community/fix-DCO.md) on every commit.
- To improve code coverage, write unit tests to test your code.
- Make sure all the tests pass. If you face any issues, please let us know!

On a side note, format your Go code with `golangci-lint` followed by `goimports` (use `make lint` and `make goimports`), and Python code with `black` and `isort` (use `make fmt`). 
If make targets are not available, you can manually format the code.
Refer to [Effective Go](https://golang.org/doc/effective_go), [Black](https://github.com/psf/black), and [Isort](https://github.com/PyCQA/isort) for full coding standards.

As you become more involved with the project, you may be added as a contributor to the repos you're working on, but there is a medium term effort to move all development to forks ✨.

## Documentation

Flyte uses Sphinx for documentation. `protoc-gen-doc` is used to generate the documentation from `.proto` files.

Sphinx spans multiple repositories under [flyteorg](https://github.com/flyteorg). It uses reStructured Text (rst) files to store the documentation content. 
For API- and code-related content, it extracts docstrings from the code files. 

To get started, refer to the [reStructuredText reference](https://www.sphinx-doc.org/en/master/usage/restructuredtext/index.html#rst-index). 

For minor edits that don’t require a local setup, you can edit the GitHub page in the documentation to propose improvements.

### Intersphinx

[Intersphinx](https://www.sphinx-doc.org/en/master/usage/extensions/intersphinx.html) can generate automatic links to the documentation of objects in other projects.

To establish a reference to any other documentation from Flyte or within it, use Intersphinx. 

To do so, create an `intersphinx_mapping` in the `conf.py` file which should be present in the respective `docs` repository. 
For example, `rsts` is the docs repository for the `flyte` repo.

For example:

```python
intersphinx_mapping = {
    "python": ("https://docs.python.org/3", None),
    "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/master/", None),
}
```

The key refers to the name used to refer to the file (while referencing the documentation), and the URL denotes the precise location. 

You can also cross-reference multiple Python objects. Check out this [section](https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects) to learn more. 

## Component Reference

To understand how the below components interact with each other, refer to [Understand the lifecycle of a workflow](https://docs.flyte.org/en/latest/concepts/workflow_lifecycle.html).

<img src="https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/contribution_guide/dependency_graph.png" alt="Dependency graph between various flyteorg repos">

### `flyte`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flyte) |
| **Purpose**: Deployment, Documentation, and Issues |
| **Languages**: Kustomize & RST |
  
> For the `flyte` repo, run the following command in the repo's root to generate documentation locally.
> ```
> make -C rsts html
> ```

### `flyteidl`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flyteidl) |
| **Purpose**: Flyte workflow specification is in [protocol buffers](https://developers.google.com/protocol-buffers) which forms the core of Flyte |
| **Language**: Protobuf |
| **Guidelines**: Refer to the [README](https://github.com/flyteorg/flyteidl#generate-code-from-protobuf) |
 
### `flytepropeller`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flytepropeller) | [Code Reference](https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller) |
| **Purpose**: Kubernetes-native operator |
| **Language**: Go |
| **Guidelines:** <ul><li>Check for Makefile in the root repo</li><li>Run the following commands:<ul><li>`make generate`</li><li>`make test_unit`</li><li>`make link`</li></ul></li><li>To compile, run `make compile`</li></ul> |

### `flyteadmin`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flyteadmin) | [Code Reference](https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin) |
| **Purpose**: Control Plane |
| **Language**: Go |
| **Guidelines**: <ul><li>Check for Makefile in the root repo</li><li>If the service code has to be tested, run it locally:<ul><li>`make compile`</li><li>`make server`</li></ul></li><li>To seed data locally:<ul><li>`make compile`</li><li>`make seed_projects`</li><li>`make migrate`</li></ul></li><li>To run integration tests locally:<ul><li>`make integration`</li><li>(or to run in containerized dockernetes): `make k8s_integration`</li></ul></li></ul> |


### `flytekit`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flytekit) |
| **Purpose**: Python SDK & Tools |
| **Language**: Python |
| **Guidelines**: Refer to the [Flytekit Contribution Guide](https://docs.flyte.org/projects/flytekit/en/latest/contributing.html) |

### `flyteconsole`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flyteconsole) |
| **Purpose**: Admin Console |
| **Language**: Typescript |
| **Guidelines**: Refer to the [README](https://github.com/flyteorg/flyteconsole/blob/master/README.md) |

### `datacatalog`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/datacatalog) | [Code Reference](https://pkg.go.dev/mod/github.com/flyteorg/datacatalog) |
| **Purpose**: Manage Input & Output Artifacts |
| **Language**: Go |

### `flyteplugins`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flyteplugins) | [Code Reference](https://pkg.go.dev/mod/github.com/flyteorg/flyteplugins) |
| **Purpose**: Flyte Plugins |
| **Language**: Go |
| **Guidelines**: <ul><li>Check for Makefile in the root repo</li><li>Run the following commands:<ul><li>`make generate`</li><li>`make test_unit`</li><li>`make link`</li></ul></ul> |

### `flytestdlib`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flytestdlib) |
| **Purpose**: Standard Library for Shared Components |
| **Language**: Go |

### `flytesnacks`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flytesnacks) |
| **Purpose**: Examples, Tips, and Tricks to use Flytekit SDKs |
| **Language**: Python (In the future, Java examples will be added) |
| **Guidelines**: Refer to the [Flytesnacks Contribution Guide](https://docs.flyte.org/projects/cookbook/en/latest/contribute.html) |

### `flytectl`

[]()  | 
| ------|
| [Repo](https://github.com/flyteorg/flytectl) |
| **Purpose**: A standalone Flyte CLI |
| **Language**: Go |
| **Guidelines**: Refer to the [FlyteCTL Contribution Guide](https://docs.flyte.org/projects/flytectl/en/stable/contribute.html) |

## Recommended iteration cycle

As you may have already read in other parts of the documentation, the [Flyte repository](https://github.com/flyteorg/flyte) includes Go code that integrates all backend components (admin, propeller, data catalog, console) into a single executable. The Flyte team is currently working on consolidating the core backend repositories into one repository, which is expected to be completed by 2023. In the meantime, you can contribute to the individual repositories and then merge your changes into the [Flyte repository](https://github.com/flyteorg/flyte). This setup is suitable for Go-based backend development, but it has not been tested for Flyteconsole development, which has a different development cycle. Nonetheless, this setup allows you to run the Flyte binary from your IDE, enabling you to debug your code effectively by setting breakpoints. Additionally, this setup connects you to all other resources in the demo environment, such as PostgreSQL and RDS.

### Dev mode cluster

To launch the dependencies, teardown any old sandboxes you may have, and then run:

```
flytectl demo start --dev
```

This command will launch the demo environment without running Flyte. By doing so, developers can run Flyte later on their host machine.

### Set up Flyte configuration

- Copy the file `flyte-single-binary-local.yaml` to `~/.flyte/local-dev-config.yaml`.
- Replace occurrences of `$HOME` with the actual path of your home directory.

### Cluster resources

One of the configuration entries you will notice is `cluster_resources.templatePath`. This folder should contain the templates that the cluster resource controller will use. To begin, you can create a file called `~/.flyte/cluster-resource-templates/00_namespace.yaml` with the following content:

```yaml
clusterResourceTemplates:
  inline:
    001_namespace.yaml: |
      apiVersion: v1
      kind: Namespace
      metadata:
        name: '{{ namespace }}'
```

### Pull console artifacts

Run the following command from the base folder of the Flyte repository to pull in the static assets for Flyteconsole:

```
make cmd/single/dist
```

### Build & iterate

To bring in the code of the component you are testing, use the command go get `github.com/flyteorg/<component>&gitsha`. Once you have done that, you can run the following command:

```
POD_NAMESPACE=flyte go run -tags console cmd/main.go start --config ~/.flyte/local-dev-config.yaml
```

The `POD_NAMESPACE` environment variable is necessary for the webhook to function correctly. You can also create a build target in your IDE with the same command.

Once it is up and running, you can access Flyte hosted by your local machine by going to `localhost:30080/console`. The Docker host mapping is used to obtain the correct IP address for your local host.

## File an Issue

We use [GitHub Issues](https://github.com/flyteorg/flyte/issues) for issue tracking. The following issue types are available for filing an issue:

* [Plugin Request](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=untriaged%2Cplugins&template=backend-plugin-request.md&title=%5BPlugin%5D)
* [Bug Report](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2C+untriaged&template=bug_report.md&title=%5BBUG%5D+)
* [Documentation Bug/Update Request](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2C+untriaged&template=docs_issue.md&title=%5BDocs%5D)
* [Core Feature Request](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged&template=feature_request.md&title=%5BCore+Feature%5D)
* [Flytectl Feature Request](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+flytectl&template=flytectl_issue.md&title=%5BFlytectl+Feature%5D)
* [Housekeeping](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=housekeeping&template=housekeeping_template.md&title=%5BHousekeeping%5D+)
* [UI Feature Request](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+ui&template=ui_feature_request.md&title=%5BUI+Feature%5D)

If none of the above fit your requirements, file a [blank](https://github.com/flyteorg/flyte/issues/new) issue.
Also, add relevant labels to your issue. For example, if you are filing a Flytekit plugin request, add the `flytekit` label.

For feedback at any point in the contribution process, feel free to reach out to us on [Slack](https://slack.flyte.org/). 

