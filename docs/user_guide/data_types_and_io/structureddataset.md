(structured_dataset)=

# StructuredDataset

```{eval-rst}
.. tags:: Basic, DataFrame
```
```{currentmodule} flytekit.types.structured
```

As with most type systems, Python has primitives, container types like maps and tuples, and support for user-defined structures. However, while there’s a rich variety of dataframe classes (Pandas, Spark, Pandera, etc.), there’s no native Python type that represents a dataframe in the abstract. This is the gap that the {py:class}`StructuredDataset` type is meant to fill. It offers the following benefits:

- Eliminate boilerplate code you would otherwise need to write to serialize/deserialize from file objects into dataframe instances,
- Eliminate additional inputs/outputs that convey metadata around the format of the tabular data held in those files,
- Add flexibility around how dataframe files are loaded,
- Offer a range of dataframe specific functionality - enforce compatibility of different schemas
  (not only at compile time, but also runtime since type information is carried along in the literal),
   store third-party schema definitions, and potentially in the future, render sample data, provide summary statistics, etc.

## Usage

To use the `StructuredDataset` type, import `pandas` and define a task that returns a Pandas Dataframe.
Flytekit will detect the Pandas DataFrame return signature and convert the interface for the task to
the {py:class}`StructuredDataset` type.

## Example

This example demonstrates how to work with a structured dataset using Flyte entities.

```{note}
To use the `StructuredDataset` type, you only need to import `pandas`. The other imports specified below are only necessary for this specific example.
```

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the dependencies for the example:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 1-18
```

Define a task that returns a Pandas DataFrame.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:pyobject: generate_pandas_df
```

Using this simplest form, however, the user is not able to set the additional dataframe information alluded to above,

- Column type information
- Serialized byte format
- Storage driver and location
- Additional third party schema information

This is by design as we wanted the default case to suffice for the majority of use-cases, and to require
as few changes to existing code as possible. Specifying these is simple, however, and relies on Python variable annotations,
which is designed explicitly to supplement types with arbitrary metadata.

## Column type information
If you want to extract a subset of actual columns of the dataframe and specify their types for type validation,
you can just specify the column names and their types in the structured dataset type annotation.

First, initialize column types you want to extract from the `StructuredDataset`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 30-31
```

Define a task that opens a structured dataset by calling `all()`.
When you invoke `all()` with ``pandas.DataFrame``, the Flyte engine downloads the parquet file on S3, and deserializes it to `pandas.DataFrame`.
Keep in mind that you can invoke ``open()`` with any dataframe type that's supported or added to structured dataset.
For instance, you can use ``pa.Table`` to convert the Pandas DataFrame to a PyArrow table.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 41-51
```

The code may result in runtime failures if the columns do not match.
The input ``df`` has ``Name``, ``Age`` and ``Height`` columns, whereas the output structured dataset will only have the ``Age`` column.

## Serialized byte format
You can use a custom serialization format to serialize your dataframes.
Here's how you can register the Pandas to CSV handler, which is already available,
and enable the CSV serialization by annotating the structured dataset with the CSV format:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 57-71
```

## Storage driver and location
By default, the data will be written to the same place that all other pointer-types (FlyteFile, FlyteDirectory, etc.) are written to.
This is controlled by the output data prefix option in Flyte which is configurable on multiple levels.

That is to say, in the simple default case, Flytekit will,

- Look up the default format for say, Pandas dataframes,
- Look up the default storage location based on the raw output prefix setting,
- Use these two settings to select an encoder and invoke it.

So what's an encoder? To understand that, let's look into how the structured dataset plugin works.

## Inner workings of a structured dataset plugin

Two things need to happen with any dataframe instance when interacting with Flyte:

- Serialization/deserialization from/to the Python instance to bytes (in the format specified above).
- Transmission/retrieval of those bits to/from somewhere.

Each structured dataset plugin (called encoder or decoder) needs to perform both of these steps.
Flytekit decides which of the loaded plugins to invoke based on three attributes:

- The byte format
- The storage location
- The Python type in the task or workflow signature.

These three keys uniquely identify which encoder (used when converting a dataframe in Python memory to a Flyte value,
e.g. when a task finishes and returns a dataframe) or decoder (used when hydrating a dataframe in memory from a Flyte value,
e.g. when a task starts and has a dataframe input) to invoke.

However, it is awkward to require users to use `typing.Annotated` on every signature.
Therefore, Flytekit has a default byte-format for every Python dataframe type registered with flytekit.

## The `uri` argument

BigQuery `uri` allows you to load and retrieve data from cloud using the `uri` argument.
The `uri` comprises of the bucket name and the filename prefixed with `gs://`.
If you specify BigQuery `uri` for structured dataset, BigQuery creates a table in the location specified by the `uri`.
The `uri` in structured dataset reads from or writes to S3, GCP, BigQuery or any storage.

