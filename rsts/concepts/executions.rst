.. _divedeep-executions:

##########
Executions
##########

Typical flow using flytectl
-----------------------------
* When you request an execution of a Workflow using the UI, Flyte CLI or other stateless systems, the system first calls the
  getLaunchPlan endpoint and retrieves a Launch Plan matching the name for a version. The Launch Plan definition includes the definitions of all the input variables declared for the Workflow.
* The user-side component then ensures that all required inputs are supplied and requests the FlyteAdmin service for an execution
* The Flyte Admin service validates the inputs, making sure that they are all specified and, if required, within the declared bounds.
* Flyte Admin then fetches the previously validated and compiled workflow closure and translates it to an executable format with all of the inputs.
* This executable Workflow is then launched on Kubernetes with an execution record in the database.

.. image:: https://raw.githubusercontent.com/lyft/flyte/assets/img/flyte_wf_execution_overview.svg?sanitize=true

.. toctree::
  :caption: Execution Details
  :maxdepth: 1

  state_machine
  execution_timeline
  observability
  dynamic_spec
  catalog
  customizable_resources
