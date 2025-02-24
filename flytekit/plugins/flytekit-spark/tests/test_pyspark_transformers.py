import pyspark
from flytekitplugins.spark import PySparkPipelineModelTransformer
from flytekitplugins.spark.task import Spark
from pyspark.ml import Pipeline, PipelineModel
from pyspark.ml.feature import Imputer

import flytekit
from flytekit import task, workflow
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.types.structured.structured_dataset import StructuredDatasetTransformerEngine


def test_type_resolution():
    assert type(TypeEngine.get_transformer(PipelineModel)) == PySparkPipelineModelTransformer


def test_basic_get():
    ctx = FlyteContextManager.current_context()
    e = StructuredDatasetTransformerEngine()
    prot = e._protocol_from_type_or_prefix(ctx, pyspark.sql.DataFrame, uri="/tmp/blah")
    en = e.get_encoder(pyspark.sql.DataFrame, prot, "")
    assert en is not None


def test_pipeline_model_compatibility():
    @task(task_config=Spark())
    def my_dataset() -> pyspark.sql.DataFrame:
        session = flytekit.current_context().spark_session
        df = session.createDataFrame([("Megan", 2.0), ("Wayne", float("nan")), ("Dennis", 8.0)], ["name", "age"])
        return df

    @task(task_config=Spark())
    def my_pipleline(df: pyspark.sql.DataFrame) -> PipelineModel:
        imputer = Imputer(inputCols=["age"], outputCols=["imputed_age"])
        pipeline = Pipeline(stages=[imputer]).fit(df)
        return pipeline

    @task(task_config=Spark())
    def run_pipeline(df: pyspark.sql.DataFrame, pipeline: PipelineModel) -> int:
        imputed_df = pipeline.transform(df)

        return imputed_df.filter(imputed_df["imputed_age"].isNull()).count()

    @workflow
    def my_wf() -> int:
        df = my_dataset()
        pipeline = my_pipleline(df=df)

        return run_pipeline(df=df, pipeline=pipeline)

    res = my_wf()
    assert res == 0
