######################
Contributing to Flyte
######################

Thank you for taking the time to contribute to Flyte! 
Please read our `Code of Conduct <https://lfprojects.org/policies/code-of-conduct/>`__ before contributing to Flyte.

Here are some guidelines for you to follow, which will make your first and follow-up contributions easier.

TL;DR: Find the repo-specific contribution guidelines in the `Component Reference <#component-reference>`__ section.

💻 Code
=======

An issue tagged with ``good first issue`` is the best place to start for first-time contributors. You can find them `here <https://github.com/flyteorg/flyte/labels/good%20first%20issue>`__.

**Appetizer for every repo: Fork and clone the concerned repository. Create a new branch on your fork and make the required changes. Create a pull request once your work is ready for review.** 

.. note::
    To open a pull request, follow `this guide <https://guides.github.com/activities/forking/>`__. 

Example PR for your reference: `GitHub PR <https://github.com/flyteorg/flytepropeller/pull/242>`__. 
A couple of checks are introduced to help maintain the robustness of the project. 

#. To get through DCO, sign off on every commit (`Reference <https://github.com/src-d/guide/blob/master/developer-community/fix-DCO.md>`__) 
#. To improve code coverage, write unit tests to test your code
#. Make sure all the tests pass. If you face any issues, please let us know

On a side note, format your Go code with ``golangci-lint`` followed by ``goimports`` (use ``make lint`` and ``make goimports``), and Python code with ``black`` and ``isort`` (use ``make fmt``). 
If make targets are not available, you can manually format the code.
Refer to `Effective Go <https://golang.org/doc/effective_go>`__, `Black <https://github.com/psf/black>`__, and `Isort <https://github.com/PyCQA/isort>`__ for full coding standards.

Good news -- you can be added as a committer to any ``flyteorg`` repo as you become more involved with the project ✨.

📃 Documentation
================

Flyte uses Sphinx for documentation. ``protoc-gen-doc`` is used to generate the documentation from ``.proto`` files.

Sphinx spans multiple repositories under `flyteorg <https://github.com/flyteorg>`__. It uses reStructured Text (rst) files to store the documentation content. 
For API- and code-related content, it extracts docstrings from the code files. 

To get started, refer to the `reStructuredText reference <https://www.sphinx-doc.org/en/master/usage/restructuredtext/index.html#rst-index>`__. 

For minor edits that don’t require a local setup, you can edit the GitHub page in the documentation to propose improvements.

The edit option can be found at the bottom of a page, as shown below.

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/contribution_guide/docs_edit.png
    :alt: GitHub edit option for Documentation
    :align: center
    :figclass: align-center

Intersphinx
***********

`Intersphinx <https://www.sphinx-doc.org/en/master/usage/extensions/intersphinx.html>`__ can generate automatic links to the documentation of objects in other projects.

To establish a reference to any other documentation from Flyte or within it, use Intersphinx. 

To do so, create an ``intersphinx_mapping`` in the ``conf.py`` file which should be present in the respective ``docs`` repository. 
For example, ``rsts`` is the docs repository for the ``flyte`` repo.

For example:

.. code-block:: python

    intersphinx_mapping = {
        "python": ("https://docs.python.org/3", None),
        "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/master/", None),
    }

The key refers to the name used to refer to the file (while referencing the documentation), and the URL denotes the precise location. 

Here is an example using ``:std:doc``:
 
* Direct reference
.. code-block:: text

    Task: :std:doc:`generated/flytekit.task`

Output:

Task: :std:doc:`generated/flytekit.task`

* Custom name
.. code-block:: text

    :std:doc:`Using custom words <generated/flytekit.task>`

Output:

:std:doc:`Using custom words <generated/flytekit.task>`

|

You can cross-reference multiple Python objects. Check out this `section <https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects>`__ to learn more. 

|

For instance, `task` decorator in flytekit uses the ``func`` role.

.. code-block:: text

    Link to flytekit code :py:func:`flytekit:flytekit.task`

Output:

Link to flytekit code :py:func:`flytekit:flytekit.task`

|

Here are a couple more examples.

.. code-block:: text

    :py:mod:`Module <python:typing>`
    :py:class:`Class <python:typing.Type>`
    :py:data:`Data <python:typing.Callable>`
    :py:func:`Function <python:typing.cast>`
    :py:meth:`Method <python:pprint.PrettyPrinter.format>`

Output:

:py:mod:`Module <python:typing>`

:py:class:`Class <python:typing.Type>`

:py:data:`Data <python:typing.Callable>`

:py:func:`Function <python:typing.cast>`

:py:meth:`Method <python:pprint.PrettyPrinter.format>`

🧱 Component Reference
======================

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/contribution_guide/dependency_graph.png
    :alt: Dependency graph between various flyteorg repos
    :align: center
    :figclass: align-center

    The dependency graph between various flyte.org repos


