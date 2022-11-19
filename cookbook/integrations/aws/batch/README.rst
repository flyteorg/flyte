##########
AWS Batch
##########

.. tags:: Data, Integration, AWS, Advanced

Executing Batch Job
=======================
Flyte backend can be connected with batch. Once enabled, it allows you to run regular task on AWS batch.

This section provides a guide on how to use the AWS Batch Plugin using flytekit python.

Installation
------------

To use the flytekit batch plugin simply run the following:

.. prompt:: bash

    pip install flytekitplugins-awsbatch

Configuring the backend to get AWS Batch working
------------------------------------------------
[Follow this guide to setting up the AWS Batch Plugin](https://docs.flyte.org/en/latest/deployment/plugin_setup/aws/batch.html#deployment-plugin-setup-aws-array)


Quick Start
------------------------
This plugin allows you to run batch tasks on AWS and only requires you to change a few lines of code.
We can then move workflow execution from Kubernetes to AWS.

.. testcode:: awsbatch-quickstart
    from flytekitplugins.awsbatch import AWSBatchConfig

    config = AWSBatch(
        parameters={"codec": "mp4"},
        platformCapabilities="EC2",
        propagateTags=True,
        retryStrategy={"attempts": 10},
        tags={"hello": "world"},
        timeout={"attemptDurationSeconds": 60},
    )

    @task(task_config=config)
    def t1(a: int) -> str:
        return str(a)
