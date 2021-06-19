.. _flytectl_create_execution:

flytectl create execution
-------------------------

Create execution resources

Synopsis
~~~~~~~~



Create the executions for given workflow/task in a project and domain.

There are three steps in generating an execution.

- Generate the execution spec file using the get command.
- Update the inputs for the execution if needed.
- Run the execution by passing in the generated yaml file.

The spec file should be generated first and then run the execution using the spec file.
You can reference the flytectl get task for more details

::

 flytectl get tasks -d development -p flytectldemo core.advanced.run_merge_sort.merge  --version v2 --execFile execution_spec.yaml

The generated file would look similar to this

.. code-block:: yaml

	 iamRoleARN: ""
	 inputs:
	   sorted_list1:
	   - 0
	   sorted_list2:
	   - 0
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 task: core.advanced.run_merge_sort.merge
	 version: "v2"


The generated file can be modified to change the input values.

.. code-block:: yaml

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
	 version: "v2"

And then can be passed through the command line.
Notice the source and target domain/projects can be different.
The root project and domain flags of -p and -d should point to task/launch plans project/domain.

::

 flytectl create execution --execFile execution_spec.yaml -p flytectldemo -d development --targetProject flytesnacks

Also an execution can be relaunched by passing in current execution id.

::

 flytectl create execution --relaunch ffb31066a0f8b4d52b77 -p flytectldemo -d development

Generic data types are also supported for execution in similar way.Following is sample of how the inputs need to be specified while creating the execution.
As usual the spec file should be generated first and then run the execution using the spec file.

::

 flytectl get task -d development -p flytectldemo  core.type_system.custom_objects.add --execFile adddatanum.yaml

The generated file would look similar to this. Here you can see empty values dumped for generic data type x and y. 

::

    iamRoleARN: ""
    inputs:
      "x": {}
      "y": {}
    kubeServiceAcct: ""
    targetDomain: ""
    targetProject: ""
    task: core.type_system.custom_objects.add
    version: v3

Modified file with struct data populated for x and y parameters for the task core.type_system.custom_objects.add

::

  iamRoleARN: "arn:aws:iam::123456789:role/dummy"
  inputs:
    "x":
      "x": 2
      "y": ydatafory
      "z":
        1 : "foo"
        2 : "bar"
    "y":
      "x": 3
      "y": ydataforx
      "z":
        3 : "buzz"
        4 : "lightyear"
  kubeServiceAcct: ""
  targetDomain: ""
  targetProject: ""
  task: core.type_system.custom_objects.add
  version: v3

Usage


::

  flytectl create execution [flags]

Options
~~~~~~~

::

      --execFile string          file for the execution params.If not specified defaults to <<workflow/task>_name>.execution_spec.yaml
  -h, --help                     help for execution
      --iamRoleARN string        iam role ARN AuthRole for launching execution.
      --kubeServiceAcct string   kubernetes service account AuthRole for launching execution.
      --relaunch string          execution id to be relaunched.
      --targetDomain string      project where execution needs to be created.If not specified configured domain would be used.
      --targetProject string     project where execution needs to be created.If not specified configured project would be used.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_create` 	 - Used for creating various flyte resources including tasks/workflows/launchplans/executions/project.

