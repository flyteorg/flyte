"""
.. _spark_horovod_keras:

Data-Parallel Distributed Training Using Horovod on Spark
---------------------------------------------------------

When time- and compute-intensive deep learning workloads need to be trained efficiently, data-parallel distributed training comes to the rescue.
This technique parallelizes the data and requires sharing of weights between different worker nodes involved in the distributed training after every epoch, which ensures that all worker nodes train a consistent model.
Overall, data-parallel distributed training can help speed up the execution time.

In this tutorial, we will understand how data-parallel distributed training works with Flyte, Horovod, and Spark.

We will forecast sales using the Rossmann store sales dataset. As the data preparation step, we will process the data using Spark, a data processing engine. To improve the speed and ease of distributed training, we will use Horovod, a distributed deep learning training framework.
Lastly, we will build a Keras model and perform distributed training using Horovod's `KerasEstimator API <https://github.com/horovod/horovod/blob/8d34c85ce7ec76e81fb3be99418b0e4d35204dc3/horovod/spark/keras/estimator.py#L88>`__.

Before executing the code, create `work_dir`, an s3 bucket.

Let's get started with the example!

"""
# %%
# First, let's import the required packages into the environment.
import datetime
import os
import pathlib
import subprocess
import sys
from dataclasses import dataclass
from distutils.version import LooseVersion
from typing import Any, Dict, List, Tuple

import flytekit
import horovod.spark.keras as hvd
import pyspark
import pyspark.sql.functions as F
import pyspark.sql.types as T
import tensorflow as tf
import tensorflow.keras.backend as K
from dataclasses_json import dataclass_json
from flytekit import Resources, task, workflow
from flytekit.types.directory import FlyteDirectory
from flytekitplugins.spark import Spark
from horovod.spark.common.backend import SparkBackend
from horovod.spark.common.store import Store
from horovod.tensorflow.keras.callbacks import BestModelCheckpoint
from pyspark import Row
from tensorflow.keras.layers import BatchNormalization, Concatenate, Dense, Dropout, Embedding, Flatten, Input, Reshape

# %%
# We define two variables to represent categorical and continuous columns in the dataset.
CATEGORICAL_COLS = [
    "Store",
    "State",
    "DayOfWeek",
    "Year",
    "Month",
    "Day",
    "Week",
    "CompetitionMonthsOpen",
    "Promo2Weeks",
    "StoreType",
    "Assortment",
    "PromoInterval",
    "CompetitionOpenSinceYear",
    "Promo2SinceYear",
    "Events",
    "Promo",
    "StateHoliday",
    "SchoolHoliday",
]

CONTINUOUS_COLS = [
    "CompetitionDistance",
    "Max_TemperatureC",
    "Mean_TemperatureC",
    "Min_TemperatureC",
    "Max_Humidity",
    "Mean_Humidity",
    "Min_Humidity",
    "Max_Wind_SpeedKm_h",
    "Mean_Wind_SpeedKm_h",
    "CloudCover",
    "trend",
    "trend_de",
    "BeforePromo",
    "AfterPromo",
    "AfterStateHoliday",
    "BeforeStateHoliday",
    "BeforeSchoolHoliday",
    "AfterSchoolHoliday",
]

# %%
# Next, let's initialize a data class to store the hyperparameters that will be used with the model (``epochs``, ``learning_rate``, ``batch_size``, etc.).


@dataclass_json
@dataclass
class Hyperparameters:
    batch_size: int = 100
    sample_rate: float = 0.01
    learning_rate: float = 0.0001
    num_proc: int = 2
    epochs: int = 100
    local_checkpoint_file: str = "checkpoint.h5"
    local_submission_csv: str = "submission.csv"


# %%
# Downloading the Data
# ====================
#
# We define a task to download the data into a ``FlyteDirectory``.
@task(
    cache=True,
    cache_version="0.1",
)
def download_data(dataset: str) -> FlyteDirectory:
    # create a directory named 'data'
    print("==============")
    print("Downloading data")
    print("==============")

    working_dir = flytekit.current_context().working_directory
    data_dir = pathlib.Path(os.path.join(working_dir, "data"))
    data_dir.mkdir(exist_ok=True)

    # download the dataset
    download_subprocess = subprocess.run(
        [
            "curl",
            dataset,
        ],
        check=True,
        capture_output=True,
    )

    # untar the data
    subprocess.run(
        [
            "tar",
            "-xz",
            "-C",
            data_dir,
        ],
        input=download_subprocess.stdout,
    )

    # return the directory populated with Rossmann data files
    return FlyteDirectory(path=str(data_dir))


