import datetime as dt
import os
import pathlib
from typing import Any, Dict, List, Optional, Type, Union

import pandas as pd
import pytest
from flyteidl.core.types_pb2 import SimpleType
from flytekitplugins.pydantic import BaseModelTransformer
from flytekitplugins.pydantic.commons import PYDANTIC_SUPPORTED_FLYTE_TYPES
from pydantic import BaseModel, Extra

import flytekit
from flytekit.core import context_manager
from flytekit.core.type_engine import TypeEngine
from flytekit.types import directory
from flytekit.types.file import file


class TrainConfig(BaseModel):
    """Config BaseModel for testing purposes."""

    batch_size: int = 32
    lr: float = 1e-3
    loss: str = "cross_entropy"

    class Config:
        extra = Extra.forbid


class Config(BaseModel):
    """Config BaseModel for testing purposes with an optional type hint."""

    model_config: Optional[Union[Dict[str, TrainConfig], TrainConfig]] = TrainConfig()


class ConfigWithDatetime(BaseModel):
    """Config BaseModel for testing purposes with datetime type hint."""

    datetime: dt.datetime = dt.datetime.now()


class NestedConfig(BaseModel):
    """Nested config BaseModel for testing purposes."""

    files: "ConfigWithFlyteFiles"
    dirs: "ConfigWithFlyteDirs"
    df: "ConfigWithPandasDataFrame"
    datetime: "ConfigWithDatetime" = ConfigWithDatetime()

    def __eq__(self, __value: object) -> bool:
        return isinstance(__value, NestedConfig) and all(
            getattr(self, attr) == getattr(__value, attr) for attr in ["files", "dirs", "df", "datetime"]
        )


class ConfigRequired(BaseModel):
    """Config BaseModel for testing purposes with required attribute."""

    model_config: Union[Dict[str, TrainConfig], TrainConfig]


class ConfigWithFlyteFiles(BaseModel):
    """Config BaseModel for testing purposes with flytekit.files.FlyteFile type hint."""

    flytefiles: List[file.FlyteFile]

    def __eq__(self, __value: object) -> bool:
        return isinstance(__value, ConfigWithFlyteFiles) and all(
            pathlib.Path(self_file).read_text() == pathlib.Path(other_file).read_text()
            for self_file, other_file in zip(self.flytefiles, __value.flytefiles)
        )


class ConfigWithFlyteDirs(BaseModel):
    """Config BaseModel for testing purposes with flytekit.directory.FlyteDirectory type hint."""

    flytedirs: List[directory.FlyteDirectory]

    def __eq__(self, __value: object) -> bool:
        return isinstance(__value, ConfigWithFlyteDirs) and all(
            os.listdir(self_dir) == os.listdir(other_dir)
            for self_dir, other_dir in zip(self.flytedirs, __value.flytedirs)
        )


class ConfigWithPandasDataFrame(BaseModel):
    """Config BaseModel for testing purposes with pandas.DataFrame type hint."""

    df: pd.DataFrame

    def __eq__(self, __value: object) -> bool:
        return isinstance(__value, ConfigWithPandasDataFrame) and self.df.equals(__value.df)


class ChildConfig(Config):
    """Child class config BaseModel for testing purposes."""

    d: List[int] = [1, 2, 3]


NestedConfig.update_forward_refs()


@pytest.mark.parametrize(
    "python_type,kwargs",
    [
        (Config, {}),
        (ConfigRequired, {"model_config": TrainConfig()}),
        (TrainConfig, {}),
        (ConfigWithFlyteFiles, {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]}),
        (ConfigWithFlyteDirs, {"flytedirs": ["tests/folder/"]}),
        (ConfigWithPandasDataFrame, {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})}),
        (
            NestedConfig,
            {
                "files": {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]},
                "dirs": {"flytedirs": ["tests/folder/"]},
                "df": {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})},
            },
        ),
    ],
)
def test_transform_round_trip(python_type: Type, kwargs: Dict[str, Any]):
    """Test that a (de-)serialization roundtrip results in the identical BaseModel."""

    ctx = context_manager.FlyteContextManager().current_context()

    type_transformer = BaseModelTransformer()

    python_value = python_type(**kwargs)

    literal_value = type_transformer.to_literal(
        ctx,
        python_value,
        python_type,
        type_transformer.get_literal_type(python_value),
    )

    reconstructed_value = type_transformer.to_python_value(ctx, literal_value, type(python_value))

    assert reconstructed_value == python_value


