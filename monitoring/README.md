# Monitoring assets

Grafana dashboards for Flyte v2.

```
dashboards/flyte-execution.json   RPC latency, throughput and error rates per service
```

## What it charts

The RPC call duration histograms emitted by otelconnect v0.10.0 over OTLP
(`rpc.server.call.duration` and `rpc.client.call.duration`) are exported to Prometheus
as `rpc_server_call_duration_seconds` and `rpc_client_call_duration_seconds`.
Durations and histogram buckets are in seconds; `_count` supplies call rates.
The dashboard groups by `service_name` and the fully-qualified `rpc_method`
(e.g. `flyteidl2.workflow.EventsProxyService/Record`), and selects failures with
`error_type!=""`. The former request/response size metrics and `rpc_service`
label are no longer emitted. It expects a Prometheus
datasource with **uid `prometheus`** — the default that kube-prometheus-stack
creates. Point it elsewhere by editing the datasource uid, or by importing
through the Grafana UI and picking a datasource.

Metrics only reach Prometheus if flyte2 exports them, so the deployment needs an
`otel` config section pointing at a collector, and the collector's Prometheus
exporter needs to be scraped.

## Using it

**Devbox** — bundled, nothing to do:

```bash
make devbox-run
make devbox-monitoring     # http://localhost:30300/d/oss/flyte-execution
```

**A cluster running kube-prometheus-stack** (typical AWS/GCP install) — the
Grafana sidecar provisions any ConfigMap carrying the `grafana_dashboard: "1"`
label, from any namespace:

```bash
kubectl create configmap flyte-execution-dashboard \
  --from-file=monitoring/dashboards/ -n default \
  --dry-run=client -o yaml | \
  kubectl label -f - --local grafana_dashboard=1 --dry-run=client -o yaml | \
  kubectl apply -f -
```

**Any other Grafana** — import `dashboards/flyte-execution.json` through the UI
(Dashboards → New → Import), or mount it via file provisioning.

## Editing

Edit in Grafana, then export via **Share → Export → Save to file** and replace
the JSON here, so changes are reviewable as a JSON diff. Keep the `uid` stable
(`oss`) — links to `/d/oss/flyte-execution` depend on it.
