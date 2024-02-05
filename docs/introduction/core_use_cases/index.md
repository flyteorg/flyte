(getting_started_core_use_cases)=

# Core Use Cases

This section of the **Getting Started** documentation will take you through the
core use cases for which Flyte is designed. Within the context of these guides,
we're going to assume that the discipline of data science can be broken down into
at least three specializations: data engineering, machine learning (or
statistical modeling more broadly), and analytics.

The purpose of these guides is to provide you, the data and ML practitioner,
some practical and simple examples of how Flyte can help you in your daily
practice.

```{list-table}
:header-rows: 0
:widths: 10 30

* - {doc}`🛠 Data Engineering <data_engineering>`
  - Create an ETL workflow for processing data with SQLAlchemy and Pandas.
* - {doc}`🤖 Machine Learning <machine_learning>`
  - Train a classifier with Scikit-Learn and Pandas.
* - {doc}`📈 Analytics <analytics>`
  - Develop a data cleaning and plotting pipeline with Plotly and Pandas.
```

```{admonition} Learn more
:class: important

Check out more examples in the {ref}`Tutorials <tutorials>` section, which
includes examples of Flyte in specific domains like
{ref}`bioinformatics <bioinformatics>`.
```

```{toctree}
:maxdepth: -1
:hidden:

data_engineering
machine_learning
analytics
```