# %%
# Data Preprocessing
# =====================
#
# 1. Let's start with cleaning and preparing the Google trend data. We create new 'Date' and 'State' columns using PySpark's ``withColumn``. These columns, in addition to other features, will contribute to the prediction of sales.
def prepare_google_trend(
    google_trend_csv: pyspark.sql.DataFrame,
) -> pyspark.sql.DataFrame:
    google_trend_all = google_trend_csv.withColumn(
        "Date", F.regexp_extract(google_trend_csv.week, "(.*?) -", 1)
    ).withColumn(
        "State", F.regexp_extract(google_trend_csv.file, "Rossmann_DE_(.*)", 1)
    )

    # map state NI -> HB,NI to align with other data sources
    google_trend_all = google_trend_all.withColumn(
        "State",
        F.when(google_trend_all.State == "NI", "HB,NI").otherwise(
            google_trend_all.State
        ),
    )

    # expand dates
    return expand_date(google_trend_all)


# %%
# 2. Next, we set a few date-specific values in the DataFrame to analyze the seasonal effects on sales.
def expand_date(df: pyspark.sql.DataFrame) -> pyspark.sql.DataFrame:
    df = df.withColumn("Date", df.Date.cast(T.DateType()))
    return (
        df.withColumn("Year", F.year(df.Date))
        .withColumn("Month", F.month(df.Date))
        .withColumn("Week", F.weekofyear(df.Date))
        .withColumn("Day", F.dayofmonth(df.Date))
    )


# %%
# 3. We retrieve the number of days before/after a special event (such as a promo or holiday). This data helps analyze how the sales may vary before/after a special event.
def add_elapsed(df: pyspark.sql.DataFrame, cols: List[str]) -> pyspark.sql.DataFrame:
    def add_elapsed_column(col, asc):
        def fn(rows):
            last_store, last_date = None, None
            for r in rows:
                if last_store != r.Store:
                    last_store = r.Store
                    last_date = r.Date
                if r[col]:
                    last_date = r.Date
                fields = r.asDict().copy()
                fields[("After" if asc else "Before") + col] = (r.Date - last_date).days
                yield Row(**fields)

        return fn

    # repartition: rearrange the rows in the DataFrame based on the partitioning expression
    # sortWithinPartitions: sort every partition in the DataFrame based on specific columns
    # mapPartitions: apply the 'add_elapsed_column' method to each partition in the dataset, and convert the partitions into a DataFrame
    df = df.repartition(df.Store)
    for asc in [False, True]:
        sort_col = df.Date.asc() if asc else df.Date.desc()
        rdd = df.sortWithinPartitions(df.Store.asc(), sort_col).rdd
        for col in cols:
            rdd = rdd.mapPartitions(add_elapsed_column(col, asc))
        df = rdd.toDF()
    return df


