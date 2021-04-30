Executing Jupyter Notebooks as tasks
=========================================

It is possible to run a Jupyter notebook as a Flyte task using `papermill <https://github.com/nteract/papermill>`_.
Papermill executes the notebook as a whole, so before using this plugin, it is essential to construct your notebook as recommended by papermill.

Salient points
---------------

#. This plugin can be used for any task - type.
  - It can be straight-up python code, which can be a tensorflow model, a data transformation, etc - but things that run in a
container and you would typically write in a ``@task``.
  - It can be a ``dynamic task``
  - It can be a any other plugin like ``spark``, ``sagemaker`` etc, **ensure that the plugin is installed as well**
#. Flytekit will execute the notebook and capture the output notebook as an *.ipynb* and an HTML rendered notebook as well
#. Flytekit will pass the inputs into the notebook as long as you have the first cell annotated as ``parameters`` and inputs are specified
#. Flytekit will read the outputs from the notebook, as long as you use annotate the notebook with ``outputs`` and outputs are specified


Installation
------------

To use the flytekit papermill plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-papermill


