import os
import random
import shutil
import tempfile
from uuid import UUID

import fsspec
import mock
import pytest
from s3fs import S3FileSystem

from flytekit.configuration import Config, DataConfig, S3Config
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.data_persistence import FileAccessProvider, get_fsspec_storage_options, s3_setup_args
from flytekit.types.directory.types import FlyteDirectory

local = fsspec.filesystem("file")
root = os.path.abspath(os.sep)


@mock.patch("google.auth.compute_engine._metadata")  # to prevent network calls
@mock.patch("flytekit.core.data_persistence.UUID")
def test_path_getting(mock_uuid_class, mock_gcs):
    mock_uuid_class.return_value.hex = "abcdef123"

    # Testing with raw output prefix pointing to a local path
    loc_sandbox = os.path.join(root, "tmp", "unittest")
    loc_data = os.path.join(root, "tmp", "unittestdata")
    local_raw_fp = FileAccessProvider(local_sandbox_dir=loc_sandbox, raw_output_prefix=loc_data)
    r = local_raw_fp.get_random_string()
    rr = local_raw_fp.join(local_raw_fp.raw_output_prefix, r)
    assert rr == os.path.join(root, "tmp", "unittestdata", "abcdef123")
    rr = local_raw_fp.join(local_raw_fp.raw_output_prefix, r, local_raw_fp.get_file_tail("/fsa/blah.csv"))
    assert rr == os.path.join(root, "tmp", "unittestdata", "abcdef123", "blah.csv")

    # Test local path and directory
    assert local_raw_fp.get_random_local_path() == os.path.join(root, "tmp", "unittest", "local_flytekit", "abcdef123")
    assert local_raw_fp.get_random_local_path("xjiosa/blah.txt") == os.path.join(
        root, "tmp", "unittest", "local_flytekit", "abcdef123", "blah.txt"
    )
    assert local_raw_fp.get_random_local_directory() == os.path.join(
        root, "tmp", "unittest", "local_flytekit", "abcdef123"
    )

    # Recursive paths
    assert "file:///abc/happy/", "s3://my-s3-bucket/bucket1/" == local_raw_fp.recursive_paths(
        "file:///abc/happy/", "s3://my-s3-bucket/bucket1/"
    )
    assert "file:///abc/happy/", "s3://my-s3-bucket/bucket1/" == local_raw_fp.recursive_paths(
        "file:///abc/happy", "s3://my-s3-bucket/bucket1"
    )

    # Test with remote pointed to s3.
    s3_fa = FileAccessProvider(local_sandbox_dir=loc_sandbox, raw_output_prefix="s3://my-s3-bucket")
    r = s3_fa.get_random_string()
    rr = s3_fa.join(s3_fa.raw_output_prefix, r)
    assert rr == "s3://my-s3-bucket/abcdef123"
    # trailing slash should make no difference
    s3_fa = FileAccessProvider(local_sandbox_dir=loc_sandbox, raw_output_prefix="s3://my-s3-bucket/")
    r = s3_fa.get_random_string()
    rr = s3_fa.join(s3_fa.raw_output_prefix, r)
    assert rr == "s3://my-s3-bucket/abcdef123"

    # Testing with raw output prefix pointing to file://
    # Skip tests for windows
    if os.name != "nt":
        file_raw_fp = FileAccessProvider(local_sandbox_dir=loc_sandbox, raw_output_prefix="file:///tmp/unittestdata")
        r = file_raw_fp.get_random_string()
        rr = file_raw_fp.join(file_raw_fp.raw_output_prefix, r)
        rr = file_raw_fp.strip_file_header(rr)
        assert rr == os.path.join(root, "tmp", "unittestdata", "abcdef123")
        r = file_raw_fp.get_random_string()
        rr = file_raw_fp.join(file_raw_fp.raw_output_prefix, r, file_raw_fp.get_file_tail("/fsa/blah.csv"))
        rr = file_raw_fp.strip_file_header(rr)
        assert rr == os.path.join(root, "tmp", "unittestdata", "abcdef123", "blah.csv")

    g_fa = FileAccessProvider(local_sandbox_dir=loc_sandbox, raw_output_prefix="gs://my-s3-bucket/")
    r = g_fa.get_random_string()
    rr = g_fa.join(g_fa.raw_output_prefix, r)
    assert rr == "gs://my-s3-bucket/abcdef123"


