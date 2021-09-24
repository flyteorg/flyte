######################
Contributing to Flyte
######################

Thank you for taking the time to contribute to Flyte! Here are some guidelines for you to follow, which will make your first and follow-up contributions easier.

.. note::
    Please read our `Code of Conduct <https://lfprojects.org/policies/code-of-conduct/>`__ before contributing to Flyte.

Code
====
An issue tagged with ``good first issue`` is the best place to start for first-time contributors. You can find them `here <https://github.com/flyteorg/flyte/labels/good%20first%20issue>`__.

To take a step ahead, check out the repositories available under `flyteorg <https://github.com/flyteorg>`__.

**Appetizer (for every repo): Fork and clone the concerned repository. Create a new branch on your fork and make the required changes. Create a pull request once your work is ready for review.** 

.. note::
    Note: To open a pull request, follow this `guide <https://guides.github.com/activities/forking/>`__.

*A piece of good news -- You can be added as a committer to any ``flyteorg`` repo as you become more involved with the project.*

Example PR for your reference: `GitHub PR <https://github.com/flyteorg/flytepropeller/pull/242>`__. A couple of checks are introduced to help in maintaining the robustness of the project. 

#. To get through DCO, sign off on every commit. (`Reference <https://github.com/src-d/guide/blob/master/developer-community/fix-DCO.md>`__) 
#. To improve code coverage, write unit tests to test your code.
#. Make sure all the tests pass. If you face any issues, please let us know.

.. note::
    Format your Go code with ``golangci-lint`` followed by ``goimports`` (we used the same in the `Makefile <https://github.com/flyteorg/flytepropeller/blob/eaf084934de5d630cd4c11aae15ecae780cc787e/boilerplate/lyft/golang_test_targets/Makefile#L11-L19>`__), and Python code with ``black`` (use ``make fmt`` command which contains both black and isort). 
    Refer to `Effective Go <https://golang.org/doc/effective_go>`_ and `Black <https://github.com/psf/black>`_ for full coding standards.

Component Reference
===================

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/contribution_guide/dependency_graph.png
    :alt: Dependency Graph between various flyteorg repos
    :align: center
    :figclass: align-center

    The dependency graph between various flyteorg repos


``flyte``
*********

.. list-table::

    * - `Repo <https://github.com/lyft/flyte>`__
    * - **Purpose**: Deployment, Documentation, and Issues 
    * - **Languages**: Kustomize & RST

``flyteidl``
************

.. list-table::

    * - `Repo <https://github.com/lyft/flyteidl>`__
    * - **Purpose**: The Flyte Workflow specification in `protocol buffers <https://developers.google.com/protocol-buffers>`__ which forms the core of Flyte
    * - **Language**: Protobuf
    * - **Setup**: Refer to the `README <https://github.com/flyteorg/flyteidl#generate-code-from-protobuf>`__
 
``flytepropeller``
******************

.. list-table::

    * - `Repo <https://github.com/lyft/flytepropeller>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flytepropeller>`__
    * - **Purpose**: Deployment, Documentation, and Issues 
    * - **Languages**: Kustomize & RST
    * - **Setup:**

        * Check for the Makefile in the root repo
        * Run the following commands:
           * ``make generate``
           * ``make test_unit``
           * ``make link``
        * To compile, run ``make compile``

``flyteadmin``
**************

.. list-table::

    * - `Repo <https://github.com/lyft/flyteadmin>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin>`__
    * - **Purpose**: Control Plane
    * - **Language**: Go
    * - **Setup**:

        * Check for the Makefile in the root repo
        * If the service code has to be tested, run it locally:
            * ``make compile``
            * ``make server``
        * To seed data locally:
            * ``make compile``
            * ``make seed_projects``
            * ``make migrate``
        * To run integration tests locally:
            * ``make integration``
            * (or, to run in containerized dockernetes): ``make k8s_integration``

``flytekit``
************

.. list-table::

    * - `Repo <https://github.com/lyft/flytekit>`__
    * - **Purpose**: Python SDK & Tools
    * - **Language**: Python
    * - **Setup**: Refer to the `Flytekit Contribution Guide <https://docs.flyte.org/projects/flytekit/en/latest/contributing.html>`__

``flyteconsole``
****************

.. list-table::

    * - `Repo <https://github.com/lyft/flyteconsole>`__
    * - **Purpose**: Admin Console
    * - **Language**: Typescript
    * - **Setup**: Refer to the `README <https://github.com/flyteorg/flyteconsole#running-flyteconsole>`__

``datacatalog``
***************

.. list-table::

    * - `Repo <https://github.com/lyft/datacatalog>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/datacatalog>`__
    * - **Purpose**: Manage Input & Output Artifacts
    * - **Language**: Go

``flyteplugins``
****************

.. list-table::

    * - `Repo <https://github.com/lyft/flyteplugins>`__ | `Code Reference <https://pkg.go.dev/mod/github.com/flyteorg/flyteplugins>`__
    * - **Purpose**: Flyte Plugins
    * - **Language**: Go
    * - **Setup**:

        * Check for the Makefile in the root repo
        * Run the following commands:
            * ``make generate``
            * ``make test_unit``
            * ``make link``

