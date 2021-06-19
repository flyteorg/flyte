.. _flytectl_create_project:

flytectl create project
-----------------------

Create project resources

Synopsis
~~~~~~~~



Create the projects.(project,projects can be used interchangeably in these commands)

::

 bin/flytectl create project --name flytesnacks --id flytesnacks --description "flytesnacks description"  --labels app=flyte

Create Project by definition file. Note: The name shouldn't contain any whitespace characters'
::

 bin/flytectl create project --file project.yaml 

.. code-block:: yaml

    id: "project-unique-id"
    name: "Name"
    labels:
     app: flyte
    description: "Some description for the project"



::

  flytectl create project [flags]

Options
~~~~~~~

::

      --description string      description for the project specified as argument.
      --file string             file for the project definition.
  -h, --help                    help for project
      --id string               id for the project specified as argument.
      --labels stringToString   labels for the project specified as argument. (default [])
      --name string             name for the project specified as argument.

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

