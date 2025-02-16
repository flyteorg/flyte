from __future__ import annotations

import gzip
import hashlib
import os
import posixpath
import subprocess as _subprocess
import tarfile
import tempfile
import typing
from typing import Optional

import click

from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.utils import timeit
from flytekit.tools.ignore import DockerIgnore, GitIgnore, IgnoreGroup, StandardIgnore
from flytekit.tools.script_mode import tar_strip_file_attributes

FAST_PREFIX = "fast"
FAST_FILEENDING = ".tar.gz"


def fast_package(source: os.PathLike, output_dir: os.PathLike, deref_symlinks: bool = False) -> os.PathLike:
    """
    Takes a source directory and packages everything not covered by common ignores into a tarball
    named after a hexdigest of the included files.
    :param os.PathLike source:
    :param os.PathLike output_dir:
    :param bool deref_symlinks: Enables dereferencing symlinks when packaging directory
    :return os.PathLike:
    """
    ignore = IgnoreGroup(source, [GitIgnore, DockerIgnore, StandardIgnore])
    digest = compute_digest(source, ignore.is_ignored)
    archive_fname = f"{FAST_PREFIX}{digest}{FAST_FILEENDING}"

    if output_dir is None:
        output_dir = tempfile.mkdtemp()
        click.secho(f"No output path provided, using a temporary directory at {output_dir} instead", fg="yellow")

    archive_fname = os.path.join(output_dir, archive_fname)

    with tempfile.TemporaryDirectory() as tmp_dir:
        tar_path = os.path.join(tmp_dir, "tmp.tar")
        with tarfile.open(tar_path, "w", dereference=deref_symlinks) as tar:
            files: typing.List[str] = os.listdir(source)
            for ws_file in files:
                tar.add(
                    os.path.join(source, ws_file),
                    arcname=ws_file,
                    filter=lambda x: ignore.tar_filter(tar_strip_file_attributes(x)),
                )
        with gzip.GzipFile(filename=archive_fname, mode="wb", mtime=0) as gzipped:
            with open(tar_path, "rb") as tar_file:
                gzipped.write(tar_file.read())

    return archive_fname


def compute_digest(source: os.PathLike, filter: Optional[callable] = None) -> str:
    """
    Walks the entirety of the source dir to compute a deterministic md5 hex digest of the dir contents.
    :param os.PathLike source:
    :param Ignore ignore:
    :return Text:
    """
    hasher = hashlib.md5()
    for root, _, files in os.walk(source, topdown=True):
        files.sort()

        for fname in files:
            abspath = os.path.join(root, fname)
            relpath = os.path.relpath(abspath, source)
            if filter:
                if filter(relpath):
                    continue

            _filehash_update(abspath, hasher)
            _pathhash_update(relpath, hasher)

    return hasher.hexdigest()


def _filehash_update(path: os.PathLike, hasher: hashlib._Hash) -> None:
    blocksize = 65536
    with open(path, "rb") as f:
        bytes = f.read(blocksize)
        while bytes:
            hasher.update(bytes)
            bytes = f.read(blocksize)


def _pathhash_update(path: os.PathLike, hasher: hashlib._Hash) -> None:
    path_list = path.split(os.sep)
    hasher.update("".join(path_list).encode("utf-8"))


def get_additional_distribution_loc(remote_location: str, identifier: str) -> str:
    """
    :param Text remote_location:
    :param Text identifier:
    :return Text:
    """
    return posixpath.join(remote_location, "{}.{}".format(identifier, "tar.gz"))


@timeit("Download distribution")
def download_distribution(additional_distribution: str, destination: str):
    """
    Downloads a remote code distribution and overwrites any local files.
    :param Text additional_distribution:
    :param os.PathLike destination:
    """
    if not os.path.isdir(destination):
        raise ValueError("Destination path is required to download distribution and it should be a directory")
    # NOTE the os.path.join(destination, ''). This is to ensure that the given path is in fact a directory and all
    # downloaded data should be copied into this directory. We do this to account for a difference in behavior in
    # fsspec, which requires a trailing slash in case of pre-existing directory.
    FlyteContextManager.current_context().file_access.get_data(additional_distribution, os.path.join(destination, ""))
    tarfile_name = os.path.basename(additional_distribution)
    if not tarfile_name.endswith(".tar.gz"):
        raise RuntimeError("Unrecognized additional distribution format for {}".format(additional_distribution))

    # This will overwrite the existing user flyte workflow code in the current working code dir.
    result = _subprocess.run(
        ["tar", "-xvf", os.path.join(destination, tarfile_name), "-C", destination],
        stdout=_subprocess.PIPE,
    )
    result.check_returncode()
