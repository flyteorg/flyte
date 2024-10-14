(contribute_docs)=

# Contributing documentation

```{eval-rst}
.. tags:: Contribute, Basic
```

Whether you're a novice or experienced software engineer, data scientist, or machine learning
practitioner, we welcome your contributions to the Flyte documentation!

The Flyte documentation comprises the following types:

* **{ref}`User guide <userguide>` documentation:** Conceptual and procedural documentation about using Flyte features to accomplish tasks.
* **{ref}`Tutorials <tutorials>`:** Longer, more advanced guides that use multiple Flyte features to solve real-world problems. Some tutorials may require extra setup, while others can only run on larger clusters.
* **{ref}`Integrations examples <integrations>`:** These examples showcase how to use the Flyte plugins that integrate with the broader data and machine learning ecosystem.
* **{ref}`Deployment documentation <deployment>`:** Guidance on deploying and configuring the Flyte backend.
* **{doc}`API documentation <../../api/index>`:** flytekit, flytectl, and flyteidl documentation.

For minor edits that don't require a local setup, you can edit the page in GitHub page to propose improvements.

## Contributing to user guide and deployment documentation

To update user guide or deployment documentation, edit the corresponding files in the [flyte repository](https://github.com/flyteorg/flyte/tree/master/docs/user_guide).

### Code in user guide documentation

If you want to include tested, runnable example code in user guide documentation, you will need to add your code to the examples directory of the [flytesnacks repository](https://github.com/flyteorg/flytesnacks). Write your code in regular Python, with regular comments. You can use the `rli` ([remoteliteralinclude](https://github.com/wpilibsuite/sphinxext-remoteliteralinclude/blob/main/README.md)) directive to include snippets of code from your example Python file.

```{important}
Unlike the comments in tutorials and integrations examples Python files, the comments in user guide Python files in flytesnacks **will not** be transformed into Markdown and processed into HTML for the docs site. All prose documentation for the user guide is contained in the Flyte repo, in the `docs/user_guide` directory.
```

## Contributing to API documentation

* **flytekit:** See the [flytekit repository](https://github.com/flyteorg/flytekit). Documentation consists of content in submodule `__init__.py` files, `rst` files, and docstrings in classes and methods.
* **flytectl:** See the [flyte repository](https://github.com/flyteorg/flyte/tree/master/flytectl). Documentation consists of `rst` files in the `flytectl/docs` directory and comments in code files.
* **flyteidl:** See the [flyte repository](https://github.com/flyteorg/flyte/tree/master/flyteidl). `protoc-gen-doc` is used to generate this documentation from `.proto` files.

## Contributing tutorials and integrations examples

Follow the steps in {ref}`Contributing tutorials or integrations examples <contribute_examples>`.

## Docs formatting

Our documentation contains a mix of Markdown and reStructuredText.

### Markdown

User guide documentation and tutorials and integrations examples uses MyST Markdown. For more information, see the [MyST syntax documentation](https://mystmd.org/guide/syntax-overview). 

### reStructuredText

Deployment and API docs mostly use reStructured Text. For more information, see the [reStructuredText documentation](https://www.sphinx-doc.org/en/master/usage/restructuredtext/basics.html).

### Python objects

You can cross-reference multiple Python modules, functions, classes, methods, and global data in documentations. For more information, see the [Sphinx documentation](https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects).

