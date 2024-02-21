from functools import wraps as _wraps
from sys import exc_info as _exc_info
from traceback import format_tb as _format_tb

from flytekit.exceptions import base as _base_exceptions
from flytekit.exceptions import system as _system_exceptions
from flytekit.exceptions import user as _user_exceptions
from flytekit.models.core import errors as _error_model


class FlyteScopedException(Exception):
    def __init__(self, context, exc_type, exc_value, exc_tb, top_trim=0, bottom_trim=0, kind=None):
        self._exc_type = exc_type
        self._exc_value = exc_value
        self._exc_tb = exc_tb
        self._top_trim = top_trim
        self._bottom_trim = bottom_trim
        self._context = context
        self._kind = kind
        super(FlyteScopedException, self).__init__(str(self.value))

    @property
    def verbose_message(self):
        tb = self.traceback
        to_trim = self._top_trim
        while to_trim > 0 and tb.tb_next is not None:
            tb = tb.tb_next

        top_tb = tb
        limit = 0
        while tb is not None:
            limit += 1
            tb = tb.tb_next
        limit = max(0, limit - self._bottom_trim)

        lines = _format_tb(top_tb, limit=limit)
        lines = [line.rstrip() for line in lines]
        lines = "\n".join(lines).split("\n")
        traceback_str = "\n    ".join([""] + lines)

        format_str = "Traceback (most recent call last):\n" "{traceback}\n" "\n" "Message:\n" "\n" "    {message}"
        return format_str.format(traceback=traceback_str, message=f"{self.type.__name__}: {self.value}")

    def __str__(self):
        return str(self.value)

    @property
    def value(self):
        if isinstance(self._exc_value, FlyteScopedException):
            return self._exc_value.value
        return self._exc_value

    @property
    def traceback(self):
        if isinstance(self._exc_value, FlyteScopedException):
            return self._exc_value.traceback
        return self._exc_tb

    @property
    def type(self):
        if isinstance(self._exc_value, FlyteScopedException):
            return self._exc_value.type
        return self._exc_type

    @property
    def error_code(self):
        """
        :rtype: Text
        """
        if isinstance(self._exc_value, FlyteScopedException):
            return self._exc_value.error_code

        if hasattr(type(self._exc_value), "error_code"):
            return type(self._exc_value).error_code
        return "{}:Unknown".format(self._context)

    @property
    def kind(self) -> int:
        """
        :rtype: int
        """
        if self._kind is not None:
            # If kind is overridden, return it.
            return self._kind
        elif isinstance(self._exc_value, FlyteScopedException):
            # Otherwise, go lower in the scope to find the kind of exception.
            return self._exc_value.kind
        elif isinstance(self._exc_value, _base_exceptions.FlyteRecoverableException):
            # If it is an exception that is recoverable, we return it as such.
            return _error_model.ContainerError.Kind.RECOVERABLE
        else:
            # The remaining exceptions are considered unrecoverable.
            return _error_model.ContainerError.Kind.NON_RECOVERABLE


class FlyteScopedSystemException(FlyteScopedException):
    def __init__(self, exc_type, exc_value, exc_tb, **kwargs):
        super(FlyteScopedSystemException, self).__init__("SYSTEM", exc_type, exc_value, exc_tb, **kwargs)

    @property
    def verbose_message(self):
        """
        :rtype: Text
        """
        base_msg = super(FlyteScopedSystemException, self).verbose_message
        base_msg += "\n\nSYSTEM ERROR! Contact platform administrators."
        return base_msg


class FlyteScopedUserException(FlyteScopedException):
    def __init__(self, exc_type, exc_value, exc_tb, **kwargs):
        super(FlyteScopedUserException, self).__init__("USER", exc_type, exc_value, exc_tb, **kwargs)

    @property
    def verbose_message(self):
        """
        :rtype: Text
        """
        base_msg = super(FlyteScopedUserException, self).verbose_message
        base_msg += "\n\nUser error."
        return base_msg


_NULL_CONTEXT = 0
_USER_CONTEXT = 1
_SYSTEM_CONTEXT = 2

# Keep the stack with a null-context so we never have to range check when peeking back.
_CONTEXT_STACK = [_NULL_CONTEXT]


