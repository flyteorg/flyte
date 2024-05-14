(decks)=

# Decks (TODO merge "Visualizing task input and output" content below)

```{eval-rst}
.. tags:: UI, Intermediate
```

The Decks feature enables you to obtain customizable and default visibility into your tasks.
Think of it as a visualization tool that you can utilize within your Flyte tasks.

Decks are equipped with a variety of {ref}`renderers <deck_renderer>`,
such as FrameRenderer and MarkdownRenderer. These renderers produce HTML files.
As an example, FrameRenderer transforms a DataFrame into an HTML table, and MarkdownRenderer converts Markdown text into HTML.

Each task has a minimum of three decks: input, output and default.
The input/output decks are used to render the input/output data of tasks,
while the default deck can be used to render line plots, scatter plots or Markdown text.
Additionally, you can create new decks to render your data using custom renderers.

:::{note}
Flyte Decks is an opt-in feature; to enable it, set `enable_deck` to `True` in the task parameters.
:::

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 1-4
```

We create a new deck named `pca` and render Markdown content along with a
[PCA](https://en.wikipedia.org/wiki/Principal_component_analysis) plot.

You can begin by initializing an {ref}`ImageSpec <image_spec_example>` object to encompass all the necessary dependencies.
This approach automatically triggers a Docker build, alleviating the need for you to manually create a Docker image.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 15-19
```

:::{important}
Replace `ghcr.io/flyteorg` with a container registry you've access to publish to.
To upload the image to the local registry in the demo cluster, indicate the registry as `localhost:30000`.
:::

Note the usage of `append` to append the Plotly deck to the Markdown deck.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:pyobject: pca_plot
```

:::{Important}
To view the log output locally, the `FLYTE_SDK_LOGGING_LEVEL` environment variable should be set to 20.
:::

The following is the expected output containing the path to the `deck.html` file:

```
{"asctime": "2023-07-11 13:16:04,558", "name": "flytekit", "levelname": "INFO", "message": "pca_plot task creates flyte deck html to file:///var/folders/6f/xcgm46ds59j7g__gfxmkgdf80000gn/T/flyte-0_8qfjdd/sandbox/local_flytekit/c085853af5a175edb17b11cd338cbd61/deck.html"}
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_deck_plot_local.webp
:alt: Flyte deck plot
:class: with-shadow
:::

Once you execute this task on the Flyte cluster, you can access the deck by clicking the _Flyte Deck_ button:

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_deck_button.png
:alt: Flyte deck button
:class: with-shadow
:::

(deck_renderer)=

## Deck renderer

The Deck renderer is an integral component of the Deck plugin, which offers both default and personalized task visibility.
Within the Deck, an array of renderers is present, responsible for generating HTML files.

These renderers showcase HTML in the user interface, facilitating the visualization and documentation of task-associated data.

In the Flyte context, a collection of deck objects is stored.
When the task connected with a deck object is executed, these objects employ renderers to transform data into HTML files.

### Available renderers

#### Frame renderer

Creates a profile report from a Pandas DataFrame.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 44-51
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_frame_renderer.png
:alt: Frame renderer
:class: with-shadow
:::



#### Top-frame renderer

Renders DataFrame as an HTML table.
This renderer doesn't necessitate plugin installation since it's accessible within the flytekit library.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 57-64
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_top_frame_renderer.png
:alt: Top frame renderer
:class: with-shadow
:::

#### Markdown renderer

Converts a Markdown string into HTML, producing HTML as a Unicode string.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:pyobject: markdown_renderer
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_markdown_renderer.png
:alt: Markdown renderer
:class: with-shadow
:::

#### Box renderer

Groups rows of DataFrame together into a
box-and-whisker mark to visualize their distribution.

Each box extends from the first quartile (Q1) to the third quartile (Q3).
The median (Q2) is indicated by a line within the box.
Typically, the whiskers extend to the edges of the box,
plus or minus 1.5 times the interquartile range (IQR: Q3-Q1).

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 85-91
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_box_renderer.png
:alt: Box renderer
:class: with-shadow
:::

#### Image renderer

Converts a {ref}`FlyteFile <files>` or `PIL.Image.Image` object into an HTML string,
where the image data is encoded as a base64 string.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 97-111
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_image_renderer.png
:alt: Image renderer
:class: with-shadow
:::

#### Table renderer

Converts a Pandas dataframe into an HTML table.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 115-123
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_table_renderer.png
:alt: Table renderer
:class: with-shadow
:::

#### Source code renderer

Converts source code to HTML and renders it as a Unicode string on the deck.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 128-141
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_source_code_renderer.png
:alt: Source code renderer
:class: with-shadow
:::

### Contribute to renderers

Don't hesitate to integrate a new renderer into
[renderer.py](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-deck-standard/flytekitplugins/deck/renderer.py)
if your deck renderers can enhance data visibility.
Feel encouraged to open a pull request and play a part in enhancing the Flyte deck renderer ecosystem!

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/development_lifecycle/


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

## Enabling Flyte decks

To enable Flyte decks, simply set `disable_deck=False` in the `@task` decorator:

```{code-cell} ipython3
import pandas as pd
from flytekit import task, workflow


@task(disable_deck=False)
def iris_data() -> pd.DataFrame:
    ...
```

Specifying this flag indicates that Decks should be rendered whenever this task
is invoked.

## Rendering task inputs and outputs

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

## Rendering in-line decks

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

## Custom renderers

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

## What's next?

In this guide, you learned how to generate static HTML reports to gain more
visibility into Flyte tasks. In the next guide, you'll learn how to optimize
your tasks via caching, retries, parallelization, resource allocation, and
plugins.