# %%
# 4. We define a function to merge several Spark DataFrames into a single DataFrame to create training and test data.
def prepare_df(
    df: pyspark.sql.DataFrame,
    store_csv: pyspark.sql.DataFrame,
    store_states_csv: pyspark.sql.DataFrame,
    state_names_csv: pyspark.sql.DataFrame,
    google_trend_csv: pyspark.sql.DataFrame,
    weather_csv: pyspark.sql.DataFrame,
) -> pyspark.sql.DataFrame:
    num_rows = df.count()

    # expand dates
    df = expand_date(df)

    # create new columns in the DataFrame by filtering out special events(promo/holiday where sales was zero or store was closed).
    df = (
        df.withColumn("Open", df.Open != "0")
        .withColumn("Promo", df.Promo != "0")
        .withColumn("StateHoliday", df.StateHoliday != "0")
        .withColumn("SchoolHoliday", df.SchoolHoliday != "0")
    )

    # merge store information
    store = store_csv.join(store_states_csv, "Store")
    df = df.join(store, "Store")

    # merge Google Trend information
    google_trend_all = prepare_google_trend(google_trend_csv)
    df = df.join(google_trend_all, ["State", "Year", "Week"]).select(
        df["*"], google_trend_all.trend
    )

    # merge in Google Trend for whole Germany
    google_trend_de = google_trend_all[
        google_trend_all.file == "Rossmann_DE"
    ].withColumnRenamed("trend", "trend_de")
    df = df.join(google_trend_de, ["Year", "Week"]).select(
        df["*"], google_trend_de.trend_de
    )

    # merge weather
    weather = weather_csv.join(
        state_names_csv, weather_csv.file == state_names_csv.StateName
    )
    df = df.join(weather, ["State", "Date"])

    # fix null values
    df = (
        df.withColumn(
            "CompetitionOpenSinceYear",
            F.coalesce(df.CompetitionOpenSinceYear, F.lit(1900)),
        )
        .withColumn(
            "CompetitionOpenSinceMonth",
            F.coalesce(df.CompetitionOpenSinceMonth, F.lit(1)),
        )
        .withColumn("Promo2SinceYear", F.coalesce(df.Promo2SinceYear, F.lit(1900)))
        .withColumn("Promo2SinceWeek", F.coalesce(df.Promo2SinceWeek, F.lit(1)))
    )

    # days and months since the competition has been open, cap it to 2 years
    df = df.withColumn(
        "CompetitionOpenSince",
        F.to_date(
            F.format_string(
                "%s-%s-15", df.CompetitionOpenSinceYear, df.CompetitionOpenSinceMonth
            )
        ),
    )
    df = df.withColumn(
        "CompetitionDaysOpen",
        F.when(
            df.CompetitionOpenSinceYear > 1900,
            F.greatest(
                F.lit(0),
                F.least(F.lit(360 * 2), F.datediff(df.Date, df.CompetitionOpenSince)),
            ),
        ).otherwise(0),
    )
    df = df.withColumn(
        "CompetitionMonthsOpen", (df.CompetitionDaysOpen / 30).cast(T.IntegerType())
    )

    # days and weeks of promotion, cap it to 25 weeks
    df = df.withColumn(
        "Promo2Since",
        F.expr(
            'date_add(format_string("%s-01-01", Promo2SinceYear), (cast(Promo2SinceWeek as int) - 1) * 7)'
        ),
    )
    df = df.withColumn(
        "Promo2Days",
        F.when(
            df.Promo2SinceYear > 1900,
            F.greatest(
                F.lit(0), F.least(F.lit(25 * 7), F.datediff(df.Date, df.Promo2Since))
            ),
        ).otherwise(0),
    )
    df = df.withColumn("Promo2Weeks", (df.Promo2Days / 7).cast(T.IntegerType()))

    # ensure that no row was lost through inner joins
    assert num_rows == df.count(), "lost rows in joins"
    return df


# %%
# 5. We build a dictionary of sorted, distinct categorical variables to create an embedding layer in our Keras model.
def build_vocabulary(df: pyspark.sql.DataFrame) -> Dict[str, List[Any]]:
    vocab = {}
    for col in CATEGORICAL_COLS:
        values = [r[0] for r in df.select(col).distinct().collect()]
        col_type = type([x for x in values if x is not None][0])
        default_value = col_type()
        vocab[col] = sorted(values, key=lambda x: x or default_value)
    return vocab


# %%
# 6. Next, we cast continuous columns to float as part of data preprocessing.
def cast_columns(df: pyspark.sql.DataFrame, cols: List[str]) -> pyspark.sql.DataFrame:
    for col in cols:
        df = df.withColumn(col, F.coalesce(df[col].cast(T.FloatType()), F.lit(0.0)))
    return df


# %%
# 7. Lastly, define a function that returns a list of values based on a key.
def lookup_columns(
    df: pyspark.sql.DataFrame, vocab: Dict[str, List[Any]]
) -> pyspark.sql.DataFrame:
    def lookup(mapping):
        def fn(v):
            return mapping.index(v)

        return F.udf(fn, returnType=T.IntegerType())

    for col, mapping in vocab.items():
        df = df.withColumn(col, lookup(mapping)(df[col]))
    return df


