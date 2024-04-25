# Model training

Understand how machine learning models can be trained from within Flyte, with an added advantage of orchestration benefits.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Diabetes Classification <pima_diabetes/index>`
  - Train an XGBoost model on the Pima Indians Diabetes Dataset.
* - {doc}`House Price Regression <house_price_prediction/index>`
  - Use dynamic workflows to train a multiregion house price prediction model using XGBoost.
* - {doc}`MNIST Classification <mnist_classifier/index>`
  - Train a neural network on MNIST with PyTorch and W&B
* - {doc}`NLP Processing with Gensim <nlp_processing/index>`
  - Word embedding and topic modelling on lee background corpus with Gensim
* - {doc}`Forecast Sales Using Rossmann Store Sales <forecasting_sales/index>`
  - Forecast sales data with data-parallel distributed training using Horovod on Spark.
```

```{toctree}
:maxdepth: 1
:hidden:

pima_diabetes/index
house_price_prediction/index
mnist_classifier/index
nlp_processing/index
forecasting_sales/index
```