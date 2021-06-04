Papermill
=========

It is possible to run a Jupyter notebook as a Flyte task using `papermill <https://github.com/nteract/papermill>`_.
Papermill executes the notebook as a whole, so before using this plugin, it is essential to construct your notebook as recommended by papermill.

Salient points
---------------

1. This plugin can be used for any task - type.
  - It can be straight-up python code, which can be a tensorflow model, a data transformation, etc., but typically run in a container and written in a ``@task``.
  - It can be a ``dynamic task``.
  - It can be a any other plugin like ``spark``, ``sagemaker``, etc., **but ensure that the plugin is installed**.
2. Flytekit will execute the notebook and capture the output notebook as an *.ipynb* and an HTML rendered notebook as well.
3. Flytekit will pass the inputs into the notebook, as long as the first cell is annotated as ``parameters`` and the inputs are specified.
4. Flytekit will read the outputs from the notebook, as long as annotate the notebook is used with ``outputs`` and the outputs are specified.


Installation
------------

To use the flytekit papermill plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-papermill


