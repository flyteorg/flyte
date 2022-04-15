Diabetes Classification
------------------------

The workflow demonstrates how to train an XGBoost model. The workflow is designed for the `Pima Indian Diabetes dataset <https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names>`__.

An example dataset is available `here <https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv>`__.

Why a Workflow?
================
One common question when you read through the example would be whether it is really required to split the training of XGBoost into multiple steps. The answer is complicated, but let us try and understand the pros and cons of doing so.

Pros:
^^^^^

- Each task/step is standalone and can be used for other pipelines
- Each step can be unit tested
- Data splitting, cleaning and processing can be done using a more scalable system like Spark
- State is always saved between steps, so it is cheap to recover from failures, especially if ``caching=True``
- High visibility

Cons:
^^^^^

- Performance for small datasets is a concern because the intermediate data is durably stored and the state is recorded, and each step is essentially a checkpoint

Steps of the Pipeline
======================

1. Gather data and split it into training and validation sets
2. Fit the actual model
3. Run a set of predictions on the validation set. The function is designed to be more generic, it can be used to simply predict given a set of observations (dataset)
4. Calculate the accuracy score for the predictions


Takeaways
===========

- Usage of FlyteSchema Type. Schema type allows passing a type safe vector from one task to task. The vector is directly loaded into a pandas dataframe. We could use an unstructured Schema (By simply omitting the column types). This will allow any data to be accepted by the training algorithm.
- We pass the file (that is auto-loaded) as a CSV input.


Walkthrough
====================

Run workflows in this directory with the custom-built base image:

```shell
pyflyte run --remote diabetes.py:diabetes_xgboost_model --image ghcr.io/flyteorg/flytecookbook:pima_diabetes-latest
```


