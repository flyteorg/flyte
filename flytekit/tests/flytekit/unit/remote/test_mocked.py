import mock

from flytekit.core.task import task


@mock.patch("flytekit.remote.remote.FlyteRemote")
def test_mocking_remote(mock_remote) -> None:
    """
    This is a test that showing one way to mock fetched tasks, since the flytekit.testing elements don't work on remote
    entities.
    """

    @task
    def t1() -> float:
        return 6.62607015e-34

    @task
    def t2() -> bool:
        return False

    mock_remote.return_value.fetch_task.side_effect = [t1, t2]
    from . import wf_with_remote

    x = wf_with_remote.hello_wf(a=3)
    assert x == (6.62607015e-34, False)
