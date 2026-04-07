import pytest

def pytest_addoption(parser):
    parser.addoption("--flytesnacks_release_tag", required=True)
    parser.addoption("--priorities", required=True)
    parser.addoption("--config_file", required=True)
    parser.addoption(
        "--return_non_zero_on_failure",
        action="store_true",
        default=False,
        help="Return a non-zero exit status if any workflow fails",
    )
    parser.addoption(
        "--terminate_workflow_on_failure",
        action="store_true",
        default=False,
        help="Abort failing workflows upon exit",
    )
    parser.addoption(
        "--test_project_name", 
        default="flytesnacks",
        help="Name of project to run functional tests on"
    )
    parser.addoption(
        "--test_project_domain", 
        default="development", 
        help="Name of domain in project to run functional tests on"
    )
    parser.addoption(
        "--cluster_pool_name",
        required=False,
        type=str,
        default=None,
    )

@pytest.fixture
def setup_flytesnacks_env(pytestconfig):
    return {
        "flytesnacks_release_tag": pytestconfig.getoption("--flytesnacks_release_tag"),
        "priorities": pytestconfig.getoption("--priorities"),
        "config_file": pytestconfig.getoption("--config_file"),
        "return_non_zero_on_failure": pytestconfig.getoption("--return_non_zero_on_failure"),
        "terminate_workflow_on_failure": pytestconfig.getoption("--terminate_workflow_on_failure"),
        "test_project_name": pytestconfig.getoption("--test_project_name"),
        "test_project_domain": pytestconfig.getoption("--test_project_domain"),
        "cluster_pool_name": pytestconfig.getoption("--cluster_pool_name"),
    }
