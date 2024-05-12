import typing

from grafanalib.core import (MILLISECONDS_FORMAT, NO_FORMAT, OPS_FORMAT,
                             PERCENT_FORMAT, SECONDS_FORMAT, SHORT_FORMAT,
                             BarGauge, Dashboard, DataSourceInput, Graph, Row,
                             Target, YAxes, YAxis, single_y_axis)

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
            description="The number of golang goroutines available to accept new work from the main workqueue. Each worker can process one item from the workqueue at a time.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum(flyte:propeller:all:free_workers_count) by (kubernetes_pod_name)",
                    refId="A",
                ),
            ],
            yAxes=YAxes(
                YAxis(format=NO_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def round_latency_per_wf() -> Graph:
        return Graph(
            title=f"Round traverse latency per workflow",
            description="Round latency by workflow name. When there are streaks within one round each iteration is measured separately.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f"sum(flyte:propeller:all:round:raw_ms) by (wf)",
                    refId="A",
                    legendFormat="{{wf}}",
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def round_latency() -> Graph:
        return Graph(
            title=f"Round Latency (includes streak rounds)",
            description="Round latency breakdown. When there are streaks within one round each iteration is measured separately.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum(flyte:propeller:all:round:raw_unlabeled_ms) by (quantile)",
                    refId="A",
                    legendFormat="traverse-P{{quantile}}",
                ),
                Target(
                    expr="avg(flyte:propeller:all:round:raw_unlabeled_ms_sum/flyte:propeller:all:round:raw_unlabeled_ms_count)",
                    refId="B",
                    legendFormat="traverse-mean",
                ),
                Target(
                    expr="sum(flyte:propeller:all:round:round_time_ms) by (quantile)",
                    refId="C",
                    legendFormat="total-P{{quantile}}",
                ),
                Target(
                    expr="avg(flyte:propeller:all:round:round_time_unlabeled_ms_sum/flyte:propeller:all:round:round_time_unlabeled_ms_count)",
                    refId="D",
                    legendFormat="total-mean",
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def error_breakdown() -> Graph:
        return Graph(
            title="Error rate breakdown",
            description="Error rates for each type of failure that may occur as propeller is traversing the workflow DAG.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum(rate(flyte:propeller:all:round:system_error_unlabeled[5m]))",
                    refId="A",
                    legendFormat="system",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:abort_error_unlabeled[5m]))",
                    refId="B",
                    legendFormat="abort",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:panic_unlabeled[5m]))",
                    refId="C",
                    legendFormat="panic",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:not_found[5m]))",
                    refId="D",
                    legendFormat="not-found",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:skipped[5m]))",
                    refId="E",
                    legendFormat="skipped",
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def streak_rate() -> Graph:
        return Graph(
            title="Streak rate",
            description="Streaks are when propeller iterates over the same workflow multiple times in a single round. This is an optimisation technique.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum(rate(flyte:propeller:all:round:streak_length_unlabeled[5m]))",
                    refId="A",
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def round_rates() -> Graph:
        return Graph(
            title="Round success/error rate",
            description="Round success, error and total rates. Also includes total rate including streaks within a single round. The streak rate graph should match the difference between the totals with and without streaks.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum(rate(flyte:propeller:all:round:success_count[5m]))",
                    refId="A",
                    legendFormat="success",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:error_count[5m]))",
                    refId="B",
                    legendFormat="error",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:round_total_ms_count[5m]))",
                    refId="C",
                    legendFormat="total",
                ),
                Target(
                    expr="sum(rate(flyte:propeller:all:round:round_time_unlabeled_ms_count[5m]))",
                    refId="D",
                    legendFormat="total-including-streaks",
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
            description="Count of currently running workflows running per project",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f"sum(flyte:propeller:all:collector:flyteworkflow) by (project)",
                    refId="A",
                ),
            ],
            yAxes=YAxes(
                YAxis(format=NO_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def enqueued_workflows() -> Graph:
        return Graph(
            title="Workflow enqueue rate",
            description="Rate at which workflows are being enqueued by flytepropeller. These enqueues all pass through the sub-queue before going back into the main queue.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum by(type) (rate(flyte:propeller:all:wf_enqueue[5m]))",
                    refId="A",
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
            title=f"Plugin Success/Failure rate",
            description="Success vs failure rate for the various plugins used for each node e.g. k8s plugin or spark plugin.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr='sum(rate({__name__=~"flyte:propeller:all:plugin:.*_success_unlabeled"}[5m]))',
                    refId="A",
                    legendFormat="success",
                ),
                Target(
                    expr='sum(rate({__name__=~"flyte:propeller:all:plugin:.*_failure_unlabeled"}[5m]))',
                    refId="B",
                    legendFormat="failure",
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
                    expr=f"sum(flyte:propeller:all:node:node_exec_latency_us) by (quantile, wf) / 1000",
                    refId="A",
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
                    expr=f"sum(flyte:propeller:all:node:event_recording:success_duration_ms) by (quantile, wf)",
                    refId="A",
                ),
                Target(
                    expr=f"sum(flyte:propeller:all:node:event_recording:failure_duration_ms) by (quantile, wf)",
                    refId="B",
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
                    expr=f"sum(flyte:propeller:all:node:node_input_latency_ms) by (quantile, wf)",
                    refId="A",
                ),
            ],
            yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
        )

    @staticmethod
    def metastore_failures():
        # Copy counts sum(rate(flyte:propeller:all:metastore:copy:overall_unlabeled_ms_count[5m]))
        return Graph(
            title=f"Metastore failure rate",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f"sum(rate(flyte:propeller:all:metastore:head_failure_unlabeled[5m]))",
                    legendFormat="head-failure",
                    refId="A",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:metastore:bad_container_unlabeled[5m]))",
                    legendFormat="bad-container",
                    refId="B",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:metastore:bad_key_unlabeled[5m]))",
                    legendFormat="bad-key",
                    refId="C",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:metastore:read_failure_unlabeled[5m]))",
                    legendFormat="read-failure",
                    refId="D",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:metastore:write_failure_unlabeled[5m]))",
                    legendFormat="write-failure",
                    refId="E",
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
                    expr=f"(sum(rate(flyte:propeller:all:metastore:cache_hit[{interval}m])) * 100) / (sum(rate(flyte:propeller:all:metastore:cache_miss[{interval}m])) + sum(rate(flyte:propeller:all:metastore:cache_hit[{interval}m])))",
                    refId="A",
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
                            expr=f"sum(flyte:propeller:all:metastore:copy:overall_unlabeled_ms) by (quantile)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore write latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:metastore:write_ms) by (quantile, wf)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore read open latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:metastore:read_open_ms) by (quantile, wf)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore head latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:metastore:head_ms) by (quantile, wf)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title=f"Metastore fetch latency by workflow",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:metastore:proto_fetch_ms) by (quantile, wf)",
                            legendFormat="proto-fetch-P{{quantile}}",
                            refId="A",
                        ),
                        Target(
                            expr="sum(flyte:propeller:all:metastore:remote_fetch_ms) by (quantile, wf)",
                            legendFormat="remote-fetch-P{{quantile}}",
                            refId="B",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
            ],
        )

    @staticmethod
    def admin_launcher_cache() -> Graph:
        return Graph(
            title="Admin Launcher cache hit/miss rate",
            description="Cache hit rate when admin launcher is queried for the status of a workflow. Admin launcher will update status of workflows in the cache on a polling interval.",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f"sum(rate(flyte:propeller:all:admin_launcher:cache_hit[5m]))",
                    legendFormat="hit",
                    refId="A",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:admin_launcher:cache_miss[5m]))",
                    legendFormat="miss",
                    refId="B",
                ),
            ],
            yAxes=YAxes(
                YAxis(format=OPS_FORMAT),
                YAxis(format=SHORT_FORMAT),
            ),
        )

    @staticmethod
    def dynamic_wf_build() -> typing.List[Graph]:
        return [
            Graph(
                title="Dynamic workflow build latency",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(flyte:propeller:all:node:build_dynamic_workflow_us) by (quantile, wf) / 1000",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="Dynamic workflow build rate",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:node:build_dynamic_workflow_us_count[5m])) by (wf)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=OPS_FORMAT),
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
                        expr=f"sum(flyte:propeller:all:task:event_recording:success_duration_ms) by (quantile, wf)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="task event recording rate",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:task:event_recording:success_duration_ms_count[5m])) by (wf)",
                        legendFormat="success-{{wf}}",
                        refId="A",
                    ),
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:task:event_recording:failure_duration_ms_count[5m])) by (wf)",
                        legendFormat="failure-{{wf}}",
                        refId="B",
                    ),
                ],
                yAxes=single_y_axis(format=OPS_FORMAT),
            ),
        ]

    @staticmethod
    def node_event_recording() -> typing.List[Graph]:
        return [
            Graph(
                title="node event recording latency",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(flyte:propeller:all:node:event_recording:success_duration_ms) by (quantile, wf)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="node event recording rate",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:node:event_recording:success_duration_ms_count[5m])) by (wf)",
                        legendFormat="success-{{wf}}",
                        refId="A",
                    ),
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:node:event_recording:failure_duration_ms_count[5m])) by (wf)",
                        legendFormat="failure-{{wf}}",
                        refId="B",
                    ),
                ],
                yAxes=single_y_axis(format=OPS_FORMAT),
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
                        expr=f"sum(flyte:propeller:all:workflow:event_recording:success_duration_ms) by (quantile, wf)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title="wf event recording rate",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:workflow:event_recording:success_duration_ms_count[5m])) by (wf)",
                        legendFormat="success",
                        refId="A",
                    ),
                    Target(
                        expr=f"sum(rate(flyte:propeller:all:workflow:event_recording:failure_duration_ms_count[5m])) by (wf)",
                        legendFormat="failure",
                        refId="B",
                    ),
                ],
                yAxes=single_y_axis(format=OPS_FORMAT),
            ),
        ]

    @staticmethod
    def grpc_latency_histogram() -> Graph:
        return BarGauge(
            title="All GRPC calls latency",
            calc="sum",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr="sum by(le) (rate(grpc_client_handling_seconds_bucket[5m]))",
                    refId="A",
                    format="heatmap",
                    legendFormat=r"{{le}}",
                ),
            ],
            displayMode="gradient",
            orientation="vertical",
            max=200,
        )

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
                            expr=f"sum(flyte:propeller:all:wf_update_latency_ms) by (quantile)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
                Graph(
                    title="etcD write rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_update_latency_ms_count[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="etcD write conflict rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_update_conflict[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="etcD write failure rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_update_failed[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="etcD write too large rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_too_large[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
            ],
        )

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
    def grpc_metrics(collapse: bool) -> Row:
        r = Row(
            title="GRPC latency metrics",
            collapse=collapse,
            panels=[
                FlytePropeller.grpc_latency_histogram(),
            ],
        )
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
            title="node event recording error rate breakdown",
            dataSource=DATASOURCE,
            targets=[
                Target(
                    expr=f"sum(rate(flyte:propeller:all:node:perma_system_error_duration_unlabeled_ms_count[5m]))",
                    legendFormat="system error",
                    refId="A",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:node:perma_user_error_duration_unlabeled_ms[5m]))",
                    legendFormat="user error",
                    refId="B",
                ),
                Target(
                    expr=f"sum(rate(flyte:propeller:all:node:perma_unknown_error_duration_unlabeled_ms[5m]))",
                    legendFormat="unknown error",
                    refId="C",
                ),
            ],
            yAxes=single_y_axis(format=OPS_FORMAT),
        )

    @staticmethod
    def queue_metrics(collapse: bool) -> Row:
        return Row(
            title="FlytePropeller Queue metrics",
            collapse=collapse,
            panels=[
                Graph(
                    title="Add rate to queue",
                    description="Rate at which items are actually added to the queue. If an item is already on the queue attempting to add it will be a no-op. Usually there is also rate limiting that will delay items from being added to the queue.",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(rate(flyte:propeller:all:main_adds[5m]))",
                            legendFormat="main",
                            refId="A",
                        ),
                        Target(
                            expr="sum(rate(flyte:propeller:all:sub_adds[5m]))",
                            legendFormat="sub",
                            refId="B",
                        ),
                        Target(
                            expr="sum(rate(flyte:propeller:all:admin_launcher:_adds[5m]))",
                            legendFormat="admin_launcher",
                            refId="C",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Add rate before rate limiting and deduplication",
                    dataSource=DATASOURCE,
                    description="Tracks every rate limited add synchronously before rate limiting delays or deduplication",
                    targets=[
                        Target(
                            expr="sum(rate(flyte:propeller:all:main_retries[5m]))",
                            legendFormat="main",
                            refId="A",
                        ),
                        Target(
                            expr="sum(rate(flyte:propeller:all:sub_retries[5m]))",
                            legendFormat="sub",
                            refId="B",
                        ),
                        Target(
                            expr="sum(rate(flyte:propeller:all:admin_launcher:_retries[5m]))",
                            legendFormat="admin_launcher",
                            refId="C",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Unprocessed Queue depth",
                    description="The number of items that are currently in the queue but have not been processed yet.",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:main_depth)",
                            legendFormat="main",
                            refId="A",
                        ),
                        Target(
                            expr="sum(flyte:propeller:all:sub_depth)",
                            legendFormat="sub",
                            refId="B",
                        ),
                        Target(
                            expr="sum(flyte:propeller:all:admin_launcher:_depth)",
                            legendFormat="admin_launcher",
                            refId="C",
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Seconds of unfinished work in progress",
                    description="Sum of the current in progress time of every in progress item in the queue.",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr="sum(flyte:propeller:all:main_unfinished_work_s)",
                            legendFormat="main",
                            refId="A",
                        ),
                        Target(
                            expr="sum(flyte:propeller:all:sub_unfinished_work_s)",
                            legendFormat="sub",
                            refId="B",
                        ),
                        Target(
                            expr="sum(flyte:propeller:all:admin_launcher:_unfinished_work_s)",
                            legendFormat="admin_launcher",
                            refId="C",
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
                FlytePropeller.round_rates(),
                FlytePropeller.error_breakdown(),
                FlytePropeller.streak_rate(),
                FlytePropeller.plugin_success_vs_failures(),
                FlytePropeller.round_latency(),
                FlytePropeller.round_latency_per_wf(),
                FlytePropeller.workflows_per_project(),
                FlytePropeller.enqueued_workflows(),
            ],
        )

    @staticmethod
    def workflow_latency_per_wf(latency_type: str) -> typing.List[Graph]:
        latency_name = latency_type.replace(":", " ").capitalize()
        return [
            Graph(
                title=f"{latency_name} latency per workflow",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(flyte:propeller:all:{latency_type}_latency_ms) by (wf)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
            Graph(
                title=f"{latency_name} latency by quantile",
                dataSource=DATASOURCE,
                targets=[
                    Target(
                        expr=f"sum(flyte:propeller:all:{latency_type}_latency_unlabeled_ms) by (quantile)",
                        refId="A",
                    ),
                ],
                yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
            ),
        ]

    @staticmethod
    def workflow_latencies(collapse: bool) -> Row:
        return Row(
            title="Workflow latencies",
            collapse=collapse,
            panels=[
                panel
                for panels in [
                    FlytePropeller.workflow_latency_per_wf(latency_type)
                    for latency_type in [
                        "workflow:acceptance",
                        "node:transition",
                        "node:queueing",
                        "workflow:completion",
                    ]
                ]
                for panel in panels
            ],
        )

    @staticmethod
    def k8s_pod_informers(collapse: bool) -> Row:
        return Row(
            title="K8s Pod Informer stats",
            collapse=collapse,
            panels=[
                Graph(
                    title=f"Update event rate from informer",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:node:container:container:informer_update[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title=f"Update events drop rate becacuse they have the same resource version",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:node:container:container:informer_update_dropped[5m]))",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
            ],
        )

    @staticmethod
    def workflow_garbage_collection(collapse: bool) -> Row:
        return Row(
            title="Workflow garbage collection",
            collapse=collapse,
            panels=[
                Graph(
                    title="Namespace successful garbage collections",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(flyte:propeller:all:gc_success)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Namespace failed garbage collections",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(flyte:propeller:all:gc_failure)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Total completed garbage collection rounds",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(flyte:propeller:all:gc_latency_ms_count)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=SHORT_FORMAT),
                ),
                Graph(
                    title="Average garbage collection round duration",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(flyte:propeller:all:gc_latency_ms_sum) / sum(flyte:propeller:all:gc_latency_ms_count)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=MILLISECONDS_FORMAT),
                ),
            ],
        )

    @staticmethod
    def workflowstore(collapse: bool) -> Row:
        return Row(
            title="Workflow store",
            collapse=collapse,
            panels=[
                Graph(
                    title="Stale workflows rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_stale_unlabeled[5m])) by (quantile)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Evict workflows rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_stale_unlabeled[5m])) by (quantile)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
                Graph(
                    title="Workflow redundant updates rate",
                    dataSource=DATASOURCE,
                    targets=[
                        Target(
                            expr=f"sum(rate(flyte:propeller:all:wf_redundant_unlabeled[5m])) by (quantile)",
                            refId="A",
                        ),
                    ],
                    yAxes=single_y_axis(format=OPS_FORMAT),
                ),
            ],
        )

    @staticmethod
    def create_all_rows(interval: int = 5) -> typing.List[Row]:
        return [
            FlytePropeller.core_metrics(interval, False),
            FlytePropeller.queue_metrics(False),
            FlytePropeller.metastore_metrics(interval, True),
            FlytePropeller.metastore_latencies(True),
            FlytePropeller.node_metrics(True),
            FlytePropeller.perf_metrics(True),
            FlytePropeller.grpc_metrics(True),
            FlytePropeller.workflow_latencies(False),
            FlytePropeller.wf_store_latency(False),
            FlytePropeller.k8s_pod_informers(True),
            FlytePropeller.workflowstore(True),
            FlytePropeller.workflow_garbage_collection(True),
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
