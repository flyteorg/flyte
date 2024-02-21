import pytest as _pytest

from flytekit.models import project


def test_project_with_default_state():
    obj = project.Project("project_id", "project_name", "project_description")
    assert obj.id == "project_id"
    assert obj.name == "project_name"
    assert obj.description == "project_description"
    assert obj.state == project.Project.ProjectState.ACTIVE
    assert obj == project.Project.from_flyte_idl(obj.to_flyte_idl())


@_pytest.mark.parametrize(
    "state", [project.Project.ProjectState.ARCHIVED, project.Project.ProjectState.SYSTEM_GENERATED]
)
def test_project_with_state(state):
    obj = project.Project("project_id", "project_name", "project_description", state=state)
    assert obj.id == "project_id"
    assert obj.name == "project_name"
    assert obj.description == "project_description"
    assert obj.state == state
    assert obj == project.Project.from_flyte_idl(obj.to_flyte_idl())
