.. _concepts-tasks:

Tasks
=====

Definition
----------

Tasks encapsulate fully independent units of execution. Flyte language exposes an extensible model
to express that in an execution-independent language. Flyte contains first class task plugins that take
care of executing these tasks.

Real world examples
-------------------

Query a data store
  Using :ref:`Hive tasks <hive-task-type>` to retrieve data into dataframes so that subsequent tasks can process them.

Transform data
  Using :ref:`Container tasks <container-task-type>` to transform data collected/computed earlier, users can develop a task as
  a simple Lambda function with specified inputs and outputs represented as a container with entrypoints, with specific compute, memory and
  gpu requirements.

Map-reduce massive jobs
  Using :ref:`Spark programs <container-task-type>` with their cluster configuration and compute requirements

Hyperparameter tuning task
  Using a system like Katib, users can execute a task that needs multiple iterations and leads to multiple other containers to execute.

A distributed or single container Tensorflow Job

Characteristics
---------------

In abstract, a task in the system is characterized by:

1. A :ref:`project and domain <concepts-projects>` combination,
2. A unique unicode name (we recommend it not to exceed 32 characters), and
3. A version string.
4. *Optional* Task interface definition

   In order for tasks to exchange data with each other, a task can define a signature (much like a function/method
   signature in programming languages). A task interface defines the :ref:`input and output variables <language-variables>` 
   as well as their :ref:`types <language-types>`.

Requirements
------------

When deciding whether a unit of execution conistitutes a Flyte Task or not. Consider the following:

- Is there a well-defined graceful/successful exit criteria for the task? A task is expected to exit after finishing processing
  its inputs.

- Is it repeatable? Under certain circumstances, a task might be retried, rerun... etc. with the same inputs. It's expected
  to produce the same outputs every single time. For example, avoid using random number generators with current clock as seed
  but opt to using a system-provided clock as the seed. 

- Is it a pure function? i.e. does it have side effects that are not known to the system (e.g. calls a web-service). It's strongly
  advisable to avoid side-effects in tasks. When side-effects are required, ensure that those operations are idempotent.

Types
-----
Since it's impossible to define the unit of execution of a task the same way for all kinds of tasks, Flyte allows different task
types in the system. Flyte comes with a set of defined, battle tested task types but also allows for a very flexible model to
introducing new :ref:`user-defined task types <contributor-extending>`. Read more about various supported task types :ref:`here <tasktypes>`.

Fault tolerance
---------------
In any distributed system failure is inevitable, allowing users to design a fault-tolerant system (e.g. workflow) is an inherent goal of Flyte. At a high level, tasks offer two parameters to control how to handle that:

Retries
  Tasks can define a retry strategy to let the system know how to handle failures (e.g. retry 3 times on any errors).

Timeouts
  In order for the system to ensure it's always making progress, tasks must be guaranteed to end. The system defines a default timeout
  period for tasks. It's also possible for task authors to define a timeout period after which the task is marked as failure. Note that
  a timed-out task will be retried if it has a retry strategy defined.

Memoization
-----------
Flyte supports memoization for task outputs to ensure identical invocations of a task are not repeatedly executed wasting compute resources.
For more information on memoization please refer to :ref:`Task Cache <features-task_cache>`.
