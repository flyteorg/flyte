import typing

from flytekit.exceptions.base import FlyteException as _FlyteException
from flytekit.exceptions.base import FlyteRecoverableException as _Recoverable


class FlyteUserException(_FlyteException):
    _ERROR_CODE = "USER:Unknown"


class FlyteTypeException(FlyteUserException, TypeError):
    _ERROR_CODE = "USER:TypeError"

    @staticmethod
    def _is_a_container(value):
        return isinstance(value, list) or isinstance(value, tuple) or isinstance(value, set)

    @classmethod
    def _create_verbose_message(cls, received_type, expected_type, received_value=None, additional_msg=None):
        if received_value is not None:
            return "Type error!  Received: {} with value: {}, Expected{}: {}. {}".format(
                received_type,
                received_value,
                " one of" if FlyteTypeException._is_a_container(expected_type) else "",
                expected_type,
                additional_msg or "",
            )
        else:
            return "Type error!  Received: {}, Expected{}: {}. {}".format(
                received_type,
                " one of" if FlyteTypeException._is_a_container(expected_type) else "",
                expected_type,
                additional_msg or "",
            )

    def __init__(self, received_type, expected_type, additional_msg=None, received_value=None):
        super(FlyteTypeException, self).__init__(
            self._create_verbose_message(
                received_type,
                expected_type,
                received_value=received_value,
                additional_msg=additional_msg,
            )
        )


class FlyteValueException(FlyteUserException, ValueError):
    _ERROR_CODE = "USER:ValueError"

    @classmethod
    def _create_verbose_message(cls, received_value, error_message):
        return "Value error!  Received: {}. {}".format(received_value, error_message)

    def __init__(self, received_value, error_message):
        super(FlyteValueException, self).__init__(self._create_verbose_message(received_value, error_message))


class FlyteAssertion(FlyteUserException, AssertionError):
    _ERROR_CODE = "USER:AssertionError"


class FlyteValidationException(FlyteAssertion):
    _ERROR_CODE = "USER:ValidationError"


class FlyteDisapprovalException(FlyteAssertion):
    _ERROR_CODE = "USER:ResultNotApproved"


class FlyteEntityAlreadyExistsException(FlyteAssertion):
    _ERROR_CODE = "USER:EntityAlreadyExists"


class FlyteEntityNotExistException(FlyteAssertion):
    _ERROR_CODE = "USER:EntityNotExist"


class FlyteTimeout(FlyteAssertion):
    _ERROR_CODE = "USER:Timeout"


class FlyteRecoverableException(FlyteUserException, _Recoverable):
    _ERROR_CODE = "USER:Recoverable"


class FlyteAuthenticationException(FlyteAssertion):
    _ERROR_CODE = "USER:AuthenticationError"


class FlyteInvalidInputException(FlyteUserException):
    _ERROR_CODE = "USER:BadInputToAPI"

    def __init__(self, request: typing.Any):
        self.request = request
        super().__init__()


class FlytePromiseAttributeResolveException(FlyteAssertion):
    _ERROR_CODE = "USER:PromiseAttributeResolveError"
