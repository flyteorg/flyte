Example of a MultiStep Linear Workflow
======================================

Introduction:
-------------
The workflow is a simple multistep xgboost  trainer. It is not required to split the training of xgboost into multiple steps, but there are pros and cons of doing so.

Pros:
 - Each task/step is standalone and can be used for various other pipelines
 - Each step can be unit tested
 - Data splitting, cleaning etc can be done using a more scalable system like Spark
 - State is always saved between steps, so it is cheap to recover from failures, especially if caching=True
 - Visibility is high

Cons:
 - Performance for small datasets is a concern. The reason is, the intermediate data is durably stored and the state recorded. Each step is essnetially a checkpoint

Steps of the Pipeline
----------------------
 - Step1: Gather data and split it into training and validation sets
 - Step2: Fit the actual model
 - Step3: Run a set of predictions on the validation set. The function is designed to be more generic, it can be used to simply predict given a set of observations (dataset)
 - Step4: Calculate the accuracy score for the predictions

The workflow is designed for a dataset as defined in 
https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names

An example dataset is available at
https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv

Important things to note:
-------------------------
- Usage of Schema Type. Schema type allows passing a type safe row vector from one task to task. The row vector is also directly loaded into a pandas dataframe
  We could use an unstructured Schema (By simply omiting the column types). this will allow any data to be accepted by the train algorithm.

- We pass the file as a CSV input. The file is auto-loaded.

