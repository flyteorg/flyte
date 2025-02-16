from typing import List


class DBTHandledError(Exception):
    """
    DBTHandledError wraps error logs and message from command execution that returns ``exit code 1``.

    Parameters
    ----------
    message : str
        Error message.
    logs : list of str
        Logs produced by the command execution.

    Attributes
    ----------
    message : str
        Error message.
    logs : list of str
        Logs produced by the command execution.
    """

    def __init__(self, message: str, logs: List[str]):
        self.logs = logs
        self.message = message


class DBTUnhandledError(Exception):
    """
    DBTUnhandledError wraps error logs and message from command execution that returns ``exit code 2``.

    Parameters
    ----------
    message : str
        Error message.
    logs : list of str
        Logs produced by the command execution.

    Attributes
    ----------
    message : str
        Error message.
    logs : list of str
        Logs produced by the command execution.
    """

    def __init__(self, message: str, logs: List[str]):
        self.logs = logs
        self.message = message
