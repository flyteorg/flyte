# Documenting workflows

```{eval-rst}
.. tags:: Basic
```

Well-documented code significantly improves code readability.
Flyte enables the use of docstrings to document your code.
Docstrings are stored in [FlyteAdmin](https://docs.flyte.org/en/latest/concepts/admin.html)
and displayed on the UI.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the relevant libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/documenting_workflows.py
:caption: basics/documenting_workflows.py
:lines: 1-3
```

We import the `slope` and `intercept` tasks from the `workflow.py` file.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/documenting_workflows.py
:caption: basics/documenting_workflows.py
:lines: 6
```

## Sphinx-style docstring

An example to demonstrate Sphinx-style docstring.

The initial section of the docstring provides a concise overview of the workflow.
The subsequent section provides a comprehensive explanation.
The last part of the docstring outlines the parameters and return type.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/documenting_workflows.py
:caption: basics/documenting_workflows.py
:pyobject: sphinx_docstring_wf
```

## NumPy-style docstring

An example to demonstrate NumPy-style docstring.

The first part of the docstring provides a concise overview of the workflow.
The next section offers a comprehensive description.
The third section of the docstring details all parameters along with their respective data types.
The final section of the docstring explains the return type and its associated data type.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/documenting_workflows.py
:caption: basics/documenting_workflows.py
:pyobject: numpy_docstring_wf
```

## Google-style docstring

An example to demonstrate Google-style docstring.

The initial section of the docstring offers a succinct one-liner summary of the workflow.
The subsequent section of the docstring provides an extensive explanation.
The third segment of the docstring outlines the parameters and return type,
including their respective data types.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/documenting_workflows.py
:caption: basics/documenting_workflows.py
:pyobject: google_docstring_wf
```

Here are two screenshots showcasing how the description appears on the UI:
1. On the workflow page, you'll find the short description:
:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/document_wf_short.png
:alt: Short description
:class: with-shadow
:::

2. If you click into the workflow, you'll see the long description in the basic information section:
:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/document_wf_long.png
:alt: Long description
:class: with-shadow
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics
