import subprocess
import sys
from functools import wraps
from typing import Callable, Optional

from flytekit.loggers import logger

from .constants import MAX_IDLE_SECONDS


def jupyter(
    _task_function: Optional[Callable] = None,
    max_idle_seconds: Optional[int] = MAX_IDLE_SECONDS,
    port: Optional[int] = 8888,
    enable: Optional[bool] = True,
    notebook_dir: Optional[str] = "/root",
    pre_execute: Optional[Callable] = None,
    post_execute: Optional[Callable] = None,
):
    def wrapper(fn):
        if not enable:
            return fn

        @wraps(fn)
        def inner_wrapper(*args, **kwargs):
            # 0. Executes the pre_execute function if provided.
            if pre_execute is not None:
                pre_execute()
                logger.info("Pre execute function executed successfully!")

            # 1. Launches and monitors the Jupyter Notebook server.
            # The following line starts a Jupyter Notebook server with specific configurations:
            #   - '--port': Specifies the port number on which the server will listen for connections.
            #   - '--notebook-dir': Sets the directory where Jupyter Notebook will look for notebooks.
            #   - '--NotebookApp.token='': Disables token-based authentication by setting an empty token.
            logger.info("Start the jupyter notebook server...")
            cmd = f"jupyter notebook --port {port} --notebook-dir={notebook_dir} --NotebookApp.token=''"

            #   - '--NotebookApp.shutdown_no_activity_timeout': Sets the maximum duration of inactivity
            #     before shutting down the Jupyter Notebook server automatically.
            # When shutdown_no_activity_timeout is 0, it means there is no idle timeout and it is always running.
            if max_idle_seconds:
                cmd += f" --NotebookApp.shutdown_no_activity_timeout={max_idle_seconds}"

            logger.info(cmd)
            process = subprocess.Popen(cmd, shell=True)

            # 2. Wait for the process to finish
            process.wait()

            # 3. Exit after subprocess has finished
            if post_execute is not None:
                post_execute()
                logger.info("Post execute function executed successfully!")
            sys.exit()

        return inner_wrapper

    # for the case when the decorator is used without arguments
    if _task_function is not None:
        return wrapper(_task_function)
    # for the case when the decorator is used with arguments
    else:
        return wrapper
