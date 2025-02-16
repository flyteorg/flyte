from flytekit.exceptions import base as _base_exceptions


class FlyteSystemException(_base_exceptions.FlyteRecoverableException):
    _ERROR_CODE = "SYSTEM:Unknown"


class FlyteNotImplementedException(FlyteSystemException, NotImplementedError):
    _ERROR_CODE = "SYSTEM:NotImplemented"


class FlyteEntrypointNotLoadable(FlyteSystemException):
    _ERROR_CODE = "SYSTEM:UnloadableCode"

    @classmethod
    def _create_verbose_message(cls, task_module, task_name=None, additional_msg=None):
        if task_name is None:
            return "Entrypoint is not loadable!  Could not load the module: '{task_module}'{additional_msg}".format(
                task_module=task_module,
                additional_msg=" due to error: {}".format(additional_msg) if additional_msg is not None else ".",
            )
        else:
            return (
                "Entrypoint is not loadable!  Could not find the task: '{task_name}' in '{task_module}'"
                "{additional_msg}".format(
                    task_module=task_module,
                    task_name=task_name,
                    additional_msg="." if additional_msg is None else " due to error: {}".format(additional_msg),
                )
            )

    def __init__(self, task_module, task_name=None, additional_msg=None):
        super(FlyteSystemException, self).__init__(
            self._create_verbose_message(task_module, task_name=task_name, additional_msg=additional_msg)
        )


class FlyteSystemAssertion(FlyteSystemException, AssertionError):
    _ERROR_CODE = "SYSTEM:AssertionError"


class FlyteAgentNotFound(FlyteSystemException, AssertionError):
    _ERROR_CODE = "SYSTEM:AgentNotFound"
