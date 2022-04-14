Diabetes Classification
------------------------

The workflow demonstrates how to train an XGBoost model. The workflow is designed for the `Pima Indian Diabetes dataset <https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names>`__.

An example dataset is available `here <https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv>`__.

Why a Workflow?
================
One common question when you read through the example might be - is it really required to split the training of xgboost into multiple steps. The answer is complicated, but let us try and understand what advantages and disadvantages of doing so.

Pros:
^^^^^

- Each task/step is standalone and can be used for various other pipelines
- Each step can be unit tested
- Data splitting, cleaning etc can be done using a more scalable system like Spark
- State is always saved between steps, so it is cheap to recover from failures, especially if caching=True
- Visibility is high

Cons:
^^^^^

- Performance for small datasets is a concern. The reason is, the intermediate data is durably stored and the state recorded. Each step is essnetially a checkpoint

Steps of the Pipeline
======================

1. Gather data and split it into training and validation sets
2. Fit the actual model
3. Run a set of predictions on the validation set. The function is designed to be more generic, it can be used to simply predict given a set of observations (dataset)
4. Calculate the accuracy score for the predictions


Takeaways
===========

- Usage of FlyteSchema Type. Schema type allows passing a type safe vector from one task to task. The vector is also directly loaded into a pandas dataframe. We could use an unstructured Schema (By simply omiting the column types). this will allow any data to be accepted by the train algorithm.
- We pass the file as a CSV input. The file is auto-loaded.


Walkthrough
====================

Run workflows in this directory with the custom-built base image like so:

```shell
pyflyte run --remote diabetes.py:diabetes_xgboost_model --image ghcr.io/flyteorg/flytecookbook:pima_diabetes-latest
```


