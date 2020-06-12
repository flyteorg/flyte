.. _recipe-6:

###############################################
How do I write a simple Presto Task?
###############################################

Unfortunately, while the examples here technically do work, there is not currently an open-source Presto client. The one that ships with Flyte plugins is a `no-op client. <https://github.com/lyft/flyteplugins/blob/master/go/tasks/plugins/presto/client/noop_presto_client.go>`__

***********
Presto Task
***********
The Presto task is fundamentally different than the other tasks that we have in that there is no ``@`` decorator. There is nothing to decorate, because there is no container being run during the execution of a Presto task, just the query on a Presto cluster.

Declaration
============

The arguments to a Presto task are the following:

  * ``task_inputs``
    These are the inputs to the task. In the example, the ``inputs()`` helper function usually used to decorate a ``@python_task`` is used to return the correct types.

  * ``statement``
    This is where your select statement should go. This query is interpolated. For instance, you can do ``SELECT * FROM table WHERE ds='{{ .Inputs.ds }}'`` if ``ds`` was declared as an input.

  * ``output_schema``
    This is a Flytekit ``Schema`` object, not to be confused with the Presto schema in which the query will run.

  * ``routing_group``
    This is the Presto routing group.

  * ``catalog``
    This is the Presto catalog, and can be omitted if included in the query.

  * ``schema``
    This is the Presto schema, and can be omitted if included in the query.

With respect to cataloging, please note that there are three implicit inputs to the Presto task. These are the routing group, catalog, and schema. That is, if the task is declared with ``discoverable=True``, then the catalog used forms part of the input of the task. If you run the task once, and it gets saved, and you change the catalog, but not the query statement, Flyte will not use the previously stored results.

Outputs
========
The Presto task always has exactly one output under the name ``results``.

Please see the included example `presto.py <presto.py>`__ for a complete description.
