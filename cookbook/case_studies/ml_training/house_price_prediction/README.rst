House Price Regression
-----------------------

A house price prediction model is used to predict a house’s price, given the required inputs.
In this example we'll train a model to predict the price of a house in multiple regions.

Generally, this could be accomplished using any regression model in machine learning.

Where does Flyte fit in?
========================
- Orchestrates the machine learning pipeline
- Can cache the output state between the steps (tasks as per Flyte)
- Easier backtracking to the error source
- Provides a Rich UI (if the Flyte backend is enabled) to view and manage the pipeline

A typical house price prediction model isn’t dynamic, but a task has to be dynamic when multiple regions are involved. 

To learn more about dynamic workflows, refer to `Write a dynamic task <https://docs.flyte.org/projects/cookbook/en/latest/auto_core_intermediate/dynamics.html>`__.

Dataset
=======
There is no built-in dataset that could be employed to build this model. A dataset has to be created, possibly using this reference model on `Github <https://github.com/awslabs/amazon-sagemaker-examples/blob/master/advanced_functionality/multi_model_xgboost_home_value/xgboost_multi_model_endpoint_home_value.ipynb>`__.

The dataset will have the following columns:
- Price
- House Size
- Number of Bedrooms
- Year Built
- Number of Bathrooms
- Number of Garage Spaces
- Lot Size

Steps to Build the Machine Learning Pipeline
============================================
- Generate dataset and split it into train, validation, and test datasets
- Fit the XGBoost model on to the data
- Generate predictions

Steps to Make the Pipeline Flyte-Compatible
===========================================
- Create two Python files to segregate the house price prediction logic. One consists of the logic per region, and the other is for multiple regions
- Define a couple of helper functions that are to be used while defining Flyte tasks and workflows
- Define three Flyte tasks -- to generate and split the data, fit the model, and generate predictions. If there are multiple regions, the tasks are dynamic
- Define a workflow to call the dynamic tasks in a specified order

Takeaways
=========
- An in-depth dive into dynamic workflows
- How the Flyte type-system works

Code Walkthrough
================
