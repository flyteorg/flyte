"""Handle errors in elastic training jobs."""
import os

RECOVERABLE_ERROR_FILE_NAME = "recoverable_error"


def create_recoverable_error_file() -> None:
    """Create a file to signal to the agent process that an exception in the worker process is recoverable.

    Torch's `elastic_launch` gives the agent process access to exceptions raised in the worker
    processes only as strings in an error file. Instead of parsing this error file in the agent process for
    the string `FlyteRecoverableException` - which would not detect exceptions inheriting from
    `FlyteRecoverableException` - we create a file in the worker process to signal to the agent process
    that the exception is recoverable. The file is created in the directory where the default
    torch elastic error file is written.

    Raises:
        ValueError: If the environment variable `TORCHELASTIC_ERROR_FILE` is not set.
    """
    error_file_path = os.environ.get("TORCHELASTIC_ERROR_FILE")
    if error_file_path is None:
        raise ValueError("`TORCHELASTIC_ERROR_FILE` environment variable not set")

    recoverable_error_file_path = os.path.join(os.path.dirname(error_file_path), RECOVERABLE_ERROR_FILE_NAME)
    with open(recoverable_error_file_path, "w") as f:
        f.write("")


def is_recoverable_worker_error(failure) -> bool:
    """Check if the error in the worker process is recoverable.

    The error is considered recoverable if the directory containing the torch elastic error file contains
    a file named `recoverable_error`.

    Args:
        failure(torch.distributed.elastic.multiprocessing.errors.ProcessFailure): The error in the worker process.

    Returns:
        bool: True if the error is recoverable, False otherwise.
    """
    error_file_path = failure.error_file
    error_file_dir = os.path.dirname(error_file_path)
    recoverable_error_file_path = os.path.join(error_file_dir, RECOVERABLE_ERROR_FILE_NAME)
    return os.path.exists(recoverable_error_file_path)
