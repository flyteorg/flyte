from typing import Dict

import pytest

import flytekit.models.task as _task_models
from flytekit import Resources
from flytekit.core.resources import convert_resources_to_resource_model

_ResourceName = _task_models.Resources.ResourceName


def test_convert_no_requests_no_limits():
    resource_model = convert_resources_to_resource_model(requests=None, limits=None)
    assert isinstance(resource_model, _task_models.Resources)
    assert resource_model.requests == []
    assert resource_model.limits == []


@pytest.mark.parametrize(
    argnames=("resource_dict", "expected_resource_name"),
    argvalues=(
        ({"cpu": "2"}, _ResourceName.CPU),
        ({"mem": "1Gi"}, _ResourceName.MEMORY),
        ({"gpu": "1"}, _ResourceName.GPU),
        ({"ephemeral_storage": "123Mb"}, _ResourceName.EPHEMERAL_STORAGE),
    ),
    ids=("CPU", "MEMORY", "GPU", "EPHEMERAL_STORAGE"),
)
def test_convert_requests(resource_dict: Dict[str, str], expected_resource_name: _task_models.Resources):
    assert len(resource_dict) == 1
    expected_resource_value = list(resource_dict.values())[0]

    requests = Resources(**resource_dict)
    resources_model = convert_resources_to_resource_model(requests=requests)

    assert len(resources_model.requests) == 1
    request = resources_model.requests[0]
    assert isinstance(request, _task_models.Resources.ResourceEntry)
    assert request.name == expected_resource_name
    assert request.value == expected_resource_value
    assert len(resources_model.limits) == 0


@pytest.mark.parametrize(
    argnames=("resource_dict", "expected_resource_name"),
    argvalues=(
        ({"cpu": "2"}, _ResourceName.CPU),
        ({"mem": "1Gi"}, _ResourceName.MEMORY),
        ({"gpu": "1"}, _ResourceName.GPU),
        ({"ephemeral_storage": "123Mb"}, _ResourceName.EPHEMERAL_STORAGE),
    ),
    ids=("CPU", "MEMORY", "GPU", "EPHEMERAL_STORAGE"),
)
def test_convert_limits(resource_dict: Dict[str, str], expected_resource_name: _task_models.Resources):
    assert len(resource_dict) == 1
    expected_resource_value = list(resource_dict.values())[0]

    requests = Resources(**resource_dict)
    resources_model = convert_resources_to_resource_model(limits=requests)

    assert len(resources_model.limits) == 1
    limit = resources_model.limits[0]
    assert isinstance(limit, _task_models.Resources.ResourceEntry)
    assert limit.name == expected_resource_name
    assert limit.value == expected_resource_value
    assert len(resources_model.requests) == 0


def test_incorrect_type_resources():
    with pytest.raises(AssertionError):
        Resources(cpu=1)  # type: ignore
    with pytest.raises(AssertionError):
        Resources(mem=1)  # type: ignore
    with pytest.raises(AssertionError):
        Resources(gpu=1)  # type: ignore
    with pytest.raises(AssertionError):
        Resources(ephemeral_storage=1)  # type: ignore
