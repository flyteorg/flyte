---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_visualizing_task_input_and_output)=

# Visualizing task input and output

Flyte {py:class}`~flytekit.deck.Deck`s are a first-class construct in
Flyte, allowing you to generate static HTML reports associated with any of the
outputs materialized within your tasks.

You can think of Decks as stacks of HTML snippets that are logically grouped by
tabs. By default, every task has three decks: an **input**, an **output**, and a
**default** deck.

Flyte materializes Decks via `Renderer`s, which are specific implementations of
how to generate an HTML report from some Python object.

## Enabling Flyte Decks

To enable Flyte Decks, simply set `disable_deck=False` in the `@task` decorator:

```{code-cell} ipython3
import pandas as pd
from flytekit import task, workflow


@task(disable_deck=False)
def iris_data() -> pd.DataFrame:
    ...
```

Specifying this flag indicates that Decks should be rendered whenever this task
is invoked.

## Rendering Task Inputs and Outputs

By default, Flyte will render the inputs and outputs of tasks with the built-in
renderers in the corresponding **input** and **output** {py:class}`~flytekit.deck.Deck`s,
respectively. In the following task, we load the iris dataset using the `plotly` package.

```{code-cell} ipython3

import plotly.express as px
from typing import Optional

from flytekit import task, workflow


@task(disable_deck=False)
def iris_data(
    sample_frac: Optional[float] = None,
    random_state: Optional[int] = None,
) -> pd.DataFrame:
    data = px.data.iris()
    if sample_frac is not None:
        data = data.sample(frac=sample_frac, random_state=random_state)
    return data


@workflow
def wf(
    sample_frac: Optional[float] = None,
    random_state: Optional[int] = None,
):
    iris_data(sample_frac=sample_frac, random_state=random_state)
```

Then, invoking the workflow containing a deck-enabled task will render the
following reports for the input and output data in an HTML file, which you can
see in the logs:

```{code-cell} ipython3
---
tags: [remove-input]
---

# this is an unrendered cell, used to capture the logs in order to render the
# Flyte Decks directly in the docs.
import datetime
import logging
import os
import re
import shutil
from pythonjsonlogger import jsonlogger
from IPython.display import HTML
from pathlib import Path


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

def cp_deck(src):
    src = Path(src)
    target = Path.cwd() / "_flyte_decks" / src.parent.name
    target.mkdir(parents=True, exist_ok=True)
    shutil.copy(src, target)
    return target / "deck.html"


logger = logging.getLogger("flytekit")
logger.setLevel(20)

deck_filter = DeckFilter()
logger.addFilter(deck_filter)
```

```{code-cell} ipython3
---
tags: [remove-output]
---
wf(sample_frac=1.0, random_state=42)
```

```{code-cell} ipython3
---
tags: [remove-input]
---
logger.removeFilter(deck_filter)
for log in deck_filter.logs:
    print(log)
```

```{note}
To see where the HTML file is written to when you run the deck-enabled tasks
locally, you need to set the `FLYTE_SDK_LOGGING_LEVEL` environment variable
to `20`. Doing so will emit logs that look like the above print statement,
where the `deck.html` filepath can be found in the `message` key.
```

## Rendering In-line Decks

You can render Decks inside the task function body by using the **default**
deck, which you can access with the {py:func}`~flytekit.current_context`
function. In the following example, we extend the `iris_data` task with:

- A markdown snippet to provide more context about what the task does.
- A boxplot of the `sepal_length` variable using {py:class}`~flytekitplugins.deck.renderer.BoxRenderer`,
  which leverages the `plotly` package to auto-generate a set of plots and
  summary statistics from the dataframe.

```{code-cell} ipython3
import flytekit
from flytekitplugins.deck.renderer import MarkdownRenderer, BoxRenderer

@task(disable_deck=False)
def iris_data(
    sample_frac: Optional[float] = None,
    random_state: Optional[int] = None,
) -> pd.DataFrame:
    data = px.data.iris()
    if sample_frac is not None:
        data = data.sample(frac=sample_frac, random_state=random_state)

    md_text = (
        "# Iris Dataset\n"
        "This task loads the iris dataset using the  `plotly` package."
    )
    flytekit.current_context().default_deck.append(MarkdownRenderer().to_html(md_text))
    flytekit.Deck("box plot", BoxRenderer("sepal_length").to_html(data))
    return data
```

This will create new tab in the Flyte Deck HTML view named **default**, which
should contain the markdown text we specified.

(getting_started_customer_renderers)=

## Custom Renderers

What if we don't want to show raw data values in the Flyte Deck? We can create a
pandas dataframe renderer that summarizes the data instead of showing raw values
by creating a custom renderer. A renderer is essentially a class with a
`to_html` method.

```{code-cell} ipython
class DataFrameSummaryRenderer:

    def to_html(self, df: pd.DataFrame) -> str:
        assert isinstance(df, pd.DataFrame)
        return df.describe().to_html()
```

Then we can use the `Annotated` type to override the default renderer of the
`pandas.DataFrame` type:

```{code-cell} ipython3
---
tags: [remove-output]
---

try:
    from typing import Annotated
except ImportError:
    from typing_extensions import Annotated


@task(disable_deck=False)
def iris_data(
    sample_frac: Optional[float] = None,
    random_state: Optional[int] = None,
) -> Annotated[pd.DataFrame, DataFrameSummaryRenderer()]:
    data = px.data.iris()
    if sample_frac is not None:
        data = data.sample(frac=sample_frac, random_state=random_state)

    md_text = (
        "# Iris Dataset\n"
        "This task loads the iris dataset using the  `plotly` package."
    )
    flytekit.current_context().default_deck.append(MarkdownRenderer().to_html(md_text))
    flytekit.Deck("box plot", BoxRenderer("sepal_length").to_html(data))
    return data
```

Finally, we can run the workflow and embed the resulting html file by parsing
out the filepath from logs:

```{code-cell} ipython3
---
tags: [remove-input]
---
import warnings

@workflow
def wf(
    sample_frac: Optional[float] = None,
    random_state: Optional[int] = None,
):
    iris_data(sample_frac=sample_frac, random_state=random_state)

deck_filter = DeckFilter()
logger.addFilter(deck_filter)

with warnings.catch_warnings():
    warnings.simplefilter("ignore")
    wf(sample_frac=1.0, random_state=42)

logger.removeFilter(deck_filter)
HTML(filename=cp_deck(deck_filter.deck_files["iris_data"]))
```

As you can see above, Flyte renders in-line decks in the order in which they are
called in the task function body: **default**, then **box plot**, which contain
the markdown description and box plot, respectively.

The built-in decks for the **input** arguments for primitive types like `int`
and `float` are barebones, simply showing the values, and the **output** argument
contains the output of our custom `DataFrameSummaryRenderer` to show a summary
of our dataset.

```{admonition} Learn more
:class: important

Flyte Decks are simple to customize, as long as you can render the Python object
into some HTML representation.

Learn more about Flyte Decks in the {ref}`User Guide <decks>`.
```

## What's Next?

In this guide, you learned how to generate static HTML reports to gain more
visibility into Flyte tasks. In the next guide, you'll learn how to optimize
your tasks via caching, retries, parallelization, resource allocation, and
plugins.
