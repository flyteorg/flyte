from collections import OrderedDict

import mock
import pytest
from flytekitplugins.flyteinteractive import (
    CODE_TOGETHER_CONFIG,
    CODE_TOGETHER_EXTENSION,
    COPILOT_CONFIG,
    COPILOT_EXTENSION,
    DEFAULT_CODE_SERVER_DIR_NAMES,
    DEFAULT_CODE_SERVER_EXTENSIONS,
    DEFAULT_CODE_SERVER_REMOTE_PATHS,
    VIM_CONFIG,
    VIM_EXTENSION,
    VscodeConfig,
    jupyter,
    vscode,
)
from flytekitplugins.flyteinteractive.vscode_lib.constants import EXIT_CODE_SUCCESS
from flytekitplugins.flyteinteractive.vscode_lib.decorator import (
    get_code_server_info,
    get_installed_extensions,
    is_extension_installed,
)

from flytekit import task, workflow
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.context_manager import ExecutionState
from flytekit.tools.translator import get_serializable_task


@pytest.fixture
def mock_local_execution():
    with mock.patch.object(ExecutionState, "is_local_execution", return_value=True) as mock_func:
        yield mock_func


@pytest.fixture
def mock_remote_execution():
    with mock.patch.object(ExecutionState, "is_local_execution", return_value=False) as mock_func:
        yield mock_func


@pytest.fixture
def mock_code_server_info_dict():
    return {"arm64": "Arm server info", "amd64": "AMD server info"}


@pytest.fixture
def vscode_patches():
    with mock.patch("multiprocessing.Process") as mock_process, mock.patch(
        "flytekitplugins.flyteinteractive.vscode_lib.decorator.prepare_interactive_python"
    ) as mock_prepare_interactive_python, mock.patch(
        "flytekitplugins.flyteinteractive.vscode_lib.decorator.exit_handler"
    ) as mock_exit_handler, mock.patch(
        "flytekitplugins.flyteinteractive.vscode_lib.decorator.download_vscode"
    ) as mock_download_vscode, mock.patch("signal.signal") as mock_signal, mock.patch(
        "flytekitplugins.flyteinteractive.vscode_lib.decorator.prepare_resume_task_python"
    ) as mock_prepare_resume_task_python, mock.patch(
        "flytekitplugins.flyteinteractive.vscode_lib.decorator.prepare_launch_json"
    ) as mock_prepare_launch_json:
        yield (
            mock_process,
            mock_prepare_interactive_python,
            mock_exit_handler,
            mock_download_vscode,
            mock_signal,
            mock_prepare_resume_task_python,
            mock_prepare_launch_json,
        )


def test_vscode_remote_execution(vscode_patches, mock_remote_execution):
    (
        mock_process,
        mock_prepare_interactive_python,
        mock_exit_handler,
        mock_download_vscode,
        mock_signal,
        mock_prepare_resume_task_python,
        mock_prepare_launch_json,
    ) = vscode_patches

    @task
    @vscode
    def t():
        return

    @workflow
    def wf():
        t()

    wf()
    mock_download_vscode.assert_called_once()
    mock_process.assert_called_once()
    mock_exit_handler.assert_called_once()
    mock_prepare_interactive_python.assert_called_once()
    mock_signal.assert_called_once()
    mock_prepare_resume_task_python.assert_called_once()
    mock_prepare_launch_json.assert_called_once()


def test_vscode_remote_execution_but_disable(vscode_patches, mock_remote_execution):
    (
        mock_process,
        mock_prepare_interactive_python,
        mock_exit_handler,
        mock_download_vscode,
        mock_signal,
        mock_prepare_resume_task_python,
        mock_prepare_launch_json,
    ) = vscode_patches

    @task
    @vscode(enable=False)
    def t():
        return

    @workflow
    def wf():
        t()

    wf()
    mock_download_vscode.assert_not_called()
    mock_process.assert_not_called()
    mock_exit_handler.assert_not_called()
    mock_prepare_interactive_python.assert_not_called()
    mock_signal.assert_not_called()
    mock_prepare_resume_task_python.assert_not_called()
    mock_prepare_launch_json.assert_not_called()


def test_vscode_local_execution(vscode_patches, mock_local_execution):
    (
        mock_process,
        mock_prepare_interactive_python,
        mock_exit_handler,
        mock_download_vscode,
        mock_signal,
        mock_prepare_resume_task_python,
        mock_prepare_launch_json,
    ) = vscode_patches

    @task
    @vscode
    def t():
        return

    @workflow
    def wf():
        t()

    wf()
    mock_download_vscode.assert_not_called()
    mock_process.assert_not_called()
    mock_exit_handler.assert_not_called()
    mock_prepare_interactive_python.assert_not_called()
    mock_signal.assert_not_called()
    mock_prepare_resume_task_python.assert_not_called()
    mock_prepare_launch_json.assert_not_called()


