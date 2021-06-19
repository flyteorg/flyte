.. _flytectl_get:

flytectl get
------------

Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.

Synopsis
~~~~~~~~



Example get projects.
::

 flytectl get project


Options
~~~~~~~

::

  -h, --help   help for get

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
* :doc:`flytectl_get_cluster-resource-attribute` 	 - Gets matchable resources of cluster resource attributes
* :doc:`flytectl_get_execution` 	 - Gets execution resources
* :doc:`flytectl_get_execution-cluster-label` 	 - Gets matchable resources of execution cluster label
* :doc:`flytectl_get_execution-queue-attribute` 	 - Gets matchable resources of execution queue attributes
* :doc:`flytectl_get_launchplan` 	 - Gets launch plan resources
* :doc:`flytectl_get_plugin-override` 	 - Gets matchable resources of plugin override
* :doc:`flytectl_get_project` 	 - Gets project resources
* :doc:`flytectl_get_task` 	 - Gets task resources
* :doc:`flytectl_get_task-resource-attribute` 	 - Gets matchable resources of task attributes
* :doc:`flytectl_get_workflow` 	 - Gets workflow resources

