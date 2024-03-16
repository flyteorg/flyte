"""
======================================
:mod:`flytekit.core.data_persistence`
======================================

.. currentmodule:: flytekit.core.data_persistence

The Data persistence module is used by core flytekit and most of the core TypeTransformers to manage data fetch & store,
between the durable backend store and the runtime environment. This is designed to be a pluggable system, with a default
simple implementation that ships with the core.

.. autosummary::
   :toctree: generated/
   :template: custom.rst
   :nosignatures:

   FileAccessProvider

"""
import io
import os
import pathlib
import tempfile
import typing
from typing import Any, Dict, Optional, Union, cast
from uuid import UUID

import fsspec
from fsspec.utils import get_protocol
from typing_extensions import Unpack

from flytekit import configuration
from flytekit.configuration import DataConfig
from flytekit.core.local_fsspec import FlyteLocalFileSystem
from flytekit.core.utils import timeit
from flytekit.exceptions.user import FlyteAssertion, FlyteValueException
from flytekit.interfaces.random import random
from flytekit.loggers import logger

# Refer to https://github.com/fsspec/s3fs/blob/50bafe4d8766c3b2a4e1fc09669cf02fb2d71454/s3fs/core.py#L198
# for key and secret
_FSSPEC_S3_KEY_ID = "key"
_FSSPEC_S3_SECRET = "secret"
_ANON = "anon"

Uploadable = typing.Union[str, os.PathLike, pathlib.Path, bytes, io.BufferedReader, io.BytesIO, io.StringIO]


def s3_setup_args(s3_cfg: configuration.S3Config, anonymous: bool = False) -> Dict[str, Any]:
    kwargs: Dict[str, Any] = {
        "cache_regions": True,
    }
    if s3_cfg.access_key_id:
        kwargs[_FSSPEC_S3_KEY_ID] = s3_cfg.access_key_id

    if s3_cfg.secret_access_key:
        kwargs[_FSSPEC_S3_SECRET] = s3_cfg.secret_access_key

    # S3fs takes this as a special arg
    if s3_cfg.endpoint is not None:
        kwargs["client_kwargs"] = {"endpoint_url": s3_cfg.endpoint}

    if anonymous:
        kwargs[_ANON] = True

    return kwargs


def azure_setup_args(azure_cfg: configuration.AzureBlobStorageConfig, anonymous: bool = False) -> Dict[str, Any]:
    kwargs: Dict[str, Any] = {}

    if azure_cfg.account_name:
        kwargs["account_name"] = azure_cfg.account_name
    if azure_cfg.account_key:
        kwargs["account_key"] = azure_cfg.account_key
    if azure_cfg.client_id:
        kwargs["client_id"] = azure_cfg.client_id
    if azure_cfg.client_secret:
        kwargs["client_secret"] = azure_cfg.client_secret
    if azure_cfg.tenant_id:
        kwargs["tenant_id"] = azure_cfg.tenant_id
    kwargs[_ANON] = anonymous
    return kwargs


def get_fsspec_storage_options(
    protocol: str, data_config: typing.Optional[DataConfig] = None, anonymous: bool = False, **kwargs
) -> Dict[str, Any]:
    data_config = data_config or DataConfig.auto()

    if protocol == "file":
        return {"auto_mkdir": True, **kwargs}
    if protocol == "s3":
        return {**s3_setup_args(data_config.s3, anonymous=anonymous), **kwargs}
    if protocol == "gs":
        if anonymous:
            kwargs["token"] = _ANON
        return kwargs
    if protocol in ("abfs", "abfss"):
        return {**azure_setup_args(data_config.azure, anonymous=anonymous), **kwargs}
    return {}


