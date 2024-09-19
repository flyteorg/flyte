from flytekit.exceptions import base, system


def test_flyte_system_exception():
    try:
        raise system.FlyteSystemException("bad")
    except Exception as e:
        assert str(e) == "bad"
        assert isinstance(type(e), base._FlyteCodedExceptionMetaclass)
        assert type(e).error_code == "SYSTEM:Unknown"
        assert isinstance(e, base.FlyteException)


def test_flyte_not_implemented_exception():
    try:
        raise system.FlyteNotImplementedException("I'm lazy so I didn't implement this.")
    except Exception as e:
        assert str(e) == "I'm lazy so I didn't implement this."
        assert isinstance(e, NotImplementedError)
        assert type(e).error_code == "SYSTEM:NotImplemented"
        assert isinstance(e, system.FlyteSystemException)


def test_flyte_entrypoint_not_loadable_exception():
    try:
        raise system.FlyteEntrypointNotLoadable("fake.module")
    except Exception as e:
        assert str(e) == "Entrypoint is not loadable!  Could not load the module: 'fake.module'."
        assert type(e).error_code == "SYSTEM:UnloadableCode"
        assert isinstance(e, system.FlyteSystemException)

    try:
        raise system.FlyteEntrypointNotLoadable("fake.module", task_name="secret_task")
    except Exception as e:
        assert str(e) == "Entrypoint is not loadable!  Could not find the task: 'secret_task' in 'fake.module'."
        assert type(e).error_code == "SYSTEM:UnloadableCode"
        assert isinstance(e, system.FlyteSystemException)

    try:
        raise system.FlyteEntrypointNotLoadable("fake.module", additional_msg="Shouldn't have used a fake module!")
    except Exception as e:
        assert (
            str(e) == "Entrypoint is not loadable!  Could not load the module: 'fake.module' "
            "due to error: Shouldn't have used a fake module!"
        )
        assert type(e).error_code == "SYSTEM:UnloadableCode"
        assert isinstance(e, system.FlyteSystemException)

    try:
        raise system.FlyteEntrypointNotLoadable(
            "fake.module",
            task_name="secret_task",
            additional_msg="Shouldn't have used a fake module!",
        )
    except Exception as e:
        assert (
            str(e) == "Entrypoint is not loadable!  Could not find the task: 'secret_task' in 'fake.module' "
            "due to error: Shouldn't have used a fake module!"
        )
        assert type(e).error_code == "SYSTEM:UnloadableCode"
        assert isinstance(e, system.FlyteSystemException)


def test_flyte_system_assertion():
    try:
        raise system.FlyteSystemAssertion("I assert that the system messed up.")
    except Exception as e:
        assert str(e) == "I assert that the system messed up."
        assert type(e).error_code == "SYSTEM:AssertionError"
        assert isinstance(e, system.FlyteSystemException)
        assert isinstance(e, AssertionError)
