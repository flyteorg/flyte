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

For minor edits that don't require a local setup, you can edit the page in GitHub page to propose improvements.}

In the ``flyteorg/flyte`` root directory you can run ``make dev-docs`` to build the documentation locally. The generated documentation will be in the ``docs/_build/html`` directory.

**Setup process**

1. First you need to make sure you can run linux/amd64 container
2. From the root of your fork, run the following commands to build the documentation and serve it locally

.. prompt:: bash $

 make dev-docs
 python -m http.server --directory docs/_build/html

3. Go to http://localhost:8000 to see the documentation.

**Supported environment variables of** ``make dev-docs``

* ``DEV_DOCS_WATCH``: If set, the docs will be built and served using `sphinx-autobuild <https://github.com/sphinx-doc/sphinx-autobuild>`__ for live updates.
* ``FLYTEKIT_LOCAL_PATH``: If set, the local path to flytekit will be used instead of the source code from the ``flyteorg/flytekit repo``.
* ``FLYTECTL_LOCAL_PATH``: If set, the local path to flytectl will be used instead of the source code from the ``flyteorg/flytectl repo``.
* ``FLYTESNACKS_LOCAL_PATH``: If set, the local path to flytesnacks will be used instead of the source code from the ``flyteorg/flytesnacks`` repo.

For example, to use the local flytekit source code instead of the source code from the ``flyteorg/flytekit`` repo, run ``export FLYTEKIT_LOCAL_PATH=/path/to/flytekit`` before running ``make dev-docs``.

**Alternative conda setup steps**

* Install ``conda``.
    *  We recommend Miniconda installed with an `official installer <https://docs.conda.io/projects/miniconda/en/latest/index.html#latest-miniconda-installer-links>`__.
* Install `conda-lock <https://github.com/conda/conda-lock>`__.
* In the ``flyteorg/flyte`` root directory run:
    * ``conda-lock install --name monodocs-env monodocs-environment.lock.yaml``
    * ``conda activate monodocs-env``
    * ``pip install ./flyteidl``

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

### Quickstart

Flyte Documentation is primarily maintained in two locations: [flyte](https://github.com/flyteorg/flyte) and [flytesnacks](https://github.com/flyteorg/flytesnacks).

#### Tips
The following are some tips to include various content:
* **Images**
	Flyte maintain all static resources in [static-resources-repo](https://github.com/flyteorg/static-resources).
	You should upload your images to this repo and open the PR, and then refer to the image in the documentation.
	Notice that the image URL should be in the format `https://raw.githubusercontent.com/flyteorg/static-resources/<git sha of your commit in PR>/<your image path>`.
* **Source code references (Link format)** <br>
	`.rst` example:
	```{code-block}
	.. raw:: html

		a href="https://github.com/flyteorg/<source repo name>/blob/<git sha>/<target file path>#L<from line>-L<to line>">View source code on GitHub</a>
	```
	
	`.md` example:
	```{code-block}
 	[View source code on GitHub]("https://github.com/flyteorg/<source repo name>/blob/<git sha>/<target file path>#L<from line>-L<to line>")
	```
* **Source code references (Embedded format)** <br>
	`.rst` example:
	```{code-block}
  .. literalinclude:: /examples/<target file path>
		:lines: <from line>-<to line>
	```

	`.md` example:
	````{code-block}
	```{literalinclude} /examples/<target file path>
		:lines: <from line>-<to line>
	```
	````

This way, the nested code block is properly displayed without breaking the Markdown structure.

#### Open a pull request
[This is an example PR](https://github.com/flyteorg/flyte/pull/5844)

Each time you update your PR, it triggers the CI build, so there’s no need to build the docs locally. Flyte uses the CI process `"docs/readthedocs.org:flyte"`, which builds the documentation after each PR.
Be sure to include the following CI-build preview link in your PR description so reviewers can easily preview the changes:
```{code-block}
https://flyte--<PR number>.org.readthedocs.build/en/<PR number>/<relative path>.html
```
The relative path is based on the `docs` directory.
For example, if the full path is `flyte/docs/user_guide/advanced_composition/chaining_flyte_entities.md`, then the relative path would be `user_guide/advanced_composition/chaining_flyte_entities` + `.html`.

#### Important note
In the `flytesnacks` repository, most Python comments using `# xxxx` are not imported into the documentation. 
You may notice some overlap between `flytesnacks` and `flyte` docs, but what is displayed primarily comes from the`flyte` repository.

Otherwise, take care of the following points:
````{important}
* Make sure `:lines:` are aligned correctly.
* Use gitsha to specify the example code instead of using master branch or relative path, as this ensures 100% accuracy.
* Build the documentation by submitting a PR instead of building it locally. 
* For `flytesnacks`, run `make fmt` before submitting the PR.
* Before uploading commits, use `git commit -s` to sign off. This step is often forgotten during the first submission.
* Run `codespell` on the modified files to check for any spelling mistakes before pushing.
* When using reference code or images, use gitsha along with GitHub raw content links.
````
