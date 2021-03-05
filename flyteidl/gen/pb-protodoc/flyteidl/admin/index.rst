Flyte Admin Service entities
============================

These are the control plane entities that can be used to communication with the
Flyte Admin service over gRPC or REST. The endpoint specification is defined in the
`admin service spec <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`__.

.. toctree::
	:maxdepth: 1
	:caption: admin
	:name: admintoc

	common.proto
	event.proto
	execution.proto
	launch_plan.proto
	matchable_resource.proto
	node_execution.proto
	notification.proto
	project.proto
	project_domain_attributes.proto
	schedule.proto
	task.proto
	task_execution.proto
	workflow.proto
	workflow_attributes.proto
