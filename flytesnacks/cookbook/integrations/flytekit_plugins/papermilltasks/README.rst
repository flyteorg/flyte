Papermill
=========

.. tags:: Integration, Jupyter, Intermediate

It is possible to run a Jupyter notebook as a Flyte task using `papermill <https://github.com/nteract/papermill>`_.
Papermill executes the notebook as a whole, so before using this plugin, it is essential to construct your notebook as
recommended by papermill. When using this plugin, there are a few important things to keep in mind:

1. This plugin can be used for any task - type.
    - It can be python code, which can be a tensorflow model, a data transformation, etc - but things that run in a container
      and you would typically write in a ``@task``.
    - It can be a :py:func:`~flytekit.dynamic` workflow.
    - It can be a any other plugin like ``spark``, ``sagemaker`` etc, **ensure that the plugin is installed as well**
2. Flytekit will execute the notebook and capture the output notebook as an *.ipynb* file and an HTML rendered notebook as well
3. Flytekit will pass the inputs into the notebook as long as you have the first cell annotated as ``parameters`` and inputs are specified
4. Flytekit will read the outputs from the notebook, as long as you use annotate the notebook with ``outputs`` and outputs are specified


Installation
------------

To use the flytekit papermill plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-papermill
