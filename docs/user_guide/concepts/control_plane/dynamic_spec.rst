.. _divedeep-dynamic-spec:

Dynamic Job Spec
================

.. tags:: Basic, Design

A dynamic job spec is a subset of the entire workflow spec that defines a set of tasks, workflows, nodes, and output bindings that control how the job should assemble its outputs.

This spec is currently only supported as an intermediate step in running Dynamic Tasks.

.. code-block:: protobuf
   :caption: Dynamic job spec in Protobuf

    message DynamicJobSpec {
        repeated Node nodes = 1;
        int64 min_successes = 2;
        repeated Binding outputs = 3;

        repeated TaskTemplate tasks = 4;
        repeated WorkflowTemplate subworkflows = 5;
    }

.. _divedeep-dynamic-tasks:

Tasks
-----

Defines one or more :ref:`Tasks <divedeep-tasks>` that can then be referenced in the spec.

.. _divedeep-dynamic-subworkflows:

Subworkflows
------------

Defines zero or more :ref:`Workflows <divedeep-workflows>` that can then be referenced in the spec.

.. _divedeep-dynamic-nodes:

Nodes
-----

Defines one or more :ref:`Nodes <divedeep-nodes>` that can run in parallel to produce the final outputs of the spec.

.. _divedeep-dynamic-outputs:

Outputs
-------

Defines one or more binding that instructs engine on how to assemble the final outputs.