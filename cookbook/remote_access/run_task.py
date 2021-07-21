"""
Running a Task
--------------------

This is multi-steps process as well where we create an execution spec file, update the spec file and then create the execution.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__

**Generate execution spec file** ::

    flytectl get tasks -d development -p flytectldemo core.advanced.run_merge_sort.merge  --latest --execFile exec_spec.yaml

**Update the input spec file for arguments to the workflow** ::

            iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
            inputs:
              sorted_list1:
              - 2
              - 4
              - 6
              sorted_list2:
              - 1
              - 3
              - 5
            kubeServiceAcct: ""
            targetDomain: ""
            targetProject: ""
            task: core.advanced.run_merge_sort.merge
            version: "v1"

**Create execution using the exec spec file** ::

    flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml


**Monitor the execution by providing the execution id from create command** ::

    flytectl get execution -p flytesnacks -d development <execid>

"""
