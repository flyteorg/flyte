(docs_contribute)=

# Contributing tutorials and integration examples

```{eval-rst}
.. tags:: Contribute, Basic
```

[Tutorials](https://docs.flyte.org/flytesnacks/tutorials/) and [integration examples](https://docs.flyte.org/flytesnacks/integrations/) help the community learn about the rich set of features that Flyte offers, and we are constantly improving them with your help!

Whether you're a novice or experienced software engineer, data scientist, or machine learning practitioner, we welcome your contributions!

* {ref}`Tutorials <tutorials>` use multiple Flyte features to solve real-world problems. Tutorials can be complex, and might require extra setup or only be executable on larger clusters.
* {ref}`Integration examples <integrations>` showcase how to use the Flyte plugins that integrate with the broader data and machine learning ecosystem.

The first step to contributing a tutorial or integration example is to open up a
[documentation issue](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2Cuntriaged&template=docs_issue.yaml&title=%5BDocs%5D+) to articulate the kind of example you want to write. The Flyte maintainers will help you figure out where your example would fit best.

## Creating a tutorial or integration example

:::{admonition} Prerequisites
Follow the {ref}`env_setup` guide to get your development environment ready.
:::

The tutorials and integration examples live in the `examples` directory in flytesnacks. Each
subdirectory contains a self-contained project:

```{code-block} bash
examples
...
├── bigquery_plugin
├── blast
├── chatgpt_agent
...
```

Additionally, the index page for the `tutorials` and `integrations` sections lives in the `/docs` directory in the flytesnacks repository.

### Adding a script to an existing tutorial or integration example

If you add an example script to an existing tutorial or integration, add the example to the `auto-examples-toc` section of the `README.md` file in the tutorial or integration subdirectory:

````{code-block}
```{auto-examples-toc}
...
my_new_example
```
````

### Creating a new tutorial or integration example

If you're creating a new tutorial or integration example, you'll need to set up a new example project.

Clone the `flytesnacks` repository:

```{prompt} bash
git clone git@github.com:flyteorg/flytesnacks.git
```

In the `flytesnacks` root directory, create an example project:

```{prompt} bash
cd flytesnacks
./scripts/create-example-project new_example_project
```

This will create a new directory under `examples`:

```{code-block} bash
examples/new_example_project
├── Dockerfile
├── README.md
├── new_example_project
│   ├── __init__.py
│   └── example.py
├── requirements.in
└── requirements.txt
```

### Creating Python examples

#### Tutorial or integration examples

If you are writing a tutorial or integration example, write your example Python script in [percent format](https://jupytext.readthedocs.io/en/latest/formats.html#the-percent-format), which allows you to interleave Python code and Markdown in the same file. Each code cell should be delimited by `# %%`, and each Markdown cell should be
delimited with `# %% [markdown]`.

```{code-block} python
# %%
print("Hello World!")

# %% [markdown]
# This is a Markdown cell

# %%
print("This is another code cell")
```

Markdown cells have access to Sphinx directives through the
[Myst Markdown](https://myst-parser.readthedocs.io/en/latest/) format,
which is a flavor of markdown that makes it easier to write documentation while
giving you the utilities of sphinx. `flytesnacks` uses the
[myst-nb](https://myst-nb.readthedocs.io/en/latest/) and
[jupytext](https://github.com/mwouts/jupytext) packages to interpret the
python files as rst-compatible files.

### Writing examples: explain what the code does

Following the [literate programming](https://en.wikipedia.org/wiki/Literate_programming) paradigm, make sure to
interleave explanations in the `*.py` files containing the code example.

:::{admonition} A Simple Example
:class: tip

Here's a code snippet that defines a function that takes two positional arguments and one keyword argument:

```python
def function(x, y, z=3):
    return x + y * z
```

As you can see, `function` adds the two first arguments and multiplies the sum with the third keyword
argument. Can you think of a better name for this `function`?
:::

Explanations don't have to be this detailed for such a simple example, but you can imagine how this makes for a better
reading experience for more complicated examples.

### Creating examples in other formats

Writing examples in `.py` files is preferred since they are easily tested and
packaged, but `flytesnacks` also supports examples written in `.ipynb` and
`.md` files in myst markdown format. This is useful in the following cases:

- `.ipynb`: When a `.py` example needs a companion jupyter notebook as a task, e.g.
  to illustrate the use of {py:class}`~flytekitplugins.papermill.NotebookTask`s,
  or when an example is intended to be run from a notebook.
- `.md`: When a piece of documentation doesn't require testable or packaged
  flyte tasks/workflows, an example page can be written as a myst markdown file.

**Note:** If you want to add Markdown files to a user guide example project, add them to the [flyte repository](https://github.com/flyteorg/flyte/tree/master/docs/user_guide) instead.

## Writing a README

The `README.md` file needs to capture the _what_, _why_, and _how_ of the example.

- What is the integration about? Its features, etc.
- Why do we need this integration? How is it going to benefit the Flyte users?
- Showcase the uniqueness of the integration
- How to install the plugin?

Finally, **for tutorials and integrations only**, write a `auto-examples-toc` directive at the bottom of the file:

````{code-block}
```{auto-examples-toc}
example_01
example_02
example_03
```
````

Where `example_01`, `example_02`, and `example_03` are the python module
names of the examples under the `new_example_project` directory. These can also
be the names of the `.ipynb` or `.md` files (but without the file extension).

:::{tip}
Refer to any subdirectory in the `examples` directory
:::

## Test your code

If the example code can be run locally, just use `python <my_file>.py` to run it.

### Testing on a cluster

Install {doc}`flytectl <../api/flytectl/index>`, the commandline interface for flyte.

:::{note}
Learn more about installation and configuration of Flytectl [here](https://docs.flyte.org/en/latest/flytectl/index.html).
:::

Start a Flyte demo cluster with:

```
flytectl demo start
```

### Testing the `basics` project examples on a local demo cluster

In this example, we'll build the `basics` project:

```{prompt} bash
# from flytesnacks root directory
cd examples/basics
```

Build the container:

```{prompt} bash
docker build . --tag "basics:v1" -f Dockerfile
```

Package the examples by running:

```{prompt} bash
pyflyte --pkgs basics package --image basics:v1 -f
```

Register the examples by running

```{prompt} bash
flytectl register files \
   -p flytesnacks \
   -d development \
   --archive flyte-package.tgz \
   --version v1
```

Visit `https://localhost:30081/console` to view the Flyte console, which consists
of the examples present in the `flytesnacks/core` directory.

### Updating dependencies

:::{admonition} Prerequisites
Install [pip-tools](https://pypi.org/project/pip-tools/) in your development
environment with:

```{prompt} bash
pip install pip-tools
```

:::

If you've updated the dependencies of the project, update the `requirements.txt`
file by running:

```{prompt} bash
pip-compile requirements.in --upgrade --verbose --resolver=backtracking
```

### Rebuild the image

If you've updated the source code or dependencies of the project, and rebuild
the image with:

```{prompt} bash
docker build . --tag "basics:v2" -f core/Dockerfile
pyflyte --pkgs basics package --image basics:v2 -f
flytectl register files \
    -p flytesnacks \
    -d development \
    --archive flyte-package.tgz \
    --version v2
```

Refer to {ref}`this guide <getting_started_package_register>`
if the code in itself is updated and requirements.txt is the same.

## Pre-commit hooks

We use [pre-commit](https://pre-commit.com/) to automate linting and code formatting on every commit.
Configured hooks include [ruff](https://github.com/astral-sh/ruff) to ensure newlines are added to the end of files, and there is proper spacing in files.

We run all those hooks in CI, but if you want to run them locally on every commit, run `pre-commit install` after
installing the dev environment requirements. In case you want to disable `pre-commit` hooks locally, run
`pre-commit uninstall`. More info [here](https://pre-commit.com/).

### Formatting

We use [ruff](https://github.com/astral-sh/ruff) to autoformat code. They
are configured as git hooks in `pre-commit`. Run `make fmt` to format your code.

### Spell-checking

We use [codespell](https://github.com/codespell-project/codespell) to catch common misspellings. Run
`make spellcheck` to spell-check the changes.

## Update Documentation Pages

The `docs/conf.py` contains the sphinx configuration for building the
`flytesnacks` documentation.

At build-time, the `flytesnacks` sphinx build system will convert all of the
projects in the `examples` directory into `docs/auto_examples`, and will be
available in the documentation.

::::{important}

The docs build system will convert the `README.md` files in each example
project into a `index.md` file, so you can reference the root page of each
example project, e.g., in myst markdown format, you can write a table-of-content
directive like so:

:::{code-block}

```{toc}
auto_examples/basics/index
```

:::

::::

If you've created a new example project, you'll need to add the `index` page
in the table of contents in `docs/index.md` to make sure the project
shows up in the documentation. Additonally, you'll need to update the appropriate
`list-table` directive in `docs/userguide.md`, `docs/tutorials.md`, or
`docs/integrations.md` so that it shows up in the respective section of the
documentation.

## Build the documentation locally

Verify that the code and documentation look as expected:

- Learn about the documentation tools [here](https://docs.flyte.org/en/latest/community/contribute.html#documentation)
- Install the requirements by running `pip install -r docs-requirements.txt`.
- Run `make -C docs html`

  ```{tip}
  To run a fresh build, run `make -C docs clean html`.
  ```

- Open the HTML pages present in the `docs/_build` directory in the browser with
  `open docs/_build/index.html`

## Create a pull request

Create the pull request, then ensure that the docs are rendered correctly by clicking on the documentation check.

```{image} https://raw.githubusercontent.com/flyteorg/static-resources/main/common/test_docs_link.png
:alt: Docs link in a PR
```

You can refer to [this PR](https://github.com/flyteorg/flytesnacks/pull/332) for the exact changes required.
