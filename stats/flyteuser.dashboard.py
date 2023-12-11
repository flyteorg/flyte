import typing
from grafanalib.core import (
    Alert, AlertCondition, Dashboard, Graph,
    GreaterThan, OP_AND, OPS_FORMAT, Row, RTYPE_SUM, SECONDS_FORMAT,
    SHORT_FORMAT, single_y_axis, Target, TimeRange, YAxes, YAxis, MILLISECONDS_FORMAT, Templating, Template,
    DataSourceInput
)

DATASOURCE_NAME = "DS_PROM"
DATASOURCE = "${%s}" % DATASOURCE_NAME


class FlyteUserDashboard(object):

    @staticmethod
    def workflow_stats(collapse: bool) -> Row:
        return Row(
            title="Workflow Stats",
            collapse=collapse,
            panels=[
                Graph(
                    title="Accepted Workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:accepted{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Successful Workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:success_duration_ms_count{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Failed Workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:failure_duration_ms_count{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Aborted Workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:workflow_aborted{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Successful workflow execution time by Quantile",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:workflow:success_duration_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by (quantile)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title="Failed workflow execution time by Quantile",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:workflow:failure_duration_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by (quantile)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title="Node queuing latency by Quantile",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:node:queueing_latency_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by (quantile)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
            ])

    @staticmethod
    def quota_stats(collapse: bool) -> Row:
        return Row(
            title="Kubernetes Quota Usage stats",
            collapse=collapse,
            panels=[
                Graph(
                    title="CPU Limits vs usage",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='kube_resourcequota{resource="limits.cpu", namespace="$project-$domain", type="hard"}',
                            refId='A',
                            legendFormat="max cpu",
                        ),
                        Target(
                            expr='kube_resourcequota{resource="limits.cpu", namespace="$project-$domain", type="used"}',
                            refId='B',
                            legendFormat="used cpu",
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Mem Limits vs usage",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='kube_resourcequota{resource="limits.memory", namespace="$project-$domain", type="hard"}',
                            refId='A',
                            legendFormat="max mem",
                        ),
                        Target(
                            expr='kube_resourcequota{resource="limits.memory", namespace="$project-$domain", type="used"}',
                            refId='B',
                            legendFormat="used mem",
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
            ])

    @staticmethod
    def resource_stats(collapse: bool) -> Row:
        return Row(
            title="Task stats",
            collapse=collapse,
            panels=[
                Graph(
                    title="Pending tasks",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(kube_pod_container_status_waiting * on(pod) group_left(label_execution_id, label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_execution_id !="",namespace=~"$project-$domain",label_workflow_name=~"$workflow"}) by (namespace, label_execution_id, label_task_name, label_node_id, label_workflow_name) > 0',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Memory Usage Percentage",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(100 * max(container_memory_rss{image!=""} * on(pod) group_left(label_execution_id, label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_execution_id !="",namespace=~"$project-$domain",label_workflow_name=~"$workflow"} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_execution_id, label_task_name, label_node_id, label_workflow_name) / max(kube_pod_container_resource_limits_memory_bytes{container!=""} * on(pod) group_left(label_execution_id, label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_execution_id !=""} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_execution_id, label_task_name, label_node_id, label_workflow_name)) > 0',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="CPU Usage Percentage",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(100* sum(rate(container_cpu_usage_seconds_total{image!=""}[2m]) * on(pod) group_left(label_execution_id, label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_execution_id !="",namespace=~"$project-$domain",label_workflow_name=~"$workflow"} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_execution_id, label_task_name, label_node_id, label_workflow_name) / sum(kube_pod_container_resource_limits_cpu_cores{container!=""} * on(pod) group_left(label_execution_id, label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_execution_id !=""} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_execution_id, label_task_name, label_node_id, label_workflow_name)) > 0',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
            ])

    @staticmethod
    def errors(collapse: bool) -> Row:
        return Row(
            title="Error (System vs user)",
            collapse=collapse,
            panels=[
                Graph(
                    title="User errors",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:node:user_error_duration_ms_count{project=~"$project",domain=~"$domain",wf=~"$project:$domain:$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="System errors",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:node:system_error_duration_ms_count{project=~"$project",domain=~"$domain",wf=~"$project:$domain:$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
            ])

    @staticmethod
    def create_all_rows(interval: int) -> typing.List[Row]:
        return [
            FlyteUserDashboard.workflow_stats(False),
            FlyteUserDashboard.quota_stats(True),
            FlyteUserDashboard.resource_stats(True),
            FlyteUserDashboard.errors(True),
        ]


project_template = Template(
    name="project",
    dataSource=DATASOURCE,
    query="label_values(flyte:propeller:all:collector:flyteworkflow, project)",
    sort=True,
)

domain_template = Template(
    name="domain",
    dataSource=DATASOURCE,
    query="label_values(flyte:propeller:all:collector:flyteworkflow, domain)",
    sort=True,
)

wf_template = Template(
    name="workflow",
    dataSource=DATASOURCE,
    query="label_values(flyte:propeller:all:collector:flyteworkflow, wf)",
    sort=True,
    allValue=".*",
    includeAll=True,
)

dashboard = Dashboard(
    tags=["flyte", "prometheus", "flyteuser", "flyte-user"],
    inputs=[
        DataSourceInput(
            name=DATASOURCE_NAME,
            label="Prometheus",
            description="Prometheus server that connects to Flyte",
            pluginId="prometheus",
            pluginName="Prometheus",
        ),
    ],
    editable=False,
    title="Flyte User Dashboard (via Prometheus)",
    rows=FlyteUserDashboard.create_all_rows(interval=5),
    templating=Templating(list=[
        project_template,
        domain_template,
        wf_template,
    ]),
    description="Flyte User Dashboard. This is great to get a birds-eye and drill down view of executions in your Flyte cluster. Useful for the user.",
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
