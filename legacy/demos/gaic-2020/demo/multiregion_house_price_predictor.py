import os

from flytekit.sdk.tasks import python_task, inputs, outputs, dynamic_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output

from demo.house_price_predictor import generate_data, save_to_file, save_to_dir, fit, predict


@inputs(locations=Types.List(Types.String), number_of_houses_per_location=Types.Integer, seed=Types.Integer)
@outputs(train=Types.List(Types.MultiPartCSV), val=Types.List(Types.MultiPartCSV), test=Types.List(Types.CSV))
@python_task(cache=True, cache_version="0.1", memory_request="200Mi")
def generate_and_split_data_multiloc(wf_params, locations, number_of_houses_per_location, seed, train, val, test):
    train_sets = []
    val_sets = []
    test_sets = []
    for loc in locations:
        _train, _val, _test = generate_data(loc, number_of_houses_per_location, seed)
        dir = "multi_data"
        os.makedirs(dir, exist_ok=True)
        train_sets.append(save_to_dir(dir, "train", _train))
        val_sets.append(save_to_dir(dir, "val", _val))
        test_sets.append(save_to_file(dir, "test", _test))
    train.set(train_sets)
    val.set(val_sets)
    test.set(test_sets)


@inputs(multi_train=Types.List(Types.MultiPartCSV))
@outputs(multi_models=Types.List(Types.Blob))
@dynamic_task(cache=True, cache_version="0.1", memory_request="200Mi")
def parallel_fit(wf_params, multi_train, multi_models):
    models = []
    for train in multi_train:
        t = fit(train=train)
        yield t
        models.append(t.outputs.model)
    multi_models.set(models)


@inputs(multi_test=Types.List(Types.CSV), multi_models=Types.List(Types.Blob))
@outputs(predictions=Types.List(Types.List(Types.Float)), accuracies=Types.List(Types.Float))
@dynamic_task(cache_version='1.1', cache=True, memory_request="200Mi")
def parallel_predict(wf_params, multi_test, multi_models, predictions, accuracies):
    preds = []
    accs = []
    for test, model in zip(multi_test, multi_models):
        p = predict(test=test, model_ser=model)
        yield p
        preds.append(p.outputs.predictions)
        accs.append(p.outputs.accuracy)
    predictions.set(preds)
    accuracies.set(accs)


@workflow_class
class MultiRegionHousePricePredictionModelTrainer(object):
    """
    This pipeline trains an XGBoost model, also generated synthetic data and runs predictions against test dataset
    """
    regions = Input(Types.List(Types.String), default=["SFO", "SEA", "DEN"],
                    help="Regions for where to train the model.")
    seed = Input(Types.Integer, default=7, help="Seed to use for splitting.")
    num_houses_per_region = Input(Types.Integer, default=1000,
                                  help="Number of houses to generate data for in each region")

    # the actual algorithm
    split = generate_and_split_data_multiloc(locations=regions, number_of_houses_per_location=num_houses_per_region,
                                             seed=seed)
    fit_task = parallel_fit(multi_train=split.outputs.train)
    predicted = parallel_predict(multi_models=fit_task.outputs.multi_models, multi_test=split.outputs.test)

    # Outputs: joblib seralized models per region and accuracy of the model per region
    # Note we should make this into a map, but for demo we will output a simple list
    models = Output(fit_task.outputs.multi_models, sdk_type=Types.List(Types.Blob))
    accuracies = Output(predicted.outputs.accuracies, sdk_type=Types.List(Types.Float))
