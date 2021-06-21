Data Cleaning
-------------
Feature Engineering off-late has become one of the most prominent topics in Machine Learning. 
It is the process of transforming raw data into features that better represent the underlying problem to the predictive models, resulting in improved model accuracy on unseen data.

This tutorial will implement data cleaning of SQLite3 data, which does both data imputation and univariate feature selection. These are so-called feature engineering techniques.

Why SQLite3?
============
SQLite3 is written such that the task doesn't depend on the user's image. It basically:

- Shifts the burden of writing the Dockerfile from the user using the task in workflows, to the author of the task type
- Allows the author to optimize the image that the task runs
- Works locally and remotely

.. note::

  SQLite3 container is special; the definition of the Python classes themselves is bundled in Flytekit, hence we just use the Flytekit image.

.. tip::

  SQLite3 is being used to showcase the example of using a ``TaskTemplate``. This is the same for SQLAlchemy. As for Athena, BigQuery, Hive plugins, a container is not required. The queries are registered with FlyteAdmin and sent directly to the respective engines.

Where does Flyte fit in?
========================
Flyte provides a way to train models and perform feature engineering as a single pipeline. 

.. admonition:: What's so special about this example?

  The pipeline doesn't build a container as such; it re-uses the pre-built task containers to construct the workflow!

Dataset
=======
We'll be using the horse colic dataset wherein we'll determine if the lesion of the horse was surgical or not. This is a modified version of the original dataset.

The dataset will have the following columns:

.. list-table:: Horse Colic Features
    :widths: 25 25 25

    * - surgery
      - Age
      - Hospital Number
    * - rectal temperature
      - pulse
      - respiratory rate
    * - temperature of extremities
      - peripheral pulse
      - mucous membranes
    * - capillary refill time
      - pain
      - peristalsis
    * - abdominal distension
      - nasogastric tube
      - nasogastric reflux
    * - nasogastric reflux PH
      - rectal examination
      - abdomen
    * - packed cell volume
      - total protein
      - abdominocentesis appearance
    * - abdomcentesis total protein
      - outcome
      - surgical lesion

The horse colic dataset will be a compressed zip file consisting of the SQLite DB.

Steps to Build the Pipeline
===========================
- Define two feature engineering tasks -- "data imputation" and "univariate feature selection"
- Reference the tasks in the actual file
- Define an SQLite3 Task and generate FlyteSchema
- Pass the inputs through an imperative workflow to validate the dataset
- Return the resultant DataFrame

Takeaways
=========
The example we're trying to demonstrate is a simple feature engineering job that you can seamlessly construct with Flyte. Here's what the nitty-gritties are:

#. Source data is from SQL-like data sources
#. Procreated feature transforms
#. Ability to create a low-code platform 
#. TaskTemplate within an imperative workflow

.. tip:: 

  If you're a data scientist, you needn't worry about the infrastructure overhead. Flyte provides an easy-to-use interface which looks just like a typical library.

Code Walkthrough
================
