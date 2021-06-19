.. _flytectl_update:

flytectl update
---------------

Used for updating flyte resources eg: project.

Synopsis
~~~~~~~~



Currently this command only provides subcommands to update project.
Takes input project which need to be archived or unarchived. Name of the project to be updated is mandatory field.
Example update project to activate it.
::

 bin/flytectl update project -p flytesnacks --activateProject


Options
~~~~~~~

::

  -h, --help   help for update

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl` 	 - flyetcl CLI tool
* :doc:`flytectl_update_cluster-resource-attribute` 	 - Updates matchable resources of cluster attributes
* :doc:`flytectl_update_execution-cluster-label` 	 - Updates matchable resources of execution cluster label
* :doc:`flytectl_update_execution-queue-attribute` 	 - Updates matchable resources of execution queue attributes
* :doc:`flytectl_update_launchplan` 	 - Updates launch plan metadata
* :doc:`flytectl_update_plugin-override` 	 - Updates matchable resources of plugin overrides
* :doc:`flytectl_update_project` 	 - Updates project resources
* :doc:`flytectl_update_task` 	 - Updates task metadata
* :doc:`flytectl_update_task-resource-attribute` 	 - Updates matchable resources of task attributes
* :doc:`flytectl_update_workflow` 	 - Updates workflow metadata