class FileAccessProvider(object):
    """
    This is the class that is available through the FlyteContext and can be used for persisting data to the remote
    durable store.
    """

    def __init__(
        self,
        local_sandbox_dir: Union[str, os.PathLike],
        raw_output_prefix: str,
        data_config: typing.Optional[DataConfig] = None,
    ):
        """
        Args:
            local_sandbox_dir: A local temporary working directory, that should be used to store data
            raw_output_prefix:
            data_config:
        """
        # Local access
        if local_sandbox_dir is None or local_sandbox_dir == "":
            raise ValueError("FileAccessProvider needs to be created with a valid local_sandbox_dir")
        local_sandbox_dir_appended = os.path.join(local_sandbox_dir, "local_flytekit")
        self._local_sandbox_dir = pathlib.Path(local_sandbox_dir_appended)
        self._local_sandbox_dir.mkdir(parents=True, exist_ok=True)
        self._local = fsspec.filesystem(None)

        self._data_config = data_config if data_config else DataConfig.auto()
        self._default_protocol = get_protocol(str(raw_output_prefix))
        self._default_remote = cast(fsspec.AbstractFileSystem, self.get_filesystem(self._default_protocol))
        if os.name == "nt" and raw_output_prefix.startswith("file://"):
            raise FlyteAssertion("Cannot use the file:// prefix on Windows.")
        self._raw_output_prefix = (
            raw_output_prefix
            if raw_output_prefix.endswith(self.sep(self._default_remote))
            else raw_output_prefix + self.sep(self._default_remote)
        )

    @property
    def raw_output_prefix(self) -> str:
        return self._raw_output_prefix

    @property
    def data_config(self) -> DataConfig:
        return self._data_config

    @property
    def raw_output_fs(self) -> fsspec.AbstractFileSystem:
        """
        Returns a file system corresponding to the provided raw output prefix
        """
        return self._default_remote

    def get_filesystem(
        self, protocol: typing.Optional[str] = None, anonymous: bool = False, **kwargs
    ) -> fsspec.AbstractFileSystem:
        if not protocol:
            return self._default_remote
        if protocol == "file":
            kwargs["auto_mkdir"] = True
            return FlyteLocalFileSystem(**kwargs)
        elif protocol == "s3":
            s3kwargs = s3_setup_args(self._data_config.s3, anonymous=anonymous)
            s3kwargs.update(kwargs)
            return fsspec.filesystem(protocol, **s3kwargs)  # type: ignore
        elif protocol == "gs":
            if anonymous:
                kwargs["token"] = _ANON
            return fsspec.filesystem(protocol, **kwargs)  # type: ignore

        storage_options = get_fsspec_storage_options(
            protocol=protocol, anonymous=anonymous, data_config=self._data_config, **kwargs
        )

        return fsspec.filesystem(protocol, **storage_options)

    def get_filesystem_for_path(self, path: str = "", anonymous: bool = False, **kwargs) -> fsspec.AbstractFileSystem:
        protocol = get_protocol(path)
        return self.get_filesystem(protocol, anonymous=anonymous, **kwargs)

    @staticmethod
    def is_remote(path: Union[str, os.PathLike]) -> bool:
        """
        Deprecated. Let's find a replacement
        """
        protocol = get_protocol(str(path))
        if protocol is None:
            return False
        return protocol != "file"

    @property
    def local_sandbox_dir(self) -> os.PathLike:
        """
        This is a context based temp dir.
        """
        return self._local_sandbox_dir

    @property
    def local_access(self) -> fsspec.AbstractFileSystem:
        return self._local

    @staticmethod
    def strip_file_header(path: str, trim_trailing_sep: bool = False) -> str:
        """
        Drops file:// if it exists from the file
        """
        if path.startswith("file://"):
            return path.replace("file://", "", 1)
        return path

    @staticmethod
    def recursive_paths(f: str, t: str) -> typing.Tuple[str, str]:
        # Only apply the join if the from_path isn't already a file. But we can do this check only
        # for local files, otherwise assume it's a directory and add /'s as usual
        if get_protocol(f) == "file":
            local_fs = fsspec.filesystem("file")
            if local_fs.exists(f) and local_fs.isdir(f):
                print("Adding trailing sep to")
                f = os.path.join(f, "")
            else:
                print("Not adding trailing sep")
        else:
            f = os.path.join(f, "")
        t = os.path.join(t, "")
        return f, t

    def sep(self, file_system: typing.Optional[fsspec.AbstractFileSystem]) -> str:
        if file_system is None or file_system.protocol == "file":
            return os.sep
        if isinstance(file_system.protocol, tuple) or isinstance(file_system.protocol, list):
            if "file" in file_system.protocol:
                return os.sep
        return file_system.sep

    def exists(self, path: str) -> bool:
        try:
            file_system = self.get_filesystem_for_path(path)
            return file_system.exists(path)
        except OSError as oe:
            logger.debug(f"Error in exists checking {path} {oe}")
            anon_fs = self.get_filesystem(get_protocol(path), anonymous=True)
            if anon_fs is not None:
                logger.debug(f"Attempting anonymous exists with {anon_fs}")
                return anon_fs.exists(path)
            raise oe

    def get(self, from_path: str, to_path: str, recursive: bool = False, **kwargs):
        file_system = self.get_filesystem_for_path(from_path)
        if recursive:
            from_path, to_path = self.recursive_paths(from_path, to_path)
        try:
            if os.name == "nt" and file_system.protocol == "file" and recursive:
                import shutil

                return shutil.copytree(
                    self.strip_file_header(from_path), self.strip_file_header(to_path), dirs_exist_ok=True
                )
            logger.info(f"Getting {from_path} to {to_path}")
            dst = file_system.get(from_path, to_path, recursive=recursive, **kwargs)
            if isinstance(dst, (str, pathlib.Path)):
                return dst
            return to_path
        except OSError as oe:
            logger.debug(f"Error in getting {from_path} to {to_path} rec {recursive} {oe}")
            if not file_system.exists(from_path):
                raise FlyteValueException(from_path, "File not found")
            file_system = self.get_filesystem(get_protocol(from_path), anonymous=True)
            if file_system is not None:
                logger.debug(f"Attempting anonymous get with {file_system}")
                return file_system.get(from_path, to_path, recursive=recursive, **kwargs)
            raise oe

    def put(self, from_path: str, to_path: str, recursive: bool = False, **kwargs):
        file_system = self.get_filesystem_for_path(to_path)
        from_path = self.strip_file_header(from_path)
        if recursive:
            # Only check this for the local filesystem
            if file_system.protocol == "file" and not file_system.isdir(from_path):
                raise FlyteAssertion(f"Source path {from_path} is not a directory")
            if os.name == "nt" and file_system.protocol == "file":
                import shutil

                return shutil.copytree(
                    self.strip_file_header(from_path), self.strip_file_header(to_path), dirs_exist_ok=True
                )
            from_path, to_path = self.recursive_paths(from_path, to_path)
        dst = file_system.put(from_path, to_path, recursive=recursive, **kwargs)
        if isinstance(dst, (str, pathlib.Path)):
            return dst
        else:
            return to_path

    def put_raw_data(
        self,
        lpath: Uploadable,
        upload_prefix: Optional[str] = None,
        file_name: Optional[str] = None,
        read_chunk_size_bytes: int = 1024,
        encoding: str = "utf-8",
        **kwargs,
    ) -> str:
        """
        This is a more flexible version of put that accepts a file-like object or a string path.
        Writes to the raw output prefix only. If you want to write to another fs use put_data or get the fsspec
        file system directly.
        FYI: Currently the raw output prefix set by propeller is already unique per retry and looks like
             s3://my-s3-bucket/data/o4/feda4e266c748463a97d-n0-0

        If lpath is a folder, then recursive will be set.
        If lpath is a streamable, then it can only be a single file.

        Writes to:
            <raw output prefix>/<upload_prefix>/<file_name>

        :param lpath: A file-like object or a string path
        :param upload_prefix: A prefix to add to the path, see above for usage, can be an "". If None then a random
            string will be generated
        :param file_name: A file name to add to the path. If None, then the file name will be the tail of the path if
            lpath is a file, or a random string if lpath is a buffer
        :param read_chunk_size_bytes: If lpath is a buffer, this is the chunk size to read from it
        :param encoding: If lpath is a io.StringIO, this is the encoding to use to encode it to binary.
        :param kwargs: Additional kwargs are passed into the the fsspec put() call or the open() call
        :return: Returns the final path data was written to.
        """
        # First figure out what the destination path should be, then call put.
        upload_prefix = self.get_random_string() if upload_prefix is None else upload_prefix
        to_path = self.join(self.raw_output_prefix, upload_prefix)
        if file_name:
            to_path = self.join(to_path, file_name)
        else:
            if isinstance(lpath, str) or isinstance(lpath, os.PathLike) or isinstance(lpath, pathlib.Path):
                to_path = self.join(to_path, self.get_file_tail(str(lpath)))
            else:
                to_path = self.join(to_path, self.get_random_string())

        # If lpath is a file, then use put.
        if isinstance(lpath, str) or isinstance(lpath, os.PathLike) or isinstance(lpath, pathlib.Path):
            p = pathlib.Path(lpath)
            from_path = str(lpath)
            if not p.exists():
                raise FlyteAssertion(f"File {from_path} does not exist")
            elif p.is_symlink():
                raise FlyteAssertion(f"File {from_path} is a symlink, can't upload")
            if p.is_dir():
                logger.debug(f"Detected directory {from_path}, using recursive put")
                r = self.put(from_path, to_path, recursive=True, **kwargs)
            else:
                logger.debug(f"Detected file {from_path}, call put non-recursive")
                r = self.put(from_path, to_path, **kwargs)
            return r or to_path

        # raw bytes
        if isinstance(lpath, bytes):
            fs = self.get_filesystem_for_path(to_path)
            with fs.open(to_path, "wb", **kwargs) as s:
                s.write(lpath)
            return to_path

        # If lpath is a buffered reader of some kind
        if isinstance(lpath, io.BufferedReader) or isinstance(lpath, io.BytesIO):
            if not lpath.readable():
                raise FlyteAssertion("Buffered reader must be readable")
            fs = self.get_filesystem_for_path(to_path)
            lpath.seek(0)
            with fs.open(to_path, "wb", **kwargs) as s:
                while data := lpath.read(read_chunk_size_bytes):
                    s.write(data)
            return to_path

        if isinstance(lpath, io.StringIO):
            if not lpath.readable():
                raise FlyteAssertion("Buffered reader must be readable")
            fs = self.get_filesystem_for_path(to_path)
            lpath.seek(0)
            with fs.open(to_path, "wb", **kwargs) as s:
                while data_str := lpath.read(read_chunk_size_bytes):
                    s.write(data_str.encode(encoding))
            return to_path

        raise FlyteAssertion(f"Unsupported lpath type {type(lpath)}")

    @staticmethod
    def get_random_string() -> str:
        return UUID(int=random.getrandbits(128)).hex

    @staticmethod
    def get_file_tail(file_path_or_file_name: str) -> str:
        _, tail = os.path.split(file_path_or_file_name)
        return tail

    def join(
        self,
        *args: Unpack[str],  # type: ignore
        unstrip: bool = False,
        fs: typing.Optional[fsspec.AbstractFileSystem] = None,
    ) -> str:
        # todo add a check here for flyte fs
        fs = fs or self.raw_output_fs
        if len(args) == 0:
            raise ValueError("Must provide at least one argument")
        base, tails = args[0], list(args[1:])
        if get_protocol(base) not in str(fs.protocol):
            logger.warning(f"joining {base} with incorrect fs {fs.protocol} vs {get_protocol(base)}")
        if base.endswith(fs.sep):  # noqa
            base = base[:-1]
        l = [base]
        l.extend(tails)
        f = fs.sep.join(l)
        if unstrip:
            f = fs.unstrip_protocol(f)
        return f

    def get_random_local_path(self, file_path_or_file_name: typing.Optional[str] = None) -> str:
        """
        Use file_path_or_file_name, when you want a random directory, but want to preserve the leaf file name
        """
        key = UUID(int=random.getrandbits(128)).hex
        tail = ""
        if file_path_or_file_name:
            _, tail = os.path.split(file_path_or_file_name)
        if tail:
            return os.path.join(self._local_sandbox_dir, key, tail)
        return os.path.join(self._local_sandbox_dir, key)

    def get_random_local_directory(self) -> str:
        _dir = self.get_random_local_path(None)
        pathlib.Path(_dir).mkdir(parents=True, exist_ok=True)
        return _dir

    def get_random_remote_path(self, file_path_or_file_name: typing.Optional[str] = None) -> str:
        if file_path_or_file_name:
            return self.join(
                self.raw_output_prefix,
                self.get_random_string(),
                self.get_file_tail(file_path_or_file_name),
            )
        return self.join(
            self.raw_output_prefix,
            self.get_random_string(),
        )

    def get_random_remote_directory(self) -> str:
        return self.join(
            self.raw_output_prefix,
            self.get_random_string(),
        )

    def download_directory(self, remote_path: str, local_path: str, **kwargs):
        """
        Downloads directory from given remote to local path
        """
        return self.get_data(remote_path, local_path, is_multipart=True)

    def download(self, remote_path: str, local_path: str, **kwargs):
        """
        Downloads from remote to local
        """
        return self.get_data(remote_path, local_path, **kwargs)

    def upload(self, file_path: str, to_path: str, **kwargs):
        """
        :param Text file_path:
        :param Text to_path:
        """
        return self.put_data(file_path, to_path, **kwargs)

    def upload_directory(self, local_path: str, remote_path: str, **kwargs):
        """
        :param Text local_path:
        :param Text remote_path:
        """
        return self.put_data(local_path, remote_path, is_multipart=True, **kwargs)

    def get_data(self, remote_path: str, local_path: str, is_multipart: bool = False, **kwargs):
        """
        :param remote_path:
        :param local_path:
        :param is_multipart:
        """
        try:
            pathlib.Path(local_path).parent.mkdir(parents=True, exist_ok=True)
            with timeit(f"Download data to local from {remote_path}"):
                self.get(remote_path, to_path=local_path, recursive=is_multipart, **kwargs)
        except Exception as ex:
            raise FlyteAssertion(
                f"Failed to get data from {remote_path} to {local_path} (recursive={is_multipart}).\n\n"
                f"Original exception: {str(ex)}"
            )

    def put_data(
        self, local_path: Union[str, os.PathLike], remote_path: str, is_multipart: bool = False, **kwargs
    ) -> str:
        """
        The implication here is that we're always going to put data to the remote location, so we .remote to ensure
        we don't use the true local proxy if the remote path is a file://

        :param local_path:
        :param remote_path:
        :param is_multipart:
        """
        try:
            local_path = str(local_path)
            with timeit(f"Upload data to {remote_path}"):
                put_result = self.put(cast(str, local_path), remote_path, recursive=is_multipart, **kwargs)
                # This is an unfortunate workaround to ensure that we return the correct path for the remote location
                # Callers of this put_data function in flytekit have been changed to assign the remote path to the
                # output
                # of this function, so we want to make sure we don't change it unless we need to.
                if remote_path.startswith("flyte://"):
                    return put_result
                return remote_path
        except Exception as ex:
            raise FlyteAssertion(
                f"Failed to put data from {local_path} to {remote_path} (recursive={is_multipart}).\n\n"
                f"Original exception: {str(ex)}"
            ) from ex


flyte_tmp_dir = tempfile.mkdtemp(prefix="flyte-")
default_local_file_access_provider = FileAccessProvider(
    local_sandbox_dir=os.path.join(flyte_tmp_dir, "sandbox"),
    raw_output_prefix=os.path.join(flyte_tmp_dir, "raw"),
    data_config=DataConfig.auto(),
)
