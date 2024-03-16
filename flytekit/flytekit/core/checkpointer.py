import io
import tempfile
import typing
from abc import abstractmethod
from pathlib import Path


class Checkpoint(object):
    """
    Base class for Checkpoint system. Checkpoint system allows reading and writing custom checkpoints from user
    scripts
    """

    @abstractmethod
    def prev_exists(self) -> bool:
        raise NotImplementedError("Use one of the derived classes")

    @abstractmethod
    def restore(self, path: typing.Union[Path, str]) -> typing.Optional[Path]:
        """
        Given a path, if a previous checkpoint exists, will be downloaded to this path.
        If download is successful the downloaded path is returned

        .. note:

            Download will not be performed, if the checkpoint was previously restored. The method will return the
            previously downloaded path.

        """
        raise NotImplementedError("Use one of the derived classes")

    @abstractmethod
    def save(self, cp: typing.Union[Path, str, io.BufferedReader]):
        """
        Args:
            cp: Checkpoint file (path, str path or a io.BufferedReader)

        Usage: If you have a io.BufferedReader then the following should work

        .. code-block: python

            with input_file.open(mode="rb") as b:
                checkpointer.save(b)
        """
        raise NotImplementedError("Use one of the derived classes")

    @abstractmethod
    def read(self) -> typing.Optional[bytes]:
        """
        This should only be used if there is a singular checkpoint file written. If more than one checkpoint file is
        found, this will raise a ValueError
        """
        raise NotImplementedError("Use one of the derived classes")

    @abstractmethod
    def write(self, b: bytes):
        """
        This will overwrite the checkpoint. It can be retrieved using read or restore
        """
        raise NotImplementedError("Use one of the derived classes")


class SyncCheckpoint(Checkpoint):
    """
    This class is NOT THREAD-SAFE!
    Sync Checkpoint, will synchronously checkpoint a user given file or folder.
    It will also synchronously download / restore previous checkpoints, when restore is invoked.

    TODO: Implement an async checkpoint system
    """

    SRC_LOCAL_FOLDER = "prev_cp"
    TMP_DST_PATH = "_dst_cp"

    def __init__(self, checkpoint_dest: str, checkpoint_src: typing.Optional[str] = None):
        """
        Args:
            checkpoint_src: If a previous checkpoint should exist, this path should be set to the folder that contains the checkpoint information
            checkpoint_dest: Location where the new checkpoint should be copied to
        """
        self._checkpoint_dest = checkpoint_dest
        self._checkpoint_src = checkpoint_src if checkpoint_src and checkpoint_src != "" else None
        self._td = tempfile.TemporaryDirectory()
        self._prev_download_path: typing.Optional[Path] = None

    def __del__(self):
        self._td.cleanup()

    def prev_exists(self) -> bool:
        return self._checkpoint_src is not None

    def restore(self, path: typing.Optional[typing.Union[Path, str]] = None) -> typing.Optional[Path]:
        # We have to lazy load, until we fix the imports
        from flytekit.core.context_manager import FlyteContextManager

        if self._checkpoint_src is None or self._checkpoint_src == "":
            return None

        if self._prev_download_path:
            return self._prev_download_path

        if path is None:
            p = Path(self._td.name)
            path = p.joinpath(self.SRC_LOCAL_FOLDER)
            path.mkdir(exist_ok=True)
        elif isinstance(path, str):
            path = Path(path)

        if not path.is_dir():
            raise ValueError("Checkpoints can be restored to a directory only.")

        FlyteContextManager.current_context().file_access.download_directory(self._checkpoint_src, str(path))
        self._prev_download_path = path
        return self._prev_download_path

    def save(self, cp: typing.Union[Path, str, io.BufferedReader]):
        # We have to lazy load, until we fix the imports
        from flytekit.core.context_manager import FlyteContextManager

        fa = FlyteContextManager.current_context().file_access
        if isinstance(cp, (Path, str)):
            if isinstance(cp, str):
                cp = Path(cp)
            if cp.is_dir():
                fa.upload_directory(str(cp), self._checkpoint_dest)
            else:
                fname = cp.stem + cp.suffix
                rpath = fa._default_remote.sep.join([str(self._checkpoint_dest), fname])
                fa.upload(str(cp), rpath)
            return

        if not isinstance(cp, io.IOBase):
            raise ValueError(f"Only a valid path or IOBase type (reader) should be provided, received {type(cp)}")

        p = Path(self._td.name)
        dest_cp = p.joinpath(self.TMP_DST_PATH)
        with dest_cp.open("wb") as f:
            f.write(cp.read())

        rpath = fa._default_remote.sep.join([str(self._checkpoint_dest), self.TMP_DST_PATH])
        fa.upload(str(dest_cp), rpath)

    def read(self) -> typing.Optional[bytes]:
        p = self.restore()
        if p is None:
            return None
        files = list(p.iterdir())
        if len(files) == 0:
            return None
        if len(files) > 1:
            raise ValueError(f"Expected exactly one checkpoint - found {len(files)}")
        f = files[0]
        return f.read_bytes()

    def write(self, b: bytes):
        p = io.BytesIO(b)
        f = typing.cast(io.BufferedReader, p)
        self.save(f)