Before writing DataFrame to a BigQuery table,

1. Create a [GCP account](https://cloud.google.com/docs/authentication/getting-started) and create a service account.
2. Create a project and add the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to your `.bashrc` file.
3. Create a dataset in your project.

Here's how you can define a task that converts a pandas DataFrame to a BigQuery table:

```python
@task
def pandas_to_bq() -> StructuredDataset:
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    return StructuredDataset(dataframe=df, uri="gs://<BUCKET_NAME>/<FILE_NAME>")
```

Replace `BUCKET_NAME` with the name of your GCS bucket and `FILE_NAME` with the name of the file the dataframe should be copied to.

### Note that no format was specified in the structured dataset constructor, or in the signature. So how did the BigQuery encoder get invoked?
This is because the stock BigQuery encoder is loaded into Flytekit with an empty format.
The Flytekit `StructuredDatasetTransformerEngine` interprets that to mean that it is a generic encoder
(or decoder) and can work across formats, if a more specific format is not found.

And here's how you can define a task that converts the BigQuery table to a pandas DataFrame:

```python
@task
def bq_to_pandas(sd: StructuredDataset) -> pd.DataFrame:
   return sd.open(pd.DataFrame).all()
```

:::{note}
Flyte creates a table inside the dataset in the project upon BigQuery query execution.
:::

## How to return multiple dataframes from a task?
For instance, how would a task return say two dataframes:
- The first dataframe be written to BigQuery and serialized by one of their libraries,
- The second needs to be serialized to CSV and written at a specific location in GCS different from the generic pointer-data bucket

If you want the default behavior (which is itself configurable based on which plugins are loaded),
you can work just with your current raw dataframe classes.

```python
@task
def t1() -> typing.Tuple[StructuredDataset, StructuredDataset]:
   ...
   return StructuredDataset(df1, uri="bq://project:flyte.table"), \
          StructuredDataset(df2, uri="gs://auxiliary-bucket/data")
```

If you want to customize the Flyte interaction behavior, you'll need to wrap your dataframe in a `StructuredDataset` wrapper object.

## How to define a custom structured dataset plugin?

`StructuredDataset` ships with an encoder and a decoder that handles the conversion of a
Python value to a Flyte literal and vice-versa, respectively.
Here is a quick demo showcasing how one might build a NumPy encoder and decoder,
enabling the use of a 2D NumPy array as a valid type within structured datasets.

### NumPy encoder

Extend `StructuredDatasetEncoder` and implement the `encode` function.
The `encode` function converts NumPy array to an intermediate format (parquet file format in this case).

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:pyobject: NumpyEncodingHandler
```

### NumPy decoder

Extend {py:class}`StructuredDatasetDecoder` and implement the {py:meth}`~StructuredDatasetDecoder.decode` function.
The {py:meth}`~StructuredDatasetDecoder.decode` function converts the parquet file to a `numpy.ndarray`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:pyobject: NumpyDecodingHandler
```

### NumPy renderer

Create a default renderer for numpy array, then Flytekit will use this renderer to
display schema of NumPy array on the Flyte deck.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:pyobject: NumpyRenderer
```

In the end, register the encoder, decoder and renderer with the `StructuredDatasetTransformerEngine`.
Specify the Python type you want to register this encoder with (`np.ndarray`),
the storage engine to register this against (if not specified, it is assumed to work for all the storage backends),
and the byte format, which in this case is `PARQUET`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 127-129
```

You can now use `numpy.ndarray` to deserialize the parquet file to NumPy and serialize a task's output (NumPy array) to a parquet file.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 134-149
```

:::{note}
`pyarrow` raises an `Expected bytes, got a 'int' object` error when the dataframe contains integers.
:::

You can run the code locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 153-157
```

### The nested typed columns

Like most storage formats (e.g. Avro, Parquet, and BigQuery), StructuredDataset support nested field structures.

:::{note}
Nested field StructuredDataset should be run when flytekit version > 1.11.0.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/structured_dataset.py
:caption: data_types_and_io/structured_dataset.py
:lines: 159-270
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
