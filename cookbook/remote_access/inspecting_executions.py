"""
Inspecting Workflow and Task Executions
---------------------------------------

Inspecting workflow and task executions are done in the same manner as below. For more details see the
`flytectl API reference <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_get_execution.html>`__.

Monitor the execution by providing the execution id from create command which can be task or workflow execution. ::

    flytectl get execution -p flytesnacks -d development <execid>

For more details use ``--details`` flag which shows node executions along with task executions on them. ::

    flytectl get execution -p flytesnacks -d development <execid> --details

If you prefer to see yaml/json view for the details then change the output format using the -o flag. ::

    flytectl get execution -p flytesnacks -d development <execid> --details -o yaml

To see the results of the execution you can inspect the node closure outputUri in detailed yaml output. ::

    "outputUri": "s3://my-s3-bucket/metadata/propeller/flytesnacks-development-<execid>/n0/data/0/outputs.pb"

"""
