from __future__ import annotations

import base64
import hashlib
import os
import pathlib
import random
import threading
import typing
from base64 import b64encode
from uuid import UUID

import fsspec
import requests
from flyteidl.service.dataproxy_pb2 import CreateUploadLocationResponse
from fsspec.callbacks import NoOpCallback
from fsspec.implementations.http import HTTPFileSystem
from fsspec.utils import get_protocol

from flytekit.loggers import logger
from flytekit.tools.script_mode import hash_file

if typing.TYPE_CHECKING:
    from flytekit.remote.remote import FlyteRemote

_DEFAULT_CALLBACK = NoOpCallback()
_PREFIX_KEY = "upload_prefix"
_HASHES_KEY = "hashes"
# This file system is not really a filesystem, so users aren't really able to specify the remote path,
# at least not yet.
REMOTE_PLACEHOLDER = "flyte://data"

HashStructure = typing.Dict[str, typing.Tuple[bytes, int]]


class FlytePathResolver:
    protocol = "flyte://"
    _flyte_path_to_remote_map: typing.Dict[str, str] = {}
    _lock = threading.Lock()

    @classmethod
    def resolve_remote_path(cls, flyte_uri: str) -> typing.Optional[str]:
        """
        Given a flyte uri, return the remote path if it exists or was created in current session, otherwise return None
        """
        with cls._lock:
            if flyte_uri in cls._flyte_path_to_remote_map:
                return cls._flyte_path_to_remote_map[flyte_uri]
            return None

    @classmethod
    def add_mapping(cls, flyte_uri: str, remote_path: str):
        """
        Thread safe method to dd a mapping from a flyte uri to a remote path
        """
        with cls._lock:
            cls._flyte_path_to_remote_map[flyte_uri] = remote_path


class HttpFileWriter(fsspec.spec.AbstractBufferedFile):
    def __init__(self, remote: FlyteRemote, filename: str, **kwargs):
        super().__init__(**kwargs)
        self._remote = remote
        self._filename = filename

    def _upload_chunk(self, final=False):
        """Only uploads the file at once from the buffer.
        Not suitable for large files as the buffer will blow the memory for very large files.
        Suitable for default values or local dataframes being uploaded all at once.
        """
        if final is False:
            return False
        self.buffer.seek(0)
        data = self.buffer.read()

        # h = hashlib.md5()
        # h.update(data)
        # md5 = h.digest()
        # l = len(data)
        #
        # headers = {"Content-Length": str(l), "Content-MD5": md5}

        try:
            res = self._remote.client.get_upload_signed_url(
                self._remote.default_project,
                self._remote.default_domain,
                None,
                None,
                filename_root=self._filename,
            )
            FlytePathResolver.add_mapping(self.path, res.native_url)
            resp = requests.put(res.signed_url, data=data)
            if not resp.ok:
                raise AssertionError(f"Failed to upload file {self._filename} to {res.signed_url} reason {resp.reason}")
        except Exception as e:
            raise AssertionError(f"Failed to upload file {self._filename} reason {e}")


def get_flyte_fs(remote: FlyteRemote) -> typing.Type[FlyteFS]:
    class _FlyteFS(FlyteFS):
        def __init__(self, **storage_options):
            super().__init__(remote=remote, **storage_options)

    return _FlyteFS