# %%
# The ``data_preparation`` function consolidates all the aforementioned data processing functions.
def data_preparation(
    data_dir: FlyteDirectory, hp: Hyperparameters
) -> Tuple[float, Dict[str, List[Any]], pyspark.sql.DataFrame, pyspark.sql.DataFrame]:
    print("================")
    print("Data preparation")
    print("================")

    # 'current_context' gives the handle of specific parameters in ``data_preparation`` task
    spark = flytekit.current_context().spark_session
    data_dir_path = data_dir.remote_source
    # read the CSV data into Spark DataFrame
    train_csv = spark.read.csv("%s/train.csv" % data_dir_path, header=True)
    test_csv = spark.read.csv("%s/test.csv" % data_dir_path, header=True)

    store_csv = spark.read.csv("%s/store.csv" % data_dir_path, header=True)
    store_states_csv = spark.read.csv(
        "%s/store_states.csv" % data_dir_path, header=True
    )
    state_names_csv = spark.read.csv("%s/state_names.csv" % data_dir_path, header=True)
    google_trend_csv = spark.read.csv("%s/googletrend.csv" % data_dir_path, header=True)
    weather_csv = spark.read.csv("%s/weather.csv" % data_dir_path, header=True)

    # retrieve a sampled subset of the train and test data
    if hp.sample_rate:
        train_csv = train_csv.sample(withReplacement=False, fraction=hp.sample_rate)
        test_csv = test_csv.sample(withReplacement=False, fraction=hp.sample_rate)

    # prepare the DataFrames from the CSV files
    train_df = prepare_df(
        train_csv,
        store_csv,
        store_states_csv,
        state_names_csv,
        google_trend_csv,
        weather_csv,
    ).cache()
    test_df = prepare_df(
        test_csv,
        store_csv,
        store_states_csv,
        state_names_csv,
        google_trend_csv,
        weather_csv,
    ).cache()

    # add elapsed times from the data spanning training & test datasets
    elapsed_cols = ["Promo", "StateHoliday", "SchoolHoliday"]
    elapsed = add_elapsed(
        train_df.select("Date", "Store", *elapsed_cols).unionAll(
            test_df.select("Date", "Store", *elapsed_cols)
        ),
        elapsed_cols,
    )

    # join with the elapsed times
    train_df = train_df.join(elapsed, ["Date", "Store"]).select(
        train_df["*"],
        *[prefix + col for prefix in ["Before", "After"] for col in elapsed_cols],
    )
    test_df = test_df.join(elapsed, ["Date", "Store"]).select(
        test_df["*"],
        *[prefix + col for prefix in ["Before", "After"] for col in elapsed_cols],
    )

    # filter out zero sales
    train_df = train_df.filter(train_df.Sales > 0)

    print("===================")
    print("Prepared data frame")
    print("===================")
    train_df.show()

    all_cols = CATEGORICAL_COLS + CONTINUOUS_COLS

    # select features
    train_df = train_df.select(*(all_cols + ["Sales", "Date"])).cache()
    test_df = test_df.select(*(all_cols + ["Id", "Date"])).cache()

    # build a vocabulary of categorical columns
    vocab = build_vocabulary(
        train_df.select(*CATEGORICAL_COLS)
        .unionAll(test_df.select(*CATEGORICAL_COLS))
        .cache(),
    )

    # cast continuous columns to float
    train_df = cast_columns(train_df, CONTINUOUS_COLS + ["Sales"])
    # search for a key and return a list of values based on a key
    train_df = lookup_columns(train_df, vocab)
    test_df = cast_columns(test_df, CONTINUOUS_COLS)
    test_df = lookup_columns(test_df, vocab)

    # split into training & validation
    # test set is in 2015, use the same period in 2014 from the training set as a validation set
    test_min_date = test_df.agg(F.min(test_df.Date)).collect()[0][0]
    test_max_date = test_df.agg(F.max(test_df.Date)).collect()[0][0]
    one_year = datetime.timedelta(365)
    train_df = train_df.withColumn(
        "Validation",
        (train_df.Date > test_min_date - one_year)
        & (train_df.Date <= test_max_date - one_year),
    )

    # determine max Sales number
    max_sales = train_df.agg(F.max(train_df.Sales)).collect()[0][0]

    # convert Sales to log domain
    train_df = train_df.withColumn("Sales", F.log(train_df.Sales))

    print("===================================")
    print("Data frame with transformed columns")
    print("===================================")
    train_df.show()

    print("================")
    print("Data frame sizes")
    print("================")

    # filter out column validation from the DataFrame, and get the count
    train_rows = train_df.filter(~train_df.Validation).count()
    val_rows = train_df.filter(train_df.Validation).count()
    test_rows = test_df.count()

    # print the number of rows in training, validation and test data
    print("Training: %d" % train_rows)
    print("Validation: %d" % val_rows)
    print("Test: %d" % test_rows)

    return max_sales, vocab, train_df, test_df