@pytest.fixture
def source_folder():
    # Set up source directory for testing
    parent_temp = tempfile.mkdtemp()
    src_dir = os.path.join(parent_temp, "source", "")
    nested_dir = os.path.join(src_dir, "nested")
    local.mkdir(nested_dir)
    local.touch(os.path.join(src_dir, "original.txt"))
    with open(os.path.join(src_dir, "original.txt"), "w") as fh:
        fh.write("hello original")
    local.touch(os.path.join(nested_dir, "more.txt"))
    yield src_dir
    shutil.rmtree(parent_temp)


def test_local_raw_fsspec(source_folder):
    # Test copying using raw fsspec local filesystem, should not create a nested folder
    with tempfile.TemporaryDirectory() as dest_tmpdir:
        local.put(source_folder, dest_tmpdir, recursive=True)

    new_temp_dir_2 = tempfile.mkdtemp()
    new_temp_dir_2 = os.path.join(new_temp_dir_2, "doesnotexist")
    local.put(source_folder, new_temp_dir_2, recursive=True)
    files = local.find(new_temp_dir_2)
    assert len(files) == 2


def test_local_provider(source_folder):
    # Test that behavior putting from a local dir to a local remote dir is the same whether or not the local
    # dest folder exists.
    dc = Config.for_sandbox().data_config
    with tempfile.TemporaryDirectory() as dest_tmpdir:
        provider = FileAccessProvider(local_sandbox_dir="/tmp/unittest", raw_output_prefix=dest_tmpdir, data_config=dc)
        r = provider.get_random_string()
        doesnotexist = provider.join(provider.raw_output_prefix, r)
        provider.put_data(source_folder, doesnotexist, is_multipart=True)
        files = provider.raw_output_fs.find(doesnotexist)
        assert len(files) == 2

        r = provider.get_random_string()
        exists = provider.join(provider.raw_output_prefix, r)
        provider.raw_output_fs.mkdir(exists)
        provider.put_data(source_folder, exists, is_multipart=True)
        files = provider.raw_output_fs.find(exists)
        assert len(files) == 2


def test_async_file_system():
    remote_path = "test:///tmp/test.py"
    local_path = "test.py"

    class MockAsyncFileSystem(S3FileSystem):
        def __init__(self, *args, **kwargs):
            super().__init__(args, kwargs)

        async def _put_file(self, *args, **kwargs):
            # s3fs._put_file returns None as well
            return None

        async def _get_file(self, *args, **kwargs):
            # s3fs._get_file returns None as well
            return None

        async def _lsdir(
            self,
            path,
            refresh=False,
            max_items=None,
            delimiter="/",
            prefix="",
            versions=False,
        ):
            return False

    fsspec.register_implementation("test", MockAsyncFileSystem)

    ctx = FlyteContextManager.current_context()
    dst = ctx.file_access.put(local_path, remote_path)
    assert dst == remote_path
    dst = ctx.file_access.get(remote_path, local_path)
    assert dst == local_path


@pytest.mark.sandbox_test
def test_s3_provider(source_folder):
    # Running mkdir on s3 filesystem doesn't do anything so leaving out for now
    dc = Config.for_sandbox().data_config
    provider = FileAccessProvider(
        local_sandbox_dir="/tmp/unittest", raw_output_prefix="s3://my-s3-bucket/testdata/", data_config=dc
    )
    doesnotexist = provider.join(provider.raw_output_prefix, provider.get_random_string())
    provider.put_data(source_folder, doesnotexist, is_multipart=True)
    fs = provider.get_filesystem_for_path(doesnotexist)
    files = fs.find(doesnotexist)
    assert len(files) == 2


