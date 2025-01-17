import os
import pathlib
import shutil
import tempfile
from base64 import b64encode

import fsspec
import pytest
from fsspec.implementations.http import HTTPFileSystem

from flytekit.configuration import Config
from flytekit.core.data_persistence import FileAccessProvider
from flytekit.remote.remote import FlyteRemote
from flytekit.remote.remote_fs import FlyteFS

local = fsspec.filesystem("file")


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


def test_basics():
    r = FlyteRemote(
        Config.for_sandbox(),
        default_project="flytesnacks",
        default_domain="development",
        data_upload_location="flyte://dv1/",
    )
    fs = FlyteFS(remote=r)
    assert fs.protocol == "flyte"
    assert fs.sep == "/"
    assert fs.unstrip_protocol("dv/fwu11/") == "flyte://dv/fwu11/"


@pytest.fixture
def sandbox_remote():
    r = FlyteRemote(
        Config.for_sandbox(),
        default_project="flytesnacks",
        default_domain="development",
        data_upload_location="flyte://data",
    )
    yield r


@pytest.mark.sandbox_test
def test_upl(sandbox_remote):
    encoded_md5 = b64encode(b"hi2dfsfj23333ileksa")
    resp = sandbox_remote.client.get_upload_signed_url(
        "flytesnacks", "development", content_md5=encoded_md5, filename="parent/child/1"
    )
    assert (
        resp.native_url
        == "s3://my-s3-bucket/flytesnacks/development/MFDWW6K2I5NHUWTNN54U26SNPJGTE3DTLJLXI6SZKE6T2===/parent/child/1"
    )


@pytest.mark.sandbox_test
def test_remote_upload_with_fs_directly(sandbox_remote, source_folder):
    fs = FlyteFS(remote=sandbox_remote)

    # Test uploading a folder, but without the /
    source_folder = source_folder.rstrip("/")
    res = fs.put(source_folder, "flyte://data", recursive=True)
    # hash of the structure of local folder. if source/ changed, will need to update hash
    assert res == "s3://my-s3-bucket/flytesnacks/development/GSEYDOSFXWFB5ABZB6AHZ2HK7Y======/source"

    # Test uploading a file
    res = fs.put(__file__, "flyte://data")
    assert res.startswith("s3://my-s3-bucket/flytesnacks/development")
    assert res.endswith("test_fs_remote.py")


@pytest.mark.sandbox_test
def test_fs_direct_trailing_slash(sandbox_remote):
    fs = FlyteFS(remote=sandbox_remote)

    with tempfile.TemporaryDirectory() as tmpdir:
        temp_dir = pathlib.Path(tmpdir)
        file_name = temp_dir / "test.txt"
        file_name.write_text("bla bla bla")

        # Uploading folder with a / won't include the folder name
        res = fs.put(str(temp_dir), "flyte://data", recursive=True)
        assert (
            res
            == f"s3://my-s3-bucket/flytesnacks/development/SUX2NK32ZQNO7F7DQMWYBVZQVM======/{temp_dir.name}/test.txt"
        )


@pytest.mark.sandbox_test
def test_remote_upload_with_data_persistence(sandbox_remote):
    sandbox_path = tempfile.mkdtemp()
    fp = FileAccessProvider(local_sandbox_dir=sandbox_path, raw_output_prefix="flyte://data/")

    with tempfile.NamedTemporaryFile("w", delete=False) as f:
        f.write("asdf")
        f.flush()
        # Test uploading a file and folder.
        res = fp.put(f.name, "flyte://data/", recursive=True)
        # Unlike using the RemoteFS directly, the trailing slash is automatically added by data persistence,
        # not sure why but preserving the behavior for now.
        only_file = pathlib.Path(f.name).name
        assert res == f"s3://my-s3-bucket/flytesnacks/development/O55F24U7RMLDOUI3LZ6SL4ZBEI======/{only_file}"


def test_common_matching():
    urls = [
        "s3://my-s3-bucket/flytesnacks/development/ABCYZWMPACZAJ2MABGMOZ6CCPY======/source/empty.md",
        "s3://my-s3-bucket/flytesnacks/development/ABCXKL5ZZWXY3PDLM3OONUHHME======/source/nested/more.txt",
        "s3://my-s3-bucket/flytesnacks/development/ABCXBAPBKONMADXVW5Q3J6YBWM======/source/original.txt",
    ]

    assert FlyteFS.extract_common(urls) == "s3://my-s3-bucket/flytesnacks/development"


def test_hashing(sandbox_remote, source_folder):
    fs = FlyteFS(remote=sandbox_remote)
    s = fs.get_hashes_and_lengths(pathlib.Path(source_folder))
    assert len(s) == 2
    lengths = set([x[1] for x in s.values()])
    assert lengths == {0, 14}
    fr = fs.get_filename_root(s)
    assert fr == "GSEYDOSFXWFB5ABZB6AHZ2HK7Y======"


def test_get_output():
    # just testing that https fs is doing what we expect
    temp_dir = tempfile.mkdtemp()

    spec = HTTPFileSystem(fsspec.filesystem("https"))
    from_path = "https://raw.githubusercontent.com/flyteorg/flytekit/master/setup.py"
    to_path = os.path.join(temp_dir, "copied.py")
    spec.get(from_path, to_path)
    res = pathlib.Path(to_path)
    assert res.is_file()
    assert not res.is_dir()
