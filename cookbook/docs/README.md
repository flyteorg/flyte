# FlyteSnacks Docs
Welcome to Flytesnacks documentation. To generate the documentation, first
install dev_requirements.txt, then run

```bash
make html
```

## How do the docs work?
Flytesnacks uses the concept of `Literate Programming<https://en.wikipedia.org/wiki/Literate_programming>`_  to generate the documentation from the examples themselves. To achieve this it uses the excellent `Sphinx-Gallery Plugin<https://sphinx-gallery.github.io/stable/index.html>`_ to render comments and codes inline in the docs.
To make this work, it is essential that the examples are written with comments following the Sphinx-Gallery style of coding. Some important things to note:
 - The example directory should have a README.rst.
 - The example itself should have a header comment, which should have a heading
   as well.
 - Docs interspersed in the example should proceed with `# %%` comment and then
   multiline comments should not have blank spaces between them.
  ```rst
  # %%
  # my very important comment
  #
  # some other stuff
  def foo():
    ...
  ```
 - prompts should use
   ```rst
    .. prompt::bash

       flytectl --version
   ```
