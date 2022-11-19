Pandera
=======

.. tags:: Integration, DataFrame, Data, Intermediate

Flytekit python natively supports :ref:`many data types <flytekit_to_flyte_type_mapping>`,
including a :ref:`FlyteSchema <typed_schema>` type for
type-annotating pandas dataframes. The flytekit pandera plugin provides an alternative for
defining dataframe schemas by integrating with `pandera <https://pandera.readthedocs.io/en/stable/>`_,
which is a runtime data validation tool for pandas dataframes.

Installation
------------

.. code:: bash

   pip install flytekitplugins-pandera

Quick Start
-----------

Pandera provides a flexible and expressive interface for defining schemas for
tabular data, where you can define the types and other statistical properties
of a column.

.. testcode:: pandera-quickstart

   import pandas as pd
   import pandera as pa
   from pandera.typing import DataFrame, Series

   class Schema(pa.SchemaModel):
       column_1: Series[int] = pa.Field(ge=0)
       column_2: Series[float] = pa.Field(gt=0, lt=100)
       column_3: Series[str] = pa.Field(str_startswith="prefix")

       @pa.check("column_3")
       def check_str_length(cls, series):
           return series.str.len() > 5

   @pa.check_types
   def processing_fn(df: DataFrame[Schema]) -> DataFrame[Schema]:
       df["column_1"] = df["column_1"] * 2
       df["column_2"] = df["column_2"] * 0.5
       df["column_3"] = df["column_3"] + "_suffix"
       return df

   raw_df = pd.DataFrame({
      "column_1": [1, 2, 3],
      "column_2": [1.5, 2.21, 3.9],
      "column_3": ["prefix_a", "prefix_b", "prefix_c"],
   })
   processed_df = processing_fn(raw_df)
   print(processed_df)

.. testoutput:: pandera-quickstart
   :options: +NORMALIZE_WHITESPACE

      column_1  column_2  column_3
   0         2     0.750  prefix_a_suffix
   1         4     1.105  prefix_b_suffix
   2         6     1.950  prefix_c_suffix

Informative errors are raised if invalid data is passed into ``processing_fn``,
indicating the failure case and the index where they were found in the dataframe:

.. testcode:: pandera-quickstart

   invalid_df = pd.DataFrame({
      "column_1": [-1, 2, -3],
      "column_2": [1.5, 2.21, 3.9],
      "column_3": ["prefix_a", "prefix_b", "prefix_c"],
   })
   processing_fn(invalid_df)

.. testoutput:: pandera-quickstart
   :options: +NORMALIZE_WHITESPACE, +IGNORE_EXCEPTION_DETAIL

   Traceback (most recent call last):
   ...
   pandera.errors.SchemaError: error in check_types decorator of function 'processing_fn': <Schema Column(name=column_1, type=<class 'int'>)> failed element-wise validator 0:
   <Check greater_than_or_equal_to: greater_than_or_equal_to(0)>
   failure cases:
      index  failure_case
   0      0            -1
   1      2            -3


Using Pandera with Flytekit Python
----------------------------------
