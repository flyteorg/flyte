from core.intermediate import custom_objects


def test_custom_objects():
    res = custom_objects.wf(x=10, y=20)
    assert res == custom_objects.Datum(x=30, y="1020", z={10: "10", 20: "20"})
