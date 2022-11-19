"""
Running a Task
--------------

Flytectl
========

This is a multi-step process where we create an execution spec file, update the spec file, and then create the execution.
More details can be found in the `Flytectl API reference <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__.

**Generate execution spec file** ::

    flytectl get tasks -d development -p flytesnacks workflows.example.generate_normal_df  --latest --execFile exec_spec.yaml

**Update the input spec file for arguments to the workflow** ::

            iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
            inputs:
              n: 200
              mean: 0.0
              sigma: 1.0
            kubeServiceAcct: ""
            targetDomain: ""
            targetProject: ""
            task: workflows.example.generate_normal_df
            version: "v1"

**Create execution using the exec spec file** ::

    flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml


**Monitor the execution by providing the execution id from create command** ::

    flytectl get execution -p flytesnacks -d development <execid>


FlyteRemote
===========

A task can be launched via FlyteRemote programmatically.

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config, SerializationSettings

    # FlyteRemote object is the main entrypoint to API
    remote = FlyteRemote(
        config=Config.for_endpoint(endpoint="flyte.example.net"),
        default_project="flytesnacks",
        default_domain="development",
    )

    # Get Task
    flyte_task = remote.fetch_task(name="workflows.example.generate_normal_df", version="v1")

    flyte_task = remote.register_task(
        entity=flyte_task,
        serialization_settings=SerializationSettings(image_config=None),
        version="v2",
    )

    # Run Task
    execution = remote.execute(
         flyte_task, inputs={"n": 200, "mean": 0.0, "sigma": 1.0}, execution_name="task_execution", wait=True
    )

    # Inspecting execution
    # The 'inputs' and 'outputs' correspond to the task execution.
    input_keys = execution.inputs.keys()
    output_keys = execution.outputs.keys()

"""
