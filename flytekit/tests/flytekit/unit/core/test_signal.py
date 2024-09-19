import pytest
from flyteidl.admin.signal_pb2 import Signal, SignalList
from mock import MagicMock

from flytekit.configuration import Config
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.models.core.identifier import SignalIdentifier, WorkflowExecutionIdentifier
from flytekit.remote.remote import FlyteRemote


@pytest.fixture
def remote():
    flyte_remote = FlyteRemote(config=Config.auto(), default_project="p1", default_domain="d1")
    flyte_remote._client_initialized = True
    return flyte_remote


def test_remote_list_signals(remote):
    ctx = FlyteContextManager.current_context()
    wfeid = WorkflowExecutionIdentifier("p", "d", "execid")
    signal_id = SignalIdentifier(signal_id="sigid", execution_id=wfeid).to_flyte_idl()
    lt = TypeEngine.to_literal_type(int)
    signal = Signal(
        id=signal_id,
        type=lt.to_flyte_idl(),
        value=TypeEngine.to_literal(ctx, 3, int, lt).to_flyte_idl(),
    )

    mock_client = MagicMock()
    mock_client.list_signals.return_value = SignalList(signals=[signal], token="")

    remote._client = mock_client
    res = remote.list_signals("execid", "p", "d", limit=10)
    assert len(res) == 1


def test_remote_set_signal(remote):
    mock_client = MagicMock()

    def checker(request):
        assert request.id.signal_id == "sigid"
        assert request.value.scalar.primitive.integer == 3

    mock_client.set_signal.side_effect = checker

    remote._client = mock_client
    remote.set_signal("sigid", "execid", 3)
