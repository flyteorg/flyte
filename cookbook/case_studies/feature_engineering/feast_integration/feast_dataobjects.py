"""
Feature Store Dataclass
-----------------------

This dataclass provides a unified interface to access Feast methods from within a feature store. 
"""

import os
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Dict, List, Optional, Union

import pandas as pd
from dataclasses_json import dataclass_json
from feast import FeatureStore as FeastFeatureStore
from feast.entity import Entity
from feast.feature_service import FeatureService
from feast.feature_view import FeatureView
from feast.infra.offline_stores.file import FileOfflineStoreConfig
from feast.infra.online_stores.sqlite import SqliteOnlineStoreConfig
from feast.repo_config import RepoConfig
from flytekit import FlyteContext
from flytekit.configuration import aws


@dataclass_json
@dataclass
class FeatureStoreConfig:
    registry_path: str
    project: str
    s3_bucket: str
    online_store_path: str = "online.db"


@dataclass_json
@dataclass
class FeatureStore:
    config: FeatureStoreConfig

    def _build_feast_feature_store(self):
        os.environ["FEAST_S3_ENDPOINT_URL"] = aws.S3_ENDPOINT.get()
        os.environ["AWS_ACCESS_KEY_ID"] = aws.S3_ACCESS_KEY_ID.get()
        os.environ["AWS_SECRET_ACCESS_KEY"] = aws.S3_SECRET_ACCESS_KEY.get()

        config = RepoConfig(
            registry=f"s3://{self.config.s3_bucket}/{self.config.registry_path}",
            project=self.config.project,
            # Notice the use of a custom provider.
            provider="custom_provider.provider.FlyteCustomProvider",
            offline_store=FileOfflineStoreConfig(),
            online_store=SqliteOnlineStoreConfig(path=self.config.online_store_path),
        )
        return FeastFeatureStore(config=config)

    def apply(
        self,
        objects: Union[
            Entity,
            FeatureView,
            FeatureService,
            List[Union[FeatureView, Entity, FeatureService]],
        ],
    ) -> None:
        fs = self._build_feast_feature_store()
        fs.apply(objects)

        # Applying also initializes the sqlite tables in the online store
        FlyteContext.current_context().file_access.upload(
            self.config.online_store_path,
            f"s3://{self.config.s3_bucket}/{self.config.online_store_path}",
        )

    def get_historical_features(
        self,
        entity_df: Union[pd.DataFrame, str],
        features: Optional[Union[List[str], FeatureService]] = None,
    ) -> pd.DataFrame:
        fs = self._build_feast_feature_store()
        retrieval_job = fs.get_historical_features(
            entity_df=entity_df,
            features=features,
        )
        return retrieval_job.to_df()

    def materialize(
        self,
        start_date: datetime,
        end_date: datetime,
        feature_views: Optional[List[str]] = None,
    ) -> None:
        FlyteContext.current_context().file_access.download(
            f"s3://{self.config.s3_bucket}/{self.config.online_store_path}",
            self.config.online_store_path,
        )
        fs = self._build_feast_feature_store()
        fs.materialize(
            start_date=start_date,
            end_date=end_date,
        )
        FlyteContext.current_context().file_access.upload(
            self.config.online_store_path,
            f"s3://{self.config.s3_bucket}/{self.config.online_store_path}",
        )

    def get_online_features(
        self,
        features: Union[List[str], FeatureService],
        entity_rows: List[Dict[str, Any]],
        feature_refs: Optional[List[str]] = None,
        full_feature_names: bool = False,
    ) -> Dict[str, Any]:
        FlyteContext.current_context().file_access.download(
            f"s3://{self.config.s3_bucket}/{self.config.online_store_path}",
            self.config.online_store_path,
        )
        fs = self._build_feast_feature_store()

        online_response = fs.get_online_features(
            features, entity_rows, feature_refs, full_feature_names
        )
        return online_response.to_dict()
