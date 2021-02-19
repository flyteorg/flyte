from native_typing import simple


def test_wf():
    x, y = simple.t1(a=5)
    assert x == 7
