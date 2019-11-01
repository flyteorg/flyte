// Per endpoint service Metrics.
package adminservice

import (
	"github.com/lyft/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type executionEndpointMetrics struct {
	scope promutils.Scope

	create      util.RequestMetrics
	relaunch    util.RequestMetrics
	createEvent util.RequestMetrics
	get         util.RequestMetrics
	getData     util.RequestMetrics
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

type nodeExecutionEndpointMetrics struct {
	scope promutils.Scope

	createEvent  util.RequestMetrics
	get          util.RequestMetrics
	getData      util.RequestMetrics
	list         util.RequestMetrics
	listChildren util.RequestMetrics
}

type projectEndpointMetrics struct {
	scope promutils.Scope

	register util.RequestMetrics
	list     util.RequestMetrics
}

type projectDomainEndpointMetrics struct {
	scope promutils.Scope

	update util.RequestMetrics
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

type AdminMetrics struct {
	Scope        promutils.Scope
	PanicCounter prometheus.Counter

	executionEndpointMetrics     executionEndpointMetrics
	launchPlanEndpointMetrics    launchPlanEndpointMetrics
	nodeExecutionEndpointMetrics nodeExecutionEndpointMetrics
	projectEndpointMetrics       projectEndpointMetrics
	projectDomainEndpointMetrics projectDomainEndpointMetrics
	taskEndpointMetrics          taskEndpointMetrics
	taskExecutionEndpointMetrics taskExecutionEndpointMetrics
	workflowEndpointMetrics      workflowEndpointMetrics
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
			createEvent: util.NewRequestMetrics(adminScope, "create_execution_event"),
			get:         util.NewRequestMetrics(adminScope, "get_execution"),
			getData:     util.NewRequestMetrics(adminScope, "get_execution_data"),
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
		nodeExecutionEndpointMetrics: nodeExecutionEndpointMetrics{
			scope:        adminScope,
			createEvent:  util.NewRequestMetrics(adminScope, "create_node_execution_event"),
			get:          util.NewRequestMetrics(adminScope, "get_node_execution"),
			getData:      util.NewRequestMetrics(adminScope, "get_node_execution_data"),
			list:         util.NewRequestMetrics(adminScope, "list_node_execution"),
			listChildren: util.NewRequestMetrics(adminScope, "list_children_node_executions"),
		},
		projectEndpointMetrics: projectEndpointMetrics{
			scope:    adminScope,
			register: util.NewRequestMetrics(adminScope, "register_project"),
			list:     util.NewRequestMetrics(adminScope, "list_projects"),
		},
		projectDomainEndpointMetrics: projectDomainEndpointMetrics{
			scope:  adminScope,
			update: util.NewRequestMetrics(adminScope, "update_project_domain"),
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
	}
}
