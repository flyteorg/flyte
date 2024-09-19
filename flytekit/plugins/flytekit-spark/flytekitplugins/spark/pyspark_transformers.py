from typing import Type

from pyspark.ml import PipelineModel

from flytekit import Blob, BlobMetadata, BlobType, FlyteContext, Literal, LiteralType, Scalar
from flytekit.core.type_engine import TypeEngine
from flytekit.extend import TypeTransformer


class PySparkPipelineModelTransformer(TypeTransformer[PipelineModel]):
    _TYPE_INFO = BlobType(format="binary", dimensionality=BlobType.BlobDimensionality.MULTIPART)

    def __init__(self):
        super(PySparkPipelineModelTransformer, self).__init__(name="PySparkPipelineModel", t=PipelineModel)

    def get_literal_type(self, t: Type[PipelineModel]) -> LiteralType:
        return LiteralType(blob=self._TYPE_INFO)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: PipelineModel,
        python_type: Type[PipelineModel],
        expected: LiteralType,
    ) -> Literal:
        # Must write to remote directory
        remote_dir = ctx.file_access.join(
            ctx.file_access.raw_output_prefix,
            ctx.file_access.get_random_string(),
        )
        python_val.write().overwrite().save(remote_dir)

        return Literal(scalar=Scalar(blob=Blob(uri=remote_dir, metadata=BlobMetadata(type=self._TYPE_INFO))))

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[PipelineModel]
    ) -> PipelineModel:
        remote_dir = lv.scalar.blob.uri
        return PipelineModel.load(remote_dir)


TypeEngine.register(PySparkPipelineModelTransformer())
