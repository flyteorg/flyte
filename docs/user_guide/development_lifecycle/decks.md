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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 1-4
```

We create a new deck named `pca` and render Markdown content along with a
[PCA](https://en.wikipedia.org/wiki/Principal_component_analysis) plot.

You can begin by initializing an {ref}`ImageSpec <image_spec_example>` object to encompass all the necessary dependencies.
This approach automatically triggers a Docker build, alleviating the need for you to manually create a Docker image.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 15-19
```

:::{important}
Replace `ghcr.io/flyteorg` with a container registry you've access to publish to.
To upload the image to the local registry in the demo cluster, indicate the registry as `localhost:30000`.
:::

Note the usage of `append` to append the Plotly deck to the Markdown deck.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 57-64
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_top_frame_renderer.png
:alt: Top frame renderer
:class: with-shadow
:::

#### Markdown renderer

Converts a Markdown string into HTML, producing HTML as a Unicode string.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 97-111
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_image_renderer.png
:alt: Image renderer
:class: with-shadow
:::

#### Table renderer

Converts a Pandas dataframe into an HTML table.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/decks.py
:caption: development_lifecycle/decks.py
:lines: 115-123
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_decks_table_renderer.png
:alt: Table renderer
:class: with-shadow
:::

### Contribute to renderers

Don't hesitate to integrate a new renderer into
[renderer.py](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-deck-standard/flytekitplugins/deck/renderer.py)
if your deck renderers can enhance data visibility.
Feel encouraged to open a pull request and play a part in enhancing the Flyte deck renderer ecosystem!

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/development_lifecycle/
