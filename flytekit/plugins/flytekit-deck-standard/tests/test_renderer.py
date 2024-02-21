import datetime
import tempfile

import markdown
import pandas as pd
import pytest
from flytekitplugins.deck.renderer import (
    BoxRenderer,
    FrameProfilingRenderer,
    GanttChartRenderer,
    ImageRenderer,
    MarkdownRenderer,
    SourceCodeRenderer,
    TableRenderer,
)
from PIL import Image

from flytekit.types.file import FlyteFile, JPEGImageFile, PNGImageFile

df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [1, 22]})
time_info_df = pd.DataFrame(
    [
        dict(
            Name="foo",
            Start=datetime.datetime.utcnow(),
            Finish=datetime.datetime.utcnow() + datetime.timedelta(microseconds=1000),
            WallTime=1.0,
            ProcessTime=1.0,
        )
    ]
)


def test_frame_profiling_renderer():
    renderer = FrameProfilingRenderer()
    assert "Pandas Profiling Report" in renderer.to_html(df).title()


def test_markdown_renderer():
    md_text = "#Hello Flyte\n##Hello Flyte\n###Hello Flyte"
    renderer = MarkdownRenderer()
    assert renderer.to_html(md_text) == markdown.markdown(md_text)


def test_box_renderer():
    renderer = BoxRenderer("Name")
    assert "Plotlyconfig = {Mathjaxconfig: 'Local'}" in renderer.to_html(df).title()


def create_simple_image(fmt: str):
    """Create a simple PNG image using PIL"""
    img = Image.new("RGB", (100, 100), color="black")
    tmp = tempfile.mktemp()
    img.save(tmp, fmt)
    return tmp


png_image = create_simple_image(fmt="png")
jpeg_image = create_simple_image(fmt="jpeg")


@pytest.mark.parametrize(
    "image_src",
    [
        FlyteFile(path=png_image),
        JPEGImageFile(path=jpeg_image),
        PNGImageFile(path=png_image),
        Image.open(png_image),
    ],
)
def test_image_renderer(image_src):
    renderer = ImageRenderer()
    assert "<img" in renderer.to_html(image_src)


def test_table_renderer():
    renderer = TableRenderer()
    assert "Dataframe Table-Class" in renderer.to_html(time_info_df).title()


def test_gantt_chart_renderer():
    renderer = GanttChartRenderer()
    assert "Plotlyconfig = {Mathjaxconfig: 'Local'}" in renderer.to_html(time_info_df).title()


def test_source_code_renderer():
    renderer = SourceCodeRenderer()
    source_code = "def hello_world():\n    print('Hello, world!')"
    result = renderer.to_html(source_code)

    # Assert that the result includes parts of the source code
    assert "hello_world" in result
    assert "Hello, world!" in result

    # Assert that the color #ffffff is used instead of #fff0f0
    assert "#ffffff" in result
    assert "#fff0f0" not in result
