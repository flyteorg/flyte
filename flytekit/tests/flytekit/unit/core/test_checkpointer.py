import typing
from pathlib import Path

import py.path

from flytekit.core.checkpointer import SyncCheckpoint

CHECKPOINT_FILE = "cp"


def create_folder_write_file(tmpdir: py.path.local) -> typing.Tuple[py.path.local, py.path.local, py.path.local]:
    outputs = tmpdir.mkdir("outputs")

    # Make an input test directory with one file called cp
    inputs = tmpdir.mkdir("inputs")
    input_file = inputs.join(CHECKPOINT_FILE)
    input_file.write_text("Hello!", encoding="utf-8")

    return inputs, input_file, outputs


def test_sync_checkpoint_file(tmpdir: py.path.local):
    inputs, input_file, outputs = create_folder_write_file(tmpdir)
    cp = SyncCheckpoint(checkpoint_dest=str(outputs))
    # Lets try to restore - should not work!
    assert not cp.restore("/tmp")
    # Now save
    cp.save(str(input_file))
    # Expect file in tmpdir
    expected_dst = outputs.join(CHECKPOINT_FILE)
    assert outputs.listdir() == [expected_dst]


def test_sync_checkpoint_reader(tmpdir: py.path.local):
    inputs, input_file, outputs = create_folder_write_file(tmpdir)
    cp = SyncCheckpoint(checkpoint_dest=str(outputs))
    # Lets try to restore - should not work!
    assert not cp.restore("/tmp")
    # Now save
    with input_file.open(mode="rb") as b:
        cp.save(b)
    # Expect file in tmpdir
    expected_dst = outputs.join(SyncCheckpoint.TMP_DST_PATH)
    assert outputs.listdir() == [expected_dst]


def test_sync_checkpoint_folder(tmpdir: py.path.local):
    inputs, input_file, outputs = create_folder_write_file(tmpdir)
    cp = SyncCheckpoint(checkpoint_dest=str(outputs))
    # Lets try to restore - should not work!
    assert not cp.restore("/tmp")
    # Now save
    cp.save(Path(str(inputs)))
    # Expect file in tmpdir
    expected_dst = outputs.join(CHECKPOINT_FILE)
    assert outputs.listdir() == [expected_dst]


def test_sync_checkpoint_previous(tmpdir: py.path.local):
    inputs, input_file, outputs = create_folder_write_file(tmpdir)
    cp = SyncCheckpoint(checkpoint_dest=str(outputs), checkpoint_src=str(inputs))
    scratch = tmpdir.mkdir("user_scratch")
    assert cp.restore(str(scratch)) == scratch
    assert scratch.listdir() == [scratch.join(CHECKPOINT_FILE)]

    # ensure download is not performed again
    assert cp.restore("x") == scratch