@pytest.mark.parametrize(
    "config_type,kwargs",
    [
        (Config, {"model_config": {"foo": TrainConfig(loss="mse")}}),
        (ConfigRequired, {"model_config": {"foo": TrainConfig(loss="mse")}}),
        (ConfigWithFlyteFiles, {"flytefiles": ["tests/folder/test_file1.txt"]}),
        (ConfigWithFlyteDirs, {"flytedirs": ["tests/folder/"]}),
        (ConfigWithPandasDataFrame, {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})}),
        (
            NestedConfig,
            {
                "files": {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]},
                "dirs": {"flytedirs": ["tests/folder/"]},
                "df": {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})},
            },
        ),
    ],
)
def test_pass_to_workflow(config_type: Type, kwargs: Dict[str, Any]):
    """Test passing a BaseModel instance to a workflow works."""
    cfg = config_type(**kwargs)

    @flytekit.task
    def train(cfg: config_type) -> config_type:
        return cfg

    @flytekit.workflow
    def wf(cfg: config_type) -> config_type:
        return train(cfg=cfg)

    returned_cfg = wf(cfg=cfg)  # type: ignore

    assert returned_cfg == cfg
    # TODO these assertions are not valid for all types


@pytest.mark.parametrize(
    "kwargs",
    [
        {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]},
    ],
)
def test_flytefiles_in_wf(kwargs: Dict[str, Any]):
    """Test passing a BaseModel instance to a workflow works."""
    cfg = ConfigWithFlyteFiles(**kwargs)

    @flytekit.task
    def read(cfg: ConfigWithFlyteFiles) -> str:
        with open(cfg.flytefiles[0], "r") as f:
            return f.read()

    @flytekit.workflow
    def wf(cfg: ConfigWithFlyteFiles) -> str:
        return read(cfg=cfg)  # type: ignore

    string = wf(cfg=cfg)
    assert string in {"foo\n", "bar\n"}  # type: ignore


@pytest.mark.parametrize(
    "kwargs",
    [
        {"flytedirs": ["tests/folder/"]},
    ],
)
def test_flytedirs_in_wf(kwargs: Dict[str, Any]):
    """Test passing a BaseModel instance to a workflow works."""
    cfg = ConfigWithFlyteDirs(**kwargs)

    @flytekit.task
    def listdir(cfg: ConfigWithFlyteDirs) -> List[str]:
        return os.listdir(cfg.flytedirs[0])

    @flytekit.workflow
    def wf(cfg: ConfigWithFlyteDirs) -> List[str]:
        return listdir(cfg=cfg)  # type: ignore

    dirs = wf(cfg=cfg)
    assert len(dirs) == 2  # type: ignore


def test_double_config_in_wf():
    """Test passing a BaseModel instance to a workflow works."""
    cfg1 = TrainConfig(batch_size=13)
    cfg2 = TrainConfig(batch_size=31)

    @flytekit.task
    def are_different(cfg1: TrainConfig, cfg2: TrainConfig) -> bool:
        return cfg1 != cfg2

    @flytekit.workflow
    def wf(cfg1: TrainConfig, cfg2: TrainConfig) -> bool:
        return are_different(cfg1=cfg1, cfg2=cfg2)  # type: ignore

    assert wf(cfg1=cfg1, cfg2=cfg2), wf(cfg1=cfg1, cfg2=cfg2)  # type: ignore


@pytest.mark.parametrize(
    "python_type,config_kwargs",
    [
        (Config, {}),
        (ConfigRequired, {"model_config": TrainConfig()}),
        (TrainConfig, {}),
        (ConfigWithFlyteFiles, {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]}),
        (ConfigWithFlyteDirs, {"flytedirs": ["tests/folder/"]}),
        (ConfigWithPandasDataFrame, {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})}),
        (
            NestedConfig,
            {
                "files": {"flytefiles": ["tests/folder/test_file1.txt", "tests/folder/test_file2.txt"]},
                "dirs": {"flytedirs": ["tests/folder/"]},
                "df": {"df": pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})},
            },
        ),
    ],
)
def test_dynamic(python_type: Type[BaseModel], config_kwargs: Dict[str, Any]):
    config_instance = python_type(**config_kwargs)

    @flytekit.task
    def train(cfg: BaseModel):
        print(cfg)

    @flytekit.dynamic(cache=True, cache_version="0.3")
    def sub_wf(cfg: BaseModel):
        train(cfg=cfg)

    @flytekit.workflow
    def wf():
        sub_wf(cfg=config_instance)

    wf()


def test_supported():
    assert len(PYDANTIC_SUPPORTED_FLYTE_TYPES) == 9


def test_single_df():
    ctx = context_manager.FlyteContextManager.current_context()
    lt = TypeEngine.to_literal_type(ConfigWithPandasDataFrame)
    assert lt.simple == SimpleType.STRUCT

    pyd = ConfigWithPandasDataFrame(df=pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]}))
    lit = TypeEngine.to_literal(ctx, pyd, ConfigWithPandasDataFrame, lt)
    assert lit.map is not None
    offloaded_keys = list(lit.map.literals["Serialized Flyte Objects"].map.literals.keys())
    assert len(offloaded_keys) == 1
    assert (
        lit.map.literals["Serialized Flyte Objects"].map.literals[offloaded_keys[0]].scalar.structured_dataset
        is not None
    )