# %%
# Training
# ===========
#
# We use ``KerasEstimator`` in Horovod to train our Keras model on an existing pre-processed Spark DataFrame.
# The Estimator leverages Horovod's ability to scale across multiple workers, thereby eliminating any specialized code to perform distributed training.
def train(
    max_sales: float,
    vocab: Dict[str, List[Any]],
    hp: Hyperparameters,
    work_dir: FlyteDirectory,
    train_df: pyspark.sql.DataFrame,
    working_dir: FlyteDirectory,
):
    print("==============")
    print("Model training")
    print("==============")

    # a method to determine root mean square percentage error of exponential of predictions
    def exp_rmspe(y_true, y_pred):
        """Competition evaluation metric, expects logarmithic inputs."""
        pct = tf.square((tf.exp(y_true) - tf.exp(y_pred)) / tf.exp(y_true))

        # compute mean excluding stores with zero denominator
        x = tf.reduce_sum(tf.where(y_true > 0.001, pct, tf.zeros_like(pct)))
        y = tf.reduce_sum(
            tf.where(y_true > 0.001, tf.ones_like(pct), tf.zeros_like(pct))
        )
        return tf.sqrt(x / y)

    def act_sigmoid_scaled(x):
        """Sigmoid scaled to logarithm of maximum sales scaled by 20%."""
        return tf.nn.sigmoid(x) * tf.math.log(max_sales) * 1.2

    # NOTE: exp_rmse and act_sigmoid_scaled functions are not placed at the module level
    # this is because we cannot explicitly send max_sales as an argument to act_sigmoid_scaled since it is an activation function
    # two of them are custom objects, and placing one at the module level and the other within the function doesn't really add up

    all_cols = CATEGORICAL_COLS + CONTINUOUS_COLS
    CUSTOM_OBJECTS = {"exp_rmspe": exp_rmspe, "act_sigmoid_scaled": act_sigmoid_scaled}

    # disable GPUs when building the model to prevent memory leaks
    if LooseVersion(tf.__version__) >= LooseVersion("2.0.0"):
        # See https://github.com/tensorflow/tensorflow/issues/33168
        os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
    else:
        K.set_session(tf.Session(config=tf.ConfigProto(device_count={"GPU": 0})))

    # build the Keras model
    inputs = {col: Input(shape=(1,), name=col) for col in all_cols}
    embeddings = [
        Embedding(len(vocab[col]), 10, input_length=1, name="emb_" + col)(inputs[col])
        for col in CATEGORICAL_COLS
    ]
    continuous_bn = Concatenate()(
        [Reshape((1, 1), name="reshape_" + col)(inputs[col]) for col in CONTINUOUS_COLS]
    )
    continuous_bn = BatchNormalization()(continuous_bn)
    x = Concatenate()(embeddings + [continuous_bn])
    x = Flatten()(x)
    x = Dense(
        1000, activation="relu", kernel_regularizer=tf.keras.regularizers.l2(0.00005)
    )(x)
    x = Dense(
        1000, activation="relu", kernel_regularizer=tf.keras.regularizers.l2(0.00005)
    )(x)
    x = Dense(
        1000, activation="relu", kernel_regularizer=tf.keras.regularizers.l2(0.00005)
    )(x)
    x = Dense(
        500, activation="relu", kernel_regularizer=tf.keras.regularizers.l2(0.00005)
    )(x)
    x = Dropout(0.5)(x)
    # specify element-wise activation
    output = Dense(1, activation=act_sigmoid_scaled)(x)
    model = tf.keras.Model([inputs[f] for f in all_cols], output)
    # display the details of the Keras model
    model.summary()

    opt = tf.keras.optimizers.Adam(lr=hp.learning_rate, epsilon=1e-3)

    # checkpoint callback to specify the options for the returned Keras model
    ckpt_callback = BestModelCheckpoint(
        monitor="val_loss", mode="auto", save_freq="epoch"
    )

    # create an object of Store class
    store = Store.create(work_dir.remote_source)
    # 'SparkBackend' uses `horovod.spark.run` to execute the distributed training function, and
    # returns a list of results by running 'train' on every worker in the cluster
    backend = SparkBackend(
        num_proc=hp.num_proc,
        stdout=sys.stdout,
        stderr=sys.stderr,
        prefix_output_with_timestamp=True,
    )
    # define a Spark Estimator that fits Keras models to a DataFrame
    keras_estimator = hvd.KerasEstimator(
        backend=backend,
        store=store,
        model=model,
        optimizer=opt,
        loss="mae",
        metrics=[exp_rmspe],
        custom_objects=CUSTOM_OBJECTS,
        feature_cols=all_cols,
        label_cols=["Sales"],
        validation="Validation",
        batch_size=hp.batch_size,
        epochs=hp.epochs,
        verbose=2,
        checkpoint_callback=ckpt_callback,
    )

    # The Estimator hides the following details:
    # 1. Binding Spark DataFrames to a deep learning training script
    # 2. Reading data into a format that can be interpreted by the training framework
    # 3. Distributed training using Horovod
    # the user would provide a Keras model to the `KerasEstimator``
    # this `KerasEstimator`` will fit the data and store it in a Spark DataFrame
    keras_model = keras_estimator.fit(train_df).setOutputCols(["Sales_output"])
    # retrieve the model training history
    history = keras_model.getHistory()
    best_val_rmspe = min(history["val_exp_rmspe"])
    print("Best RMSPE: %f" % best_val_rmspe)

    # save the trained model
    keras_model.save(os.path.join(working_dir, hp.local_checkpoint_file))
    print(
        "Written checkpoint to %s" % os.path.join(working_dir, hp.local_checkpoint_file)
    )
    # the Estimator returns a Transformer representation of the trained model once training is complete
    return keras_model


