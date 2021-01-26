from flytekit.sdk.test_utils import flyte_test
from flytekit.sdk.types import Types

from multi_step_linear import diabetes_xgboost as dxgb


@flyte_test
def test_DiabetesXGBoostModelTrainer():
    """
    This shows how each task can be unit tested and yet changed into the workflow.
    TODO: Have one test to run the entire workflow end to end.
    """

    dataset = Types.CSV.create_at_known_location(
        "https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv")

    # Test get dataset
    result = dxgb.get_traintest_splitdatabase.unit_test(dataset=dataset, seed=7, test_split_ratio=0.33)
    assert "x_train" in result
    assert "y_train" in result
    assert "x_test" in result
    assert "y_test" in result

    # Test fit
    m = dxgb.fit.unit_test(x=result["x_train"], y=result["y_train"], hyperparams=dxgb.XGBoostModelHyperparams(max_depth=4).to_dict())

    assert "model" in m

    p = dxgb.predict.unit_test(x=result["x_test"], model_ser=m["model"])

    assert "predictions" in p

    metric = dxgb.metrics.unit_test(predictions=p["predictions"], y=result["y_test"])

    assert "accuracy" in metric

    print(metric["accuracy"])
    assert metric["accuracy"] * 100.0 > 75.0