class FlyteFS(HTTPFileSystem):
    """
    Want this to behave mostly just like the HTTP file system.
    """

    sep = "/"
    protocol = "flyte"

    def __init__(
        self,
        remote: FlyteRemote,
        asynchronous: bool = False,
        **storage_options,
    ):
        super().__init__(asynchronous=asynchronous, **storage_options)
        self._remote = remote

    @property
    def fsid(self) -> str:
        return "flyte"

    async def _get_file(self, rpath, lpath, **kwargs):
        """
        Don't do anything special. If it's a flyte url, the create a download link and write to lpath,
        otherwise default to parent.
        """
        raise NotImplementedError("FlyteFS currently doesn't support downloading files.")

    def get_upload_link(
        self,
        local_file_path: str,
        remote_file_part: str,
        prefix: str,
        hashes: HashStructure,
    ) -> typing.Tuple[CreateUploadLocationResponse, int, bytes]:
        if not pathlib.Path(local_file_path).exists():
            raise AssertionError(f"File {local_file_path} does not exist")

        p = pathlib.Path(typing.cast(str, local_file_path))
        k = str(p.absolute())
        if k in hashes:
            md5_bytes, content_length = hashes[k]
        else:
            raise AssertionError(f"File {local_file_path} not found in hashes")
        upload_response = self._remote.client.get_upload_signed_url(
            self._remote.default_project,
            self._remote.default_domain,
            md5_bytes,
            remote_file_part,
            filename_root=prefix,
        )
        logger.debug(f"Resolved signed url {local_file_path} to {upload_response.native_url}")
        return upload_response, content_length, md5_bytes

    async def _put_file(
        self,
        lpath,
        rpath,
        chunk_size=5 * 2**20,
        callback=_DEFAULT_CALLBACK,
        method="put",
        **kwargs,
    ):
        """
        fsspec will call this method to upload a file. If recursive, rpath will already be individual files.
        Make the request and upload, but then how do we get the s3 paths back to the user?
        """
        # remove from kwargs otherwise super() call will fail
        p = kwargs.pop(_PREFIX_KEY)
        hashes = kwargs.pop(_HASHES_KEY)
        # Parse rpath, strip out everything that doesn't make sense.
        rpath = rpath.replace(f"{REMOTE_PLACEHOLDER}/", "", 1)
        resp, content_length, md5_bytes = self.get_upload_link(lpath, rpath, p, hashes)

        headers = {"Content-Length": str(content_length), "Content-MD5": b64encode(md5_bytes).decode("utf-8")}
        kwargs["headers"] = headers
        rpath = resp.signed_url
        FlytePathResolver.add_mapping(rpath, resp.native_url)
        logger.debug(f"Writing {lpath} to {rpath}")
        await super()._put_file(lpath, rpath, chunk_size, callback=callback, method=method, **kwargs)
        return resp.native_url

    @staticmethod
    def extract_common(native_urls: typing.List[str]) -> str:
        """
        This function that will take a list of strings and return the longest prefix that they all have in common.
        That is, if you have
            ['s3://my-s3-bucket/flytesnacks/development/ABCYZWMPACZAJ2MABGMOZ6CCPY======/source/empty.md',
             's3://my-s3-bucket/flytesnacks/development/ABCXKL5ZZWXY3PDLM3OONUHHME======/source/nested/more.txt',
             's3://my-s3-bucket/flytesnacks/development/ABCXBAPBKONMADXVW5Q3J6YBWM======/source/original.txt']
        this will return back 's3://my-s3-bucket/flytesnacks/development/'
        Note that trailing characters after a separator that just happen to be the same will also be stripped.
        """
        if len(native_urls) == 0:
            return ""
        if len(native_urls) == 1:
            return native_urls[0]

        common_prefix = ""
        shortest = min([len(x) for x in native_urls])
        x = [[native_urls[j][i] for j in range(len(native_urls))] for i in range(shortest)]
        for i in x:
            if len(set(i)) == 1:
                common_prefix += i[0]
            else:
                break

        fs = fsspec.filesystem(get_protocol(native_urls[0]))
        sep = fs.sep
        # split the common prefix on the last separator so we don't get any trailing characters.
        common_prefix = common_prefix.rsplit(sep, 1)[0]
        logger.debug(f"Returning {common_prefix} from {native_urls}")
        return common_prefix

    def get_hashes_and_lengths(self, p: pathlib.Path) -> HashStructure:
        """
        Returns a flat list of absolute file paths to their hashes and content lengths
        this output is used both for the file upload request, and to create consistently a filename root for
        uploaded folders. We'll also use it for single files just for consistency.
        If a directory then all the files in the directory will be hashed.
        If a single file then just that file will be hashed.
        Skip symlinks
        """
        if p.is_symlink():
            return {}
        if p.is_dir():
            hashes = {}
            for f in p.iterdir():
                hashes.update(self.get_hashes_and_lengths(f))
            return hashes
        else:
            md5_bytes, _, content_length = hash_file(p.resolve())
            return {str(p.absolute()): (md5_bytes, content_length)}

    @staticmethod
    def get_filename_root(file_info: HashStructure) -> str:
        """
        Given a dictionary of file paths to hashes and content lengths, return a consistent filename root.
        This is done by hashing the sorted list of file paths and then base32 encoding the result.
        If the input is empty, then generate a random string
        """
        if len(file_info) == 0:
            return UUID(int=random.getrandbits(128)).hex
        sorted_paths = sorted(file_info.keys())
        h = hashlib.md5()
        for p in sorted_paths:
            h.update(file_info[p][0])
        return base64.b32encode(h.digest()).decode("utf-8")

    async def _put(
        self,
        lpath,
        rpath,
        recursive=False,
        callback=_DEFAULT_CALLBACK,
        batch_size=None,
        **kwargs,
    ):
        """
        cp file.txt flyte://data/...
        rpath gets ignored, so it doesn't matter what it is.
        """
        if rpath != REMOTE_PLACEHOLDER:
            logger.debug(f"FlyteFS doesn't yet support specifying full remote path, ignoring {rpath}")

        # Hash everything at the top level
        file_info = self.get_hashes_and_lengths(pathlib.Path(lpath))
        prefix = self.get_filename_root(file_info)

        kwargs[_PREFIX_KEY] = prefix
        kwargs[_HASHES_KEY] = file_info
        res = await super()._put(lpath, REMOTE_PLACEHOLDER, recursive, callback, batch_size, **kwargs)
        if isinstance(res, list):
            res = self.extract_common(res)
        return res

    async def _isdir(self, path):
        return True

    def exists(self, path, **kwargs):
        raise NotImplementedError("flyte file system currently can't check if a file exists.")

    def _open(
        self,
        path,
        mode="wb",
        block_size=None,
        autocommit=None,  # XXX: This differs from the base class.
        cache_type=None,
        cache_options=None,
        size=None,
        **kwargs,
    ):
        if mode != "wb":
            raise ValueError("Only wb mode is supported")

        # Dataframes are written as multiple files, default is the first file with 00000 suffix, we should drop
        # that suffix and use the parent directory as the remote path.

        return HttpFileWriter(
            self._remote, os.path.basename(path), fs=self, path=os.path.dirname(path), mode=mode, **kwargs
        )

    def __str__(self):
        p = super().__str__()
        return f"FlyteFS({self._remote}): {p}"
