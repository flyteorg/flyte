from flytekit import lazy_module

why = lazy_module("whylogs")
constraints = lazy_module("whylogs.core.constraints")
pd = lazy_module("pandas")


class WhylogsSummaryDriftRenderer:
    """
    Creates a whylogs' Summary Drift report from two pandas DataFrames. One of them
    is the reference and the other one is the target data, meaning that this is what
    the report will compare it against.
    """

    @staticmethod
    def to_html(reference_data: pd.DataFrame, target_data: pd.DataFrame) -> str:
        """
        This static method will profile the input data and then generate an HTML report
        with the Summary Drift calculations for all the dataframe's columns

        :param reference_data: The DataFrame that will be the reference for the drift report
        :type: pandas.DataFrame

        :param target_data: The data to compare against and create the Summary Drift report
        :type target_data: pandas.DataFrame
        """

        target_view = why.log(target_data).view()
        reference_view = why.log(reference_data).view()
        viz = why.viz.NotebookProfileVisualizer()
        viz.set_profiles(target_profile_view=target_view, reference_profile_view=reference_view)
        return viz.summary_drift_report().data


class WhylogsConstraintsRenderer:
    """
    Creates a whylogs' Constraints report from a `Constraints` object. Currently our API
    requires the user to have a profiled DataFrame in place to be able to use it. Then the report
    will render a nice HTML that will let users check which constraints passed or failed their
    logic. An example constraints object definition can be written as follows:

    .. code-block:: python

        profile_view = why.log(df).view()
        builder = ConstraintsBuilder(profile_view)
        num_constraint = MetricConstraint(
                            name=f'numbers between {min_value} and {max_value} only',
                            condition=lambda x: x.min > min_value and x.max < max_value,
                            metric_selector=MetricsSelector(
                                                    metric_name='distribution',
                                                    column_name='sepal_length'
                                                    )
                        )

        builder.add_constraint(num_constraint)
        constraints = builder.build()

    Each Constraints object (builder.build() in the former example) can have as many constraints as
    desired. If you want to learn more, check out our docs and examples at https://whylogs.readthedocs.io/
    """

    @staticmethod
    def to_html(constraints: constraints.Constraints) -> str:
        viz = why.viz.NotebookProfileVisualizer()
        report = viz.constraints_report(constraints=constraints)
        return report.data
