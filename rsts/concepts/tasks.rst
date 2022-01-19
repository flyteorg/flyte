.. _divedeep-tasks:

Tasks
======

Tasks are fully independent units of execution and first-class entities of Flyte. 
They are the fundamental building blocks and extension points that encapsulate the users' code.

Characteristics
---------------

In general, a Flyte task is characterized by:

1. A combination of :ref:`divedeep-projects` and :ref:`divedeep-domains`,
2. A unique unicode name (we recommend it not to exceed 32 characters),
3. A version string, and/or
4. *Optional* Task interface definition

   For tasks to exchange data with each other, a task can define a signature (much like a function/method
   signature in programming languages). A task interface defines the input and output variables —
   :std:ref:`variablesentry <flyteidl:protos/docs/core/core:variablemap.variablesentry>`
   and their types, :std:ref:`literaltype <flyteidl:protos/docs/core/core:literaltype>`.

Can "X" Be a Flyte Task?
-------------------------

When deciding whether or not a unit of execution constitutes a Flyte task, consider the following:

- Is there a well-defined graceful/successful exit criteria for the task? A task is expected to exit after finishing processing
  its inputs.
- Is it repeatable? Under certain circumstances, a task might be retried, rerun, etc. with the same inputs. It's expected
  to produce the same outputs every single time. For example, avoid using random number generators with current clock as seed
  and instead use a system-provided clock as the seed. 
- Is it a pure function, i.e., does it have side effects that are not known to the system (e.g. calls a web-service)? It's strongly
  advisable to avoid side-effects in tasks. When side-effects are required, ensure that those operations are idempotent.

Dynamic Tasks
--------------

"Dynamic tasks" is a misnomer. 
Flyte is one-of-a-kind workflow engine that ships with the concept of truly `Dynamic Workflows <https://blog.flyte.org/dynamic-workflows-in-flyte>`__!
Users can generate workflows in reaction to user inputs or computed values at runtime. 
These executions are evaluated to generate a static graph, before execution.

Extending Task
---------------

Plugins
^^^^^^^

Flyte language exposes an extensible model to express tasks in an execution-independent language. 
It contains first-class task plugins (e.g. `Papermill <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-papermill/flytekitplugins/papermill/task.py>`__, 
`Great Expectations <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-greatexpectations/flytekitplugins/great_expectations/task.py>`__, etc.) 
that take care of executing the Flyte tasks. 
Almost any action can be implemented and introduced into Flyte as a "Plugin".

- Tasks that run queries on distributed data warehouses like Redshift, Hive, Snowflake, etc.
- Tasks that run executions on compute engines like Spark, Flink, AWS Sagemaker, AWS Batch, Kubernetes pods, jobs, etc.
- Tasks that call web services

Flyte ships with some defaults; for example, running a simple Python function does not need any hosted service. Flyte knows how to
execute these kinds of tasks on Kubernetes. It turns out these are the vast majority of tasks in ML, and Flyte is adept at 
handling an enormous scale on Kubernetes; this is achieved by implementing a unique scheduler on top of Kubernetes.

Types
^^^^^

It is impossible to define the unit of execution of a task in the same way for all kinds of tasks. Hence, Flyte allows different task
types in the system. Flyte comes with a set of defined, battle-tested task types. It also allows for a very flexible model to
:std:ref:`define new types <cookbook:plugins_extend>`.

Inherent Features
-----------------

Fault tolerance
^^^^^^^^^^^^^^^

In any distributed system, failure is inevitable. Allowing users to design a fault-tolerant system (e.g. workflow) is an inherent goal of Flyte. 
At a high level, tasks offer two parameters to achieve fault tolerance:

**Retries**
  Tasks can define a retry strategy to let the system know how to handle failures (example: retry 3 times on any kind of error).

**Timeouts**
  For the system to ensure it is always making progress, tasks must be guaranteed to end gracefully/successfully. The system defines a default timeout period for the tasks. It is also possible for task authors to define a timeout period, after which the task gets marked as failure. Note that a timed-out task will be retried if it has a retry strategy defined.

Memoization
^^^^^^^^^^^

Flyte supports memoization of task outputs to ensure that identical invocations of a task don't get executed repeatedly, wasting compute resources.
For more information on memoization, please refer to the :std:ref:`Caching Example <cookbook:sphx_glr_auto_core_flyte_basics_task_cache.py>`.