``flytestdlib``
***************

.. list-table::

    * - `Repo <https://github.com/lyft/flytestdlib>`__
    * - **Purpose**: Standard Library for Shared Components
    * - **Language**: Go

``flytesnacks``
***************

.. list-table::

    * - `Repo <https://github.com/lyft/flytesnacks>`__
    * - **Purpose**: Examples, Tips, and Tricks to use Flytekit SDKs
    * - **Language**: Python (In future, Java shall be added)
    * - **Setup**: Refer to the `README <https://github.com/flyteorg/flytesnacks#----------contribution-guide---->`__

``flytectl``
************

.. list-table::

    * - `Repo <https://github.com/lyft/flytectl>`__
    * - **Purpose**: A Standalone Flyte CLI
    * - **Language**: Go
    * - **Setup**:

        * Check for the Makefile in the root repo
        * Run the following commands:
            * ``make generate``
            * ``make test_unit``
            * ``make link``    

Issues
======
`GitHub Issues <https://github.com/flyteorg/flyte/issues>`__ is used for issue tracking. There are a variety of issue types available that you could use while filing an issue.

* `Plugin Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=untriaged%2Cplugins&template=backend-plugin-request.md&title=%5BPlugin%5D>`__
* `Bug Report <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2C+untriaged&template=bug_report.md&title=%5BBUG%5D+>`__
* `Documentation Bug/Update Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2C+untriaged&template=docs_issue.md&title=%5BDocs%5D>`__
* `Core Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged&template=feature_request.md&title=%5BCore+Feature%5D>`__
* `Flytectl Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+flytectl&template=flytectl_issue.md&title=%5BFlytectl+Feature%5D>`__
* `Housekeeping <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=housekeeping&template=housekeeping_template.md&title=%5BHousekeeping%5D+>`__
* `UI Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+ui&template=ui_feature_request.md&title=%5BUI+Feature%5D>`__

If none of the above fits your requirements, file a `blank <https://github.com/flyteorg/flyte/issues/new>`__ issue.

Documentation
=============
Flyte uses Sphinx for documentation and ``godocs`` for Golang. ``godocs`` is quite simple -- comment your code and you are good to go!

Sphinx spans across multiple repositories under the `flyteorg <https://github.com/flyteorg>`__ repository. It uses reStructured Text (rst) files to store the documentation content. For both the API and code-related content, it extracts docstrings from the code files. 

To get started, look into `reStructuredText reference <https://www.sphinx-doc.org/en/master/usage/restructuredtext/index.html#rst-index>`__. 

Docs Environment Setup
**********************

Install all the requirements from the `docs-requirements.txt` file present in the root of a repository.

.. code-block:: console

    pip install -r docs-requirements.txt

From the ``docs`` directory present in the repository root (for ``flytesnacks``, ``docs`` is present in ``flytesnacks/cookbook``), run the command:

.. code-block:: console

    make html

.. note::
    For implicit targets, run ``make -C docs html``. 

You can then view the HTML pages in the ``docs/_build`` directory.

.. note::
    For ``flyte`` repo, there is no ``docs`` directory. Instead, consider the ``rsts`` directory. To generate HTML files, run the following command in the root of the repo.

    .. code-block:: console

        make -C rsts html

For minor edits that don’t require a local setup, you can edit GitHub page in the documentation to propose improvements.

The edit option can be found at the bottom of a page, as shown below.

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/contribution_guide/docs_edit.png
    :alt: GitHub edit option for Documentation
    :align: center
    :figclass: align-center

Intersphinx
***********
`Intersphinx <https://www.sphinx-doc.org/en/master/usage/extensions/intersphinx.html>`__ can generate automatic links to the documentation of objects in other projects.

To establish a reference to any other documentation from Flyte or within it, use intersphinx. 

To do so, create an ``intersphinx_mapping`` in the ``conf.py`` file present in the ``docs/source`` directory.

For example:

.. code-block:: python

    intersphinx_mapping = {
        "python": ("https://docs.python.org/3", None),
        "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/master/", None),
    }

.. note::
    ``docs/source`` is present in the repository root. Click `here <https://github.com/flyteorg/flytekit/blob/55505c4a6f0240d8273eb16febcad64623764929/docs/source/conf.py#L194-L200>`__ to view the intersphinx configuration.

The key refers to the name used to refer to the file (while referencing the documentation), and the URL denotes the precise location. 

Here are a couple of examples that you can refer to:

1. 
.. code-block:: text

    Task: :std:doc:`generated/flytekit.task`

Output:

Task: :std:doc:`generated/flytekit.task`

2.
.. code-block:: text

    :std:doc:`Using custom words <generated/flytekit.task>`

Output:

:std:doc:`Using custom words <generated/flytekit.task>`

|

Linking to Python elements can change based on what you are linking to. Check out this `section <https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects>`__ to learn more. 

|

For instance, linking to the `task` decorator in flytekit uses the ``func`` role.

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


For feedback at any point in the contribution process, feel free to reach out to us using the links on our `Community <https://docs.flyte.org/en/latest/community/index.html>`_ page.
