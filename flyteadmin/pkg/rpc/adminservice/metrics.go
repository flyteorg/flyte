// Per endpoint service Metrics.
package adminservice

import (
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type executionEndpointMetrics struct {
	scope promutils.Scope

	create      util.RequestMetrics
	relaunch    util.RequestMetrics
	recover     util.RequestMetrics
	createEvent util.RequestMetrics
	get         util.RequestMetrics
	update      util.RequestMetrics
	getData     util.RequestMetrics
	getMetrics  util.RequestMetrics
	list        util.RequestMetrics
	terminate   util.RequestMetrics
}

type launchPlanEndpointMetrics struct {
	scope promutils.Scope

	create     util.RequestMetrics
	update     util.RequestMetrics
	get        util.RequestMetrics
	getActive  util.RequestMetrics
	list       util.RequestMetrics
	listActive util.RequestMetrics
	listIds    util.RequestMetrics
}

type namedEntityEndpointMetrics struct {
	scope promutils.Scope

	list   util.RequestMetrics
	update util.RequestMetrics
	get    util.RequestMetrics
}

type nodeExecutionEndpointMetrics struct {
	scope promutils.Scope

	createEvent  util.RequestMetrics
	get          util.RequestMetrics
	getData      util.RequestMetrics
	getMetrics   util.RequestMetrics
	list         util.RequestMetrics
	listChildren util.RequestMetrics
}

type projectEndpointMetrics struct {
	scope promutils.Scope

	register util.RequestMetrics
	list     util.RequestMetrics
	update   util.RequestMetrics
}

type attributeEndpointMetrics struct {
	scope promutils.Scope

	update util.RequestMetrics
	get    util.RequestMetrics
	delete util.RequestMetrics
	list   util.RequestMetrics
}

type taskEndpointMetrics struct {
	scope promutils.Scope

	create  util.RequestMetrics
	get     util.RequestMetrics
	list    util.RequestMetrics
	listIds util.RequestMetrics
}

type taskExecutionEndpointMetrics struct {
	scope promutils.Scope

	createEvent util.RequestMetrics
	get         util.RequestMetrics
	getData     util.RequestMetrics
	list        util.RequestMetrics
}

type workflowEndpointMetrics struct {
	scope promutils.Scope

	create  util.RequestMetrics
	get     util.RequestMetrics
	list    util.RequestMetrics
	listIds util.RequestMetrics
}

type descriptionEntityEndpointMetrics struct {
	scope promutils.Scope

	create util.RequestMetrics
	get    util.RequestMetrics
	list   util.RequestMetrics
}

type AdminMetrics struct {
	Scope        promutils.Scope
	PanicCounter prometheus.Counter

	executionEndpointMetrics               executionEndpointMetrics
	launchPlanEndpointMetrics              launchPlanEndpointMetrics
	namedEntityEndpointMetrics             namedEntityEndpointMetrics
	nodeExecutionEndpointMetrics           nodeExecutionEndpointMetrics
	projectEndpointMetrics                 projectEndpointMetrics
	projectAttributesEndpointMetrics       attributeEndpointMetrics
	projectDomainAttributesEndpointMetrics attributeEndpointMetrics
	workflowAttributesEndpointMetrics      attributeEndpointMetrics
	matchableAttributesEndpointMetrics     attributeEndpointMetrics
	taskEndpointMetrics                    taskEndpointMetrics
	taskExecutionEndpointMetrics           taskExecutionEndpointMetrics
	workflowEndpointMetrics                workflowEndpointMetrics
	descriptionEntityMetrics               descriptionEntityEndpointMetrics
}

func InitMetrics(adminScope promutils.Scope) AdminMetrics {
	return AdminMetrics{
		Scope: adminScope,
		PanicCounter: adminScope.MustNewCounter("handler_panic",
			"panics encountered while handling requests to the admin service"),

		executionEndpointMetrics: executionEndpointMetrics{
			scope:       adminScope,
			create:      util.NewRequestMetrics(adminScope, "create_execution"),
			relaunch:    util.NewRequestMetrics(adminScope, "relaunch_execution"),
			recover:     util.NewRequestMetrics(adminScope, "recover_execution"),
			createEvent: util.NewRequestMetrics(adminScope, "create_execution_event"),
			get:         util.NewRequestMetrics(adminScope, "get_execution"),
			update:      util.NewRequestMetrics(adminScope, "update_execution"),
			getData:     util.NewRequestMetrics(adminScope, "get_execution_data"),
			getMetrics:  util.NewRequestMetrics(adminScope, "get_execution_metrics"),
			list:        util.NewRequestMetrics(adminScope, "list_execution"),
			terminate:   util.NewRequestMetrics(adminScope, "terminate_execution"),
		},
		launchPlanEndpointMetrics: launchPlanEndpointMetrics{
			scope:      adminScope,
			create:     util.NewRequestMetrics(adminScope, "create_launch_plan"),
			update:     util.NewRequestMetrics(adminScope, "update_launch_plan"),
			get:        util.NewRequestMetrics(adminScope, "get_launch_plan"),
			getActive:  util.NewRequestMetrics(adminScope, "get_active_launch_plan"),
			list:       util.NewRequestMetrics(adminScope, "list_launch_plan"),
			listActive: util.NewRequestMetrics(adminScope, "list_active_launch_plans"),
			listIds:    util.NewRequestMetrics(adminScope, "list_launch_plan_ids"),
		},
		namedEntityEndpointMetrics: namedEntityEndpointMetrics{
			scope:  adminScope,
			get:    util.NewRequestMetrics(adminScope, "get_named_entity"),
			list:   util.NewRequestMetrics(adminScope, "list_named_entities"),
			update: util.NewRequestMetrics(adminScope, "update_named_entity"),
		},
		nodeExecutionEndpointMetrics: nodeExecutionEndpointMetrics{
			scope:        adminScope,
			createEvent:  util.NewRequestMetrics(adminScope, "create_node_execution_event"),
			get:          util.NewRequestMetrics(adminScope, "get_node_execution"),
			getData:      util.NewRequestMetrics(adminScope, "get_node_execution_data"),
			getMetrics:   util.NewRequestMetrics(adminScope, "get_node_execution_metrics"),
			list:         util.NewRequestMetrics(adminScope, "list_node_execution"),
			listChildren: util.NewRequestMetrics(adminScope, "list_children_node_executions"),
		},
		projectEndpointMetrics: projectEndpointMetrics{
			scope:    adminScope,
			register: util.NewRequestMetrics(adminScope, "register_project"),
			list:     util.NewRequestMetrics(adminScope, "list_projects"),
			update:   util.NewRequestMetrics(adminScope, "update_project"),
		},
		projectAttributesEndpointMetrics: attributeEndpointMetrics{
			scope:  adminScope,
			update: util.NewRequestMetrics(adminScope, "update_project_attrs"),
			get:    util.NewRequestMetrics(adminScope, "get_project_attrs"),
			delete: util.NewRequestMetrics(adminScope, "delete_project_attrs"),
		},
		projectDomainAttributesEndpointMetrics: attributeEndpointMetrics{
			scope:  adminScope,
			update: util.NewRequestMetrics(adminScope, "update_project_domain_attrs"),
			get:    util.NewRequestMetrics(adminScope, "get_project_domain_attrs"),
			delete: util.NewRequestMetrics(adminScope, "delete_project_domain_attrs"),
		},
		workflowAttributesEndpointMetrics: attributeEndpointMetrics{
			scope:  adminScope,
			update: util.NewRequestMetrics(adminScope, "update_workflow_attrs"),
			get:    util.NewRequestMetrics(adminScope, "get_workflow_attrs"),
			delete: util.NewRequestMetrics(adminScope, "delete_workflow_attrs"),
		},
		matchableAttributesEndpointMetrics: attributeEndpointMetrics{
			scope: adminScope,
			list:  util.NewRequestMetrics(adminScope, "list_matchable_resource_attrs"),
		},
		taskEndpointMetrics: taskEndpointMetrics{
			scope:   adminScope,
			create:  util.NewRequestMetrics(adminScope, "create_task"),
			get:     util.NewRequestMetrics(adminScope, "get_task"),
			list:    util.NewRequestMetrics(adminScope, "list_task"),
			listIds: util.NewRequestMetrics(adminScope, "list_task_ids"),
		},
		taskExecutionEndpointMetrics: taskExecutionEndpointMetrics{
			scope:       adminScope,
			createEvent: util.NewRequestMetrics(adminScope, "create_task_execution_event"),
			get:         util.NewRequestMetrics(adminScope, "get_task_execution"),
			getData:     util.NewRequestMetrics(adminScope, "get_task_execution_data"),
			list:        util.NewRequestMetrics(adminScope, "list_task_execution"),
		},
		workflowEndpointMetrics: workflowEndpointMetrics{
			scope:   adminScope,
			create:  util.NewRequestMetrics(adminScope, "create_workflow"),
			get:     util.NewRequestMetrics(adminScope, "get_workflow"),
			list:    util.NewRequestMetrics(adminScope, "list_workflow"),
			listIds: util.NewRequestMetrics(adminScope, "list_workflow_ids"),
		},

		descriptionEntityMetrics: descriptionEntityEndpointMetrics{
			scope:  adminScope,
			create: util.NewRequestMetrics(adminScope, "create_description_entity"),
			get:    util.NewRequestMetrics(adminScope, "get_description_entity"),
			list:   util.NewRequestMetrics(adminScope, "list_description_entity"),
		},
	}
}
