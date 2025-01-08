.. _contribute_Flyte:

#####################
Contributing to Flyte
#####################

.. tags:: Contribute, Basic

Thank you for taking the time to contribute to Flyte!

Here are some guidelines for you to follow, which will make your first and follow-up contributions easier.

TL;DR: Find the repo-specific contribution guidelines in the :ref:`Component Reference <component_reference>` section.

💻 Becoming a contributor
=========================

An issue tagged with `good first issue <https://github.com/flyteorg/flyte/labels/good%20first%20issue>`__ is the best place to start for first-time contributors.

**Fork and clone the concerned repository. Create a new branch on your fork and make the required changes. Create a pull request once your work is ready for review.**

.. note::
    To open a pull request, refer to `GitHub's guide <https://guides.github.com/activities/forking/>`__ for detailed instructions.

Example PR for your reference: `GitHub PR <https://github.com/flyteorg/flytepropeller/pull/242>`__.
Several checks are introduced to help maintain the robustness of the project:

#. To get through DCO, sign off on every commit (`Reference <https://github.com/src-d/guide/blob/master/developer-community/fix-DCO.md>`__)
#. To improve code coverage, write unit tests to test your code.
#. Make sure all the tests pass. If you face any issues, please let us know in the `#contribute <https://flyte-org.slack.com/archives/C04NJPLRWUX>`__ channel
#. Format your Go code with ``golangci-lint`` followed by ``goimports`` (use ``make lint`` and ``make goimports``)
#. Format your Python code with ``black`` and ``isort`` (use ``make fmt``).
#. If make targets are not available, you can manually format the code.

.. note::
    Refer to `Effective Go <https://golang.org/doc/effective_go>`__, `Black <https://github.com/psf/black>`__, and `Isort <https://github.com/PyCQA/isort>`__ for full coding standards.

As you become more involved with the project, you may be able to be added as a committer to the repos you're working on. Checkout the `Flyte Contributor Ladder <https://github.com/flyteorg/community/blob/main/GOVERNANCE.md#community-roles-and-path-to-maintainership>`__ to learn more. 

Before submitting your PR
**************************

We stronly encourage you to add one of these labels to your Pull Request:

- **added**: For new features.
- **changed**: For changes in existing functionality.
- **deprecated**: For soon-to-be-removed features.
- **removed**: For features being removed.
- **fixed**: For any bug fixes.
- **security**: In case of vulnerabilities

This is helpful to build human-readable release notes. `Learn more <https://keepachangelog.com/en/1.1.0/>`__

.. note::
    Learn how to apply a label to a PR in the `Github docs <https://docs.github.com/en/issues/using-labels-and-milestones-to-track-work/managing-labels#applying-a-label>`__.

🐞 File an issue
================

We use `GitHub Issues <https://github.com/flyteorg/flyte/issues>`__ for issue tracking. The following issue types are available for filing an issue:

* `Plugin Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=untriaged%2Cplugins&template=backend-plugin-request.md&title=%5BPlugin%5D>`__
* `Bug Report <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2C+untriaged&template=bug_report.md&title=%5BBUG%5D+>`__
* `Documentation Bug/Update Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2C+untriaged&template=docs_issue.md&title=%5BDocs%5D>`__
* `Core Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged&template=feature_request.md&title=%5BCore+Feature%5D>`__
* `Flytectl Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+flytectl&template=flytectl_issue.md&title=%5BFlytectl+Feature%5D>`__
* `Housekeeping <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=housekeeping&template=housekeeping_template.md&title=%5BHousekeeping%5D+>`__
* `UI Feature Request <https://github.com/flyteorg/flyte/issues/new?assignees=&labels=enhancement%2C+untriaged%2C+ui&template=ui_feature_request.md&title=%5BUI+Feature%5D>`__

If none of the above fit your requirements, file a `blank <https://github.com/flyteorg/flyte/issues/new>`__ issue.
Also, add relevant labels to your issue. For example, if you are filing a Flytekit plugin request, add the ``flytekit`` label.

For feedback at any point in the contribution process, feel free to reach out to us on `Slack <https://flyte-org.slack.com/archives/C04NJPLRWUX>`__.

.. toctree::
    :maxdepth: 1
    :hidden:

    Contributing code <contribute_code>
    Contributing docs <contribute_docs>
    Contributing tutorials or integrations examples <contribute_examples>