def test_local_provider_get_empty():
    dc = Config.for_sandbox().data_config
    with tempfile.TemporaryDirectory() as empty_source:
        with tempfile.TemporaryDirectory() as dest_folder:
            provider = FileAccessProvider(
                local_sandbox_dir="/tmp/unittest", raw_output_prefix=empty_source, data_config=dc
            )
            provider.get_data(empty_source, dest_folder, is_multipart=True)
            loc = provider.get_filesystem_for_path(dest_folder)
            src_files = loc.find(empty_source)
            assert len(src_files) == 0
            dest_files = loc.find(dest_folder)
            assert len(dest_files) == 0


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_s3_setup_args_env_empty(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    mock_os.get.return_value = None
    s3c = S3Config.auto()
    kwargs = s3_setup_args(s3c)
    assert kwargs == {"cache_regions": True}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_s3_setup_args_env_both(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "AWS_ACCESS_KEY_ID": "ignore-user",
        "AWS_SECRET_ACCESS_KEY": "ignore-secret",
        "FLYTE_AWS_ACCESS_KEY_ID": "flyte",
        "FLYTE_AWS_SECRET_ACCESS_KEY": "flyte-secret",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    kwargs = s3_setup_args(S3Config.auto())
    assert kwargs == {"key": "flyte", "secret": "flyte-secret", "cache_regions": True}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_s3_setup_args_env_flyte(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "FLYTE_AWS_ACCESS_KEY_ID": "flyte",
        "FLYTE_AWS_SECRET_ACCESS_KEY": "flyte-secret",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    kwargs = s3_setup_args(S3Config.auto())
    assert kwargs == {"key": "flyte", "secret": "flyte-secret", "cache_regions": True}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_s3_setup_args_env_aws(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "AWS_ACCESS_KEY_ID": "ignore-user",
        "AWS_SECRET_ACCESS_KEY": "ignore-secret",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    kwargs = s3_setup_args(S3Config.auto())
    # not explicitly in kwargs, since fsspec/boto3 will use these env vars by default
    assert kwargs == {"cache_regions": True}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_get_fsspec_storage_options_gcs(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "FLYTE_GCP_GSUTIL_PARALLELISM": "False",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    storage_options = get_fsspec_storage_options("gs", DataConfig.auto())
    assert storage_options == {}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_get_fsspec_storage_options_gcs_with_overrides(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "FLYTE_GCP_GSUTIL_PARALLELISM": "False",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    storage_options = get_fsspec_storage_options("gs", DataConfig.auto(), anonymous=True, other_argument="value")
    assert storage_options == {"token": "anon", "other_argument": "value"}


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_get_fsspec_storage_options_azure(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "FLYTE_AZURE_STORAGE_ACCOUNT_NAME": "accountname",
        "FLYTE_AZURE_STORAGE_ACCOUNT_KEY": "accountkey",
        "FLYTE_AZURE_TENANT_ID": "tenantid",
        "FLYTE_AZURE_CLIENT_ID": "clientid",
        "FLYTE_AZURE_CLIENT_SECRET": "clientsecret",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    storage_options = get_fsspec_storage_options("abfs", DataConfig.auto())
    assert storage_options == {
        "account_name": "accountname",
        "account_key": "accountkey",
        "client_id": "clientid",
        "client_secret": "clientsecret",
        "tenant_id": "tenantid",
        "anon": False,
    }


@mock.patch("flytekit.configuration.get_config_file")
@mock.patch("os.environ")
def test_get_fsspec_storage_options_azure_with_overrides(mock_os, mock_get_config_file):
    mock_get_config_file.return_value = None
    ee = {
        "FLYTE_AZURE_STORAGE_ACCOUNT_NAME": "accountname",
        "FLYTE_AZURE_STORAGE_ACCOUNT_KEY": "accountkey",
    }
    mock_os.get.side_effect = lambda x, y: ee.get(x)
    storage_options = get_fsspec_storage_options(
        "abfs", DataConfig.auto(), anonymous=True, account_name="other_accountname", other_argument="value"
    )
    assert storage_options == {
        "account_name": "other_accountname",
        "account_key": "accountkey",
        "anon": True,
        "other_argument": "value",
    }


def test_crawl_local_nt(source_folder):
    """
    running this to see what it prints
    """
    if os.name != "nt":  # don't
        return
    source_folder = os.path.join(source_folder, "")  # ensure there's a trailing / or \
    fd = FlyteDirectory(path=source_folder)
    res = fd.crawl()
    split = [(x, y) for x, y in res]
    print(f"NT split {split}")

    # Test crawling a directory without trailing / or \
    source_folder = source_folder[:-1]
    fd = FlyteDirectory(path=source_folder)
    res = fd.crawl()
    files = [os.path.join(x, y) for x, y in res]
    print(f"NT files joined {files}")


def test_crawl_local_non_nt(source_folder):
    """
    crawl on the source folder fixture should return for example
        ('/var/folders/jx/54tww2ls58n8qtlp9k31nbd80000gp/T/tmpp14arygf/source/', 'original.txt')
        ('/var/folders/jx/54tww2ls58n8qtlp9k31nbd80000gp/T/tmpp14arygf/source/', 'nested/more.txt')
    """
    if os.name == "nt":  # don't
        return
    source_folder = os.path.join(source_folder, "")  # ensure there's a trailing / or \
    fd = FlyteDirectory(path=source_folder)
    res = fd.crawl()
    split = [(x, y) for x, y in res]
    files = [os.path.join(x, y) for x, y in split]
    assert set(split) == {(source_folder, "original.txt"), (source_folder, os.path.join("nested", "more.txt"))}
    expected = {os.path.join(source_folder, "original.txt"), os.path.join(source_folder, "nested", "more.txt")}
    assert set(files) == expected

    # Test crawling a directory without trailing / or \
    source_folder = source_folder[:-1]
    fd = FlyteDirectory(path=source_folder)
    res = fd.crawl()
    files = [os.path.join(x, y) for x, y in res]
    assert set(files) == expected

    # Test crawling a single file
    fd = FlyteDirectory(path=os.path.join(source_folder, "original1.txt"))
    res = fd.crawl()
    files = [os.path.join(x, y) for x, y in res]
    assert len(files) == 0


@pytest.mark.sandbox_test
def test_crawl_s3(source_folder):
    """
    ('s3://my-s3-bucket/testdata/5b31492c032893b515650f8c76008cf7', 'original.txt')
    ('s3://my-s3-bucket/testdata/5b31492c032893b515650f8c76008cf7', 'nested/more.txt')
    """
    # Running mkdir on s3 filesystem doesn't do anything so leaving out for now
    dc = Config.for_sandbox().data_config
    provider = FileAccessProvider(
        local_sandbox_dir="/tmp/unittest", raw_output_prefix="s3://my-s3-bucket/testdata/", data_config=dc
    )
    s3_random_target = provider.join(provider.raw_output_prefix, provider.get_random_string())
    provider.put_data(source_folder, s3_random_target, is_multipart=True)
    ctx = FlyteContextManager.current_context()
    expected = {f"{s3_random_target}/original.txt", f"{s3_random_target}/nested/more.txt"}

    with FlyteContextManager.with_context(ctx.with_file_access(provider)):
        fd = FlyteDirectory(path=s3_random_target)
        res = fd.crawl()
        res = [(x, y) for x, y in res]
        files = [os.path.join(x, y) for x, y in res]
        assert set(files) == expected
        assert set(res) == {(s3_random_target, "original.txt"), (s3_random_target, os.path.join("nested", "more.txt"))}

        fd_file = FlyteDirectory(path=f"{s3_random_target}/original.txt")
        res = fd_file.crawl()
        files = [r for r in res]
        assert len(files) == 1


@pytest.mark.sandbox_test
def test_walk_local_copy_to_s3(source_folder):
    dc = Config.for_sandbox().data_config
    explicit_empty_folder = UUID(int=random.getrandbits(128)).hex
    raw_output_path = f"s3://my-s3-bucket/testdata/{explicit_empty_folder}"
    provider = FileAccessProvider(local_sandbox_dir="/tmp/unittest", raw_output_prefix=raw_output_path, data_config=dc)

    ctx = FlyteContextManager.current_context()
    local_fd = FlyteDirectory(path=source_folder)
    local_fd_crawl = local_fd.crawl()
    local_fd_crawl = [x for x in local_fd_crawl]
    with FlyteContextManager.with_context(ctx.with_file_access(provider)):
        fd = FlyteDirectory.new_remote()
        assert raw_output_path in fd.path

        # Write source folder files to new remote path
        for root_path, suffix in local_fd_crawl:
            new_file = fd.new_file(suffix)  # noqa
            with open(os.path.join(root_path, suffix), "rb") as r:  # noqa
                with new_file.open("w") as w:
                    print(f"Writing, t {type(w)} p {new_file.path} |{suffix}|")
                    w.write(str(r.read()))

        new_crawl = fd.crawl()
        new_suffixes = [y for x, y in new_crawl]
        assert len(new_suffixes) == 2  # should have written two files
