whylogs
=======

whylogs is an open source software that allows you to log and inspect differents aspects of your data and ML models. 
It creates efficient and mergeable statistical summaries of your datasets, called profiles, that have similar properties 
to logs produced by regular software applications.


whylogs with Flyte
------------------
The integration we've built consists on a type registration and also renderers.


whylogs Flyte Type
^^^^^^^^^^^^^^^^^^
The first part of the integration consists on the ability to pass a ``DatasetProfileView`` in and out of
the desired tasks. whylogs' DatasetProfileView is the representation of a snapshot of your dataset.
With this integration under Flyte's type system, users will benefit from profiling their desired dataset once
and be able to use the statistical representation for many validations and drift detection systems downstream.

To be able use it, pass in a ``pandas.DataFrame`` to a task and call:

.. code:: python
    @task
    def profiling_task(data: pd.DataFrame) -> DatasetProfileView:
        results = why.log(data)
        return results.view()

This will grant any downstream task the ability to ingest the profiled dataset and use
basically anything from whylogs' api, such as transforming it back to a pandas DataFrame:

.. code:: python
    @task
    def consume_profile_view(profile_view: DatasetProfileView) -> pd.DataFrame:
        return profile_view.to_pandas()


Renderers
---------
The Summary Drift Report is a neat HTML report containing information on the distribution and drift
detection of a target and a reference profile. It makes it easy for users to compare a new read dataset
against the one that was used to train the model that's in production.

To use it, simply take in the two desired ``pandas.DataFrame`` objects and call:

.. code:: python
    renderer = WhylogsSummaryDriftRenderer()
    report = renderer.to_html(target_data=new_data, reference_data=reference_data)
    flytekit.Deck("summary drift", report)

The other report that can be generated with our integration is the Constraints Report. With it, users will
have a neat view on a Flyte Deck that will give intuition on which are the passed and failing constraints, enabling
them to act quicker on potentially wrong results.

.. code:: python
    from whylogs.core.constraints.factories import greater_than_number

    @task
    def constraints_report(profile_view: DatasetProfileView) -> bool:
        builder = ConstraintsBuilder(dataset_profile_view=profile_view)
        builder.add_constraint(greater_than_number(column_name="my_column", number=10.0))

        constraints = builder.build()

        renderer = WhylogsConstraintsRenderer()
        flytekit.Deck("constraints", renderer.to_html(constraints=constraints))

        return constraints.validate()

Since we need a ``Constraints`` object to create this report, users can also return a boolean to whether their dataset
passed the validations or not, and take actions depending on this result downstream, as the code snippet above showed.
Other use-case would be to return the constraints report itself and parse it to provide more information to other
systems automatically.

.. code:: python
    constraints = builder.build()
    constraints.report()

    >> [('my_column greater than number 10.0', 0, 1)]


Installing the plugin
---------------------

In order to have the whylogs plugin installed, simply run:

.. code:: bash
    pip install flytekitplugins.whylogs

And you should then have it available to use on your environment!

With any questions or demands, feel free to join our community Slack_.

.. _Slack: http://join.slack.whylabs.ai/
