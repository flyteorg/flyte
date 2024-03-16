import gzip
import hashlib
import importlib
import os
import shutil
import tarfile
import tempfile
import typing
from pathlib import Path

from flytekit import PythonFunctionTask
from flytekit.core.tracker import get_full_module_path
from flytekit.core.workflow import ImperativeWorkflow, WorkflowBase


def compress_scripts(source_path: str, destination: str, module_name: str):
    """
    Compresses the single script while maintaining the folder structure for that file.

    For example, given the follow file structure:
    .
    ├── flyte
    │   ├── __init__.py
    │   └── workflows
    │       ├── example.py
    │       ├── another_example.py
    │       ├── yet_another_example.py
    │       └── __init__.py

    Let's say you want to compress `example.py`. In that case we specify the the full module name as
    flyte.workflows.example and that will produce a tar file that contains only that file alongside
    with the folder structure, i.e.:

    .
    ├── flyte
    │   ├── __init__.py
    │   └── workflows
    │       ├── example.py
    │       └── __init__.py

    Note: If `example.py` didn't import tasks or workflows from `another_example.py` and `yet_another_example.py`, these files were not copied to the destination..

    """
    with tempfile.TemporaryDirectory() as tmp_dir:
        destination_path = os.path.join(tmp_dir, "code")

        visited: typing.List[str] = []
        copy_module_to_destination(source_path, destination_path, module_name, visited)
        tar_path = os.path.join(tmp_dir, "tmp.tar")
        with tarfile.open(tar_path, "w") as tar:
            tmp_path: str = os.path.join(tmp_dir, "code")
            files: typing.List[str] = os.listdir(tmp_path)
            for ws_file in files:
                tar.add(os.path.join(tmp_path, ws_file), arcname=ws_file, filter=tar_strip_file_attributes)
        with gzip.GzipFile(filename=destination, mode="wb", mtime=0) as gzipped:
            with open(tar_path, "rb") as tar_file:
                gzipped.write(tar_file.read())


def copy_module_to_destination(
    original_source_path: str, original_destination_path: str, module_name: str, visited: typing.List[str]
):
    """
    Copy the module (file) to the destination directory. If the module relative imports other modules, flytekit will
    recursively copy them as well.
    """
    mod = importlib.import_module(module_name)
    full_module_name = get_full_module_path(mod, mod.__name__)
    if full_module_name in visited:
        return
    visited.append(full_module_name)

    source_path = original_source_path
    destination_path = original_destination_path
    pkgs = full_module_name.split(".")

    for p in pkgs[:-1]:
        os.makedirs(os.path.join(destination_path, p), exist_ok=True)
        destination_path = os.path.join(destination_path, p)
        source_path = os.path.join(source_path, p)
        init_file = Path(os.path.join(source_path, "__init__.py"))
        if init_file.exists():
            shutil.copy(init_file, Path(os.path.join(destination_path, "__init__.py")))

    # Ensure destination path exists to cover the case of a single file and no modules.
    os.makedirs(destination_path, exist_ok=True)
    script_file = Path(source_path, f"{pkgs[-1]}.py")
    script_file_destination = Path(destination_path, f"{pkgs[-1]}.py")
    # Build the final script relative path and copy it to a known place.
    shutil.copy(
        script_file,
        script_file_destination,
    )

    # Try to copy other files to destination if tasks or workflows aren't in the same file
    for flyte_entity_name in mod.__dict__:
        flyte_entity = mod.__dict__[flyte_entity_name]
        if (
            isinstance(flyte_entity, (PythonFunctionTask, WorkflowBase))
            and not isinstance(flyte_entity, ImperativeWorkflow)
            and flyte_entity.instantiated_in
        ):
            copy_module_to_destination(
                original_source_path, original_destination_path, flyte_entity.instantiated_in, visited
            )


# Takes in a TarInfo and returns the modified TarInfo:
# https://docs.python.org/3/library/tarfile.html#tarinfo-objects
# intended to be passed as a filter to tarfile.add
# https://docs.python.org/3/library/tarfile.html#tarfile.TarFile.add
def tar_strip_file_attributes(tar_info: tarfile.TarInfo) -> tarfile.TarInfo:
    # set time to epoch timestamp 0, aka 00:00:00 UTC on 1 January 1970
    # note that when extracting this tarfile, this time will be shown as the modified date
    tar_info.mtime = 0

    # user/group info
    tar_info.uid = 0
    tar_info.uname = ""
    tar_info.gid = 0
    tar_info.gname = ""

    # stripping paxheaders may not be required
    # see https://stackoverflow.com/questions/34688392/paxheaders-in-tarball
    tar_info.pax_headers = {}

    return tar_info


def hash_file(file_path: typing.Union[os.PathLike, str]) -> (bytes, str, int):
    """
    Hash a file and produce a digest to be used as a version
    """
    h = hashlib.md5()
    l = 0

    with open(file_path, "rb") as file:
        while True:
            # Reading is buffered, so we can read smaller chunks.
            chunk = file.read(h.block_size)
            if not chunk:
                break
            h.update(chunk)
            l += len(chunk)

    return h.digest(), h.hexdigest(), l


def _find_project_root(source_path) -> str:
    """
    Find the root of the project.
    The root of the project is considered to be the first ancestor from source_path that does
    not contain a __init__.py file.

    N.B.: This assumption only holds for regular packages (as opposed to namespace packages)
    """
    # Start from the directory right above source_path
    path = Path(source_path).parent.resolve()
    while os.path.exists(os.path.join(path, "__init__.py")):
        path = path.parent
    return str(path)
