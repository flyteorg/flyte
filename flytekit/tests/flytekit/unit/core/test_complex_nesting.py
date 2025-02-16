import os
import tempfile
from dataclasses import dataclass
from typing import List

import pytest
from dataclasses_json import DataClassJsonMixin

from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.context_manager import ExecutionState, FlyteContextManager
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.type_engine import TypeEngine
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile


@dataclass
class MyProxyConfiguration(DataClassJsonMixin):
    # File and directory paths kept as 'str' so Flyte doesn't manage these static resources
    splat_data_dir: str
    apriori_file: str


@dataclass
class MyProxyParameters(DataClassJsonMixin):
    id: str
    job_i_step: int


@dataclass
class MyAprioriConfiguration(DataClassJsonMixin):
    static_data_dir: FlyteDirectory
    external_data_dir: FlyteDirectory


@dataclass
class MyInput(DataClassJsonMixin):
    main_product: FlyteFile
    apriori_config: MyAprioriConfiguration
    proxy_config: MyProxyConfiguration
    proxy_params: MyProxyParameters


@pytest.fixture
def folders_and_files_setup():
    tmp_dir = tempfile.TemporaryDirectory()
    fd, path = tempfile.mkstemp(dir=tmp_dir.name)
    tmp_dir_static_data = tempfile.TemporaryDirectory()
    tmp_dir_external_data = tempfile.TemporaryDirectory()

    try:
        with os.fdopen(fd, "w") as tmp:
            tmp.write("Hello world")
        yield path, tmp_dir_static_data.name, tmp_dir_external_data.name
    finally:
        tmp_dir.cleanup()
        tmp_dir_static_data.cleanup()
        tmp_dir_external_data.cleanup()


@pytest.fixture
def two_sample_inputs(folders_and_files_setup):
    (file_path, static_data_path, external_data_path) = folders_and_files_setup

    main_product = FlyteFile(file_path)
    apriori = MyAprioriConfiguration(
        static_data_dir=FlyteDirectory(static_data_path),
        external_data_dir=FlyteDirectory(external_data_path),
    )
    proxy_c = MyProxyConfiguration(splat_data_dir="/tmp/proxy_splat", apriori_file="/opt/config/a_file")
    proxy_p = MyProxyParameters(id="pp_id", job_i_step=1)

    my_input = MyInput(
        main_product=main_product,
        apriori_config=apriori,
        proxy_config=proxy_c,
        proxy_params=proxy_p,
    )

    my_input_2 = MyInput(
        main_product=main_product,
        apriori_config=apriori,
        proxy_config=proxy_c,
        proxy_params=proxy_p,
    )

    yield my_input, my_input_2


def test_dataclass_complex_transform(two_sample_inputs):
    my_input = two_sample_inputs[0]
    my_input_2 = two_sample_inputs[1]

    ctx = FlyteContextManager.current_context()
    literal_type = TypeEngine.to_literal_type(MyInput)
    first_literal = TypeEngine.to_literal(ctx, my_input, MyInput, literal_type)
    assert first_literal.scalar.generic["apriori_config"] is not None

    converted_back_1 = TypeEngine.to_python_value(ctx, first_literal, MyInput)
    assert converted_back_1.apriori_config is not None

    second_literal = TypeEngine.to_literal(ctx, converted_back_1, MyInput, literal_type)
    assert second_literal.scalar.generic["apriori_config"] is not None

    converted_back_2 = TypeEngine.to_python_value(ctx, second_literal, MyInput)
    assert converted_back_2.apriori_config is not None

    input_list = [my_input, my_input_2]
    input_list_type = TypeEngine.to_literal_type(List[MyInput])
    literal_list = TypeEngine.to_literal(ctx, input_list, List[MyInput], input_list_type)
    assert literal_list.collection.literals[0].scalar.generic["apriori_config"] is not None
    assert literal_list.collection.literals[1].scalar.generic["apriori_config"] is not None


