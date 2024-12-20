from flytekitplugins.awssagemaker.models import hpo_job, training_job


def test_hyperparameter_tuning_objective():
    obj = hpo_job.HyperparameterTuningObjective(
        objective_type=hpo_job.HyperparameterTuningObjectiveType.MAXIMIZE, metric_name="test_metric"
    )
    obj2 = hpo_job.HyperparameterTuningObjective.from_flyte_idl(obj.to_flyte_idl())

    assert obj == obj2


def test_hyperparameter_job_config():
    jc = hpo_job.HyperparameterTuningJobConfig(
        tuning_strategy=hpo_job.HyperparameterTuningStrategy.BAYESIAN,
        tuning_objective=hpo_job.HyperparameterTuningObjective(
            objective_type=hpo_job.HyperparameterTuningObjectiveType.MAXIMIZE, metric_name="test_metric"
        ),
        training_job_early_stopping_type=hpo_job.TrainingJobEarlyStoppingType.AUTO,
    )

    jc2 = hpo_job.HyperparameterTuningJobConfig.from_flyte_idl(jc.to_flyte_idl())
    assert jc2.tuning_strategy == jc.tuning_strategy
    assert jc2.tuning_objective == jc.tuning_objective
    assert jc2.training_job_early_stopping_type == jc.training_job_early_stopping_type


def test_hyperparameter_tuning_job():
    rc = training_job.TrainingJobResourceConfig(
        instance_type="test_type",
        instance_count=10,
        volume_size_in_gb=25,
        distributed_protocol=training_job.DistributedProtocol.MPI,
    )
    alg = training_job.AlgorithmSpecification(
        algorithm_name=training_job.AlgorithmName.CUSTOM,
        algorithm_version="",
        input_mode=training_job.InputMode.FILE,
        input_content_type=training_job.InputContentType.TEXT_CSV,
    )
    tj = training_job.TrainingJob(
        training_job_resource_config=rc,
        algorithm_specification=alg,
    )
    hpo = hpo_job.HyperparameterTuningJob(max_number_of_training_jobs=10, max_parallel_training_jobs=5, training_job=tj)

    hpo2 = hpo_job.HyperparameterTuningJob.from_flyte_idl(hpo.to_flyte_idl())

    assert hpo.max_number_of_training_jobs == hpo2.max_number_of_training_jobs
    assert hpo.max_parallel_training_jobs == hpo2.max_parallel_training_jobs
    assert (
        hpo2.training_job.training_job_resource_config.instance_type
        == hpo.training_job.training_job_resource_config.instance_type
    )
    assert (
        hpo2.training_job.training_job_resource_config.instance_count
        == hpo.training_job.training_job_resource_config.instance_count
    )
    assert (
        hpo2.training_job.training_job_resource_config.distributed_protocol
        == hpo.training_job.training_job_resource_config.distributed_protocol
    )
    assert (
        hpo2.training_job.training_job_resource_config.volume_size_in_gb
        == hpo.training_job.training_job_resource_config.volume_size_in_gb
    )
    assert (
        hpo2.training_job.algorithm_specification.algorithm_name
        == hpo.training_job.algorithm_specification.algorithm_name
    )
    assert (
        hpo2.training_job.algorithm_specification.algorithm_version
        == hpo.training_job.algorithm_specification.algorithm_version
    )
    assert hpo2.training_job.algorithm_specification.input_mode == hpo.training_job.algorithm_specification.input_mode
    assert (
        hpo2.training_job.algorithm_specification.input_content_type
        == hpo.training_job.algorithm_specification.input_content_type
    )
