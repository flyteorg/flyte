"""
Feast Custom Provider
---------------------

Custom provider helps in handling remote Feast manifests within Flyte. It helps Flyte tasks communicate seamlessly.
"""

from datetime import datetime
from typing import Callable, List, Union

import pandas
from feast.feature_view import FeatureView
from feast.infra.local import LocalProvider
from feast.infra.offline_stores.file_source import FileSource
from feast.infra.offline_stores.offline_store import RetrievalJob
from feast.registry import Registry
from feast.repo_config import RepoConfig
from flytekit.core.context_manager import FlyteContext
from tqdm import tqdm


class FlyteCustomProvider(LocalProvider):
    def __init__(self, config: RepoConfig):
        super().__init__(config)

    def materialize_single_feature_view(
        self,
        config: RepoConfig,
        feature_view: FeatureView,
        start_date: datetime,
        end_date: datetime,
        registry: Registry,
        project: str,
        tqdm_builder: Callable[[int], tqdm],
    ) -> None:
        """
        Loads the latest feature values for a specific feature value from the offline store into the online store.
        """
        self._localize_feature_view(feature_view)

        super().materialize_single_feature_view(
            config, feature_view, start_date, end_date, registry, project, tqdm_builder
        )

    def get_historical_features(
        self,
        config: RepoConfig,
        feature_views: List[FeatureView],
        feature_refs: List[str],
        entity_df: Union[pandas.DataFrame, str],
        registry: Registry,
        project: str,
        full_feature_names: bool,
    ) -> RetrievalJob:
        """
        Returns a training dataframe from the offline store
        """
        # We substitute the remote s3 file with a reference to a local file in each feature view being requested
        for fv in feature_views:
            self._localize_feature_view(fv)

        return super().get_historical_features(
            config,
            feature_views,
            feature_refs,
            entity_df,
            registry,
            project,
            full_feature_names,
        )

    def _localize_feature_view(self, feature_view: FeatureView):
        """
        This function ensures that the `FeatureView` object points to files in the local disk
        """
        if not isinstance(feature_view.batch_source, FileSource):
            return

        # Copy parquet file to a local file
        file_source: FileSource = feature_view.batch_source
        random_local_path = (
            FlyteContext.current_context().file_access.get_random_local_path(
                file_source.path
            )
        )
        FlyteContext.current_context().file_access.get_data(
            file_source.path,
            random_local_path,
            is_multipart=True,
        )
        feature_view.batch_source = FileSource(
            path=random_local_path,
            event_timestamp_column=file_source.event_timestamp_column,
        )
