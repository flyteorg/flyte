---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(decks)=

# Decks

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

To begin, import the dependencies.

```{code-cell}
import flytekit
from flytekit import ImageSpec, task
from flytekitplugins.deck.renderer import MarkdownRenderer
from sklearn.decomposition import PCA
```

+++ {"lines_to_next_cell": 0}

We create a new deck named `pca` and render Markdown content along with a
[PCA](https://en.wikipedia.org/wiki/Principal_component_analysis) plot.

You can begin by initializing an {ref}`ImageSpec <image_spec_example>` object to encompass all the necessary dependencies.
This approach automatically triggers a Docker build, alleviating the need for you to manually create a Docker image.

```{code-cell}
custom_image = ImageSpec(name="flyte-decks-example", packages=["plotly"], registry="ghcr.io/flyteorg")

if custom_image.is_container():
    import plotly
    import plotly.express as px
```

+++ {"lines_to_next_cell": 0}

:::{important}
Replace `ghcr.io/flyteorg` with a container registry you've access to publish to.
To upload the image to the local registry in the demo cluster, indicate the registry as `localhost:30000`.
:::

Note the usage of `append` to append the Plotly deck to the Markdown deck.

```{code-cell}
@task(enable_deck=True, container_image=custom_image)
def pca_plot():
    iris_df = px.data.iris()
    X = iris_df[["sepal_length", "sepal_width", "petal_length", "petal_width"]]
    pca = PCA(n_components=3)
    components = pca.fit_transform(X)
    total_var = pca.explained_variance_ratio_.sum() * 100
    fig = px.scatter_3d(
        components,
        x=0,
        y=1,
        z=2,
        color=iris_df["species"],
        title=f"Total Explained Variance: {total_var:.2f}%",
        labels={"0": "PC 1", "1": "PC 2", "2": "PC 3"},
    )
    main_deck = flytekit.Deck("pca", MarkdownRenderer().to_html("### Principal Component Analysis"))
    main_deck.append(plotly.io.to_html(fig))
```

+++ {"lines_to_next_cell": 0}

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

```{code-cell}
import pandas as pd
from flytekitplugins.deck.renderer import FrameProfilingRenderer


@task(enable_deck=True)
def frame_renderer() -> None:
    df = pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]})
    flytekit.Deck("Frame Renderer", FrameProfilingRenderer().to_html(df=df))
```

+++ {"lines_to_next_cell": 0}

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_frame_renderer.png
:alt: Frame renderer
:class: with-shadow
:::

+++ {"lines_to_next_cell": 0}

#### Top-frame renderer

Renders DataFrame as an HTML table.
This renderer doesn't necessitate plugin installation since it's accessible within the flytekit library.

```{code-cell}
from typing import Annotated

from flytekit.deck import TopFrameRenderer


@task(enable_deck=True)
def top_frame_renderer() -> Annotated[pd.DataFrame, TopFrameRenderer(1)]:
    return pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]})
```

+++ {"lines_to_next_cell": 0}

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_top_frame_renderer.png
:alt: Top frame renderer
:class: with-shadow
:::

#### Markdown renderer

Converts a Markdown string into HTML, producing HTML as a Unicode string.

```{code-cell}
@task(enable_deck=True)
def markdown_renderer() -> None:
    flytekit.current_context().default_deck.append(
        MarkdownRenderer().to_html("You can install flytekit using this command: ```import flytekit```")
    )
```

+++ {"lines_to_next_cell": 0}

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

```{code-cell}
from flytekitplugins.deck.renderer import BoxRenderer


@task(enable_deck=True)
def box_renderer() -> None:
    iris_df = px.data.iris()
    flytekit.Deck("Box Plot", BoxRenderer("sepal_length").to_html(iris_df))
```

+++ {"lines_to_next_cell": 0}

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_box_renderer.png
:alt: Box renderer
:class: with-shadow
:::

#### Image renderer

Converts a {ref}`FlyteFile <files>` or `PIL.Image.Image` object into an HTML string,
where the image data is encoded as a base64 string.

```{code-cell}
from flytekit import workflow
from flytekit.types.file import FlyteFile
from flytekitplugins.deck.renderer import ImageRenderer


@task(enable_deck=True)
def image_renderer(image: FlyteFile) -> None:
    flytekit.Deck("Image Renderer", ImageRenderer().to_html(image_src=image))


@workflow
def image_renderer_wf(
    image: FlyteFile = "https://bit.ly/3KZ95q4",
) -> None:
    image_renderer(image=image)
```

+++ {"lines_to_next_cell": 0}

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_image_renderer.png
:alt: Image renderer
:class: with-shadow
:::

#### Table renderer

Converts a Pandas dataframe into an HTML table.

```{code-cell}
from flytekitplugins.deck.renderer import TableRenderer


@task(enable_deck=True)
def table_renderer() -> None:
    flytekit.Deck(
        "Table Renderer",
        TableRenderer().to_html(df=pd.DataFrame(data={"col1": [1, 2], "col2": [3, 4]}), table_width=50),
    )
```

+++ {"lines_to_next_cell": 0}

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_table_renderer.png
:alt: Table renderer
:class: with-shadow
:::

#### Source code renderer

Converts source code to HTML and renders it as a Unicode string on the deck.

```{code-cell}
:lines_to_next_cell: 2

import inspect

from flytekitplugins.deck.renderer import SourceCodeRenderer


@task(enable_deck=True)
def source_code_renderer() -> None:
    file_path = inspect.getsourcefile(frame_renderer.__wrapped__)
    with open(file_path, "r") as f:
        source_code = f.read()
    flytekit.Deck(
        "Source Code Renderer",
        SourceCodeRenderer().to_html(source_code),
    )
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
