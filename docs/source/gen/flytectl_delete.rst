.. _flytectl_delete:

flytectl delete
---------------

Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

Synopsis
~~~~~~~~



Example Delete executions.
::

 bin/flytectl delete execution kxd1i72850  -d development  -p flytesnacks


Options
~~~~~~~

::

  -h, --help   help for delete

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
* :doc:`flytectl_delete_cluster-resource-attribute` 	 - Deletes matchable resources of cluster attributes
* :doc:`flytectl_delete_execution` 	 - Terminate/Delete execution resources.
* :doc:`flytectl_delete_execution-cluster-label` 	 - Deletes matchable resources of execution cluster label
* :doc:`flytectl_delete_execution-queue-attribute` 	 - Deletes matchable resources of execution queue attributes
* :doc:`flytectl_delete_plugin-override` 	 - Deletes matchable resources of plugin overrides
* :doc:`flytectl_delete_task-resource-attribute` 	 - Deletes matchable resources of task attributes

