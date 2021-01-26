[Back to Cookbook Menu](../..)

# Use Jupyter and papermill to author tasks

Starting [Flytekit 0.10.12](https://github.com/lyft/flytekit/releases/tag/v0.10.12), there is now support  to run Jupyter Notebooks as Flyte Tasks.

Currently we support  `python` and `pyspark` notebooks. 
These notebook tasks support basic primitive inputs injection into the notebook as well as capturing outputs from the notebook. The notebooks are executed using [papermill](https://papermill.readthedocs.io/en/latest/).
Inputs are injected by papermill in the notebooks in a special `injected_parameters` cell and outputs are captured by `flytekit` from a specially tagged `outputs` cell. For `pyspark`,  the notebook should also contain a `conf` cell with the spark config to run with. 


## Notebook as Tasks Examples
For Spark Notebooks, you will need to have Spark plugin enabled and SparkOperator deployed on your K8s cluster. Please refer [Spark](../../../plugins/spark/README.md) for details.

1. [Python Notebook](python-notebook.ipynb) / [Flyte Notebook Workflow](python_notebook_wf.py)
2. [Spark Notebook](spark-notebook.ipynb) / [Flyte Spark Notebook Workflow](spark_notebook_wf.py)


