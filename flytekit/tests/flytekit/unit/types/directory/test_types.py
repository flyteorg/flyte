import mock

from flytekit import FlyteContext
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile


def test_new_file_dir():
    fd = FlyteDirectory(path="s3://my-bucket")
    assert fd.sep == "/"
    inner_dir = fd.new_dir("test")
    assert inner_dir.path == "s3://my-bucket/test"
    fd = FlyteDirectory(path="s3://my-bucket/")
    inner_dir = fd.new_dir("test")
    assert inner_dir.path == "s3://my-bucket/test"
    f = inner_dir.new_file("test")
    assert isinstance(f, FlyteFile)
    assert f.path == "s3://my-bucket/test/test"


def test_new_remote_dir():
    fd = FlyteDirectory.new_remote()
    assert FlyteContext.current_context().file_access.raw_output_prefix in fd.path


@mock.patch("flytekit.types.directory.types.os.name", "nt")
def test_sep_nt():
    fd = FlyteDirectory(path="file://mypath")
    assert fd.sep == "\\"
    fd = FlyteDirectory(path="s3://mypath")
    assert fd.sep == "/"
