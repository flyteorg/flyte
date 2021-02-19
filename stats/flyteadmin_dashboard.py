import typing
from grafanalib.core import (
    Alert, AlertCondition, Dashboard, Graph,
    GreaterThan, OP_AND, OPS_FORMAT, Row, RTYPE_SUM, SECONDS_FORMAT,
    SHORT_FORMAT, single_y_axis, Target, TimeRange, YAxes, YAxis
)

# ------------------------------
# For Gostats we recommend using
# Grafana dashboard ID 10826 - https://grafana.com/grafana/dashboards/10826
#

DATASOURCE = "Prometheus"


class FlyteAdmin(object):
    APIS = [
        "create_execution",
        "create_launch_plan",
        "create_task",
        "create_workflow",
        "create_node_execution",
        "create_task_execution",
        "get_execution",
        "get_launch_plan",
        "get_task",
        "get_workflow",
        "get_node_execution",
        "get_task_execution",
        "get_active_launch_plan",
        "list_execution",
        "list_launch_plan",
        "list_task",
        "list_workflow",
        "list_node_execution",
        "list_task_execution",
        "list_active_launch_plan",
    ]

    @staticmethod
    def error_codes(api: str, interval: int = 1) -> Graph:
        return Graph(
            title=f"{api} return codes",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:codes:OK[{interval}m]))',
                    legendFormat="ok",
                    refId='A',
                ),
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:codes:InvalidArgument[{interval}m]))',
                    legendFormat="invalid-args",
                    refId='B',
                ),
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:codes:AlreadyExists[{interval}m]))',
                    legendFormat="already-exists",
                    refId='C',
                ),
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:codes:FailedPrecondition[{interval}m]))',
                    legendFormat="failed-precondition",
                    refId='D',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def error_vs_success(api: str, interval: int = 1) -> Graph:
        return Graph(
            title=f"{api} success vs errors",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:errors[{interval}m]))',
                    legendFormat="errors",
                    refId='A',
                ),
                Target(
                    expr=f'sum(irate(flyte:admin:{api}:success[{interval}m]))',
                    legendFormat="success",
                    refId='B',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def api_latency(api: str, interval: int = 1) -> Graph:
        return Graph(
            title=f"{api} Latency",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:admin:{api}:duration_ms[{interval}m])) by (quantile)',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=SECONDS_FORMAT),
        )

    @staticmethod
    def create_api_row(api: str, collapse: bool, interval: int = 1) -> Row:
        return Row(
            title=f"{api} stats",
            collapse=collapse,
            panels=[
                FlyteAdmin.error_codes(api, interval),
                FlyteAdmin.error_vs_success(api, interval),
                FlyteAdmin.api_latency(api, interval),
            ]
        )

    @staticmethod
    def create_all_apis(interval: int = 5) -> typing.List[Row]:
        rows = []
        for api in FlyteAdmin.APIS:
            rows.append(FlyteAdmin.create_api_row(api, collapse=True, interval=interval))
        return rows


dashboard = Dashboard(
    editable=True,
    title="Flyte Admin Dashboard (via Prometheus)",
    rows=FlyteAdmin.create_all_apis(interval=5),
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
