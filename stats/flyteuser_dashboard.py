import typing
from grafanalib.core import (
    Alert, AlertCondition, Dashboard, Graph,
    GreaterThan, OP_AND, OPS_FORMAT, Row, RTYPE_SUM, SECONDS_FORMAT,
    SHORT_FORMAT, single_y_axis, Target, TimeRange, YAxes, YAxis, MILLISECONDS_FORMAT, Templating, Template
)

DATASOURCE = "Prometheus"


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
                            expr='sum(flyte:propeller:all:workflow:failed_duration_ms{project=~"$project", domain=~"$domain", wf=~"$workflow"}) by (quantile)',
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
            title="Task CPU Usage stats",
            collapse=collapse,
            panels=[
                Graph(
                    title="CPU usage %",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='kube_resourcequota{image!="", resource="limits.cpu", namespace="$project-$domain", type="hard"}',
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
            ])

    @staticmethod
    def create_all_rows(interval: int) -> typing.List[Row]:
        return [
            FlyteUserDashboard.workflow_stats(False),
            FlyteUserDashboard.quota_stats(True),
            FlyteUserDashboard.resource_stats(True),
        ]


project_template = Template(
    name="project",
    dataSource=DATASOURCE,
    query="label_values(flyte:propeller:all:collector:flyteworkflow, project)",
    sort=True,
    allValue=".*",
    includeAll=True,
)

domain_template = Template(
    name="domain",
    dataSource=DATASOURCE,
    query="label_values(flyte:propeller:all:collector:flyteworkflow, domain)",
    sort=True,
    allValue=".*",
    includeAll=True,
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
    editable=True,
    title="Flyte User Dashboard (via Prometheus)",
    rows=FlyteUserDashboard.create_all_rows(interval=5),
    templating=Templating(list=[
        project_template,
        domain_template,
        wf_template,
    ]),
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
