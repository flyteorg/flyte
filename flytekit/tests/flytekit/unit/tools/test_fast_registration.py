import os
import subprocess
import tarfile

import pytest

from flytekit.tools.fast_registration import (
    FAST_FILEENDING,
    FAST_PREFIX,
    compute_digest,
    fast_package,
    get_additional_distribution_loc,
)
from flytekit.tools.ignore import DockerIgnore, GitIgnore, IgnoreGroup, StandardIgnore
from tests.flytekit.unit.tools.test_ignore import make_tree


@pytest.fixture
def flyte_project(tmp_path):
    tree = {
        "data": {"large.file": "", "more.files": ""},
        "src": {
            "workflows": {
                "__pycache__": {"some.pyc": ""},
                "hello_world.py": "print('Hello World!')",
            },
        },
        "utils": {
            "util.py": "print('Hello from utils!')",
        },
        ".venv": {"lots": "", "of": "", "packages": ""},
        ".env": "supersecret",
        "some.bar": "",
        "some.foo": "",
        "keep.foo": "",
        ".gitignore": "\n".join([".env", ".venv", "# A comment", "data", "*.foo", "!keep.foo"]),
        ".dockerignore": "\n".join(["data", "*.bar", ".git"]),
    }

    make_tree(tmp_path, tree)
    os.symlink(str(tmp_path) + "/utils/util.py", str(tmp_path) + "/src/util")
    subprocess.run(["git", "init", str(tmp_path)])
    return tmp_path


def test_package(flyte_project, tmp_path):
    archive_fname = fast_package(source=flyte_project, output_dir=tmp_path)
    with tarfile.open(archive_fname) as tar:
        assert sorted(tar.getnames()) == [
            ".dockerignore",
            ".gitignore",
            "keep.foo",
            "src",
            "src/util",
            "src/workflows",
            "src/workflows/hello_world.py",
            "utils",
            "utils/util.py",
        ]
        util = tar.getmember("src/util")
        assert util.issym()
    assert str(os.path.basename(archive_fname)).startswith(FAST_PREFIX)
    assert str(archive_fname).endswith(FAST_FILEENDING)


def test_package_with_symlink(flyte_project, tmp_path):
    archive_fname = fast_package(source=flyte_project / "src", output_dir=tmp_path, deref_symlinks=True)
    with tarfile.open(archive_fname, dereference=True) as tar:
        assert sorted(tar.getnames()) == [
            "util",
            "workflows",
            "workflows/hello_world.py",
        ]
        util = tar.getmember("util")
        assert util.isfile()
    assert str(os.path.basename(archive_fname)).startswith(FAST_PREFIX)
    assert str(archive_fname).endswith(FAST_FILEENDING)


def test_digest_ignore(flyte_project):
    ignore = IgnoreGroup(flyte_project, [GitIgnore, DockerIgnore, StandardIgnore])
    digest1 = compute_digest(flyte_project, ignore.is_ignored)

    change_file = flyte_project / "data" / "large.file"
    assert ignore.is_ignored(change_file)
    change_file.write_text("I don't matter")

    digest2 = compute_digest(flyte_project, ignore.is_ignored)
    assert digest1 == digest2


def test_digest_change(flyte_project):
    ignore = IgnoreGroup(flyte_project, [GitIgnore, DockerIgnore, StandardIgnore])
    digest1 = compute_digest(flyte_project, ignore.is_ignored)

    change_file = flyte_project / "src" / "workflows" / "hello_world.py"
    assert not ignore.is_ignored(change_file)
    change_file.write_text("print('I do matter!')")

    digest2 = compute_digest(flyte_project, ignore.is_ignored)
    assert digest1 != digest2


def test_get_additional_distribution_loc():
    assert get_additional_distribution_loc("s3://my-s3-bucket/dir", "123abc") == "s3://my-s3-bucket/dir/123abc.tar.gz"
