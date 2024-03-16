from flytekit.models import filters


def test_eq_filter():
    assert filters.Equal("key", "value").to_flyte_idl() == "eq(key,value)"


def test_neq_filter():
    assert filters.NotEqual("key", "value").to_flyte_idl() == "ne(key,value)"


def test_gt_filter():
    assert filters.GreaterThan("key", "value").to_flyte_idl() == "gt(key,value)"


def test_gte_filter():
    assert filters.GreaterThanOrEqual("key", "value").to_flyte_idl() == "gte(key,value)"


def test_lt_filter():
    assert filters.LessThan("key", "value").to_flyte_idl() == "lt(key,value)"


def test_lte_filter():
    assert filters.LessThanOrEqual("key", "value").to_flyte_idl() == "lte(key,value)"


def test_value_in_filter():
    assert filters.ValueIn("key", ["1", "2", "3"]).to_flyte_idl() == "value_in(key,1;2;3)"


def test_contains_filter():
    assert filters.Contains("key", ["1", "2", "3"]).to_flyte_idl() == "contains(key,1;2;3)"


def test_filter_list():
    fl = filters.FilterList([filters.Equal("domain", "staging"), filters.NotEqual("project", "FakeProject")])

    assert fl.to_flyte_idl() == "eq(domain,staging)+ne(project,FakeProject)"