def test_two(two_sample_inputs):
    my_input = two_sample_inputs[0]
    my_input_2 = two_sample_inputs[1]

    @dynamic
    def dt1(a: List[MyInput]) -> List[FlyteFile]:
        x = []
        for aa in a:
            x.append(aa.main_product)
        return x

    with FlyteContextManager.with_context(
        FlyteContextManager.current_context().with_serialization_settings(
            SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        with FlyteContextManager.with_context(
            ctx.with_execution_state(
                ctx.execution_state.with_params(
                    mode=ExecutionState.Mode.TASK_EXECUTION,
                )
            )
        ) as ctx:
            input_literal_map = TypeEngine.dict_to_literal_map(
                ctx, d={"a": [my_input, my_input_2]}, type_hints={"a": List[MyInput]}
            )
            dynamic_job_spec = dt1.dispatch_execute(ctx, input_literal_map)
            assert len(dynamic_job_spec.literals["o0"].collection.literals) == 2


def test_str_input(folders_and_files_setup):
    proxy_c = MyProxyConfiguration(splat_data_dir="/tmp/proxy_splat", apriori_file="/opt/config/a_file")
    proxy_p = MyProxyParameters(id="pp_id", job_i_step=1)

    # Intentionally passing in the wrong type
    my_input = MyInput(
        main_product=folders_and_files_setup[0],  # noqa
        apriori_config=MyAprioriConfiguration(
            static_data_dir=FlyteDirectory("gs://my-bucket/one"),
            external_data_dir=FlyteDirectory("gs://my-bucket/two"),
        ),
        proxy_config=proxy_c,
        proxy_params=proxy_p,
    )
    ctx = FlyteContextManager.current_context()
    literal_type = TypeEngine.to_literal_type(MyInput)
    first_literal = TypeEngine.to_literal(ctx, my_input, MyInput, literal_type)
    assert first_literal.scalar.generic is not None


def test_dc_dyn_directory(folders_and_files_setup):
    proxy_c = MyProxyConfiguration(splat_data_dir="/tmp/proxy_splat", apriori_file="/opt/config/a_file")
    proxy_p = MyProxyParameters(id="pp_id", job_i_step=1)

    my_input_gcs = MyInput(
        main_product=FlyteFile(folders_and_files_setup[0]),
        apriori_config=MyAprioriConfiguration(
            static_data_dir=FlyteDirectory("gs://my-bucket/one"),
            external_data_dir=FlyteDirectory("gs://my-bucket/two"),
        ),
        proxy_config=proxy_c,
        proxy_params=proxy_p,
    )

    my_input_gcs_2 = MyInput(
        main_product=FlyteFile(folders_and_files_setup[0]),
        apriori_config=MyAprioriConfiguration(
            static_data_dir=FlyteDirectory("gs://my-bucket/three"),
            external_data_dir=FlyteDirectory("gs://my-bucket/four"),
        ),
        proxy_config=proxy_c,
        proxy_params=proxy_p,
    )

    @dynamic
    def dt1(a: List[MyInput]) -> List[FlyteDirectory]:
        x = []
        for aa in a:
            x.append(aa.apriori_config.external_data_dir)

        return x

    ctx = FlyteContextManager.current_context()
    cb = (
        ctx.new_builder()
        .with_serialization_settings(
            SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
        .with_execution_state(ctx.execution_state.with_params(mode=ExecutionState.Mode.TASK_EXECUTION))
    )
    with FlyteContextManager.with_context(cb) as ctx:
        input_literal_map = TypeEngine.dict_to_literal_map(
            ctx, d={"a": [my_input_gcs, my_input_gcs_2]}, type_hints={"a": List[MyInput]}
        )
        dynamic_job_spec = dt1.dispatch_execute(ctx, input_literal_map)
        assert dynamic_job_spec.literals["o0"].collection.literals[0].scalar.blob.uri == "gs://my-bucket/two"
        assert dynamic_job_spec.literals["o0"].collection.literals[1].scalar.blob.uri == "gs://my-bucket/four"
