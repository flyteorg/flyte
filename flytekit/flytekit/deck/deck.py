import os
import typing
from typing import Optional

from flytekit.core.context_manager import ExecutionParameters, ExecutionState, FlyteContext, FlyteContextManager
from flytekit.loggers import logger
from flytekit.tools.interactive import ipython_check

OUTPUT_DIR_JUPYTER_PREFIX = "jupyter"
DECK_FILE_NAME = "deck.html"


class Deck:
    """
    Deck enable users to get customizable and default visibility into their tasks.

    Deck contains a list of renderers (FrameRenderer, MarkdownRenderer) that can
    generate a html file. For example, FrameRenderer can render a DataFrame as an HTML table,
    MarkdownRenderer can convert Markdown string to HTML

    Flyte context saves a list of deck objects, and we use renderers in those decks to render
    the data and create an HTML file when those tasks are executed

    Each task has a least three decks (input, output, default). Input/output decks are
    used to render tasks' input/output data, and the default deck is used to render line plots,
    scatter plots or Markdown text. In addition, users can create new decks to render
    their data with custom renderers.

    .. warning::

        This feature is in beta.

    .. code-block:: python

        iris_df = px.data.iris()

        @task()
        def t1() -> str:
            md_text = '#Hello Flyte##Hello Flyte###Hello Flyte'
            m = MarkdownRenderer()
            s = BoxRenderer("sepal_length")
            deck = flytekit.Deck("demo", s.to_html(iris_df))
            deck.append(m.to_html(md_text))
            default_deck = flytekit.current_context().default_deck
            default_deck.append(m.to_html(md_text))
            return md_text


        # Use Annotated to override default renderer
        @task()
        def t2() -> Annotated[pd.DataFrame, TopFrameRenderer(10)]:
            return iris_df
    """

    def __init__(self, name: str, html: Optional[str] = ""):
        self._name = name
        self._html = html
        FlyteContextManager.current_context().user_space_params.decks.append(self)

    def append(self, html: str) -> "Deck":
        assert isinstance(html, str)
        self._html = self._html + "\n" + html
        return self

    @property
    def name(self) -> str:
        return self._name

    @property
    def html(self) -> str:
        return self._html


class TimeLineDeck(Deck):
    """
    The TimeLineDeck class is designed to render the execution time of each part of a task.
    Unlike deck class, the conversion of data to HTML is delayed until the html property is accessed.
    This approach is taken because rendering a timeline graph with partial data would not provide meaningful insights.
    Instead, the complete data set is used to create a comprehensive visualization of the execution time of each part of the task.
    """

    def __init__(self, name: str, html: Optional[str] = ""):
        super().__init__(name, html)
        self.time_info = []

    def append_time_info(self, info: dict):
        assert isinstance(info, dict)
        self.time_info.append(info)

    @property
    def html(self) -> str:
        try:
            from flytekitplugins.deck.renderer import GanttChartRenderer, TableRenderer
        except ImportError:
            warning_info = "Plugin 'flytekit-deck-standard' is not installed. To display time line, install the plugin in the image."
            logger.warning(warning_info)
            return warning_info

        if len(self.time_info) == 0:
            return ""

        import pandas

        df = pandas.DataFrame(self.time_info)
        note = """
            <p><strong>Note:</strong></p>
            <ol>
                <li>if the time duration is too small(< 1ms), it may be difficult to see on the time line graph.</li>
                <li>For accurate execution time measurements, users should refer to wall time and process time.</li>
            </ol>
        """
        # set the accuracy to microsecond
        df["ProcessTime"] = df["ProcessTime"].apply(lambda time: "{:.6f}".format(time))
        df["WallTime"] = df["WallTime"].apply(lambda time: "{:.6f}".format(time))

        gantt_chart_html = GanttChartRenderer().to_html(df)
        time_table_html = TableRenderer().to_html(
            df[["Name", "WallTime", "ProcessTime"]],
            header_labels=["Name", "Wall Time(s)", "Process Time(s)"],
        )
        return gantt_chart_html + time_table_html + note


def _get_deck(
    new_user_params: ExecutionParameters, ignore_jupyter: bool = False
) -> typing.Union[str, "IPython.core.display.HTML"]:  # type:ignore
    """
    Get flyte deck html string
    If ignore_jupyter is set to True, then it will return a str even in a jupyter environment.
    """
    deck_map = {deck.name: deck.html for deck in new_user_params.decks}
    raw_html = get_deck_template().render(metadata=deck_map)
    if not ignore_jupyter and ipython_check():
        try:
            from IPython.core.display import HTML
        except ImportError:
            ...
        return HTML(raw_html)
    return raw_html


def _output_deck(task_name: str, new_user_params: ExecutionParameters):
    ctx = FlyteContext.current_context()
    local_dir = ctx.file_access.get_random_local_directory()
    local_path = f"{local_dir}{os.sep}{DECK_FILE_NAME}"
    try:
        with open(local_path, "w", encoding="utf-8") as f:
            f.write(_get_deck(new_user_params, ignore_jupyter=True))
        logger.info(f"{task_name} task creates flyte deck html to file://{local_path}")
        if ctx.execution_state.mode == ExecutionState.Mode.TASK_EXECUTION:
            fs = ctx.file_access.get_filesystem_for_path(new_user_params.output_metadata_prefix)
            remote_path = f"{new_user_params.output_metadata_prefix}{ctx.file_access.sep(fs)}{DECK_FILE_NAME}"
            kwargs: typing.Dict[str, str] = {
                "ContentType": "text/html",  # For s3
                "content_type": "text/html",  # For gcs
            }
            ctx.file_access.put_data(local_path, remote_path, **kwargs)
    except Exception as e:
        logger.error(f"Failed to write flyte deck html with error {e}.")


def get_deck_template() -> "Template":
    from jinja2 import Environment, FileSystemLoader, select_autoescape

    root = os.path.dirname(os.path.abspath(__file__))
    templates_dir = os.path.join(root, "html")
    env = Environment(
        loader=FileSystemLoader(templates_dir),
        # 🔥 include autoescaping for security purposes
        # sources:
        # - https://jinja.palletsprojects.com/en/3.0.x/api/#autoescaping
        # - https://stackoverflow.com/a/38642558/8474894 (see in comments)
        # - https://stackoverflow.com/a/68826578/8474894
        autoescape=select_autoescape(enabled_extensions=("html",)),
    )
    return env.get_template("template.html")
