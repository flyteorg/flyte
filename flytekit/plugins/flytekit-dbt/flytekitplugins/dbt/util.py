import json
import subprocess
from typing import List

from flytekit.loggers import logger


def run_cli(cmd: List[str]) -> (int, List[str]):
    """
    Execute a CLI command in a subprocess

    Parameters
    ----------
    cmd : list of str
        Command to be executed.

    Returns
    -------
    int
        Command's exit code.
    list of str
        Logs produced by the command execution.
    """

    logs = []
    process = subprocess.Popen(cmd, stdout=subprocess.PIPE)
    for raw_line in process.stdout or []:
        line = raw_line.decode("utf-8")
        try:
            json_line = json.loads(line)
        except json.JSONDecodeError:
            logger.info(line.rstrip())
        else:
            logs.append(json_line)
            # TODO: pluck `levelname` from json_line and choose appropriate level to use
            # in flytekit logger instead of defaulting to `info`
            logger.info(line.rstrip())

    process.wait()
    return process.returncode, logs
