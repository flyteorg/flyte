---
celltoolbar: Tags
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
language_info:
  codemirror_mode:
    name: ipython
    version: 3
  file_extension: .py
  mimetype: text/x-python
  name: python
  nbconvert_exporter: python
  pygments_lexer: ipython3
  version: 3.8.6
---

# Simple Papermill Notebook

This notebook is used in the previous example as a `NotebookTask` to demonstrate
how to use the papermill plugin.

The cell below has the `parameters` tag, defining the inputs to the notebook.

```{code-cell} ipython3
:tags: [parameters]

v = 3.14
```

Then, we do some computation:

```{code-cell} ipython3
square = v*v
print(square)
```

Finally, we use the `flytekitplugins.papermill` package to record the outputs
so that flyte understands which state to serialize and pass into a downstream
task.

```{code-cell} ipython3
:tags: [outputs]

from flytekitplugins.papermill  import record_outputs

record_outputs(square=square)
```
