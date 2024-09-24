import typing
from grafanalib.core import (
    Alert, AlertCondition, Dashboard, Graph,BarChart,BarGauge,
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
                    title="Accepted Workflows (avg)",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='avg(flyte:propeller:all:workflow:accepted{project=~"$project", domain=~"$domain", wf=~"$workflow"})',
                            refId='A',
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=OPS_FORMAT),
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Workflow success rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:success_duration_ms_count{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                   yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Workflow failure rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(rate(flyte:propeller:all:workflow:failure_duration_ms_count{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Aborted Workflows (avg)",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='avg_over_time(flyte:propeller:all:workflow:workflow_aborted{project=~"$project", domain=~"$domain", wf=~"$workflow"}[5m])',
                            refId='A',
                        ),
                    ],
                      yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                BarGauge(
                    title="Successful wf execution duration by quantile",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(avg(flyte:propeller:all:workflow:success_duration_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by(quantile))/1000',
                            refId='A',
                        ),
                    ],
                    orientation='horizontal',
                   format=SECONDS_FORMAT,
                ),
                BarGauge(
                    title="Failed wf execution duration by quantile",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(avg(flyte:propeller:all:workflow:failure_duration_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by(quantile))/1000',
                            refId='A',
                        ),
                    ],
                    orientation='horizontal',
                    format=SECONDS_FORMAT,
                ),
            ])

    @staticmethod
    def quota_stats(collapse: bool) -> Row:
        return Row(
            title="Kubernetes Resource Quota Usage",
            collapse=collapse,
            panels=[
                Graph(
                    title="CPU Limit vs requested by namespace",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='kube_resourcequota{resource="limits.cpu", namespace="$project-$domain", type="hard"}',
                            refId='A',
                            legendFormat="CPU limit",
                        ),
                        Target(
                            expr='kube_resourcequota{resource="limits.cpu", namespace="$project-$domain", type="used"}',
                            refId='B',
                            legendFormat="CPU requested",
                        ),
                    ],
                    yAxes=YAxes(
                        YAxis(format=SHORT_FORMAT),
                    ),
                ),
                Graph(
                    title="Memory limit vs requested by namespace (MiB)",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(kube_resourcequota{resource="limits.memory", namespace="$project-$domain", type="hard"})*9.5367e-7',
                            refId='A',
                            legendFormat="Memory limit (MiB)",
                        ),
                        Target(
                            expr='(kube_resourcequota{resource="limits.memory", namespace="$project-$domain", type="used"})*9.5367e-7',
                            refId='B',
                            legendFormat="Memory requested (MiB)",
                        ),
                    ],
                ),
            ])

    @staticmethod
    def resource_stats(collapse: bool) -> Row:
        return Row(
            title="Task stats",
            collapse=collapse,
            panels=[
                Graph(
                    title="Pending Tasks",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(kube_pod_status_phase{phase="Pending"} * on(pod) group_left(label_task_name, label_node_id, label_workflow_name) kube_pod_labels{label_workflow_name=~"$workflow"}) by (namespace, label_task_name, label_node_id, label_workflow_name) > 0',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                BarChart(
                    title="Memory Usage per Task(%)",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(100 * (max(container_memory_working_set_bytes{container!=""} * on(pod) group_left(label_task_name, label_node_id, label_workflow_name) kube_pod_labels{namespace=~"$project-$domain",label_workflow_name=~"$workflow"} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_task_name, label_node_id, label_workflow_name) / max(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits{container!=""} * on(pod) group_left(label_task_name, label_node_id, label_workflow_name) kube_pod_labels * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_task_name, label_node_id, label_workflow_name))) > 0',
                            refId='A',
                        ),
                    ],
                    showValue='true',
                ),
                BarChart(
                    title="CPU Usage per Task(%)",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='(100 * (sum(rate(container_cpu_usage_seconds_total{image!=""}[2m]) * on(pod) group_left(label_task_name, label_node_id, label_workflow_name) kube_pod_labels{namespace=~"$project-$domain",label_workflow_name=~"$workflow"} * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_task_name, label_node_id, label_workflow_name) / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{container!=""} * on(pod) group_left(label_task_name, label_node_id, label_workflow_name) kube_pod_labels * on(pod) group_left(phase) kube_pod_status_phase{phase="Running"}) by (namespace, pod, label_task_name, label_node_id, label_workflow_name))) > 0',
                            refId='A',
                        ),
                    ],
                    showValue='true',
                ),
            ])

    @staticmethod
    def errors(collapse: bool) -> Row:
        return Row(
            title="Error (System vs User)",
            collapse=collapse,
            panels=[
                Graph(
                    title="User error rate",
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
                    title="System error rate",
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
    description="Flyte User Dashboard. It's designed to give an overview of execution status and resource consumption.",
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
