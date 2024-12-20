from typing import Type

from flytekit import BlobType, FlyteContext
from flytekit.extend import T, TypeEngine, TypeTransformer
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from whylogs.core import DatasetProfileView
from whylogs.viz.extensions.reports.profile_summary import ProfileSummaryReport


class WhylogsDatasetProfileTransformer(TypeTransformer[DatasetProfileView]):
    """
    Transforms whylogs Dataset Profile Views to and from a Schema (typed/untyped)
    """

    _TYPE_INFO = BlobType(format="binary", dimensionality=BlobType.BlobDimensionality.SINGLE)

    def __init__(self):
        super(WhylogsDatasetProfileTransformer, self).__init__("whylogs-profile-transformer", t=DatasetProfileView)

    def get_literal_type(self, t: Type[DatasetProfileView]) -> LiteralType:
        return LiteralType(blob=self._TYPE_INFO)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: DatasetProfileView,
        python_type: Type[DatasetProfileView],
        expected: LiteralType,
    ) -> Literal:
        local_dir = ctx.file_access.get_random_local_path()
        python_val.write(local_dir)
        remote_path = ctx.file_access.put_raw_data(local_dir)
        return Literal(scalar=Scalar(blob=Blob(uri=remote_path, metadata=BlobMetadata(type=self._TYPE_INFO))))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[DatasetProfileView]) -> T:
        local_dir = ctx.file_access.get_random_local_path()
        ctx.file_access.download(lv.scalar.blob.uri, local_dir)
        return DatasetProfileView.read(local_dir)

    def to_html(
        self, ctx: FlyteContext, python_val: DatasetProfileView, expected_python_type: Type[DatasetProfileView]
    ) -> str:
        report = ProfileSummaryReport(target_view=python_val)
        return report.report().data


TypeEngine.register(WhylogsDatasetProfileTransformer())
