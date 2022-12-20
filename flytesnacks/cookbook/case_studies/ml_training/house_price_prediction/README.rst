House Price Regression
----------------------

.. tags:: Data, MachineLearning, DataFrame, Intermediate

House Price Regression refers to the prediction of house prices based on various factors, using the XGBoost Regression model (in our case).
In this example, we will train our data on the XGBoost model to predict house prices in multiple regions.

.. _flyte's-role:

Where Does Flyte Fit In?
==========================
- Orchestrates the machine learning pipeline.
- Helps cache the output state between :py:func:`tasks <flytekit:flytekit.task>`.
- Easier backtracking to the error source.
- Provides a Rich UI to view and manage the pipeline.

House price prediction pipeline for one region doesn't require a :py:func:`~flytekit:flytekit.dynamic` workflow. When multiple regions are involved, to iterate through the regions at run-time and thereby build the DAG, Flyte workflow has to be :py:func:`~flytekit:flytekit.dynamic`.

.. tip::

    Refer to :ref:`sphx_glr_auto_core_control_flow_dynamics.py` section to learn more about dynamic workflows.

Dataset
========
We will create a custom dataset to build our model by referring to the `Sagemaker example <https://github.com/aws/amazon-sagemaker-examples/blob/master/advanced_functionality/multi_model_xgboost_home_value/xgboost_multi_model_endpoint_home_value.ipynb>`__.

The dataset will have the following columns:

- Price
- House Size
- Number of Bedrooms
- Year Built
- Number of Bathrooms
- Number of Garage Spaces
- Lot Size

Takeaways
===========
- An in-depth dive into dynamic workflows
- How the Flyte type-system works

Code Walkthrough
=================
