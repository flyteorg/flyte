.. _flytectl_delete_execution:

flytectl delete execution
-------------------------

Terminate/Delete execution resources.

Synopsis
~~~~~~~~



Terminate executions.(execution,executions can be used interchangeably in these commands)

Task executions can be aborted only if they are in non-terminal state i.e if they are FAILED,ABORTED or SUCCEEDED then
calling terminate on them has no effect.

Terminate a single execution with its name

::

 bin/flytectl delete execution c6a51x2l9e  -d development  -p flytesnacks

You can get executions to check its state.

::

 bin/flytectl get execution  -d development  -p flytesnacks
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | NAME (7)   | WORKFLOW NAME                                                           | TYPE     | PHASE     | STARTED                        | ELAPSED TIME  |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | c6a51x2l9e | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:13:04.680476300Z | 15.540361300s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 

Terminate multiple executions with there names
::

 bin/flytectl delete execution eeam9s8sny p4wv4hwgc4  -d development  -p flytesnacks

Similarly you can get executions to find the state of previously terminated executions.

::

 bin/flytectl get execution  -d development  -p flytesnacks
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | NAME (7)   | WORKFLOW NAME                                                           | TYPE     | PHASE     | STARTED                        | ELAPSED TIME  |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | c6a51x2l9e | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:13:04.680476300Z | 15.540361300s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | eeam9s8sny | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:14:04.803084100Z | 42.306385500s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 
 | p4wv4hwgc4 | recipes.core.basic.lp.go_greet                                          | WORKFLOW | ABORTED   | 2021-02-17T08:14:27.476307400Z | 19.727504400s |
  ------------ ------------------------------------------------------------------------- ---------- ----------- -------------------------------- --------------- 

Usage


::

  flytectl delete execution [flags]

Options
~~~~~~~

::

  -h, --help   help for execution

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config string    config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string    Specifies the Flyte project's domain.
  -o, --output string    Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string   Specifies the Flyte project.

SEE ALSO
~~~~~~~~

* :doc:`flytectl_delete` 	 - Used for terminating/deleting various flyte resources including tasks/workflows/launchplans/executions/project.

