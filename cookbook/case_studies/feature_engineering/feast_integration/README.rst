Feast Integration
-----------------

**Feature Engineering** off-late has become one of the most prominent topics in Machine Learning.
It is the process of transforming raw data into features that better represent the underlying problem to the predictive models, resulting in improved model accuracy on unseen data.

** `Feast<https://feast.dev/>`_ is an operational data system for managing and serving machine learning features to models in production.**

Flyte provides a way to train models and perform feature engineering as a single pipeline.
But, it provides no way to serve these features to production when the model matures and is ready to be served in production.

Flyte adds the capability of engineering the features and Feast provides the feature registry and online serving system. One thing that Flyte makes possible is incremental development of features and only turning on the sync to online stores when you are confident about the features

In this tutorial, we'll walk through how Feast can be used to store and retrieve features to train and test the model curated using the Flyte pipeline.

Dataset
=======
We'll be using the horse colic dataset wherein we'll determine if the lesion of the horse is surgical or not. This is a modified version of the original dataset.

The dataset will have the following columns:

.. list-table:: Horse Colic Features
    :widths: 25 25 25 25 25

    * - surgery
      - Age
      - Hospital Number
      - rectal temperature
      - pulse
    * - respiratory rate
      - temperature of extremities
      - peripheral pulse
      - mucous membranes
      - capillary refill time
    * - pain
      - peristalsis
      - abdominal distension
      - nasogastric tube
      - nasogastric reflux
    * - nasogastric reflux PH
      - rectal examination
      - abdomen
      - packed cell volume
      - total protein
    * - abdominocentesis appearance
      - abdomcentesis total protein
      - outcome
      - surgical lesion
      - timestamp

The horse colic dataset will be a compressed zip file consisting of the SQLite DB. For this example we just wanted a dataset that was available online, but this could be easily plugged into another dataset / data management system like Snowflake, Athena, Hive, BigQuery or Spark, all of those are supported by Flyte.

Takeaways
=========
The example we're trying to demonstrate is a simple feature engineering job that you can seamlessly construct with Flyte. Here's what the nitty-gritties are:

#. Source data is from SQL-like data sources
#. Procreated feature transforms
#. Ability to create a low-code platform
#. Feast integration
#. Serve features to production using Feast
#. TaskTemplate within an imperative workflow

.. tip::

  If you're a data scientist, you needn't worry about the infrastructure overhead. Flyte provides an easy-to-use interface which looks just like a typical library.
