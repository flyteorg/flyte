.. _divedeep-registration:

############
Registration
############

During registration, Flyte validates the workflow structure and saves the workflow. The registration process also updates the workflow graph. 

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/executions/flyte_wf_registration_overview.svg?sanitize=true

Typical Flow Using Flytekit
---------------------------

* The user defines tasks using the :py:mod:`Flytekit <flytekit:flytekit>` Task Definition language
* Next, they define a workflow using the :py:mod:`Flytekit <flytekit:flytekit>` Workflow definition language.
* The user uses flytekit's register CLI to compile the tasks into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. During this, the task representation is bound to a container that constitutes the code for the task. This associated entity is registered with FlyteAdmin using the registerTask API.
* Next, they use Flytekit's register CLI to compile the workflow into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. The referenced tasks are replaced by their FlyteAdmin registered Identifier, obtained in the previous step. This associated entity is registered with FlyteAdmin using the registerWorkflow API.
* Now, they can launch an execution using the FlyteAdmin launch execution API, which requires the necessary inputs provided. This is automatically done if the user uses the Flytectl to launch the execution.
* Later, they can use the FlyteAdmin read API's to get details of the execution, monitor it to completion or retrieve a historical execution
* **OR** can use the FlyteConsole to visualize the execution in real time as it progresses or visualize any historical execution. The console makes it easy to view debugging information for the execution.
* They can set specific rules such as *notification* on **failure** or **success** or publish all events in an execution to a pub-sub system.
* Lastly, they can also query the datastore to get a summary of all the executions and the compute resources consumed.

Typical Flow Without Flytekit
-----------------------------
It is possible to achieve the exact same workflow as mentioned above even if Flytekit is not available. Workflows and tasks are purely specifications and can be provided using tools like ``YAML``, ``JSON``, ``protobuf binary`` or any other programming language. 

Contributions using other tools are welcome.

Registration in the Backend
---------------------------

When FlyteAdmin receives a workflow registration request, it uses the workflow compiler to compile and validate the workflow. It also fetches all the referenced tasks and creates a complete workflow closure, which is stored in the metastore. If the workflow compilation fails, the compiler returns an error to the client.
