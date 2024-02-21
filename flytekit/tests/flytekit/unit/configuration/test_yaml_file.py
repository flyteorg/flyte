import os

import mock

from flytekit.configuration import ConfigEntry, get_config_file
from flytekit.configuration.file import LegacyConfigEntry, YamlConfigEntry
from flytekit.configuration.internal import AWS, Credentials, Images, Platform


def test_config_entry_file():
    c = ConfigEntry(
        LegacyConfigEntry("platform", "url", str), YamlConfigEntry("admin.endpoint"), lambda x: x.replace("dns:///", "")
    )
    assert c.read() is None

    cfg = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/sample.yaml"))
    assert cfg.yaml_config is not None
    assert c.read(cfg) == "flyte.mycorp.io"

    c = ConfigEntry(LegacyConfigEntry("platform", "url2", str))  # Does not exist
    assert c.read(cfg) is None


def test_config_entry_file_normal():
    # Most yaml config files will not have images, make sure that a normal one without an image section doesn't
    # return None
    cfg = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/no_images.yaml"))
    images_dict = Images.get_specified_images(cfg)
    assert images_dict == {}
    assert cfg.yaml_config is not None


@mock.patch("flytekit.configuration.file.getenv")
def test_config_entry_file_2(mock_get):
    # Test reading of the environment variable that flytectl asks users to set.
    # Can take both extensions
    sample_yaml_file_name = os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/sample.yml")

    mock_get.return_value = sample_yaml_file_name

    c = ConfigEntry(
        LegacyConfigEntry("platform", "url", str), YamlConfigEntry("admin.endpoint"), lambda x: x.replace("dns:///", "")
    )
    assert c.read() is None

    cfg = get_config_file(sample_yaml_file_name)
    assert c.read(cfg) == "flyte.mycorp.io"
    assert cfg.yaml_config is not None

    c = ConfigEntry(LegacyConfigEntry("platform", "url2", str))  # Does not exist
    assert c.read(cfg) is None


def test_real_config():
    config_file = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/sample.yaml"))
    res = Platform.INSECURE.read(config_file)
    assert res

    res = Platform.URL.read(config_file)
    assert res == "flyte.mycorp.io"

    res = AWS.S3_ACCESS_KEY_ID.read(config_file)
    assert res == "minio"

    res = AWS.S3_ENDPOINT.read(config_file)
    assert res == "http://localhost:30084"

    res = AWS.S3_SECRET_ACCESS_KEY.read(config_file)
    assert res == "miniostorage"

    res = Credentials.SCOPES.read(config_file)
    assert res == ["all"]


def test_use_ssl():
    config_file = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/nossl.yaml"))
    res = Platform.INSECURE.read(config_file)
    assert res is False
