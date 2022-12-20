from types.generic import generic_type_task


def test_generic_task():
    in1 = {
        "a": "hello",
        "b": "how are you",
        "c": ["array"],
        "d": {"nested": "value"},
    }
    output = generic_type_task.unit_test(custom=in1)
    assert output["counts"] == {
        "a": 5,
        "b": 11,
        "c": ["array"],
        "d": {"nested": "value"},
    }
