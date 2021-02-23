import typing
from grafanalib.core import (
    Alert, AlertCondition, Dashboard, Graph,
    GreaterThan, OP_AND, OPS_FORMAT, Row, RTYPE_SUM, SECONDS_FORMAT,
    SHORT_FORMAT, single_y_axis, Target, TimeRange, YAxes, YAxis, MILLISECONDS_FORMAT
)

DATASOURCE = "Prometheus"


class FlyteUserDashboard(object):

    @staticmethod
    def create_free_workers() -> Graph:
        return Graph(
            title="Free workers count",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='sum(flyte:propeller:all:free_workers_count) by (kubernetes_pod_name)',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )


dashboard = Dashboard(
    editable=True,
    title="Flyte User Dashboard (via Prometheus)",
    rows=FlyteUserDashboard.create_all_rows(interval=5),
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
