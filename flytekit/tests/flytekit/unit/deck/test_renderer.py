import sys

import pyarrow as pa
import pytest

from flytekit.deck.renderer import DEFAULT_MAX_COLS, DEFAULT_MAX_ROWS, ArrowRenderer, TopFrameRenderer


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
@pytest.mark.parametrize(
    "rows, cols, max_rows, expected_max_rows, max_cols, expected_max_cols",
    [
        (1, 1, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (10, 1, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (1, 10, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (DEFAULT_MAX_ROWS + 1, 10, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (1, DEFAULT_MAX_COLS + 1, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (10, DEFAULT_MAX_COLS + 1, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (DEFAULT_MAX_ROWS + 1, DEFAULT_MAX_COLS + 1, None, DEFAULT_MAX_ROWS, None, DEFAULT_MAX_COLS),
        (100_000, 10, 123, 123, 5, 5),
        (10_000, 1000, DEFAULT_MAX_ROWS, DEFAULT_MAX_ROWS, DEFAULT_MAX_COLS, DEFAULT_MAX_COLS),
    ],
)
def test_renderer(rows, cols, max_rows, expected_max_rows, max_cols, expected_max_cols):
    import pandas as pd

    df = pd.DataFrame({f"abc-{k}": list(range(rows)) for k in range(cols)})
    pa_df = pa.Table.from_pandas(df)

    kwargs = {}
    if max_rows is not None:
        kwargs["max_rows"] = max_rows
    if max_cols is not None:
        kwargs["max_cols"] = max_cols

    assert TopFrameRenderer(**kwargs).to_html(df) == df.to_html(max_rows=expected_max_rows, max_cols=expected_max_cols)
    assert ArrowRenderer().to_html(pa_df) == pa_df.to_string()
