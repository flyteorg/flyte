"""
Flytekit StructuredDataset
==========================================================
.. currentmodule:: flytekit.types.structured

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   StructuredDataset
   StructuredDatasetEncoder
   StructuredDatasetDecoder
"""


from flytekit.deck.renderer import ArrowRenderer, TopFrameRenderer
from flytekit.loggers import logger

from .structured_dataset import (
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetTransformerEngine,
)


def register_csv_handlers():
    from .basic_dfs import CSVToPandasDecodingHandler, PandasToCSVEncodingHandler

    StructuredDatasetTransformerEngine.register(PandasToCSVEncodingHandler(), default_format_for_type=True)
    StructuredDatasetTransformerEngine.register(CSVToPandasDecodingHandler(), default_format_for_type=True)


def register_pandas_handlers():
    import pandas as pd

    from .basic_dfs import PandasToParquetEncodingHandler, ParquetToPandasDecodingHandler

    StructuredDatasetTransformerEngine.register(PandasToParquetEncodingHandler(), default_format_for_type=True)
    StructuredDatasetTransformerEngine.register(ParquetToPandasDecodingHandler(), default_format_for_type=True)
    StructuredDatasetTransformerEngine.register_renderer(pd.DataFrame, TopFrameRenderer())


def register_arrow_handlers():
    import pyarrow as pa

    from .basic_dfs import ArrowToParquetEncodingHandler, ParquetToArrowDecodingHandler

    StructuredDatasetTransformerEngine.register(ArrowToParquetEncodingHandler(), default_format_for_type=True)
    StructuredDatasetTransformerEngine.register(ParquetToArrowDecodingHandler(), default_format_for_type=True)
    StructuredDatasetTransformerEngine.register_renderer(pa.Table, ArrowRenderer())


def register_bigquery_handlers():
    try:
        from .bigquery import (
            ArrowToBQEncodingHandlers,
            BQToArrowDecodingHandler,
            BQToPandasDecodingHandler,
            PandasToBQEncodingHandlers,
        )

        StructuredDatasetTransformerEngine.register(PandasToBQEncodingHandlers())
        StructuredDatasetTransformerEngine.register(BQToPandasDecodingHandler())
        StructuredDatasetTransformerEngine.register(ArrowToBQEncodingHandlers())
        StructuredDatasetTransformerEngine.register(BQToArrowDecodingHandler())
    except ImportError:
        logger.info(
            "We won't register bigquery handler for structured dataset because "
            "we can't find the packages google-cloud-bigquery-storage and google-cloud-bigquery"
        )