# %%
# Evaluation
# ==============
#
# We use the model transformer to forecast sales.
def test(
    keras_model,
    working_dir: FlyteDirectory,
    test_df: pyspark.sql.DataFrame,
    hp: Hyperparameters,
) -> FlyteDirectory:

    print("================")
    print("Final prediction")
    print("================")

    pred_df = keras_model.transform(test_df)
    pred_df.printSchema()
    pred_df.show(5)
    # convert from log domain to real Sales numbers
    pred_df = pred_df.withColumn("Sales_pred", F.exp(pred_df.Sales_output))

    submission_df = pred_df.select(
        pred_df.Id.cast(T.IntegerType()), pred_df.Sales_pred
    ).toPandas()
    submission_df.sort_values(by=["Id"]).to_csv(
        os.path.join(working_dir, hp.local_submission_csv), index=False
    )
    # predictions are saved to a CSV file.
    print("Saved predictions to %s" % hp.local_submission_csv)

    return working_dir


# %%
# Defining the Spark Task
# ========================
#
# Flyte provides an easy-to-use interface to specify Spark-related attributes.
# The Spark attributes need to be attached to a specific task, and just like that, Flyte can run Spark jobs natively on Kubernetes clusters!
# Within the task, let's call the data pre-processing, training, and evaluation functions.
#
# .. note::
#
#   To set up Spark, refer to :ref:`flyte-and-spark`.
#
@task(
    task_config=Spark(
        # the below configuration is applied to the Spark cluster
        spark_conf={
            "spark.driver.memory": "2000M",
            "spark.executor.memory": "2000M",
            "spark.executor.cores": "1",
            "spark.executor.instances": "2",
            "spark.driver.cores": "1",
            "spark.sql.shuffle.partitions": "16",
            "spark.worker.timeout": "120",
        }
    ),
    cache=True,
    cache_version="0.2",
    requests=Resources(mem="1Gi"),
    limits=Resources(mem="1Gi"),
)
def horovod_spark_task(
    data_dir: FlyteDirectory, hp: Hyperparameters, work_dir: FlyteDirectory
) -> FlyteDirectory:

    max_sales, vocab, train_df, test_df = data_preparation(data_dir, hp)

    # working directory will have the model and predictions as separate files
    working_dir = flytekit.current_context().working_directory

    keras_model = train(
        max_sales,
        vocab,
        hp,
        work_dir,
        train_df,
        working_dir,
    )

    # generate predictions
    return test(keras_model, working_dir, test_df, hp)


# %%
# Lastly, we define a workflow to run the pipeline.
@workflow
def horovod_spark_wf(
    dataset: str = "https://cdn.discordapp.com/attachments/545481172399030272/886952942903627786/rossmann.tgz",
    hp: Hyperparameters = Hyperparameters(),
    work_dir: FlyteDirectory = "s3://flyte-demo/horovod-tmp/",
) -> FlyteDirectory:
    data_dir = download_data(dataset=dataset)

    # work_dir corresponds to the Horovod-Spark store
    return horovod_spark_task(data_dir=data_dir, hp=hp, work_dir=work_dir)


# %%
# Running the Model Locally
# ==========================
#
# We can run the code locally too, provided Spark is enabled and the plugin is set up in the environment.
if __name__ == "__main__":
    metrics_directory = horovod_spark_wf()
    print(f"Find the model and predictions at {metrics_directory}")
