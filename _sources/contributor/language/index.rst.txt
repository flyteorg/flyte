.. _user-language:

Flyte Specification Language
============================

Flyte at the core consists of a specification language that all the components of the system can interact with. 
The specification language is written in `Protocol Buffers`_. This allows for a very portable language model. An implementation of the language
makes interaction with the system easier and brings the power of Flyte to the user. A full implementation of the language is provided in 
:ref:`Python SDK <user-sdk-python>` as well as generated :ref:`Golang types <user-sdk-go>`, C++ and JAVA.

This is a reference manual for Flyte Specification Language. For more information on how to use the provided SDK, please
refer to :ref:`SDK manual <user-sdk-python>`.

Elements
--------
.. _language-types:

Types
  Flyte Spec Language defines a flexible typing system. There are broadly three sets of types; primitives (e.g. ints, float, bool, string),
  collections (lists and maps) and use-case specific types (e.g. Schema, Blob, Struct).
  
  Collection types support any dimensionality (e.g. a list of map of list of strings).

  .. note::
    Flyte collection types are invariant. i.e. a list of Cat is not a subtype of list of Animal.

  For reference documentation on types, refer to :ref:`Types <api_file_flyteidl/core/types.proto>`

.. _language-literals:

Literals
  Literals are bound values for specific types. E.g. `5` is a literal of primitive type int with value `5`

  For reference documentation on literals, refer to :ref:`Types <api_file_flyteidl/core/literals.proto>`

.. _language-variables:

Variables
  Variables are named entities that have a type. Variables are used to name inputs and outputs of execution units (tasks, wfs... etc.)
  as well as referencing outputs of other execution units.

  For reference documentation on variables, refer to :ref:`Variable <api_msg_flyteidl.core.Variable>`

.. _language-interface:

Interface
  Every execution unit in Flyte can optionally define an interface. The interface holds information about the inputs
  and output variables. In order for **task B** to consume output **X** of **task A**, it has to declare an input
  variable of the same type (not necessarily the same name).

  For reference documentation on interfaces, refer to :ref:`Interfaces <api_file_flyteidl/core/interface.proto>`

.. _language-tasks:

Tasks
  TaskTemplate is a system entity that describes common information about the task for the rest of the system to reason
  about. TaskTemplates contain:

  - An :ref:`identifier <language-identifiers>` that globally identifies this task and allows other entities to reference
    it.
  - A string type that is used throughout the system to customize the behavior of the task execution (e.g. launch a 
    container in K8s, execute a SQL query... etc.) as well as how it gets rendered in the UI.
  - An :ref:`interface <language-interface>` that can be used to type check bindings as well as compatibility of plugging
    the task in a larger workflow.
  - An optional container that describes the versioned container image associated with the task.
  - An optional custom Struct_ used by various plugins to carry information used to customize this task's behavior.
  
  For concept documentation on tasks, refer to :ref:`task concepts <concepts-tasks>`.
  For reference documentation on tasks, refer to :ref:`Tasks <api_file_flyteidl/core/tasks.proto>`

.. _language-bindings:

Bindings
  Bindings define how a certain input of a task should receive its :ref:`literal <language-literals>` value.
  Bindings can be static (assigned at compile/registration time to a literal) or can be references to outputs of other
  nodes.

  A Binding of an input of type collection of strings can either be bound to an output of the same type of a different
  node or else individual items can be bound to outputs of type string of other tasks.
  
  e.g.
  .. code-block::
   
   let taskA return output S of type String
   let taskB return output SList of type list of string
   let taskC takes input i of type list of string

   // Bind the entire input to the entire output.
   taskC.i <- taskB.SList

   // Bind individual items
   taskC.i[0] <- taskA.S
   taskC.i[1] <- taskA.S

  For reference documentation on bindings, refer to :ref:`Types <api_msg_flyteidl.core.Binding>`

.. _language-identifiers:

