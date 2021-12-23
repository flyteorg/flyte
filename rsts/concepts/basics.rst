.. _divedeep:

#############
Core Concepts
#############

.. panels::
    :header: text-center
 
    .. link-button:: divedeep-tasks
       :type: ref
       :text: Tasks
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    A **Task** is any independent unit of processing. They can be pure functions or functions with side-effects.
    Each definition of a task also has associated configuration and requirements specifications.
 
    ---
 
    .. link-button:: divedeep-workflows
       :type: ref
       :text: Workflows
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    **Workflows** are programs that are guaranteed to eventually reach a terminal state and are represented as
    Directed Acyclic Graphs (DAGs) expressed in protobuf.

    ---

    .. link-button:: divedeep-nodes
       :type: ref
       :text: Nodes
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    A **Node** is an encapsulation of an instance of a Task and represent the unit of work, where multiple Nodes that are
    interconnected via workflows
    
    ---

    .. link-button:: divedeep-launchplans
       :type: ref
       :text: Launch Plans
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    **Launch Plans** provide a mechanism to specialize input parameters for workflows associated different schedules.

    ---

    .. link-button:: divedeep-registration
        :type: ref
        :text: Registration
        :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    **Registration** is the process of uploading a workflow and its task definitions to the FlyteAdmin service.
    Registration creates an inventory of available tasks, workflows and launchplans declared per project and domain.

    ---

    .. link-button:: divedeep-executions
        :type: ref
        :text: Executions
        :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    **Executions** are instances of workflows, nodes or tasks created in the system as a result of a user-requested
    execution or a scheduled execution.

    ---

    .. link-button:: divedeep-execution-timeline
        :type: ref
        :text: Understand how an Execution progresses
        :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    This section explains how an **Execution** progresses through the FlytePropeller execution engine.

    ---

    .. link-button:: divedeep-data-management
        :type: ref
        :text: Understand how Flyte manages data flow between tasks
        :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    This section explains what is metadata, how large amounts of raw-data is handled and how data flows between tasks.


The diagram below shows how inputs flow through tasks and workflows to produce outputs.

.. image:: ./flyte_wf_tasks_high_level.png


.. toctree::
    :maxdepth: 1
    :name: Core Concepts
    :hidden:

    tasks
    workflows_nodes
    launchplans_schedules
    executions
