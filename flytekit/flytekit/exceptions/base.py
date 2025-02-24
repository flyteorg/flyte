class _FlyteCodedExceptionMetaclass(type):
    @property
    def error_code(cls):
        return cls._ERROR_CODE


class FlyteException(Exception, metaclass=_FlyteCodedExceptionMetaclass):
    _ERROR_CODE = "UnknownFlyteException"


class FlyteRecoverableException(FlyteException):
    _ERROR_CODE = "RecoverableFlyteException"
