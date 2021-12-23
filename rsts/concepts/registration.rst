.. _divedeep-registration:

############
Registration
############

During registration, Flyte validates the workflow structure and saves the workflow. The registration process also updates the workflow graph. 

.. image:: https://raw.githubusercontent.com/lyft/flyte/assets/img/flyte_wf_registration_overview.svg?sanitize=true

Typical Flow Using Flytekit
---------------------------

 * A user defines tasks using the :py:mod:`Flytekit <flytekit:flytekit>` Task Definition language
 * A User defines a workflow using either :py:mod:`Flytekit <flytekit:flytekit>` Workflow definition lanugage.
 * The user then uses flytekit's register cli to compile the tasks into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. During this time the task representation is also bound to a container that contains the code for the task. This associated entity is registered with FlyteAdmin using the registerTask api.
 * The user then uses flytekit's register cli to compile the workflow into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. The referenced tasks are replaced by their Flyte Admin registered Identifier, obtained in the previous step. This associated entity is registered with FlyteAdmin using the registerWorkflow api.
 * The user can then launch an execution using the FlyteAdmin launch execution API, which requires the necessary inputs provided. This is automatically done if the user uses the Flyte-CLI to launch the execution.
 * The user can later use the FlyteAdmin read API's to get details of the execution,
   monitor it to completion or retrieve a historical execution
 * OR alternatively the user can use the FlyteConsole to visualize the execution in
   realtime as it progresses or visualize any historical execution. The console
   also makes it easy to view debugging information for the execution.
 * The user can set specific rules like *notification* on failure or success or
   publish all events in an execution to a pub-sub system.
 * The user can also query the datastore to get a summary of all the executions and
   the compute resources consumed.

Typical Flow without Flytekit
------------------------------------
It is possible to achieve the exact same workflow as above in case Flytekit is not available. Workflows and tasks are purely specifications and can be provided using any tool like YAML, JSON, protobuf binary or any other programming language. 

Contributions using other tools are welcome.

Registration in the Backend
---------------------------

When FlyteAdmin receives a workflow registration request, it uses the workflow compiler to compile and validate the workflow. It also fetches all the referenced tasks and creates a complete workflow closure, which is stored in the metastore. If the workflow compilation fails, the compiler returns an error to the client.
