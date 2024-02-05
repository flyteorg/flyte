---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_analytics)=

# Analytics

Flyte is ideal for data cleaning, statistical summarization, and plotting
because with `flytekit` you can leverage the rich Python ecosystem of data
processing and visualization tools.

## Cleaning Data

In this example, we are going to analyze some covid vaccination data:

```{code-cell} ipython3
import pandas as pd
import plotly
import plotly.graph_objects as go
from flytekit import Deck, task, workflow, Resources


@task(requests=Resources(mem="1Gi"))
def clean_data() -> pd.DataFrame:
    """Clean the dataset."""
    df = pd.read_csv("https://covid.ourworldindata.org/data/owid-covid-data.csv")
    filled_df = (
        df.sort_values(["people_vaccinated"], ascending=False)
        .groupby("location")
        .first()
        .reset_index()
    )[["location", "people_vaccinated", "population", "date"]]
    return filled_df
```

As you can see, we're using `pandas` for data processing, and in the task
below we use `plotly` to create a choropleth map of the percent of a country's
population that has received at least one COVID-19 vaccination.

## Rendering Plots

We can use {ref}`Flyte Decks <decks>` for rendering a static HTML report
of the map. In this case, we normalize the `people_vaccinated` by the
`population` count of each country:

```{code-cell} ipython3
@task(disable_deck=False)
def plot(df: pd.DataFrame):
    """Render a Choropleth map."""
    df["text"] = df["location"] + "<br>" + "Last updated on: " + df["date"]
    fig = go.Figure(
        data=go.Choropleth(
            locations=df["location"],
            z=df["people_vaccinated"].astype(float) / df["population"].astype(float),
            text=df["text"],
            locationmode="country names",
            colorscale="Blues",
            autocolorscale=False,
            reversescale=False,
            marker_line_color="darkgray",
            marker_line_width=0.5,
            zmax=1,
            zmin=0,
        )
    )

    fig.update_layout(
        title_text=(
          "Percent population with at least one dose of COVID-19 vaccine"
        ),
        geo_scope="world",
        geo=dict(
            showframe=False, showcoastlines=False, projection_type="equirectangular"
        ),
    )
    Deck("Choropleth Map", plotly.io.to_html(fig))


@workflow
def analytics_workflow():
    """Prepare a data analytics workflow."""
    plot(df=clean_data())
```

Running this workflow, we get an interactive plot, courtesy of `plotly`:

```{code-cell} ipython3
---
tags: [remove-input]
---

# this is an unrendered cell, used to capture the logs in order to render the
# Flyte Decks directly in the docs.
import logging
import os
import re
from pythonjsonlogger import jsonlogger
from IPython.display import HTML


class DeckFilter(logging.Filter):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.formatter = jsonlogger.JsonFormatter(
            fmt="%(asctime)s %(name)s %(levelname)s %(message)s"
        )
        self.logs = []
        self.deck_files = {}

    def filter(self, record):
        patt = "(.+) task creates flyte deck html to (.+/deck.html)"
        msg = record.getMessage()
        matches = re.match(patt, msg)

        if msg == "Connection error. Skip stats collection.":
            return False

        if matches:
            task, filepath = matches.group(1), matches.group(2)
            self.logs.append(self.formatter.format(record))
            self.deck_files[task] = re.sub("^file://", "", filepath)
        return False


logger = logging.getLogger("flytekit")
logger.setLevel(20)

deck_filter = DeckFilter()
logger.addFilter(deck_filter)
```

```{code-cell} ipython3
analytics_workflow()
```

```{code-cell} ipython3
---
tags: [remove-input]
---

import os
import shutil
from pathlib import Path

def cp_deck(src):
    src = Path(src)
    target = Path.cwd() / "_flyte_decks" / src.parent.name
    target.mkdir(parents=True, exist_ok=True)
    shutil.copy(src, target)
    return target / "deck.html"

logger.removeFilter(deck_filter)
HTML(filename=cp_deck(deck_filter.deck_files["plot"]))
```

## Custom Flyte Deck Renderers

You can also create your own {ref}`custom Flyte Deck renderers <getting_started_customer_renderers>`
to visualize data with any plotting/visualization library of your choice, as
long as you can render HTML for the objects of interest.

```{important}
Prefer other data processing frameworks? Flyte ships with
[Polars](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-polars),
{ref}`Dask <plugins-dask-k8s>`, {ref}`Modin <modin-integration>`, {ref}`Spark <plugins-spark-k8s>`,
[Vaex](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-vaex),
and [DBT](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-dbt)
integrations.

If you need to connect to a database, Flyte provides first-party
support for {ref}`AWS Athena <aws-athena>`, {ref}`Google Bigquery <big-query>`,
{ref}`Snowflake <plugins-snowflake>`, {ref}`SQLAlchemy <sql_alchemy>`, and
{ref}`SQLite3 <integrations_sql_sqlite3>`.
```
