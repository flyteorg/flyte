.. _divedeep-registration:

############
Registration
############

.. tags:: Basic, Glossary, Design

During registration, Flyte validates the workflow structure and saves the workflow. The registration process also updates the workflow graph. 

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/executions/flyte_wf_registration_overview.svg?sanitize=true

Typical Flow 
-------------
The following steps elaborate on the specifics of the registration process:

* Define the tasks using the :py:mod:`Flytekit <flytekit:flytekit>` Task Definition language.
* Define a workflow using the :py:mod:`Flytekit <flytekit:flytekit>` Workflow definition language.
* Use `flytectl register CLI <https://docs.flyte.org/en/latest/flytectl/gen/flytectl_register_files.html>`__ to compile the tasks into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. During this, the task representation is bound to a container that constitutes the code for the task. This associated entity is registered with FlyteAdmin using the registerTask API.
* Use flytectl register CLI to compile the workflow into their serialized representation as described in :std:ref:`Flyte Specification language <flyteidl:flyteidltoc>`. The referenced tasks are replaced by their FlyteAdmin registered Identifiers, obtained in the previous step. The associated entity is registered with FlyteAdmin using the registerWorkflow API.
* Launch an execution using the FlyteAdmin launch execution API, which requires the necessary inputs provided. This is automatically done if the user uses flytectl to launch the execution.
* Use the FlyteAdmin read APIs to get details of the execution, monitor it to completion, or retrieve a historical execution.
* **OR** use the FlyteConsole to visualize the execution in real time as it progresses or visualize any historical execution. The console makes it easy to view debugging information for the execution.
* Set specific rules such as *notification* on **failure** or **success** or publish all events in the execution to a pub-sub system.
* Query the datastore to get a summary of all the executions and the compute resources consumed.

.. note::
    Workflows and tasks are purely specifications and can be provided using tools like ``YAML``, ``JSON``, ``protobuf binary`` or any other programming language, and hence registration is possible using other tools. Contributions welcome!

Registration in the Backend
---------------------------

When FlyteAdmin receives a workflow registration request, it uses the workflow compiler to compile and validate the workflow. It also fetches all the referenced tasks and creates a complete workflow closure, which is stored in the metastore. If the workflow compilation fails, the compiler returns an error to the client.
