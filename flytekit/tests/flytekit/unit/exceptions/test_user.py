from flytekit.exceptions import base, user


def test_flyte_user_exception():
    try:
        raise user.FlyteUserException("bad")
    except Exception as e:
        assert str(e) == "bad"
        assert isinstance(type(e), base._FlyteCodedExceptionMetaclass)
        assert type(e).error_code == "USER:Unknown"
        assert isinstance(e, base.FlyteException)


def test_flyte_type_exception():
    try:
        raise user.FlyteTypeException("int", "float", received_value=1, additional_msg="That was a bad idea!")
    except Exception as e:
        assert str(e) == "Type error!  Received: int with value: 1, Expected: float. That was a bad idea!"
        assert isinstance(e, TypeError)
        assert type(e).error_code == "USER:TypeError"
        assert isinstance(e, user.FlyteUserException)

    try:
        raise user.FlyteTypeException(
            "int",
            ("list", "set"),
            received_value=1,
            additional_msg="That was a bad idea!",
        )
    except Exception as e:
        assert (
            str(e) == "Type error!  Received: int with value: 1, Expected one of: ('list', 'set'). That was a "
            "bad idea!"
        )
        assert isinstance(e, TypeError)
        assert type(e).error_code == "USER:TypeError"
        assert isinstance(e, user.FlyteUserException)

    try:
        raise user.FlyteTypeException("int", "float", additional_msg="That was a bad idea!")
    except Exception as e:
        assert str(e) == "Type error!  Received: int, Expected: float. That was a bad idea!"
        assert isinstance(e, TypeError)
        assert type(e).error_code == "USER:TypeError"
        assert isinstance(e, user.FlyteUserException)

    try:
        raise user.FlyteTypeException("int", ("list", "set"), additional_msg="That was a bad idea!")
    except Exception as e:
        assert str(e) == "Type error!  Received: int, Expected one of: ('list', 'set'). That was a " "bad idea!"
        assert isinstance(e, TypeError)
        assert type(e).error_code == "USER:TypeError"
        assert isinstance(e, user.FlyteUserException)


def test_flyte_value_exception():
    try:
        raise user.FlyteValueException(-1, "Expected a value > 0")
    except user.FlyteValueException as e:
        assert str(e) == "Value error!  Received: -1. Expected a value > 0"
        assert isinstance(e, ValueError)
        assert type(e).error_code == "USER:ValueError"
        assert isinstance(e, user.FlyteUserException)


def test_flyte_assert():
    try:
        raise user.FlyteAssertion("I ASSERT THAT THIS IS WRONG!")
    except user.FlyteAssertion as e:
        assert str(e) == "I ASSERT THAT THIS IS WRONG!"
        assert isinstance(e, AssertionError)
        assert type(e).error_code == "USER:AssertionError"
        assert isinstance(e, user.FlyteUserException)


def test_flyte_validation_error():
    try:
        raise user.FlyteValidationException("I validated that your stuff was wrong.")
    except user.FlyteValidationException as e:
        assert str(e) == "I validated that your stuff was wrong."
        assert isinstance(e, AssertionError)
        assert type(e).error_code == "USER:ValidationError"
        assert isinstance(e, user.FlyteUserException)
