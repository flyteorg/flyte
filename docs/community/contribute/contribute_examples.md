(contribute_examples)=
# Contribute tutorials or integrations examples

The first step to contributing a tutorial or integration example is to open a
[documentation issue](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=documentation%2Cuntriaged&template=docs_issue.yaml&title=%5BDocs%5D+) to describe the example you want to write. The Flyte maintainers will help you figure out where your tutorial or integration example would best fit.

:::{admonition} Prerequisites
Follow the {ref}`env_setup` guide to get your development environment ready.
:::

The tutorials and integrations examples live in the `examples` directory of the [flytesnacks repo](https://github.com/flyteorg/flytesnacks), where each subdirectory contains a self-contained example project that covers a particular tutorial or integration. There are also subdirectories that contain the code included in the user guide via the `rli` (remoteliteralinclude) directive.

```{code-block} bash
examples
├── README.md
├── airflow_plugin
├── athena_plugin
├── aws_batch_plugin
├── basics
├── bigquery_agent
...
```

## Adding an example script to an existing project

If you're adding a new example to an existing project, you can simply create a
new `.py` file in the appropriate directory. For example, if you want to add a new
example in the `examples/exploratory_data_analysis` project, simply do:

```{prompt} bash
touch examples/exploratory_data_analysis/my_new_example.py
```

If you are creating a new integration or tutorial example, add the example to the `README.md` file of the
example project as an entry in the `auto-examples-toc` directive:

````{code-block}
```{auto-examples-toc}
...
my_new_example
```
````

If you are creating a new user guide example, you can reference the code in the user guide documentation using the `rli` (remoteliteralinclude) directive. You do not need to add the new example to the `README.md` file of the example project.

## Creating a new example project

````{important}
If you're creating a new tutorial or integration example
that doesn't fit into any of the existing subdirectories, you'll need to set up a
new example project.

In the `flytesnacks` root directory, run the following command to create an example project:

```{prompt} bash
./scripts/create-example-project.sh new_example_project
```

This will create a new directory under `examples`:

```{code-block} bash
examples/new_example_project
├── Dockerfile
├── README.md
├── new_example_project
│   ├── __init__.py
│   └── example.py
└── requirements.in
```

````

## Creating Python examples

### Tutorial or integration examples

If you are writing a tutorial or integration example, write your example Python script in [percent format](https://jupytext.readthedocs.io/en/latest/formats.html#the-percent-format),
which allows you to interleave Python code and Markdown in the same file. Each
code cell should be delimited by `# %%`, and each Markdown cell should be
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
[MyST Markdown](https://myst-parser.readthedocs.io/en/latest/) format,
which is a flavor of Markdown that makes it easier to write documentation while
giving you the utilities of Sphinx. `flytesnacks` uses the
[myst-nb](https://myst-nb.readthedocs.io/en/latest/) and
[jupytext](https://github.com/mwouts/jupytext) packages to interpret the
Python files as rst-compatible files.

## Writing examples: explain what the code does

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

Explanations don't have to be this detailed for such a simple example, but you can imagine how this makes for a better reading experience for more complicated examples.

## Creating examples in other formats

Writing examples in `.py` files is preferred since they are easily tested and
packaged, but `flytesnacks` also supports examples written in `.ipynb` and
`.md` files in MyST Markdown format. This is useful in the following cases:

- `.ipynb`: When a `.py` example needs a companion Jupyter notebook as a task, e.g.
  to illustrate the use of {py:class}`~flytekitplugins.papermill.NotebookTask`s,
  or when an example is intended to be run from a notebook.
- `.md`: When a piece of documentation doesn't require testable or packaged
  flyte tasks/workflows, an example page can be written as a MyST Markdown file.

```{note}
If you want to add Markdown files to a user guide example project, add them to the [flyte repository](https://github.com/flyteorg/flyte/tree/master/docs/user_guide) instead.
```

## Writing a README

The `README.md` file needs to capture the _what_, _why_, and _how_ of the example.

- What is the tutorial or integration about? Its features, etc.
- Why do we need this tutorial or integration? How is it going to benefit Flyte users?
- Showcase the uniqueness of the tutorial or integration
- Integration plugin installation steps

Finally, **for tutorials and integrations only**, write an `auto-examples-toc` directive at the bottom of the file:

````{code-block}
```{auto-examples-toc}
example_01
example_02
example_03
```
````

Where `example_01`, `example_02`, and `example_03` are the Python module
names of the examples under the `new_example_project` directory. These can also
be the names of the `.ipynb` or `.md` files (but without the file extension).

:::{tip}
Refer to any subdirectory in the `examples` directory
:::

## Test your code

If the example code can be run locally, you can use `pyflyte run` to test it:

```{prompt} bash
pyflyte run <python_file>.py <task_or_workflow_name> --<arg1> <arg1_value> --<arg2> <arg2_value> ...
```

### Testing on a cluster

Install {doc}`flytectl <flytectl:index>`, the command line interface for flyte.

:::{note}
Learn more about installation and configuration of Flytectl [here](https://docs.flyte.org/en/latest/flytectl/docs_index.html).
:::

Start a Flyte demo cluster with:

```
flytectl demo start
```

#### Testing the `basics` project examples on a local demo cluster

In this example, we'll build the `basics` project.

Change to the basics directory:

```{prompt} bash
# from flytesnacks root directory
cd examples/basics
```

Build and push the container to the local Docker registry provided by the demo cluster:

```{prompt} bash
docker build . --tag "localhost:30000/basics:v1" -f Dockerfile --push
```

Run a workflow in the local demo cluster by specifying the `--image` flag
and passing the `--remote` flag:

```{prompt} bash
pyflyte run --remote \
  --image localhost:30000/basics:v1 \
  basics/hello_world.py hello_world_wf
```

Visit `https://localhost:30081/console` to view the Flyte console, which consists
of the examples present in the `flytesnacks/core` directory.

## Updating dependencies

:::{admonition} Prerequisites
Install [pip-tools](https://pypi.org/project/pip-tools/) in your development
environment with:

```{prompt} bash
pip install pip-tools
```

:::

If the project uses pinned dependencies in a `requirements.in` file, run the `pip-compile` command to create a new pinned `requirements.txt` file:

```{prompt} bash
pip-compile requirements.in --upgrade --verbose --resolver=backtracking
```

## Rebuild the image

If you've updated the source code or dependencies of the project, rebuild
the image with:

```{prompt} bash
docker build . --tag "localhost:30000/basics:v2" -f Dockerfile --push
```

Next, run the workflow again with the new image:

```{prompt} bash
pyflyte run --remote \
  --image localhost:30000/basics:v2 \
  basics/hello_world.py hello_world_wf
```

Refer to {ref}`this guide <getting_started_package_register>`
if the code itself is updated and requirements.txt is the same.

## Pre-commit hooks

We use [pre-commit](https://pre-commit.com/) to automate linting and code formatting on every commit.
Configured hooks include [ruff](https://github.com/astral-sh/ruff) to ensure newlines are added to the end of files, and there is proper spacing in files.

We run all those hooks in CI, but if you want to run them locally on every commit, run `pre-commit install` after
installing the dev environment requirements. In case you want to disable `pre-commit` hooks locally, run
`pre-commit uninstall`. More info [here](https://pre-commit.com/).

### Formatting

We use [ruff](https://github.com/astral-sh/ruff) to autoformat code. They
are configured as git hooks in `pre-commit`. Run `make fmt` to format your code.

#### Spell-checking

We use [codespell](https://github.com/codespell-project/codespell) to catch common misspellings. Run
`make spellcheck` to spell-check the changes.

## Update tutorials and integrations examples documentation pages

The `docs/conf.py` file in the flytesnacks repository contains the Sphinx configuration for building the `flytesnacks` documentation.

At build time, the `flytesnacks` Sphinx build system will convert the tutorials and integrations examples in the `examples` directory into `docs/auto_examples` and make them available in the documentation.

::::{important}

The docs build system will convert the `README.md` files in each tutorial and example integration project into an `index.md` file, allowing you to reference the root page of each
example project in MyST Markdown format.

::::

If you've created a new tutorial or integration example project, you'll need to add its `index` page
to the table of contents in `docs/tutorials/index.md` or `docs/integrations/index.md` and update the appropriate `list-table` directive in the file to make sure the example shows up in the documentation:

:::{code-block}

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Airflow agent </auto_examples/airflow_agent/index>`
  - Run Airflow jobs in your workflows with the Airflow agent.
  ...
```

...

```{toc}
/auto_examples/bigquery_agent/index
```

:::

### Build the documentation locally

Verify that the code and documentation look as expected:

- Learn about the documentation tools [here](https://docs.flyte.org/en/latest/community/contribute.html#documentation)
- Install the requirements by running `pip install -r docs-requirements.txt`.
- Run `make -C docs html`

  ```{tip}
  To run a fresh build, run `make -C docs clean html`.
  ```

- Open the HTML pages present in the `docs/_build` directory in the browser with
  `open docs/_build/index.html`

### Create a pull request

Create the pull request in the flytesnacks repository, then ensure that the docs are rendered correctly by clicking on the documentation check:

```{image} https://github.com/user-attachments/assets/581330cc-c9ab-418d-a9ae-bb3c346603ba
:alt: Docs link in a PR
```