``flyte``
*********

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyte>`__
    * - **Purpose**: Deployment, Documentation, and Issues 
    * - **Languages**: Kustomize & RST
  
.. note::
    For the ``flyte`` repo, run the following command in the repo's root to generate documentation locally.

    .. code-block:: console

        make -C rsts html

``flyteidl``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteidl>`__
    * - **Purpose**: Flyte workflow specification is in `protocol buffers <https://developers.google.com/protocol-buffers>`__ which forms the core of Flyte
    * - **Language**: Protobuf
    * - **Guidelines**: Refer to the `README <https://github.com/flyteorg/flyteidl#generate-code-from-protobuf>`__
 
``flytepropeller``
******************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytepropeller>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller>`__
    * - **Purpose**: Kubernetes-native operator
    * - **Language**: Go
    * - **Guidelines:**

        * Check for Makefile in the root repo
        * Run the following commands:
           * ``make generate``
           * ``make test_unit``
           * ``make link``
        * To compile, run ``make compile``

``flyteadmin``
**************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteadmin>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin>`__
    * - **Purpose**: Control Plane
    * - **Language**: Go
    * - **Guidelines**:

        * Check for Makefile in the root repo
        * If the service code has to be tested, run it locally:
            * ``make compile``
            * ``make server``
        * To seed data locally:
            * ``make compile``
            * ``make seed_projects``
            * ``make migrate``
        * To run integration tests locally:
            * ``make integration``
            * (or to run in containerized dockernetes): ``make k8s_integration``

``flytekit``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytekit>`__
    * - **Purpose**: Python SDK & Tools
    * - **Language**: Python
    * - **Guidelines**: Refer to the `Flytekit Contribution Guide <https://docs.flyte.org/projects/flytekit/en/latest/contributing.html>`__

``flyteconsole``
****************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteconsole>`__
    * - **Purpose**: Admin Console
    * - **Language**: Typescript
    * - **Guidelines**: Refer to the `Flyte Console Contribution Guide <https://github.com/flyteorg/flyteconsole/CONTRIBUTING.md>`__

``datacatalog``
***************

.. list-table::

    * - `Repo <https://github.com/flyteorg/datacatalog>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/datacatalog>`__
    * - **Purpose**: Manage Input & Output Artifacts
    * - **Language**: Go

``flyteplugins``
****************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flyteplugins>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteplugins>`__
    * - **Purpose**: Flyte Plugins
    * - **Language**: Go
    * - **Guidelines**:

        * Check for Makefile in the root repo
        * Run the following commands:
            * ``make generate``
            * ``make test_unit``
            * ``make link``

``flytestdlib``
***************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytestdlib>`__
    * - **Purpose**: Standard Library for Shared Components
    * - **Language**: Go

``flytesnacks``
***************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytesnacks>`__
    * - **Purpose**: Examples, Tips, and Tricks to use Flytekit SDKs
    * - **Language**: Python (In the future, Java examples will be added)
    * - **Guidelines**: Refer to the `Flytesnacks Contribution Guide <https://docs.flyte.org/projects/cookbook/en/latest/contribute.html>`__

``flytectl``
************

.. list-table::

    * - `Repo <https://github.com/flyteorg/flytectl>`__
    * - **Purpose**: A standalone Flyte CLI
    * - **Language**: Go
    * - **Guidelines**: Refer to the `FlyteCTL Contribution Guide <https://docs.flyte.org/projects/flytectl/en/stable/contribute.html>`__    

🐞 File an Issue
================

`GitHub Issues <https://github.com/flyteorg/flyte/issues>`__ is used for issue tracking. The following are available issue types that you could use for filing an issue.

* `Plugin Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=untriaged%2Cplugins&template=backend-plugin-request.md&title=%5BPlugin%5D>`__
* `Bug Report <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2C+untriaged&template=bug_report.md&title=%5BBUG%5D+>`__
* `Documentation Bug/Update Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2C+untriaged&template=docs_issue.md&title=%5BDocs%5D>`__
* `Core Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged&template=feature_request.md&title=%5BCore+Feature%5D>`__
* `Flytectl Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+flytectl&template=flytectl_issue.md&title=%5BFlytectl+Feature%5D>`__
* `Housekeeping <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=housekeeping&template=housekeeping_template.md&title=%5BHousekeeping%5D+>`__
* `UI Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+ui&template=ui_feature_request.md&title=%5BUI+Feature%5D>`__

If none of the above fit your requirements, file a `blank <https://github.com/flyteorg/flyte/issues/new>`__ issue.

|

For feedback at any point in the contribution process, feel free to reach out to us on `Slack <http://flyte-org.slack.com/>`__. 
If you haven't joined Slack yet, you can do so by clicking `this link <https://docs.google.com/forms/d/e/1FAIpQLScWPJZncyL-bBCWbuQ9HWuMzA8-r2RJc__CSWJoAYUQADN-BQ/viewform>`__.
