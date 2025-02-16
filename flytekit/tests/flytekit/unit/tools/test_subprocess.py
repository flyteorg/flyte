import mock

from flytekit.tools import subprocess


class _MockProcess(object):
    def wait(self):
        return 0


@mock.patch.object(subprocess._subprocess, "Popen")
def test_check_call(mock_call):
    mock_call.return_value = _MockProcess()
    op = subprocess.check_call(["ls", "-l"], shell=True, env={"a": "b"}, cwd="/tmp")
    assert op == 0
    mock_call.assert_called()
    assert mock_call.call_args[0][0] == ["ls", "-l"]
    assert mock_call.call_args[1]["shell"] is True
    assert mock_call.call_args[1]["env"] == {"a": "b"}
    assert mock_call.call_args[1]["cwd"] == "/tmp"


@mock.patch.object(subprocess._subprocess, "Popen")
def test_check_call_shellex(mock_call):
    mock_call.return_value = _MockProcess()
    op = subprocess.check_call("ls -l", shell=True, env={"a": "b"}, cwd="/tmp")
    assert op == 0
    assert op == 0
    mock_call.assert_called()
    assert mock_call.call_args[0][0] == ["ls", "-l"]
    assert mock_call.call_args[1]["shell"] is True
    assert mock_call.call_args[1]["env"] == {"a": "b"}
    assert mock_call.call_args[1]["cwd"] == "/tmp"
