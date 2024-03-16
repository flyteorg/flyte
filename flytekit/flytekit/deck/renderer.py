from typing import TYPE_CHECKING, Any

from typing_extensions import Protocol, runtime_checkable

from flytekit import lazy_module

if TYPE_CHECKING:
    # Always import these modules in type-checking mode or when running pytest
    import pandas
    import pyarrow
else:
    pandas = lazy_module("pandas")
    pyarrow = lazy_module("pyarrow")


@runtime_checkable
class Renderable(Protocol):
    def to_html(self, python_value: Any) -> str:
        """Convert an object(markdown, pandas.dataframe) to HTML and return HTML as a unicode string.
        Returns: An HTML document as a string.
        """
        raise NotImplementedError


DEFAULT_MAX_ROWS = 10
DEFAULT_MAX_COLS = 100


class TopFrameRenderer:
    """
    Render a DataFrame as an HTML table.
    """

    def __init__(self, max_rows: int = DEFAULT_MAX_ROWS, max_cols: int = DEFAULT_MAX_COLS):
        self._max_rows = max_rows
        self._max_cols = max_cols

    def to_html(self, df: "pandas.DataFrame") -> str:
        assert isinstance(df, pandas.DataFrame)
        return df.to_html(max_rows=self._max_rows, max_cols=self._max_cols)


class ArrowRenderer:
    """
    Render an Arrow dataframe as an HTML table.
    """

    def to_html(self, df: "pyarrow.Table") -> str:
        assert isinstance(df, pyarrow.Table)
        return df.to_string()