Identifiers
  An identifier is a globally unique structure that identifies an entity in the system. Varrious entities have different
  identifiers. Tasks, Workflows and Launchplans are all globally identified with an identifier.

  For reference documentation on identifiers, refer to :ref:`Indentifier <api_msg_flyteidl.core.Identifier>`

.. _language-conditions:

Conditional Expressions
  Flyte Spec Language supports conditional expressions of two types; logical and comparison expressions.
  Expressions are represented as a tree of boolean expressions.
  
  Logical Expressions
    A logical expression is expressed as a logical operator (only AND and OR operators are supported) and two conditional
    expressions

  Comparison Expressions
    A comparison expression is expressed as a comparison operator 
    (Equal, Not Equal, Greater Than, Greater Than or Equal, Less Than or Less Than or Equal) and two operands.

    An operand can either be a :ref:`primitive <language-literals>` or a variable name that exists as an input to the 
    node where the expression is evaluated.

  For reference documentation on conditions, refer to :ref:`Indentifier <api_file_flyteidl/core/condition.proto>`

.. _language-nodes:

Nodes
  Nodes are encapsulations around any execution unit (task, workflow, launchplan... etc.) that abstracts away details about
  the execution unit and allows interaction with other nodes with minimal information.

  Nodes define :ref:`how to bind <language-bindings>` the inputs of the underlying execution unit (e.g. task).
  They can also alter how outputs are exposed by providing output aliases to some or all of the outputs variables.

  Multiple nodes can reference the same execution unit with different input bindings.

  Dependencies between nodes is driven by bindings; e.g. NodeC depends on data from NodeA, therefore the system will
  wait to execute NodeC until NodeA has successfully finished executing.
  Nodes can optionally define execution dependencies that are not expressed in bindings.

  For more information about the different types of nodes, please refer to :ref:`node concepts <concepts-nodes>`.
  For reference documentation on nodes, refer to :ref:`Nodes <api_msg_flyteidl.core.Node>`

.. _language-workflows:

Workflows
  Workflows represent an execution scope in the system. A workflow is a directed-acyclic-graph that describes what steps
  need to be executed, in what order as well as which steps need to consume data from other steps.

  For concept documentation on workflows, refer to :ref:`node concepts <concepts-workflows>`.
  For reference documentation on workflows, refer to :ref:`Workflows <api_file_flyteidl/core/workflow.proto>`

Compiler
--------
Flyte system requires compiled Workflows to execute. A CLI is provided for convenience to compile locally. Flyte
automatically compiles all registered tasks and workflows.
The compilation process validates that all nodes and tasks interfaces match up. It then pre-computes the actual
execution plan and serializes the compiled workflow.

Properties of types and values
------------------------------

Type Identity
  Two types are equal if and only if their `Protocol Buffers`_ representation is identical.

Assignability
  Flyte types are invariant. Two variables can be assigned to each other if and only if their types
  are identical.

Extensibility
-------------

New types
  Flyte types can be extended in two ways:

  A customization of an existing type
    e.g. Creating a URL type can be represented as a Literal Type String (that contains the final URL). As far as the task
    interface is concerned, its output is of type String. A subsequent task can then have an input of type string to bind
    to that output.

    Because this approach localizes the visibility of the new type to the one task that produced it, it's hard to enforce
    any type checks (or validation for URL format... etc.) at the system layer.

    This approach won't always satsify more complex types' needs.

  Using generic types
    Flyte offers a literal type `STRUCT` that allows the passing of any valid Struct_ as a value.
    This approach is very powerful to pass other `Protocol Buffers`_ messages or custom types (represented in JSON_)
    that only :ref:`plugins <contributor-extending>` knows about.

  .. note::
    Flyte doesn't yet support subtyping, or custom strongly typed structs.


.. toctree::
    :maxdepth: 2
    :name: languagetoc
    :caption: Spec Language

.. _Protocol Buffers: https://en.wikipedia.org/wiki/Protocol_Buffers
.. _Struct: https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/struct.proto
.. _JSON: https://en.wikipedia.org/wiki/JSON