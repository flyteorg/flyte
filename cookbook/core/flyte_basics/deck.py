"""
Flyte Decks
-------------

Deck enables users to get customizable and default visibility into their tasks.

Flytekit contains various renderers (FrameRenderer, MarkdownRenderer) that
generate HTML files. For example, FrameRenderer renders a DataFrame as an HTML table,
and MarkdownRenderer converts Markdown string to HTML.

Each task has at least three decks (input, output, and default). Input/output decks are
used to render tasks' input/output data, and the default deck is used to render line plots,
scatter plots or markdown text. In addition, users can create new decks to render
their data with custom renderers.
Let's dive into an example.

"""

# %%
# Import the dependencies.
import flytekit
import pandas as pd
import plotly.express as px
from flytekit import task, workflow
from flytekit.deck.renderer import TopFrameRenderer
from flytekitplugins.deck.renderer import BoxRenderer, MarkdownRenderer
from typing_extensions import Annotated

# Fetch iris data.
iris_df = px.data.iris()


# %%
# Create a new deck called ``demo``, and use the box renderer to display the box plot on demo deck.
# Use MarkdownRenderer to render ``md_text``, and append HTML to the default deck.
@task
def t1() -> str:
    md_text = "#Hello Flyte\n##Hello Flyte\n###Hello Flyte"
    flytekit.Deck("demo", BoxRenderer("sepal_length").to_html(iris_df))
    flytekit.current_context().default_deck.append(MarkdownRenderer().to_html(md_text))
    return md_text


# %%
# .. note::
#
#   To see the log output, the ``FLYTE_SDK_LOGGING_LEVEL`` environment variable should be set to 20.

# %%
# Expected output
#
# .. prompt:: text
#
#   {"asctime": "2022-04-19 23:12:17,266", "name": "flytekit", "levelname": "INFO", "message": "t1 task creates flyte deck html to file:///tmp/flyte/20220419_231216/sandbox/local_flytekit/161e15f8c9331e83237bcf52e604697b/deck.html"}
#   {"asctime": "2022-04-19 23:12:17,283", "name": "flytekit", "levelname": "INFO", "message": "t2 task creates flyte deck html to file:///tmp/flyte/20220419_231216/sandbox/local_flytekit/6d8d1bafe04769592d7b0e212c50bd0e/deck.html"}

# %%
# Open the ``deck.html`` file.

# %%
# .. figure:: https://i.ibb.co/4Sxts8w/Screen-Shot-2022-04-19-at-11-24-55-PM.png
#   :alt: Demo Deck
#   :class: with-shadow
#
# .. figure:: https://i.ibb.co/HtH4C4F/Screen-Shot-2022-04-19-at-11-27-17-PM.png
#   :alt: Default Deck
#   :class: with-shadow
#


# %%
# Use ``Annotated`` to override the default renderer, and display top 10 rows of dataframe.
@task
def t2() -> Annotated[pd.DataFrame, TopFrameRenderer(10)]:
    return iris_df


# %%
# Open the Flyte Deck on FlyteConsole.

# %%
# .. figure:: https://i.ibb.co/7yCJnSs/flyteconsole.png
#   :alt: FlyteConsole
#   :class: with-shadow
#

# %%
# .. figure:: https://i.ibb.co/vhB8Mnz/dataframe.png
#   :alt: Dataframe
#   :class: with-shadow
#

# %%
# Define the workflow.
@workflow
def wf():
    t1()
    t2()


# %%
# Run the code locally.
if __name__ == "__main__":
    wf()


# %%
# Flyte Deck Renderer
# ===================
#
# What Is a Flyte Deck Renderer?
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# Deck renderer is a part of the Deck plugin. Deck provides default and customized task visibility. Deck contains renderers that generate HTML files.
# The renderer renders html in the `Flyte console <https://i.ibb.co/vhB8Mnz/dataframe.png>`__, which can be used to visualize and document data associated with a task.
#
# Flyte context saves a list of deck objects, and these objects use renderers to render the data as HTML files when the task associated with it is executed.
#
# How to Use a Deck Renderer?
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# You can render `custom data <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-deck-standard/flytekitplugins/deck/renderer.py>`__ by creating `new decks <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-deck-standard/flytekitplugins/deck/renderer.py#L21-L30>`__ and using custom renderers. The input to the renderer is your data and the output is a html string.
#
# Every task has a minimum of three decks:
#
# #. Input deck;
# #. Output deck; and
# #. Default deck.
#
# The input and output decks render the input and output data associated with the task.
# The default deck renders line plots, scatter plots, and markdown text.
#
# How to Contribute Your Renderer?
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# We would love any contributions to Flyte Deck. If your deck renderers can improve data visibility, feel free to add a renderer to `renderer.py <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-deck-standard/flytekitplugins/deck/renderer.py>`__. Feel free to open a pull request and help us enrich the Flyte Deck renderer ecosystem!
