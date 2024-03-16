from typing import TYPE_CHECKING, List, Optional, Union

from flytekit import lazy_module
from flytekit.types.file import FlyteFile

if TYPE_CHECKING:
    import markdown
    import pandas as pd
    import PIL.Image
    import plotly.express as px
else:
    pd = lazy_module("pandas")
    markdown = lazy_module("markdown")
    px = lazy_module("plotly.express")
    PIL = lazy_module("PIL")


class SourceCodeRenderer:
    """
    Convert Python source code to HTML, and return HTML as a unicode string.
    """

    def __init__(self, title: str = "Source Code"):
        self._title = title

    def to_html(self, source_code: str) -> str:
        """
        Convert the provided Python source code into HTML format using Pygments library.

        This method applies a colorful style and replaces the color "#fff0f0" with "#ffffff" in CSS.

        Args:
            source_code (str): The Python source code to be converted.

        Returns:
            str: The resulting HTML as a string, including CSS and highlighted source code.
        """
        from pygments import highlight
        from pygments.formatters.html import HtmlFormatter
        from pygments.lexers.python import PythonLexer

        formatter = HtmlFormatter(style="colorful")
        css = formatter.get_style_defs(".highlight").replace("#fff0f0", "#ffffff")
        html = highlight(source_code, PythonLexer(), formatter)
        return f"<style>{css}</style>{html}"


class FrameProfilingRenderer:
    """
    Generate a ProfileReport based on a pandas DataFrame
    """

    def __init__(self, title: str = "Pandas Profiling Report"):
        self._title = title

    def to_html(self, df: "pd.DataFrame") -> str:
        assert isinstance(df, pd.DataFrame)
        import ydata_profiling

        profile = ydata_profiling.ProfileReport(df, title=self._title)
        return profile.to_html()


class MarkdownRenderer:
    """Convert a markdown string to HTML and return HTML as a unicode string.

    This is a shortcut function for `Markdown` class to cover the most
    basic use case.  It initializes an instance of Markdown, loads the
    necessary extensions and runs the parser on the given text.
    """

    def to_html(self, text: str) -> str:
        return markdown.markdown(text)


class BoxRenderer:
    """
    In a box plot, rows of `data_frame` are grouped together into a
    box-and-whisker mark to visualize their distribution.

    Each box spans from quartile 1 (Q1) to quartile 3 (Q3). The second
    quartile (Q2) is marked by a line inside the box. By default, the
    whiskers correspond to the box edges +/- 1.5 times the interquartile
    range (IQR: Q3-Q1), see "points" for other options.
    """

    # More detail, see https://plotly.com/python/box-plots/
    def __init__(self, column_name):
        self._column_name = column_name

    def to_html(self, df: "pd.DataFrame") -> str:
        fig = px.box(df, y=self._column_name)
        return fig.to_html()


class ImageRenderer:
    """Converts a FlyteFile or PIL.Image.Image object to an HTML string with the image data
    represented as a base64-encoded string.
    """

    def to_html(self, image_src: Union[FlyteFile, "PIL.Image.Image"]) -> str:
        img = self._get_image_object(image_src)
        return self._image_to_html_string(img)

    @staticmethod
    def _get_image_object(image_src: Union[FlyteFile, "PIL.Image.Image"]) -> "PIL.Image.Image":
        if isinstance(image_src, FlyteFile):
            local_path = image_src.download()
            return PIL.Image.open(local_path)
        elif isinstance(image_src, PIL.Image.Image):
            return image_src
        else:
            raise ValueError("Unsupported image source type")

    @staticmethod
    def _image_to_html_string(img: "PIL.Image.Image") -> str:
        import base64
        from io import BytesIO

        buffered = BytesIO()
        img.save(buffered, format="PNG")
        img_base64 = base64.b64encode(buffered.getvalue()).decode()
        return f'<img src="data:image/png;base64,{img_base64}" alt="Rendered Image" />'


class TableRenderer:
    """
    Convert a pandas DataFrame into an HTML table.
    """

    def to_html(self, df: pd.DataFrame, header_labels: Optional[List] = None, table_width: Optional[int] = None) -> str:
        # Check if custom labels are provided and have the correct length
        if header_labels is not None and len(header_labels) == len(df.columns):
            df = df.copy()
            df.columns = header_labels

        style = f"""
            <style>
                .table-class {{
                    border: 1px solid #ccc;  /* Add a thin border around the table */
                    border-collapse: collapse;
                    font-family: Arial, sans-serif;
                    color: #333;
                    {f'width: {table_width}px;' if table_width is not None else ''}
                }}

                .table-class th, .table-class td {{
                    border: 1px solid #ccc;  /* Add a thin border around each cell */
                    padding: 8px;  /* Add some padding inside each cell */
                }}

                /* Set the background color for even rows */
                .table-class tr:nth-child(even) {{
                    background-color: #f2f2f2;
                }}

                /* Add a hover effect to the rows */
                .table-class tr:hover {{
                    background-color: #ddd;
                }}

                /* Center the column headers */
                .table-class th {{
                    text-align: center;
                }}
            </style>
        """
        return style + df.to_html(classes="table-class", index=False)


class GanttChartRenderer:
    """
    This renderer is primarily used by the timeline deck. The input DataFrame should
    have at least the following columns:
    - "Start": datetime.datetime (represents the start time)
    - "Finish": datetime.datetime (represents the end time)
    - "Name": string (the name of the task or event)
    """

    def to_html(self, df: pd.DataFrame, chart_width: Optional[int] = None) -> str:
        fig = px.timeline(df, x_start="Start", x_end="Finish", y="Name", color="Name", width=chart_width)

        fig.update_xaxes(
            tickangle=90,
            rangeslider_visible=True,
            tickformatstops=[
                dict(dtickrange=[None, 1], value="%3f ms"),
                dict(dtickrange=[1, 60], value="%S:%3f s"),
                dict(dtickrange=[60, 3600], value="%M:%S m"),
                dict(dtickrange=[3600, None], value="%H:%M h"),
            ],
        )

        # Remove y-axis tick labels and title since the time line deck space is limited.
        fig.update_yaxes(showticklabels=False, title="")

        fig.update_layout(
            autosize=True,
            # Set the orientation of the legend to horizontal and move the legend anchor 2% beyond the top of the timeline graph's vertical axis
            legend=dict(orientation="h", y=1.02),
        )

        return fig.to_html()
