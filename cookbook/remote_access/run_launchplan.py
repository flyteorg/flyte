"""
Running a Launchplan
--------------------

This is multi-steps process where we create an execution spec file, update the spec file and then create the execution.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__.

**Generate an execution spec file** ::

    flytectl get launchplan -p flytesnacks -d development myapp.workflows.example.my_wf  --execFile exec_spec.yaml

**Update the input spec file for arguments to the workflow** ::

    ....
    inputs:
        name: "adam"
    ....

**Create execution using the exec spec file** ::

    flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml


**Monitor the execution by providing the execution id from create command** ::

    flytectl get execution -p flytesnacks -d development <execid>

"""
