EDA, Feature Engineering, and Modeling With Papermill
=====================================================

.. tags:: Data, Jupyter, Intermediate

Exploratory Data Analysis (EDA) refers to the critical process of performing initial investigations on data to discover patterns,
spot anomalies, test hypotheses and check assumptions with the help of summary statistics and graphical representations.

EDA cannot be solely implemented within Flyte as it requires visual analysis of the data.
In such scenarios, we are inclined towards using a Jupyter notebook as it helps visualize and feature engineer the data.

**Now the question is, how do we leverage the power of Jupyter Notebook within Flyte to perform EDA on the data?**

Papermill
---------

`Papermill <https://papermill.readthedocs.io/en/latest/>`__ is a tool for parameterizing and executing Jupyter Notebooks.
Papermill lets you:

- parameterize notebooks
- execute notebooks

We have a pre-packaged version of Papermill with Flyte that lets you leverage the power of Jupyter Notebook within Flyte pipelines.

To install the plugin, run the following command:

.. prompt:: bash $

    pip install flytekitplugins-papermill

Examples
--------

There are three code examples that you can refer to in this tutorial:

- Run the whole pipeline (EDA + Feature Engineering + Modeling) in one notebook
- Run EDA and feature engineering in one notebook, fetch the result (EDA'ed and feature engineered-dataset),
  and model the data as a Flyte task by sending the dataset as an argument
- Run EDA and feature engineering in one notebook, fetch the result (EDA'ed and feature engineered-dataset),
  and model the data in another notebook by sending the dataset as an argument

Notebook Etiquette
^^^^^^^^^^^^^^^^^^

- If you want to send inputs and receive outputs, your Jupyter notebook has to have ``parameters`` and ``outputs`` tags, respectively.
  To set up tags in a notebook, follow this `guide <https://jupyterbook.org/content/metadata.html#adding-tags-using-notebook-interfaces>`__.
- ``parameters`` cell must only have the input variables.
- ``outputs`` cell looks like the following:

  .. code-block:: python

    from flytekitplugins.papermill import record_outputs
    record_outputs(variable_name=variable_name)

  Of course, you can have any number of variables!
- The ``inputs`` and ``outputs`` variable names in the ``NotebookTask`` must match the variable names in the notebook.

.. note::
  You will see three outputs on running the Python code files, although a single output is returned.
  One output is the executed notebook, and the other is the rendered HTML of the notebook.