def _is_base_context():
    return _CONTEXT_STACK[-2] == _NULL_CONTEXT


def _decorator(outer_f):
    """Decorate a function with signature func(wrapped, args, kwargs)."""

    @_wraps(outer_f)
    def inner_decorator(inner_f):
        @_wraps(inner_f)
        def f(*args, **kwargs):
            return outer_f(inner_f, args, kwargs)

        return f

    return inner_decorator


@_decorator
def system_entry_point(wrapped, args, kwargs):
    """
    The reason these two (see the user one below) decorators exist is to categorize non-Flyte exceptions at arbitrary
    locations. For example, while there is a separate ecosystem of Flyte-defined user and system exceptions
    (see the FlyteException hierarchy), and we can easily understand and categorize those, if flytekit comes upon
    a random ``ValueError`` or other non-flytekit defined error, how would we know if it was a bug in flytekit versus an
    error with user code or something the user called? The purpose of these decorators is to categorize those (see
    the last case in the nested try/catch below.

    Decorator for wrapping functions that enter a system context. This should decorate every method that may invoke some
    user code later on down the line. This will allow us to add differentiation between what is a user error and
    what is a system failure. Furthermore, we will clean the exception trace so as to make more sense to the
    user -- allowing them to know if they should take action themselves or pass on to the platform owners.
    We will dispatch metrics and such appropriately.
    """
    try:
        _CONTEXT_STACK.append(_SYSTEM_CONTEXT)
        if _is_base_context():
            # If this is the first time either of this decorator, or the one below is called, then we unwrap the
            # exception. The first time these decorators are used is currently in the entrypoint.py file. The scoped
            # exceptions are unwrapped because at that point, we want to return the underlying error to the user.
            try:
                return wrapped(*args, **kwargs)
            except FlyteScopedException as ex:
                raise ex.value
        else:
            try:
                return wrapped(*args, **kwargs)
            except FlyteScopedException as scoped:
                raise scoped
            except _user_exceptions.FlyteUserException:
                # Re-raise from here.
                raise FlyteScopedUserException(*_exc_info())
            except Exception:
                # This is why this function exists - arbitrary exceptions that we don't know what to do with are
                # interpreted as system errors.
                # System error, raise full stack-trace all the way up the chain.
                raise FlyteScopedSystemException(*_exc_info(), kind=_error_model.ContainerError.Kind.RECOVERABLE)
    finally:
        _CONTEXT_STACK.pop()


@_decorator
def user_entry_point(wrapped, args, kwargs):
    """
    See the comment for the system_entry_point above as well.

    Decorator for wrapping functions that enter into a user context.  This will help us differentiate user-created
    failures even when it is re-entrant into system code.

    Note: a user_entry_point can ONLY ever be called from within a @system_entry_point wrapped function, therefore,
    we can always ensure we will hit a system_entry_point to correctly reformat our exceptions.  Also, any exception
    we create here will only be handled within our system code so we don't need to worry about leaking weird exceptions
    to the user.
    """
    try:
        _CONTEXT_STACK.append(_USER_CONTEXT)
        if _is_base_context():
            # See comment at this location for system_entry_point
            fn_name = wrapped.__name__
            try:
                return wrapped(*args, **kwargs)
            except FlyteScopedException as exc:
                raise exc.type(f"Error encountered while executing '{fn_name}':\n  {exc.value}") from exc
            except Exception as exc:
                raise type(exc)(f"Error encountered while executing '{fn_name}':\n  {exc}") from exc
        else:
            try:
                return wrapped(*args, **kwargs)
            except FlyteScopedException as scoped:
                raise scoped
            except _user_exceptions.FlyteUserException:
                raise FlyteScopedUserException(*_exc_info())
            except _system_exceptions.FlyteSystemException:
                raise FlyteScopedSystemException(*_exc_info())
            except Exception:
                # This is why this function exists - arbitrary exceptions that we don't know what to do with are
                # interpreted as user exceptions.
                # This will also catch FlyteUserException re-raised by the system_entry_point handler
                raise FlyteScopedUserException(*_exc_info())
    finally:
        _CONTEXT_STACK.pop()
