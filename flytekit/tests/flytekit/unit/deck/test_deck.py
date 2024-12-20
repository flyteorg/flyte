import datetime
import sys

import pytest
from mock import mock

import flytekit
from flytekit import Deck, FlyteContextManager, task
from flytekit.deck import TopFrameRenderer
from flytekit.deck.deck import _output_deck


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_deck():
    import pandas as pd

    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [1, 22]})
    ctx = FlyteContextManager.current_context()
    ctx.user_space_params._decks = [ctx.user_space_params.default_deck]
    renderer = TopFrameRenderer()
    deck_name = "test"
    deck = Deck(deck_name)
    deck.append(renderer.to_html(df))
    assert deck.name == deck_name
    assert deck.html is not None
    assert len(ctx.user_space_params.decks) == 2

    _output_deck("test_task", ctx.user_space_params)


def test_timeline_deck():
    time_info = dict(
        Name="foo",
        Start=datetime.datetime.utcnow(),
        Finish=datetime.datetime.utcnow() + datetime.timedelta(microseconds=1000),
        WallTime=1.0,
        ProcessTime=1.0,
    )
    ctx = FlyteContextManager.current_context()
    ctx.user_space_params._decks = []
    timeline_deck = ctx.user_space_params.timeline_deck
    timeline_deck.append_time_info(time_info)
    assert timeline_deck.name == "timeline"
    assert len(timeline_deck.time_info) == 1
    assert timeline_deck.time_info[0] == time_info
    assert len(ctx.user_space_params.decks) == 1


@pytest.mark.parametrize(
    "disable_deck,expected_decks",
    [
        (None, 1),  # time line deck
        (False, 3),  # time line deck + input and output decks
        (True, 1),  # time line deck
    ],
)
def test_deck_for_task(disable_deck, expected_decks):
    ctx = FlyteContextManager.current_context()

    kwargs = {}
    if disable_deck is not None:
        kwargs["disable_deck"] = disable_deck

    @task(**kwargs)
    def t1(a: int) -> str:
        return str(a)

    t1(a=3)
    assert len(ctx.user_space_params.decks) == expected_decks


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
@pytest.mark.filterwarnings("ignore:disable_deck was deprecated")
@pytest.mark.parametrize(
    "enable_deck,disable_deck, expected_decks, expect_error",
    [
        (None, None, 2, False),  # default deck and time line deck
        (None, False, 4, False),  # default deck and time line deck + input and output decks
        (None, True, 2, False),  # default deck and time line deck
        (True, None, 4, False),  # default deck and time line deck + input and output decks
        (False, None, 2, False),  # default deck and time line deck
        (True, True, -1, True),  # Set both disable_deck and enable_deck to True and confirm that it fails
        (False, False, -1, True),  # Set both disable_deck and enable_deck to False and confirm that it fails
    ],
)
def test_deck_pandas_dataframe(enable_deck, disable_deck, expected_decks, expect_error):
    import pandas as pd

    ctx = FlyteContextManager.current_context()

    kwargs = {}
    if disable_deck is not None:
        kwargs["disable_deck"] = disable_deck

    if enable_deck is not None:
        kwargs["enable_deck"] = enable_deck

    if expect_error:
        with pytest.raises(ValueError):

            @task(**kwargs)
            def t_df(a: str) -> int:
                df = pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})
                flytekit.current_context().default_deck.append(TopFrameRenderer().to_html(df))
                return int(a)

    else:

        @task(**kwargs)
        def t_df(a: str) -> int:
            df = pd.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})
            flytekit.current_context().default_deck.append(TopFrameRenderer().to_html(df))
            return int(a)

        t_df(a="42")
        assert len(ctx.user_space_params.decks) == expected_decks


def test_deck_deprecation_warning_disable_deck():
    warn_msg = "disable_deck was deprecated in 1.10.0, please use enable_deck instead"
    with pytest.warns(FutureWarning, match=warn_msg):

        @task(disable_deck=False)
        def a():
            pass


@mock.patch("flytekit.deck.deck.ipython_check")
def test_deck_in_jupyter(mock_ipython_check):
    mock_ipython_check.return_value = True

    ctx = FlyteContextManager.current_context()
    ctx.user_space_params._decks = [ctx.user_space_params.default_deck]
    v = ctx.get_deck()
    from IPython.core.display import HTML

    assert isinstance(v, HTML)
    _output_deck("test_task", ctx.user_space_params)

    @task()
    def t1(a: int) -> str:
        return str(a)

    with flytekit.new_context() as ctx:
        t1(a=3)
        deck = ctx.get_deck()
        assert deck is not None


def test_get_deck():
    html = "你好，Flyte"
    ctx = FlyteContextManager.current_context()
    ctx.user_space_params._decks = [ctx.user_space_params.default_deck]
    ctx.user_space_params._decks[0] = flytekit.Deck("test", html)
    _output_deck("test_task", ctx.user_space_params)
