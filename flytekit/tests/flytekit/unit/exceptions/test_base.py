from flytekit.exceptions import base


def test_flyte_exception():
    try:
        raise base.FlyteException("bad")
    except Exception as e:
        assert str(e) == "bad"
        assert isinstance(type(e), base._FlyteCodedExceptionMetaclass)
        assert type(e).error_code == "UnknownFlyteException"
