import unittest

from flytekitplugins.awssagemaker.models import training_job


def test_training_job_resource_config():
    rc = training_job.TrainingJobResourceConfig(
        instance_count=1,
        instance_type="random.instance",
        volume_size_in_gb=25,
        distributed_protocol=training_job.DistributedProtocol.MPI,
    )

    rc2 = training_job.TrainingJobResourceConfig.from_flyte_idl(rc.to_flyte_idl())
    assert rc2 == rc
    assert rc2.distributed_protocol == training_job.DistributedProtocol.MPI
    assert rc != training_job.TrainingJobResourceConfig(
        instance_count=1,
        instance_type="random.instance",
        volume_size_in_gb=25,
        distributed_protocol=training_job.DistributedProtocol.UNSPECIFIED,
    )

    assert rc != training_job.TrainingJobResourceConfig(
        instance_count=1,
        instance_type="oops",
        volume_size_in_gb=25,
        distributed_protocol=training_job.DistributedProtocol.MPI,
    )


def test_metric_definition():
    md = training_job.MetricDefinition(name="test-metric", regex="[a-zA-Z]*")

    md2 = training_job.MetricDefinition.from_flyte_idl(md.to_flyte_idl())
    assert md == md2
    assert md2.name == "test-metric"
    assert md2.regex == "[a-zA-Z]*"


def test_algorithm_specification():
    case = unittest.TestCase()
    alg_spec = training_job.AlgorithmSpecification(
        algorithm_name=training_job.AlgorithmName.CUSTOM,
        algorithm_version="v100",
        input_mode=training_job.InputMode.FILE,
        metric_definitions=[training_job.MetricDefinition(name="a", regex="b")],
        input_content_type=training_job.InputContentType.TEXT_CSV,
    )

    alg_spec2 = training_job.AlgorithmSpecification.from_flyte_idl(alg_spec.to_flyte_idl())

    assert alg_spec2.algorithm_name == training_job.AlgorithmName.CUSTOM
    assert alg_spec2.algorithm_version == "v100"
    assert alg_spec2.input_mode == training_job.InputMode.FILE
    case.assertCountEqual(alg_spec.metric_definitions, alg_spec2.metric_definitions)
    assert alg_spec == alg_spec2


def test_training_job():
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

    tj2 = training_job.TrainingJob.from_flyte_idl(tj.to_flyte_idl())
    # checking tj == tj2 would return false because we don't have the __eq__ magic method defined
    assert tj.training_job_resource_config.instance_type == tj2.training_job_resource_config.instance_type
    assert tj.training_job_resource_config.instance_count == tj2.training_job_resource_config.instance_count
    assert tj.training_job_resource_config.distributed_protocol == tj2.training_job_resource_config.distributed_protocol
    assert tj.training_job_resource_config.volume_size_in_gb == tj2.training_job_resource_config.volume_size_in_gb
    assert tj.algorithm_specification.algorithm_name == tj2.algorithm_specification.algorithm_name
    assert tj.algorithm_specification.algorithm_version == tj2.algorithm_specification.algorithm_version
    assert tj.algorithm_specification.input_mode == tj2.algorithm_specification.input_mode
    assert tj.algorithm_specification.input_content_type == tj2.algorithm_specification.input_content_type
