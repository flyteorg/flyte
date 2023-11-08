import typing

from grafanalib.core import (
    Dashboard, Graph, Gauge, Stat,
    OPS_FORMAT, Row, SHORT_FORMAT, single_y_axis, Target, YAxes, YAxis, MILLISECONDS_FORMAT, DataSourceInput,
    PERCENT_FORMAT, NO_FORMAT, SECONDS_FORMAT
)

# ------------------------------
# For Gostats we recommend using
# Grafana dashboard ID 10826 - https://grafana.com/grafana/dashboards/10826
#

DATASOURCE_NAME = "DS_PROM"
DATASOURCE = "${%s}" % DATASOURCE_NAME


class FlytePropeller(object):

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

    @staticmethod
    def round_latency_per_wf(interval: int = 1) -> Graph:
        return Graph(
            title=f"round Latency per workflow",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:propeller:all:round:raw_ms[{interval}m])) by (wf)',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def round_latency(interval: int = 1) -> Graph:
        return Graph(
            title=f"round Latency by quantile",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:propeller:all:round:raw_unlabeled_ms[{interval}m])) by (quantile)',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def round_panic() -> Graph:
        return Graph(
            title="Round panic",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='sum(rate(flyte:propeller:all:round:panic_unlabeled[5m]))',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def streak_length() -> Graph:
        return Graph(
            title="Avg streak length",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='avg(flyte:propeller:all:round:streak_length_unlabeled)',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def system_errors() -> Graph:
        return Graph(
            title="System errors",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='sum(deriv(flyte:propeller:all:round:system_error_unlabeled[5m]))',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def abort_errors() -> Graph:
        return Graph(
            title="System errors",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='sum(rate(flyte:propeller:all:round:abort_error[5m]))',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def workflows_per_project() -> Graph:
        return Graph(
            title=f"Running Workflows per project",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(flyte:propeller:all:collector:flyteworkflow) by (project)',
                    refId='A',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def plugin_success_vs_failures() -> Graph:
        """
        TODO We need to convert the plugin names to be labels, so that prometheus can perform queries correctly
        """
        return Graph(
            title=f"Plugin Failures",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='{__name__=~"flyte:propeller:all:node:plugin:.*_failure_unlabeled"}',
                    refId='A',
                ),
                Target(
                    expr='{__name__=~"flyte:propeller:all:node:plugin:.*_success_unlabeled"}',
                    refId='B',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def node_exec_latency() -> Graph:
        return Graph(
            title=f"Node Exec latency quantile and workflow",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(flyte:propeller:all:node:node_exec_latency_us) by (quantile, wf) / 1000',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def node_event_recording_latency() -> Graph:
        return Graph(
            title=f"Node Event event recording latency quantile and workflow",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(flyte:propeller:all:node:event_recording:success_duration_ms) by (quantile, wf)',
                    refId='A',
                ),
                Target(
                    expr=f'sum(flyte:propeller:all:node:event_recording:failure_duration_ms) by (quantile, wf)',
                    refId='B',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def node_input_latency() -> Graph:
        return Graph(
            title=f"Node Input latency quantile and workflow",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(flyte:propeller:all:node:node_input_latency_ms) by (quantile, wf)',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def metastore_failures():
        # Copy counts sum(rate(flyte:propeller:all:metastore:copy:overall_unlabeled_ms_count[5m]))
        return Graph(
            title=f"Failures from metastore",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:propeller:all:metastore:head_failure_unlabeled[5m]))',
                    legendFormat="head-failure",
                    refId='A',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:metastore:bad_container_unlabeled[5m]))',
                    legendFormat="bad-container",
                    refId='B',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:metastore:bad_key_unlabeled[5m]))',
                    legendFormat="bad-key",
                    refId='C',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:metastore:read_failure_unlabeled[5m]))',
                    legendFormat="read-failure",
                    refId='D',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:metastore:write_failure_unlabeled[5m]))',
                    legendFormat="write-failure",
                    refId='E',
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def metastore_cache_hit_percentage(interval: int) -> Graph:
        """
        TODO replace with metric math maybe?
        """
        return Graph(
            title="cache hit percentage",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'(sum(rate(flyte:propeller:all:metastore:cache_hit[{interval}m])) * 100) / (sum(rate(flyte:propeller:all:metastore:cache_miss[{interval}m])) + sum(rate(flyte:propeller:all:metastore:cache_hit[{interval}m])))',
                    refId='A',
                ),
            ],
            yAxes=single_y_axis(format=PERCENT_FORMAT),
        )

    @staticmethod
    def metastore_latencies(collapse: bool) -> Row:
        return Row(
            title=f"Metastore latencies",
            collapse=collapse,
            panels=[
                Graph(
                    title=f"Metastore copy latency",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(flyte:propeller:all:metastore:copy:overall_unlabeled_ms) by (quantile)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore write latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:metastore:write_ms) by (quantile, wf)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore read open latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:metastore:read_open_ms) by (quantile, wf)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore head latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:metastore:head_ms) by (quantile, wf)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore fetch latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr='sum(flyte:propeller:all:metastore:proto_fetch_ms) by (quantile, wf)',
                            legendFormat="proto-fetch",
                            refId='A',
                        ),

                        Target(
                            expr='sum(flyte:propeller:all:metastore:remote_fetch_ms) by (quantile, wf)',
                            legendFormat="remote-fetch",
                            refId='B',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
            ]
        )

    @staticmethod
    def admin_launcher_cache() -> Graph:
        return Graph(
            title="Admin Launcher cache",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:propeller:all:admin_launcher:cache_hit[5m]))',
                    legendFormat="hit",
                    refId='A',
                ),

                Target(
                    expr=f'sum(rate(flyte:propeller:all:admin_launcher:cache_miss[5m]))',
                    legendFormat="miss",
                    refId='B',
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def dynamic_wf_build() -> typing.List[Graph]:
        return [
            Graph(
                title="Dynamic workflow build latency",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(flyte:propeller:all:node:build_dynamic_workflow_us) by (quantile, wf) / 1000',
                        refId='A',
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="Dynamic workflow build count",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:node:build_dynamic_workflow_us_count[5m])) by (wf)',
                        refId='A',
                    ),
                ],
                yAxes=single_y_axis(format=NO_FORMAT),
            ),
        ]

    @staticmethod
    def task_event_recording() -> typing.List[Graph]:
        return [
            Graph(
                title="task event recording latency",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(flyte:propeller:all:task:event_recording:success_duration_ms) by (quantile, wf)',
                        refId='A',
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="task event recording count",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:task:event_recording:success_duration_ms_count[5m])) by (wf)',
                        legendFormat="success wf",
                        refId='A',
                    ),
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:task:event_recording:failure_duration_ms_count[5m])) by (wf)',
                        legendFormat="failure",
                        refId='B',
                    ),
                ],
                yAxes=single_y_axis(format=NO_FORMAT),
            ),
        ]

    @staticmethod
    def node_event_recording() -> typing.List[Graph]:
        return [
            Graph(
                title="node event recording latency success",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(flyte:propeller:all:node:event_recording:success_duration_ms) by (quantile, wf)',
                        refId='A',
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="node event recording count",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:node:event_recording:success_duration_ms_count[5m])) by (wf)',
                        legendFormat="success",
                        refId='A',
                    ),
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:node:event_recording:failure_duration_ms_count[5m])) by (wf)',
                        legendFormat="failure",
                        refId='B',
                    ),
                ],
                yAxes=single_y_axis(format=NO_FORMAT),
            ),
        ]

    @staticmethod
    def wf_event_recording() -> typing.List[Graph]:
        return [
            Graph(
                title="wf event recording latency success",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(flyte:propeller:all:workflow:event_recording:success_duration_ms) by (quantile, wf)',
                        refId='A',
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="wf event recording count",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:workflow:event_recording:success_duration_ms_count[5m])) by (wf)',
                        legendFormat="success",
                        refId='A',
                    ),
                    Target(
                        expr=f'sum(rate(flyte:propeller:all:workflow:event_recording:failure_duration_ms_count[5m])) by (wf)',
                        legendFormat="failure",
                        refId='B',
                    ),
                ],
                yAxes=single_y_axis(format=NO_FORMAT),
            ),
        ]

    @staticmethod
    def wf_store_latency(collapse: bool) -> Row:
        return Row(
            title="etcD write metrics",
            collapse=collapse,
            panels=[
                Graph(
                    title="wf update etcD latency",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(flyte:propeller:all:wf_update_latency_ms) by (quantile)',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title="etcD writes",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:wf_update_latency_ms_count[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=NO_FORMAT),
                ),
                Graph(
                    title="etcD write conflicts",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:wf_update_conflict[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=NO_FORMAT),
                ),
                Graph(
                    title="etcD write fail",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:wf_update_failed[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=NO_FORMAT),
                ),
            ])

    @staticmethod
    def perf_metrics(collapse: bool) -> Row:
        r = Row(
            title="Perf metrics",
            collapse=collapse,
            panels=[],
        )
        r.panels.extend(FlytePropeller.wf_event_recording())
        r.panels.extend(FlytePropeller.node_event_recording())
        r.panels.extend(FlytePropeller.task_event_recording())
        r.panels.extend(FlytePropeller.dynamic_wf_build())
        r.panels.append(FlytePropeller.admin_launcher_cache())
        return r

    @staticmethod
    def metastore_metrics(interval: int, collapse: bool) -> Row:
        return Row(
            title="Metastore failures and cache",
            collapse=collapse,
            panels=[
                FlytePropeller.metastore_cache_hit_percentage(interval),
                FlytePropeller.metastore_failures(),
            ],
        )

    @staticmethod
    def node_errors() -> Graph:
        return Graph(
            title="node event recording count",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f'sum(rate(flyte:propeller:all:node:perma_system_error_duration_unlabeled_ms_count[5m]))',
                    legendFormat="system error",
                    refId='A',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:node:perma_user_error_duration_unlabeled_ms[5m]))',
                    legendFormat="user error",
                    refId='B',
                ),
                Target(
                    expr=f'sum(rate(flyte:propeller:all:node:perma_unknown_error_duration_unlabeled_ms[5m]))',
                    legendFormat="user error",
                    refId='C',
                ),
            ],
            yAxes=single_y_axis(format=NO_FORMAT),
        )

    @staticmethod
    def queue_metrics(collapse: bool) -> Row:
        return Row(
            title="FlytePropeller Queue metrics",
            collapse=collapse,
            panels=[
                Graph(
                    title="WF Adds to main queue",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:main_adds[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Unprocessed Queue depth",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:main_depth[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Item retries",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:main_retries[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Seconds of unfinished work in progress",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'flyte:propeller:all:main_unfinished_work_s',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SECONDS_FORMAT),
                ),
                Graph(
                    title="Workqueue work average duration",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:main_work_duration_us_sum[5m]) / rate(flyte:propeller:all:main_work_duration_us_count[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SECONDS_FORMAT),
                ),
                Graph(
                    title="Duration for which an item stays in queue - avg",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f'sum(rate(flyte:propeller:all:main_queue_latency_us_sum[5m]) / rate(flyte:propeller:all:main_queue_latency_us_count[5m]))',
                            refId='A',
                        ),
                    ],
                    yAxes=single_y_axis(format=SECONDS_FORMAT),
                ),
            ],
        )

    @staticmethod
    def node_metrics(collapse: bool) -> Row:
        return Row(
            title="Node Metrics",
            collapse=collapse,
            panels=[
                FlytePropeller.node_exec_latency(),
                FlytePropeller.node_input_latency(),
                FlytePropeller.node_event_recording_latency(),
                FlytePropeller.node_errors(),
            ],
        )

    @staticmethod
    def core_metrics(interval: int, collapse: bool) -> Row:
        return Row(
            title="Core metrics",
            collapse=collapse,
            panels=[
                FlytePropeller.create_free_workers(),
                FlytePropeller.abort_errors(),
                FlytePropeller.system_errors(),
                FlytePropeller.plugin_success_vs_failures(),
                FlytePropeller.round_latency(interval),
                FlytePropeller.round_latency_per_wf(interval),
                FlytePropeller.round_panic(),
                FlytePropeller.workflows_per_project(),
            ],
        )

    @staticmethod
    def create_all_rows(interval: int = 5) -> typing.List[Row]:
        return [
            FlytePropeller.core_metrics(interval, False),
            FlytePropeller.metastore_metrics(interval, True),
            FlytePropeller.metastore_latencies(True),
            FlytePropeller.node_metrics(True),
            FlytePropeller.perf_metrics(True),
            FlytePropeller.wf_store_latency(False),
            FlytePropeller.queue_metrics(True),
        ]


dashboard = Dashboard(
    tags=["flyte", "prometheus", "flytepropeller", "flyte-dataplane"],
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
    title="Flyte Propeller Dashboard (via Prometheus)",
    rows=FlytePropeller.create_all_rows(interval=5),
    description="Flyte Propeller Dashboard. This is great for monitoring FlytePropeller / Flyte data plane deployments. This is mostly useful for the Flyte deployment maintainer",
).auto_panel_ids()

if __name__ == "__main__":
    print(dashboard.to_json_data())
