import os
import subprocess
import sys

from flytekit.tools.script_mode import compress_scripts, hash_file

MAIN_WORKFLOW = """
from flytekit import task, workflow
from wf1.test import t1

@workflow
def my_wf() -> str:
    return "hello world"
"""

IMPERATIVE_WORKFLOW = """
from flytekit import Workflow, task

@task
def t1(a: int):
    print(a)


wf = Workflow(name="my.imperative.workflow.example")
wf.add_workflow_input("a", int)
node_t1 = wf.add_entity(t1, a=wf.inputs["a"])
"""

T1_TASK = """
from flytekit import task
from wf2.test import t2


@task()
def t1() -> str:
    print("hello")
    return "hello"
"""

T2_TASK = """
from flytekit import task

@task()
def t2() -> str:
    print("hello")
    return "hello"
"""


def test_deterministic_hash(tmp_path):
    workflows_dir = tmp_path / "workflows"
    workflows_dir.mkdir()

    # Create dummy init file
    open(workflows_dir / "__init__.py", "a").close()
    # Write a dummy workflow
    workflow_file = workflows_dir / "hello_world.py"
    workflow_file.write_text(MAIN_WORKFLOW)

    imperative_workflow_file = workflows_dir / "imperative_wf.py"
    imperative_workflow_file.write_text(IMPERATIVE_WORKFLOW)

    t1_dir = tmp_path / "wf1"
    t1_dir.mkdir()
    open(t1_dir / "__init__.py", "a").close()
    t1_file = t1_dir / "test.py"
    t1_file.write_text(T1_TASK)

    t2_dir = tmp_path / "wf2"
    t2_dir.mkdir()
    open(t2_dir / "__init__.py", "a").close()
    t2_file = t2_dir / "test.py"
    t2_file.write_text(T2_TASK)

    destination = tmp_path / "destination"

    sys.path.append(str(workflows_dir.parent))
    compress_scripts(str(workflows_dir.parent), str(destination), "workflows.hello_world")

    digest, hex_digest, _ = hash_file(destination)

    # Try again to assert digest determinism
    destination2 = tmp_path / "destination2"
    compress_scripts(str(workflows_dir.parent), str(destination2), "workflows.hello_world")
    digest2, hex_digest2, _ = hash_file(destination)

    assert digest == digest2
    assert hex_digest == hex_digest2

    test_dir = tmp_path / "test"
    test_dir.mkdir()

    result = subprocess.run(
        ["tar", "-xvf", destination, "-C", test_dir],
        stdout=subprocess.PIPE,
    )
    result.check_returncode()
    assert len(next(os.walk(test_dir))[1]) == 3

    compress_scripts(str(workflows_dir.parent), str(destination), "workflows.imperative_wf")
