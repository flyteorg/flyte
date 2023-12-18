.. _divedeep-executions:

##########
Executions
##########

.. tags:: Basic, Glossary

**Executions** are instances of workflows, nodes or tasks created in the system as a result of a user-requested execution or a scheduled execution.

Typical Flow Using Flytectl
---------------------------

* When an execution of a workflow is triggered using UI/Flytecli/other stateless systems, the system first calls the ``getLaunchPlan`` endpoint and retrieves a launch plan matching the given version. The launch plan definition includes definitions of all input variables declared for the workflow.
* The user-side component then ensures that all the required inputs are supplied and requests the FlyteAdmin service for an execution.
* The FlyteAdmin service validates the inputs, ensuring that they are all specified and, if required, within the declared bounds.
* FlyteAdmin then fetches the previously validated and compiled workflow closure and translates it to an executable format with all the inputs.
* This executable workflow is launched on Kubernetes with an execution record in the database.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/executions/flyte_wf_execution_overview.svg?sanitize=true