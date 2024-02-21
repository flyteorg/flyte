import pandas
import pytest
from flytekitplugins.dolt.schema import DoltConfig, DoltTable, DoltTableNameTransformer
from google.protobuf.struct_pb2 import Struct

from flytekit.models.literals import Literal, Scalar


def test_dolt_table_to_python_value(mocker):
    mocker.patch("dolt_integrations.core.save", return_value=True)
    table = DoltTable(data=pandas.DataFrame(), config=DoltConfig(db_path="p"))

    lv = DoltTableNameTransformer.to_literal(
        self=None,
        ctx=None,
        python_val=table,
        python_type=DoltTable,
        expected=None,
    )

    assert lv.scalar.generic["config"]["branch_conf"] is None


def test_dolt_table_to_literal(mocker):
    df = pandas.DataFrame()
    mocker.patch("dolt_integrations.core.load", return_value=None)
    mocker.patch("doltcli.Dolt", return_value=None)
    mocker.patch("pandas.read_csv", return_value=df)

    s = Struct()
    s.update({"config": {"db_path": "", "tablename": "t"}})
    lv = Literal(Scalar(generic=s))

    res = DoltTableNameTransformer.to_python_value(
        self=None,
        ctx=None,
        lv=lv,
        expected_python_type=DoltTable,
    )

    assert res.data.equals(df)


def test_dolt_table_to_python_value_error():
    with pytest.raises(AssertionError):
        DoltTableNameTransformer.to_literal(
            self=None,
            ctx=None,
            python_val=None,
            python_type=DoltTable,
            expected=None,
        )


def test_dolt_table_to_literal_error():
    s = Struct()
    s.update({"dummy": "data"})
    lv = Literal(Scalar(generic=s))

    with pytest.raises(ValueError):
        DoltTableNameTransformer.to_python_value(
            self=None,
            ctx=None,
            lv=lv,
            expected_python_type=DoltTable,
        )