def test_vscode_run_task_first_succeed(mock_remote_execution):
    @task
    @vscode(run_task_first=True)
    def t(a: int, b: int) -> int:
        return a + b

    @workflow
    def wf(a: int, b: int) -> int:
        out = t(a=a, b=b)
        return out

    res = wf(a=10, b=5)
    assert res == 15


def test_vscode_run_task_first_fail(vscode_patches, mock_remote_execution):
    (
        mock_process,
        mock_prepare_interactive_python,
        mock_exit_handler,
        mock_download_vscode,
        mock_signal,
        mock_prepare_resume_task_python,
        mock_prepare_launch_json,
    ) = vscode_patches

    @task
    @vscode
    def t(a: int, b: int):
        dummy = a // b  # noqa: F841
        return

    @workflow
    def wf(a: int, b: int):
        t(a=a, b=b)

    wf(a=10, b=0)
    mock_download_vscode.assert_called_once()
    mock_process.assert_called_once()
    mock_exit_handler.assert_called_once()
    mock_prepare_interactive_python.assert_called_once()
    mock_signal.assert_called_once()
    mock_prepare_resume_task_python.assert_called_once()
    mock_prepare_launch_json.assert_called_once()


@mock.patch("flytekitplugins.flyteinteractive.jupyter_lib.decorator.subprocess.Popen")
@mock.patch("flytekitplugins.flyteinteractive.jupyter_lib.decorator.sys.exit")
def test_jupyter(mock_exit, mock_popen):
    @task
    @jupyter
    def t():
        return

    @workflow
    def wf():
        t()

    wf()
    mock_popen.assert_called_once()
    mock_exit.assert_called_once()


def test_is_extension_installed():
    installed_extensions = [
        "ms-python.python",
        "ms-toolsai.jupyter",
        "ms-toolsai.jupyter-keymap",
        "ms-toolsai.jupyter-renderers",
        "ms-toolsai.vscode-jupyter-cell-tags",
        "ms-toolsai.vscode-jupyter-slideshow",
    ]
    config = VscodeConfig()
    for extension in config.extension_remote_paths:
        assert is_extension_installed(extension, installed_extensions)


def test_vscode_config():
    config = VscodeConfig()
    assert config.code_server_remote_paths == DEFAULT_CODE_SERVER_REMOTE_PATHS
    assert config.code_server_dir_names == DEFAULT_CODE_SERVER_DIR_NAMES
    assert config.extension_remote_paths == DEFAULT_CODE_SERVER_EXTENSIONS

    code_together_config = CODE_TOGETHER_CONFIG
    assert code_together_config.code_server_remote_paths == DEFAULT_CODE_SERVER_REMOTE_PATHS
    assert code_together_config.code_server_dir_names == DEFAULT_CODE_SERVER_DIR_NAMES
    assert code_together_config.extension_remote_paths == DEFAULT_CODE_SERVER_EXTENSIONS + [CODE_TOGETHER_EXTENSION]

    copilot_config = COPILOT_CONFIG
    assert copilot_config.code_server_remote_paths == DEFAULT_CODE_SERVER_REMOTE_PATHS
    assert copilot_config.code_server_dir_names == DEFAULT_CODE_SERVER_DIR_NAMES
    assert copilot_config.extension_remote_paths == DEFAULT_CODE_SERVER_EXTENSIONS + [COPILOT_EXTENSION]

    vim_config = VIM_CONFIG
    assert vim_config.code_server_remote_paths == DEFAULT_CODE_SERVER_REMOTE_PATHS
    assert vim_config.code_server_dir_names == DEFAULT_CODE_SERVER_DIR_NAMES
    assert vim_config.extension_remote_paths == DEFAULT_CODE_SERVER_EXTENSIONS + [VIM_EXTENSION]

    all_extensions_config = VscodeConfig()
    all_extensions_config.add_extensions([CODE_TOGETHER_EXTENSION, COPILOT_EXTENSION, VIM_EXTENSION])
    assert CODE_TOGETHER_EXTENSION in all_extensions_config.extension_remote_paths
    assert COPILOT_EXTENSION in all_extensions_config.extension_remote_paths
    assert VIM_EXTENSION in all_extensions_config.extension_remote_paths


def test_vscode_config_add_extensions():
    additional_extensions = [COPILOT_EXTENSION, VIM_EXTENSION, CODE_TOGETHER_EXTENSION]

    config = VscodeConfig()
    config.add_extensions(additional_extensions)

    for extension in additional_extensions:
        assert extension in config.extension_remote_paths

    additional_extension = "test_str_extension"
    config.add_extensions(additional_extension)
    assert additional_extension in config.extension_remote_paths


