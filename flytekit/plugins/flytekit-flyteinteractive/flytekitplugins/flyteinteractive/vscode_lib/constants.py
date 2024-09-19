import os

# Where the code-server tar and plugins are downloaded to
EXECUTABLE_NAME = "code-server"
DOWNLOAD_DIR = "/tmp/code-server"
HOURS_TO_SECONDS = 60 * 60
DEFAULT_UP_SECONDS = 10 * HOURS_TO_SECONDS  # 10 hours
DEFAULT_CODE_SERVER_REMOTE_PATHS = {
    "amd64": "https://github.com/coder/code-server/releases/download/v4.18.0/code-server-4.18.0-linux-amd64.tar.gz",
    "arm64": "https://github.com/coder/code-server/releases/download/v4.18.0/code-server-4.18.0-linux-arm64.tar.gz",
}
DEFAULT_CODE_SERVER_EXTENSIONS = [
    "https://raw.githubusercontent.com/flyteorg/flytetools/master/flytekitplugins/flyin/ms-python.python-2023.20.0.vsix",
    "https://raw.githubusercontent.com/flyteorg/flytetools/master/flytekitplugins/flyin/ms-toolsai.jupyter-2023.9.100.vsix",
]
DEFAULT_CODE_SERVER_DIR_NAMES = {
    "amd64": "code-server-4.18.0-linux-amd64",
    "arm64": "code-server-4.18.0-linux-arm64",
}
# Default max idle seconds to terminate the vscode server
HOURS_TO_SECONDS = 60 * 60
MAX_IDLE_SECONDS = 10 * HOURS_TO_SECONDS  # 10 hours

# Duration to pause the checking of the heartbeat file until the next one
HEARTBEAT_CHECK_SECONDS = 60

# The path is hardcoded by code-server
# https://coder.com/docs/code-server/latest/FAQ#what-is-the-heartbeat-file
HEARTBEAT_PATH = os.path.expanduser("~/.local/share/code-server/heartbeat")

INTERACTIVE_DEBUGGING_FILE_NAME = "flyteinteractive_interactive_entrypoint.py"
RESUME_TASK_FILE_NAME = "flyteinteractive_resume_task.py"
# Config keys to store in task template
VSCODE_TYPE_KEY = "flyteinteractive_type"
VSCODE_PORT_KEY = "flyteinteractive_port"

# Context attribute name of the task function's source file path
TASK_FUNCTION_SOURCE_PATH = "TASK_FUNCTION_SOURCE_PATH"

# Subprocess constants
EXIT_CODE_SUCCESS = 0