def test_vscode_with_args(vscode_patches, mock_remote_execution):
    (
        mock_process,
        mock_prepare_interactive_python,
        mock_exit_handler,
        mock_download_vscode,
        mock_signal,
        mock_prepare_resume_task_python,
        mock_prepare_launch_json,
    ) = vscode_patches

    @task
    @vscode
    def t():
        return

    @workflow
    def wf():
        t()

    wf()

    mock_download_vscode.assert_called_once()
    mock_process.assert_called_once()
    mock_exit_handler.assert_called_once()
    mock_prepare_interactive_python.assert_called_once()
    mock_signal.assert_called_once()
    mock_prepare_resume_task_python.assert_called_once()
    mock_prepare_launch_json.assert_called_once()


def test_vscode_extra_config(mock_remote_execution):
    @vscode(
        max_idle_seconds=100,
        port=8081,
        enable=True,
        pre_execute=None,
        post_execute=None,
        config=None,
    )
    def t():
        return

    assert t.get_extra_config()["link_type"] == "vscode"
    assert t.get_extra_config()["port"] == "8081"


def test_serialize_vscode(mock_remote_execution):
    @task
    @vscode(
        max_idle_seconds=100,
        port=8081,
        enable=True,
        pre_execute=None,
        post_execute=None,
        config=None,
    )
    def t():
        return

    default_image = Image(name="default", fqn="docker.io/xyz", tag="some-git-hash")
    default_image_config = ImageConfig(default_image=default_image)
    default_serialization_settings = SerializationSettings(
        project="p", domain="d", version="v", image_config=default_image_config
    )

    serialized_task = get_serializable_task(OrderedDict(), default_serialization_settings, t)
    assert serialized_task.template.config == {"link_type": "vscode", "port": "8081"}


@mock.patch("platform.machine", return_value="aarch64")
def test_arm_platform(mock_machine, mock_code_server_info_dict):
    assert get_code_server_info(mock_code_server_info_dict) == "Arm server info"


@mock.patch("platform.machine", return_value="x86_64")
def test_amd_platform(mock_machine, mock_code_server_info_dict):
    assert get_code_server_info(mock_code_server_info_dict) == "AMD server info"


@mock.patch("platform.machine", return_value="Unsupported machine info")
def test_platform_unsupported(mock_machine, mock_code_server_info_dict):
    with pytest.raises(
        ValueError,
        match="Automatic download is only supported on AMD64 and ARM64 architectures. If you are using a different architecture, please visit the code-server official website to manually download the appropriate version for your image.",
    ):
        get_code_server_info(mock_code_server_info_dict)


@mock.patch("subprocess.run")
def test_get_installed_extensions_succeeded(mock_run):
    # Set up the mock process
    mock_process = mock.Mock()
    mock_process.returncode = EXIT_CODE_SUCCESS
    mock_process.stdout = (
        "ms-python.python\n"
        "ms-toolsai.jupyter\n"
        "ms-toolsai.jupyter-keymap\n"
        "ms-toolsai.jupyter-renderers\n"
        "ms-toolsai.vscode-jupyter-cell-tags\n"
        "ms-toolsai.vscode-jupyter-slideshow\n"
    )
    mock_run.return_value = mock_process

    installed_extensions = get_installed_extensions()

    # Verify the correct command was called
    mock_run.assert_called_once_with(["code-server", "--list-extensions"], capture_output=True, text=True)

    # Assert that the output matches the expected list of extensions
    expected_extensions = [
        "ms-python.python",
        "ms-toolsai.jupyter",
        "ms-toolsai.jupyter-keymap",
        "ms-toolsai.jupyter-renderers",
        "ms-toolsai.vscode-jupyter-cell-tags",
        "ms-toolsai.vscode-jupyter-slideshow",
    ]
    assert installed_extensions == expected_extensions


@mock.patch("subprocess.run")
def test_get_installed_extensions_failed(mock_run):
    # Set up the mock process
    mock_process = mock.Mock()
    mock_process.returncode = 1
    mock_process.stdout = (
        "ms-python.python\n"
        "ms-toolsai.jupyter\n"
        "ms-toolsai.jupyter-keymap\n"
        "ms-toolsai.jupyter-renderers\n"
        "ms-toolsai.vscode-jupyter-cell-tags\n"
        "ms-toolsai.vscode-jupyter-slideshow\n"
    )
    mock_run.return_value = mock_process

    installed_extensions = get_installed_extensions()

    mock_run.assert_called_once_with(["code-server", "--list-extensions"], capture_output=True, text=True)

    expected_extensions = []
    assert installed_extensions == expected_extensions
